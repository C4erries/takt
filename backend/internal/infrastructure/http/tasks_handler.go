package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	taskApp "takt/internal/application/task"
	domainTask "takt/internal/domain/task"
	"takt/internal/infrastructure/http/generated"
)

func mapDomainTaskToGenerated(t *domainTask.Task) generated.Task {
	var dueDate *openapi_types.Date
	if t.DueDate != nil {
		dueDate = &openapi_types.Date{Time: *t.DueDate}
	}

	return generated.Task{
		Id:                  t.ID,
		WorkspaceId:         t.WorkspaceID,
		ProjectId:           t.ProjectID,
		ColumnId:            t.ColumnID,
		Title:               t.Title,
		DescriptionMarkdown: t.DescriptionMarkdown,
		IsCompleted:         t.IsCompleted,
		CompletedAt:         t.CompletedAt,
		DueDate:             dueDate,
		Priority:            generated.TaskPriority(t.Priority),
		FocusOrder:          t.FocusOrder,
		Position:            float64(t.Position),
		CreatedAt:           t.CreatedAt,
		UpdatedAt:           t.UpdatedAt,
	}
}

func (h *Handler) GetInboxTasks(c echo.Context) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	tasks, err := h.taskService.GetInboxTasks(c.Request().Context(), workspaceID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to fetch inbox tasks",
		})
	}

	resp := make([]generated.Task, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, mapDomainTaskToGenerated(t))
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateInboxTask(c echo.Context) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	var req generated.CreateInboxTaskJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid request payload",
		})
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Title cannot be empty",
		})
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		d := req.DueDate.Time
		dueDate = &d
	}

	priority := domainTask.PriorityNone
	if req.Priority != nil && *req.Priority != "" {
		priority = domainTask.Priority(*req.Priority)
	}

	desc := ""
	if req.DescriptionMarkdown != nil {
		desc = *req.DescriptionMarkdown
	}

	t, err := h.taskService.CreateInboxTask(c.Request().Context(), taskApp.CreateInboxTaskInput{
		WorkspaceID:         workspaceID,
		Title:               title,
		DescriptionMarkdown: desc,
		Priority:            priority,
		DueDate:             dueDate,
	})
	if err != nil {
		if errors.Is(err, domainTask.ErrInvalidTaskTitle) || errors.Is(err, domainTask.ErrInvalidPriority) {
			return c.JSON(http.StatusBadRequest, generated.Error{
				Code:    "BAD_REQUEST",
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create task",
		})
	}

	return c.JSON(http.StatusCreated, mapDomainTaskToGenerated(t))
}

func (h *Handler) ToggleTaskCompletion(c echo.Context, id openapi_types.UUID) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	t, err := h.taskService.ToggleTask(c.Request().Context(), id, workspaceID)
	if err != nil {
		if errors.Is(err, domainTask.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{
				Code:    "NOT_FOUND",
				Message: "Task not found in workspace",
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to toggle task",
		})
	}

	return c.JSON(http.StatusOK, mapDomainTaskToGenerated(t))
}

// ToggleTask provides backward compatibility for callers not using ServerInterface directly.
func (h *Handler) ToggleTask(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid task ID format",
		})
	}
	return h.ToggleTaskCompletion(c, id)
}

func (h *Handler) DeleteTask(c echo.Context, id openapi_types.UUID) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{
			Code:    "UNAUTHORIZED",
			Message: err.Error(),
		})
	}

	if err := h.taskService.DeleteTask(c.Request().Context(), id, workspaceID); err != nil {
		if errors.Is(err, domainTask.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, generated.Error{
				Code:    "NOT_FOUND",
				Message: "Task not found in workspace",
			})
		}
		return c.JSON(http.StatusInternalServerError, generated.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete task",
		})
	}

	return c.NoContent(http.StatusNoContent)
}
