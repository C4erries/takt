package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	attachmentApp "takt/internal/application/attachment"
	automationApp "takt/internal/application/automation"
	projectApp "takt/internal/application/project"
	searchApp "takt/internal/application/search"
	taskApp "takt/internal/application/task"
	workspaceApp "takt/internal/application/workspace"
	domainTask "takt/internal/domain/task"
	domainWorkspace "takt/internal/domain/workspace"
	"takt/internal/infrastructure/http/generated"
)

var _ generated.ServerInterface = (*Handler)(nil)

var startTime = time.Now()

// DBChecker defines an interface for pinging the database.
type DBChecker interface {
	Ping(ctx context.Context) error
}

// Handler holds HTTP dependencies and handlers.
type Handler struct {
	dbChecker         DBChecker
	version           string
	isProduction      bool
	workspaceService  workspaceApp.UseCases
	taskService       taskApp.UseCases
	projectService    projectApp.UseCases
	automationService automationApp.UseCases
	attachmentService attachmentApp.UseCases
	searchService     searchApp.UseCases
	maxAttachmentSize int64
	jwtManager        *JWTManager
}

// NewHandler constructs a new HTTP Handler.
func NewHandler(
	dbChecker DBChecker,
	version string,
	isProduction bool,
	workspaceService workspaceApp.UseCases,
	taskService taskApp.UseCases,
	projectService projectApp.UseCases,
	jwtManager *JWTManager,
	automationService ...automationApp.UseCases,
) *Handler {
	var autoSvc automationApp.UseCases
	if len(automationService) > 0 {
		autoSvc = automationService[0]
	}
	return &Handler{
		dbChecker:         dbChecker,
		version:           version,
		isProduction:      isProduction,
		workspaceService:  workspaceService,
		taskService:       taskService,
		projectService:    projectService,
		automationService: autoSvc,
		jwtManager:        jwtManager,
		maxAttachmentSize: 52428800, // 50 MB default
	}
}

// SetAttachmentService configures the attachment service and size limit.
func (h *Handler) SetAttachmentService(svc attachmentApp.UseCases, maxSizeBytes int64) {
	h.attachmentService = svc
	if maxSizeBytes > 0 {
		h.maxAttachmentSize = maxSizeBytes
	}
}

// SetSearchService configures the search application service.
func (h *Handler) SetSearchService(svc searchApp.UseCases) {
	h.searchService = svc
}

// JWTManager returns configured JWT manager.
func (h *Handler) JWTManager() *JWTManager {
	return h.jwtManager
}

// GetHealth handles GET /health and GET /api/v1/health.
func (h *Handler) GetHealth(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	dbStatus, overallStatus := generated.Disconnected, generated.HealthStatusDegraded
	if h.dbChecker != nil && h.dbChecker.Ping(ctx) == nil {
		dbStatus, overallStatus = generated.Connected, generated.HealthStatusOk
	}

	uptime := int(time.Since(startTime).Seconds())
	resp := generated.Health{
		Status:        overallStatus,
		Version:       h.version,
		UptimeSeconds: &uptime,
		Database: struct {
			Status generated.HealthDatabaseStatus `json:"status"`
		}{
			Status: dbStatus,
		},
	}

	return c.JSON(http.StatusOK, resp)
}

// HealthCheck provides backward compatibility for health checking.
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.GetHealth(c)
}

// --- Auth handlers ---

func mapUserToGenerated(u *domainWorkspace.User) generated.User {
	return generated.User{
		Id:          u.ID,
		Email:       openapi_types.Email(u.Email),
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
	}
}

func mapWorkspaceToGenerated(w *domainWorkspace.UserWorkspace) generated.Workspace {
	return generated.Workspace{
		Id:        w.ID,
		Name:      w.Name,
		Role:      generated.WorkspaceRole(w.Role),
		OwnerId:   w.OwnerID,
		CreatedAt: w.CreatedAt,
	}
}

func jsonErr(c echo.Context, status int, code, msg string) error {
	return c.JSON(status, generated.Error{Code: code, Message: msg})
}

