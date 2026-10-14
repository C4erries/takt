package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	taskApp "takt/internal/application/task"
	domainTask "takt/internal/domain/task"
)

type CreateInboxTaskRequest struct {
	Title               string `json:"title"`
	DescriptionMarkdown string `json:"descriptionMarkdown"`
	Priority            string `json:"priority"`
	DueDate             string `json:"dueDate"`
}

type TaskResponse struct {
	ID                  uuid.UUID           `json:"id"`
	WorkspaceID         uuid.UUID           `json:"workspaceId"`
	ProjectID           *uuid.UUID          `json:"projectId"`
	ColumnID            *uuid.UUID          `json:"columnId"`
	Title               string              `json:"title"`
	DescriptionMarkdown string              `json:"descriptionMarkdown"`
	IsCompleted         bool                `json:"isCompleted"`
	CompletedAt         *time.Time          `json:"completedAt"`
	DueDate             *string             `json:"dueDate"`
	Priority            domainTask.Priority `json:"priority"`
	FocusOrder          *int                `json:"focusOrder"`
	Position            int                 `json:"position"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
}

func mapDomainTaskToResponse(t *domainTask.Task) TaskResponse {
	var dueDateStr *string
	if t.DueDate != nil {
		s := t.DueDate.Format("2006-01-02")
		dueDateStr = &s
	}

	return TaskResponse{
		ID:                  t.ID,
		WorkspaceID:         t.WorkspaceID,
		ProjectID:           t.ProjectID,
		ColumnID:            t.ColumnID,
		Title:               t.Title,
		DescriptionMarkdown: t.DescriptionMarkdown,
		IsCompleted:         t.IsCompleted,
		CompletedAt:         t.CompletedAt,
		DueDate:             dueDateStr,
		Priority:            t.Priority,
		FocusOrder:          t.FocusOrder,
		Position:            t.Position,
		CreatedAt:           t.CreatedAt,
		UpdatedAt:           t.UpdatedAt,
	}
}

func (h *Handler) GetInboxTasks(c echo.Context) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"code":    "UNAUTHORIZED",
			"message": err.Error(),
		})
	}

	tasks, err := h.taskService.GetInboxTasks(c.Request().Context(), workspaceID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch inbox tasks",
		})
	}

	resp := make([]TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, mapDomainTaskToResponse(t))
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateInboxTask(c echo.Context) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"code":    "UNAUTHORIZED",
			"message": err.Error(),
		})
	}

	var req CreateInboxTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "BAD_REQUEST",
			"message": "Invalid request payload",
		})
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "BAD_REQUEST",
			"message": "Title cannot be empty",
		})
	}

	var dueDate *time.Time
	if strings.TrimSpace(req.DueDate) != "" {
		parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.DueDate))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Invalid dueDate format, expected YYYY-MM-DD",
			})
		}
		dueDate = &parsedDate
	}

	priority := domainTask.Priority(strings.TrimSpace(req.Priority))
	if priority == "" {
		priority = domainTask.PriorityNone
	}

	t, err := h.taskService.CreateInboxTask(c.Request().Context(), taskApp.CreateInboxTaskInput{
		WorkspaceID:         workspaceID,
		Title:               req.Title,
		DescriptionMarkdown: req.DescriptionMarkdown,
		Priority:            priority,
		DueDate:             dueDate,
	})
	if err != nil {
		if errors.Is(err, domainTask.ErrInvalidTaskTitle) || errors.Is(err, domainTask.ErrInvalidPriority) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"code":    "BAD_REQUEST",
				"message": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create task",
		})
	}

	return c.JSON(http.StatusCreated, mapDomainTaskToResponse(t))
}

func (h *Handler) ToggleTask(c echo.Context) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"code":    "UNAUTHORIZED",
			"message": err.Error(),
		})
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "BAD_REQUEST",
			"message": "Invalid task ID format",
		})
	}

	t, err := h.taskService.ToggleTask(c.Request().Context(), taskID, workspaceID)
	if err != nil {
		if errors.Is(err, domainTask.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "Task not found in workspace",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to toggle task",
		})
	}

	return c.JSON(http.StatusOK, mapDomainTaskToResponse(t))
}

func (h *Handler) DeleteTask(c echo.Context) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"code":    "UNAUTHORIZED",
			"message": err.Error(),
		})
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "BAD_REQUEST",
			"message": "Invalid task ID format",
		})
	}

	if err := h.taskService.DeleteTask(c.Request().Context(), taskID, workspaceID); err != nil {
		if errors.Is(err, domainTask.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "Task not found in workspace",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to delete task",
		})
	}

	return c.NoContent(http.StatusNoContent)
}
