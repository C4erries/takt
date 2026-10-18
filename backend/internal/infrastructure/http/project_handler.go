package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	projectApp "takt/internal/application/project"
	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
	"takt/internal/infrastructure/http/generated"
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
		errors.Is(err, domainProject.ErrInvalidBoardID),
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
	return c.JSON(status, generated.Error{Code: code, Message: msg})
}

func mapDomainProjectToGenerated(p *domainProject.Project) generated.Project {
	return generated.Project{
		Id:          p.ID,
		WorkspaceId: p.WorkspaceID,
		Name:        p.Name,
		Description: p.Description,
		Color:       p.Color,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func mapDomainBoardToGenerated(b *domainProject.Board) generated.Board {
	return generated.Board{
		Id:        b.ID,
		ProjectId: b.ProjectID,
		Name:      b.Name,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

func mapDomainColumnToGenerated(col *domainProject.Column) generated.Column {
	return generated.Column{
		Id:             col.ID,
		BoardId:        col.BoardID,
		Name:           col.Name,
		Position:       col.Position,
		StatusCategory: generated.ColumnStatusCategory(col.StatusCategory),
		IsInbox:        col.IsInbox,
		CreatedAt:      col.CreatedAt,
		UpdatedAt:      col.UpdatedAt,
	}
}

func (h *Handler) ListProjects(ctx echo.Context, workspaceId openapi_types.UUID) error {
	projects, err := h.projectService.GetProjects(ctx.Request().Context(), workspaceId)
	if err != nil {
		return h.respondDomainError(ctx, err)
	}
	resp := make([]generated.Project, len(projects))
	for i, p := range projects {
		resp[i] = mapDomainProjectToGenerated(p)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateProject(ctx echo.Context, workspaceId openapi_types.UUID) error {
	var req generated.CreateProjectJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid payload"})
	}
	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}
	color := ""
	if req.Color != nil {
		color = *req.Color
	}
	p, err := h.projectService.CreateProject(ctx.Request().Context(), projectApp.CreateProjectInput{
		WorkspaceID: workspaceId,
		Name:        req.Name,
		Description: desc,
		Color:       color,
	})
	if err != nil {
		return h.respondDomainError(ctx, err)
	}
	return ctx.JSON(http.StatusCreated, mapDomainProjectToGenerated(p))
}

func (h *Handler) GetProject(ctx echo.Context, workspaceId openapi_types.UUID, projectId openapi_types.UUID) error {
	p, err := h.projectService.GetProject(ctx.Request().Context(), projectId, workspaceId)
	if err != nil {
		return h.respondDomainError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, mapDomainProjectToGenerated(p))
}

func (h *Handler) UpdateProject(ctx echo.Context, workspaceId openapi_types.UUID, projectId openapi_types.UUID) error {
	var req generated.UpdateProjectJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid payload"})
	}
	name, desc, color := "", "", ""
	if req.Name != nil {
		name = *req.Name
	}
	if req.Description != nil {
		desc = *req.Description
	}
	if req.Color != nil {
		color = *req.Color
	}
	p, err := h.projectService.UpdateProject(ctx.Request().Context(), projectApp.UpdateProjectInput{
		ProjectID:   projectId,
		WorkspaceID: workspaceId,
		Name:        name,
		Description: desc,
		Color:       color,
	})
	if err != nil {
		return h.respondDomainError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, mapDomainProjectToGenerated(p))
}

func (h *Handler) DeleteProject(ctx echo.Context, workspaceId openapi_types.UUID, projectId openapi_types.UUID) error {
	if err := h.projectService.DeleteProject(ctx.Request().Context(), projectId, workspaceId); err != nil {
		return h.respondDomainError(ctx, err)
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (h *Handler) ListProjectBoards(ctx echo.Context, workspaceId openapi_types.UUID, projectId openapi_types.UUID) error {
	boards, err := h.projectService.GetProjectBoards(ctx.Request().Context(), projectId, workspaceId)
	if err != nil {
		return h.respondDomainError(ctx, err)
	}
	resp := make([]generated.Board, len(boards))
	for i, b := range boards {
		resp[i] = mapDomainBoardToGenerated(b)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateProjectBoard(ctx echo.Context, workspaceId openapi_types.UUID, projectId openapi_types.UUID) error {
	var req generated.CreateProjectBoardJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid payload"})
	}
	b, err := h.projectService.CreateBoard(ctx.Request().Context(), projectId, workspaceId, req.Name)
	if err != nil {
		return h.respondDomainError(ctx, err)
	}
	return ctx.JSON(http.StatusCreated, mapDomainBoardToGenerated(b))
}

func (h *Handler) ListBoardColumns(ctx echo.Context, workspaceId openapi_types.UUID, boardId openapi_types.UUID) error {
	cols, err := h.projectService.GetBoardColumns(ctx.Request().Context(), boardId, workspaceId)
	if err != nil {
		return h.respondDomainError(ctx, err)
	}
	resp := make([]generated.Column, len(cols))
	for i, col := range cols {
		resp[i] = mapDomainColumnToGenerated(col)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateBoardColumn(ctx echo.Context, workspaceId openapi_types.UUID, boardId openapi_types.UUID) error {
	var req generated.CreateBoardColumnJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid payload"})
	}
	pos := float64(0)
	if req.Position != nil {
		pos = *req.Position
	}
	statusCat := domainProject.StatusCategoryTodo
	if req.StatusCategory != nil {
		statusCat = domainProject.StatusCategory(*req.StatusCategory)
	}
	col, err := h.projectService.CreateColumn(ctx.Request().Context(), projectApp.CreateColumnInput{
		WorkspaceID:    workspaceId,
		BoardID:        boardId,
		Name:           req.Name,
		Position:       pos,
		StatusCategory: statusCat,
	})
	if err != nil {
		return h.respondDomainError(ctx, err)
	}
	return ctx.JSON(http.StatusCreated, mapDomainColumnToGenerated(col))
}

func (h *Handler) UpdateBoardColumn(ctx echo.Context, workspaceId openapi_types.UUID, boardId openapi_types.UUID, columnId openapi_types.UUID) error {
	var req generated.UpdateBoardColumnJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid payload"})
	}
	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	pos := float64(0)
	if req.Position != nil {
		pos = *req.Position
	}
	statusCat := domainProject.StatusCategory("")
	if req.StatusCategory != nil {
		statusCat = domainProject.StatusCategory(*req.StatusCategory)
	}
	col, err := h.projectService.UpdateColumn(ctx.Request().Context(), projectApp.UpdateColumnInput{
		WorkspaceID:    workspaceId,
		BoardID:        req.BoardId,
		ColumnID:       columnId,
		Name:           name,
		Position:       pos,
		StatusCategory: statusCat,
		IsInbox:        req.IsInbox,
	})
	if err != nil {
		return h.respondDomainError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, mapDomainColumnToGenerated(col))
}

func (h *Handler) DeleteBoardColumn(ctx echo.Context, workspaceId openapi_types.UUID, boardId openapi_types.UUID, columnId openapi_types.UUID) error {
	if err := h.projectService.DeleteColumn(ctx.Request().Context(), boardId, columnId, workspaceId); err != nil {
		return h.respondDomainError(ctx, err)
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (h *Handler) MoveTaskToColumn(ctx echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID) error {
	var req generated.MoveTaskToColumnJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid payload"})
	}
	t, err := h.projectService.MoveTask(ctx.Request().Context(), projectApp.MoveTaskInput{
		WorkspaceID: workspaceId,
		TaskID:      taskId,
		ColumnID:    req.ColumnId,
		Position:    req.Position,
		ProjectID:   req.ProjectId,
	})
	if err != nil {
		return h.respondDomainError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, mapDomainTaskToGenerated(t))
}

func (h *Handler) ListProjectTasks(ctx echo.Context, workspaceId openapi_types.UUID, projectId openapi_types.UUID) error {
	tasks, err := h.projectService.GetProjectTasks(ctx.Request().Context(), projectId, workspaceId)
	if err != nil {
		return h.respondDomainError(ctx, err)
	}
	resp := make([]generated.Task, len(tasks))
	for i, t := range tasks {
		resp[i] = mapDomainTaskToGenerated(t)
	}
	return ctx.JSON(http.StatusOK, resp)
}