func (h *Handler) Register(c echo.Context) error {
	var req generated.RegisterJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return jsonErr(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload")
	}

	email := strings.TrimSpace(string(req.Email))
	displayName := strings.TrimSpace(req.DisplayName)

	if email == "" || req.Password == "" || displayName == "" {
		return jsonErr(c, http.StatusBadRequest, "BAD_REQUEST", "Email, password and displayName are required")
	}

	out, err := h.workspaceService.Register(c.Request().Context(), workspaceApp.RegisterInput{
		Email:       email,
		Password:    req.Password,
		DisplayName: displayName,
	})
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrUserAlreadyExists) {
			return jsonErr(c, http.StatusConflict, "CONFLICT", "User with this email already exists")
		}
		if errors.Is(err, domainWorkspace.ErrInvalidPassword) ||
			errors.Is(err, domainWorkspace.ErrInvalidEmail) ||
			errors.Is(err, domainWorkspace.ErrInvalidDisplayName) {
			return jsonErr(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
		return jsonErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register user")
	}

	SetAuthCookie(c, out.Token, h.isProduction)

	return c.JSON(http.StatusCreated, generated.AuthResponse{
		Token:     out.Token,
		User:      mapUserToGenerated(out.User),
		Workspace: mapWorkspaceToGenerated(out.Workspace),
	})
}

func (h *Handler) Login(c echo.Context) error {
	var req generated.LoginJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return jsonErr(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload")
	}

	email := strings.TrimSpace(string(req.Email))
	if email == "" || req.Password == "" {
		return jsonErr(c, http.StatusBadRequest, "BAD_REQUEST", "Email and password are required")
	}

	out, err := h.workspaceService.Login(c.Request().Context(), workspaceApp.LoginInput{
		Email:    email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrInvalidCredentials) || errors.Is(err, domainWorkspace.ErrUserNotFound) {
			return jsonErr(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid email or password")
		}
		return jsonErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to log in")
	}

	SetAuthCookie(c, out.Token, h.isProduction)

	return c.JSON(http.StatusOK, generated.AuthResponse{
		Token:     out.Token,
		User:      mapUserToGenerated(out.User),
		Workspace: mapWorkspaceToGenerated(out.Workspace),
	})
}

func (h *Handler) Logout(c echo.Context) error {
	ClearAuthCookie(c)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *Handler) GetMe(c echo.Context) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return jsonErr(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}

	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return jsonErr(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}

	out, err := h.workspaceService.GetMe(c.Request().Context(), userID, workspaceID)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrUserNotFound) || errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return jsonErr(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		}
		return jsonErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch user profile")
	}

	return c.JSON(http.StatusOK, generated.MeResponse{
		User:      mapUserToGenerated(out.User),
		Workspace: mapWorkspaceToGenerated(out.Workspace),
	})
}

// SearchWorkspace handles GET /workspaces/{workspace_id}/search.
func (h *Handler) SearchWorkspace(c echo.Context, workspaceId openapi_types.UUID, params generated.SearchWorkspaceParams) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return jsonErr(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}

	if h.searchService == nil {
		return jsonErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Search service is unavailable")
	}

	res, err := h.searchService.Search(c.Request().Context(), workspaceId, userID, params.Q)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return jsonErr(c, http.StatusForbidden, "FORBIDDEN", "Access to workspace denied")
		}
		if errors.Is(err, domainWorkspace.ErrWorkspaceNotFound) {
			return jsonErr(c, http.StatusNotFound, "NOT_FOUND", "Workspace not found")
		}
		return jsonErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to execute search")
	}

	respTasks := make([]generated.Task, 0, len(res.Tasks))
	for _, t := range res.Tasks {
		respTasks = append(respTasks, mapDomainTaskToGenerated(t))
	}

	respProjects := make([]generated.Project, 0, len(res.Projects))
	for _, p := range res.Projects {
		respProjects = append(respProjects, mapDomainProjectToGenerated(p))
	}

	return c.JSON(http.StatusOK, generated.SearchResponse{
		Tasks:    respTasks,
		Projects: respProjects,
	})
}

