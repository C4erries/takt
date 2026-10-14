package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	projectApp "takt/internal/application/project"
	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
)

type ProjectResponse struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type BoardResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"projectId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ColumnResponse struct {
	ID             uuid.UUID `json:"id"`
	BoardID        uuid.UUID `json:"boardId"`
	Name           string    `json:"name"`
	Position       float64   `json:"position"`
	StatusCategory string    `json:"statusCategory"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type (
	CreateProjectHTTPRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}
	UpdateProjectHTTPRequest = CreateProjectHTTPRequest

	CreateColumnHTTPRequest struct {
		Name           string  `json:"name"`
		Position       float64 `json:"position"`
		StatusCategory string  `json:"statusCategory"`
	}
	UpdateColumnHTTPRequest = CreateColumnHTTPRequest

	MoveTaskHTTPRequest struct {
		ColumnID  uuid.UUID  `json:"columnId"`
		Position  float64    `json:"position"`
		ProjectID *uuid.UUID `json:"projectId"`
	}
)

func (h *Handler) respondDomainError(c echo.Context, err error) error {
	status, code := http.StatusInternalServerError, "INTERNAL_ERROR"
	switch {
	case errors.Is(err, domainProject.ErrColumnNotEmpty):
		status, code = http.StatusConflict, "COLUMN_NOT_EMPTY"
	case errors.Is(err, domainProject.ErrProjectNotFound):
		status, code = http.StatusNotFound, "PROJECT_NOT_FOUND"
	case errors.Is(err, domainProject.ErrBoardNotFound):
		status, code = http.StatusNotFound, "BOARD_NOT_FOUND"
	case errors.Is(err, domainProject.ErrColumnNotFound):
		status, code = http.StatusNotFound, "COLUMN_NOT_FOUND"
	case errors.Is(err, domainTask.ErrTaskNotFound):
		status, code = http.StatusNotFound, "TASK_NOT_FOUND"
	case errors.Is(err, domainProject.ErrInvalidProjectName),
		errors.Is(err, domainProject.ErrInvalidProjectColor),
		errors.Is(err, domainProject.ErrInvalidBoardName),
		errors.Is(err, domainProject.ErrInvalidColumnName),
		errors.Is(err, domainProject.ErrInvalidStatusCategory),
		errors.Is(err, domainProject.ErrInvalidPosition):
		status, code = http.StatusBadRequest, "BAD_REQUEST"
	case errors.Is(err, domainProject.ErrAccessDenied):
		status, code = http.StatusForbidden, "FORBIDDEN"
	}
	msg := err.Error()
	if status == http.StatusInternalServerError {
		msg = "Internal server error"
	}
	return c.JSON(status, map[string]string{"code": code, "message": msg})
}

func parseParam(c echo.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(c.Param(param)))
	if err != nil {
		return uuid.Nil, c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "BAD_REQUEST",
			"message": "Invalid " + param + " format",
		})
	}
	return id, nil
}

func toProjectResponse(p *domainProject.Project) ProjectResponse {
	return ProjectResponse{
		ID: p.ID, WorkspaceID: p.WorkspaceID, Name: p.Name,
		Description: p.Description, Color: p.Color,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func (h *Handler) ListProjects(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	projects, err := h.projectService.GetProjects(c.Request().Context(), wsID)
	if err != nil {
		return h.respondDomainError(c, err)
	}
	resp := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		resp[i] = toProjectResponse(p)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateProject(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	var req CreateProjectHTTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "BAD_REQUEST", "message": "Invalid payload"})
	}
	p, err := h.projectService.CreateProject(c.Request().Context(), projectApp.CreateProjectInput{
		WorkspaceID: wsID, Name: req.Name, Description: req.Description, Color: req.Color,
	})
	if err != nil {
		return h.respondDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, toProjectResponse(p))
}

func (h *Handler) GetProject(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	projID, err := parseParam(c, "project_id")
	if err != nil {
		return err
	}
	p, err := h.projectService.GetProject(c.Request().Context(), projID, wsID)
	if err != nil {
		return h.respondDomainError(c, err)
	}
	return c.JSON(http.StatusOK, toProjectResponse(p))
}

func (h *Handler) UpdateProject(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	projID, err := parseParam(c, "project_id")
	if err != nil {
		return err
	}
	var req UpdateProjectHTTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "BAD_REQUEST", "message": "Invalid payload"})
	}
	p, err := h.projectService.UpdateProject(c.Request().Context(), projectApp.UpdateProjectInput{
		ProjectID: projID, WorkspaceID: wsID, Name: req.Name, Description: req.Description, Color: req.Color,
	})
	if err != nil {
		return h.respondDomainError(c, err)
	}
	return c.JSON(http.StatusOK, toProjectResponse(p))
}

func (h *Handler) DeleteProject(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	projID, err := parseParam(c, "project_id")
	if err != nil {
		return err
	}
	if err := h.projectService.DeleteProject(c.Request().Context(), projID, wsID); err != nil {
		return h.respondDomainError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) ListProjectBoards(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	projID, err := parseParam(c, "project_id")
	if err != nil {
		return err
	}
	boards, err := h.projectService.GetProjectBoards(c.Request().Context(), projID, wsID)
	if err != nil {
		return h.respondDomainError(c, err)
	}
	resp := make([]BoardResponse, len(boards))
	for i, b := range boards {
		resp[i] = BoardResponse{ID: b.ID, ProjectID: b.ProjectID, Name: b.Name, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt}
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateProjectBoard(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	projID, err := parseParam(c, "project_id")
	if err != nil {
		return err
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "BAD_REQUEST", "message": "Invalid payload"})
	}
	b, err := h.projectService.CreateBoard(c.Request().Context(), projID, wsID, req.Name)
	if err != nil {
		return h.respondDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, BoardResponse{ID: b.ID, ProjectID: b.ProjectID, Name: b.Name, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt})
}

func (h *Handler) ListBoardColumns(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	boardID, err := parseParam(c, "board_id")
	if err != nil {
		return err
	}
	cols, err := h.projectService.GetBoardColumns(c.Request().Context(), boardID, wsID)
	if err != nil {
		return h.respondDomainError(c, err)
	}
	resp := make([]ColumnResponse, len(cols))
	for i, col := range cols {
		resp[i] = ColumnResponse{
			ID: col.ID, BoardID: col.BoardID, Name: col.Name,
			Position: col.Position, StatusCategory: string(col.StatusCategory),
			CreatedAt: col.CreatedAt, UpdatedAt: col.UpdatedAt,
		}
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateBoardColumn(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	boardID, err := parseParam(c, "board_id")
	if err != nil {
		return err
	}
	var req CreateColumnHTTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "BAD_REQUEST", "message": "Invalid payload"})
	}
	col, err := h.projectService.CreateColumn(c.Request().Context(), projectApp.CreateColumnInput{
		WorkspaceID: wsID, BoardID: boardID, Name: req.Name,
		Position: req.Position, StatusCategory: domainProject.StatusCategory(req.StatusCategory),
	})
	if err != nil {
		return h.respondDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, ColumnResponse{
		ID: col.ID, BoardID: col.BoardID, Name: col.Name,
		Position: col.Position, StatusCategory: string(col.StatusCategory),
		CreatedAt: col.CreatedAt, UpdatedAt: col.UpdatedAt,
	})
}

func (h *Handler) UpdateBoardColumn(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	boardID, err := parseParam(c, "board_id")
	if err != nil {
		return err
	}
	colID, err := parseParam(c, "column_id")
	if err != nil {
		return err
	}
	var req UpdateColumnHTTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "BAD_REQUEST", "message": "Invalid payload"})
	}
	col, err := h.projectService.UpdateColumn(c.Request().Context(), projectApp.UpdateColumnInput{
		WorkspaceID: wsID, BoardID: boardID, ColumnID: colID,
		Name: req.Name, Position: req.Position, StatusCategory: domainProject.StatusCategory(req.StatusCategory),
	})
	if err != nil {
		return h.respondDomainError(c, err)
	}
	return c.JSON(http.StatusOK, ColumnResponse{
		ID: col.ID, BoardID: col.BoardID, Name: col.Name,
		Position: col.Position, StatusCategory: string(col.StatusCategory),
		CreatedAt: col.CreatedAt, UpdatedAt: col.UpdatedAt,
	})
}

func (h *Handler) DeleteBoardColumn(c echo.Context) error {
	boardID, _ := uuid.Parse(strings.TrimSpace(c.Param("board_id")))
	return h.deleteColumnWithBoard(c, boardID)
}

func (h *Handler) DeleteColumnDirect(c echo.Context) error {
	return h.deleteColumnWithBoard(c, uuid.Nil)
}

func (h *Handler) deleteColumnWithBoard(c echo.Context, boardID uuid.UUID) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	colID, err := parseParam(c, "column_id")
	if err != nil {
		return err
	}
	if err := h.projectService.DeleteColumn(c.Request().Context(), boardID, colID, wsID); err != nil {
		return h.respondDomainError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) MoveTaskToColumn(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	taskID, err := parseParam(c, "task_id")
	if err != nil {
		return err
	}
	var req MoveTaskHTTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "BAD_REQUEST", "message": "Invalid payload"})
	}
	t, err := h.projectService.MoveTask(c.Request().Context(), projectApp.MoveTaskInput{
		WorkspaceID: wsID, TaskID: taskID, ColumnID: req.ColumnID,
		Position: req.Position, ProjectID: req.ProjectID,
	})
	if err != nil {
		return h.respondDomainError(c, err)
	}
	return c.JSON(http.StatusOK, mapDomainTaskToResponse(t))
}

func (h *Handler) ListProjectTasks(c echo.Context) error {
	wsID, err := parseParam(c, "workspace_id")
	if err != nil {
		return err
	}
	projID, err := parseParam(c, "project_id")
	if err != nil {
		return err
	}
	tasks, err := h.projectService.GetProjectTasks(c.Request().Context(), projID, wsID)
	if err != nil {
		return h.respondDomainError(c, err)
	}
	resp := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		resp[i] = mapDomainTaskToResponse(t)
	}
	return c.JSON(http.StatusOK, resp)
}

