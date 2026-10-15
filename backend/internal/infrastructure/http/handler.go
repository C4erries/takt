package http

import (
	"context"
	"errors"
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

	dbStatus := generated.Connected
	overallStatus := generated.HealthStatusOk

	if h.dbChecker != nil {
		if err := h.dbChecker.Ping(ctx); err != nil {
			dbStatus = generated.Disconnected
			overallStatus = generated.HealthStatusDegraded
		}
	} else {
		dbStatus = generated.Disconnected
		overallStatus = generated.HealthStatusDegraded
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

func (h *Handler) Register(c echo.Context) error {
	var req generated.RegisterJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid request payload",
		})
	}

	email := strings.TrimSpace(string(req.Email))
	displayName := strings.TrimSpace(req.DisplayName)

	if email == "" || req.Password == "" || displayName == "" {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Email, password and displayName are required",
		})
	}

	out, err := h.workspaceService.Register(c.Request().Context(), workspaceApp.RegisterInput{
		Email:       email,
		Password:    req.Password,
		DisplayName: displayName,
	})
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrUserAlreadyExists) {
			return c.JSON(http.StatusConflict, generated.Error{
				Code:    "CONFLICT",
				Message: "User with this email already exists",
			})
		}
		if errors.Is(err, domainWorkspace.ErrInvalidPassword) ||
			errors.Is(err, domainWorkspace.ErrInvalidEmail) ||
			errors.Is(err, domainWorkspace.ErrInvalidDisplayName) {
			return c.JSON(http.StatusBadRequest, generated.Error{
				Code:    "BAD_REQUEST",
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to register user",
		})
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
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid request payload",
		})
	}

	email := strings.TrimSpace(string(req.Email))
	if email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Email and password are required",
		})
	}

	out, err := h.workspaceService.Login(c.Request().Context(), workspaceApp.LoginInput{
		Email:    email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrInvalidCredentials) || errors.Is(err, domainWorkspace.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, generated.Error{
				Code:    "UNAUTHORIZED",
				Message: "Invalid email or password",
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to log in",
		})
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
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	out, err := h.workspaceService.GetMe(c.Request().Context(), userID, workspaceID)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrUserNotFound) || errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusUnauthorized, generated.Error{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to fetch user profile",
		})
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
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	if h.searchService == nil {
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Search service is unavailable",
		})
	}

	res, err := h.searchService.Search(c.Request().Context(), workspaceId, userID, params.Q)
	if err != nil {
		if errors.Is(err, domainWorkspace.ErrAccessDenied) {
			return c.JSON(http.StatusForbidden, generated.Error{
				Code:    "FORBIDDEN",
				Message: "Access to workspace denied",
			})
		}
		if errors.Is(err, domainWorkspace.ErrWorkspaceNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{
				Code:    "NOT_FOUND",
				Message: "Workspace not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to execute search",
		})
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