func respondTaskDomainError(c echo.Context, err error) error {
	status, code := http.StatusInternalServerError, "INTERNAL_ERROR"
	switch {
	case errors.Is(err, domainTask.ErrTaskNotFound), errors.Is(err, domainTask.ErrTagNotFound):
		status, code = http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, domainTask.ErrInvalidTaskTitle), errors.Is(err, domainTask.ErrInvalidPriority),
		errors.Is(err, domainTask.ErrInvalidTagName), errors.Is(err, domainTask.ErrInvalidTagColor),
		errors.Is(err, domainTask.ErrInvalidFocusOrder), errors.Is(err, domainTask.ErrSelfParenting),
		errors.Is(err, domainTask.ErrInvalidColumnID), errors.Is(err, domainTask.ErrInvalidTaskID),
		errors.Is(err, domainTask.ErrInvalidActivityType):
		status, code = http.StatusBadRequest, "BAD_REQUEST"
	case errors.Is(err, domainTask.ErrTagAlreadyAttached):
		status, code = http.StatusConflict, "TAG_ALREADY_ATTACHED"
	}
	msg := err.Error()
	if status == http.StatusInternalServerError {
		log.Printf("[ERROR 500] %s %s: %+v", c.Request().Method, c.Request().RequestURI, err)
		msg = "Internal server error"
	}
	return jsonErr(c, status, code, msg)
}

func mapDomainTagToGenerated(tag *domainTask.Tag) generated.Tag {
	return generated.Tag{
		Id:          tag.ID,
		WorkspaceId: tag.WorkspaceID,
		Name:        tag.Name,
		ColorHex:    tag.ColorHex,
		CreatedAt:   tag.CreatedAt,
	}
}

func mapDomainTaskToGenerated(t *domainTask.Task) generated.Task {
	var dueDate *openapi_types.Date
	if t.DueDate != nil {
		dueDate = &openapi_types.Date{Time: *t.DueDate}
	}
	var tags *[]generated.Tag
	if t.Tags != nil {
		mapped := make([]generated.Tag, len(t.Tags))
		for i, tag := range t.Tags {
			mapped[i] = mapDomainTagToGenerated(tag)
		}
		tags = &mapped
	}
	var subCount, subDoneCount *int
	if t.SubtasksCount > 0 {
		subCount = &t.SubtasksCount
	}
	if t.SubtasksCompletedCount > 0 || t.SubtasksCount > 0 {
		subDoneCount = &t.SubtasksCompletedCount
	}
	return generated.Task{
		Id: t.ID, WorkspaceId: t.WorkspaceID, ParentId: t.ParentID, ProjectId: t.ProjectID,
		ColumnId: t.ColumnID, Title: t.Title, DescriptionMarkdown: t.DescriptionMarkdown,
		IsCompleted: t.IsCompleted, CompletedAt: t.CompletedAt, DueDate: dueDate,
		Priority: generated.TaskPriority(t.Priority), FocusOrder: t.FocusOrder, Position: float64(t.Position),
		SubtasksCount: subCount, SubtasksCompletedCount: subDoneCount, Tags: tags,
		IsArchived: t.IsArchived, ArchivedAt: t.ArchivedAt, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}

func mapDomainTasksToGenerated(tasks []*domainTask.Task) []generated.Task {
	resp := make([]generated.Task, len(tasks))
	for i, t := range tasks {
		resp[i] = mapDomainTaskToGenerated(t)
	}
	return resp
}

func mapDomainActivityToGenerated(a *domainTask.TaskActivity) generated.TaskActivity {
	var payload *map[string]interface{}
	if a.Payload != nil {
		p := a.Payload
		payload = &p
	}
	return generated.TaskActivity{
		Id:           a.ID,
		TaskId:       a.TaskID,
		UserId:       a.UserID,
		UserName:     a.UserName,
		ActivityType: a.ActivityType,
		Payload:      payload,
		CreatedAt:    a.CreatedAt,
	}
}

func mapDomainActivitiesToGenerated(activities []*domainTask.TaskActivity) []generated.TaskActivity {
	resp := make([]generated.TaskActivity, len(activities))
	for i, a := range activities {
		resp[i] = mapDomainActivityToGenerated(a)
	}
	return resp
}

