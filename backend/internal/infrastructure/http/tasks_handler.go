package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

func respondTaskDomainError(c echo.Context, err error) error {
	status, code := http.StatusInternalServerError, "INTERNAL_ERROR"
	switch {
	case errors.Is(err, domainTask.ErrTaskNotFound), errors.Is(err, domainTask.ErrTagNotFound):
		status, code = http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, domainTask.ErrInvalidTaskTitle),
		errors.Is(err, domainTask.ErrInvalidPriority),
		errors.Is(err, domainTask.ErrInvalidTagName),
		errors.Is(err, domainTask.ErrInvalidTagColor),
		errors.Is(err, domainTask.ErrInvalidFocusOrder):
		status, code = http.StatusBadRequest, "BAD_REQUEST"
	case errors.Is(err, domainTask.ErrTagAlreadyAttached):
		status, code = http.StatusConflict, "TAG_ALREADY_ATTACHED"
	}
	msg := err.Error()
	if status == http.StatusInternalServerError {
		msg = "Internal server error"
	}
	return c.JSON(status, generated.Error{Code: code, Message: msg})
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
		Tags:                tags,
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
		return respondTaskDomainError(c, err)
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
		return respondTaskDomainError(c, err)
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
		return respondTaskDomainError(c, err)
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
		return respondTaskDomainError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) GetTodayTasks(c echo.Context, workspaceId openapi_types.UUID) error {
	tasks, err := h.taskService.GetTodayTasks(c.Request().Context(), workspaceId)
	if err != nil {
		return respondTaskDomainError(c, err)
	}

	resp := make([]generated.Task, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, mapDomainTaskToGenerated(t))
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateTask(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID) error {
	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request().Body)
		c.Request().Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	var req generated.UpdateTaskJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid request payload",
		})
	}

	input := taskApp.UpdateTaskInput{
		TaskID:              taskId,
		WorkspaceID:         workspaceId,
		Title:               req.Title,
		DescriptionMarkdown: req.DescriptionMarkdown,
		FocusOrder:          req.FocusOrder,
		ProjectID:           req.ProjectId,
		ColumnID:            req.ColumnId,
	}

	if len(bodyBytes) > 0 {
		var rawMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &rawMap); err == nil {
			if val, exists := rawMap["dueDate"]; exists && val == nil {
				input.ClearDueDate = true
			} else if val, exists := rawMap["due_date"]; exists && val == nil {
				input.ClearDueDate = true
			}
		}
	}

	if req.DueDate != nil {
		d := req.DueDate.Time
		input.DueDate = &d
	}
	if req.Priority != nil {
		p := domainTask.Priority(*req.Priority)
		input.Priority = &p
	}

	t, err := h.taskService.UpdateTask(c.Request().Context(), input)
	if err != nil {
		return respondTaskDomainError(c, err)
	}

	return c.JSON(http.StatusOK, mapDomainTaskToGenerated(t))
}

func (h *Handler) SetTaskFocus(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID) error {
	var req generated.SetTaskFocusJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid request payload",
		})
	}

	t, err := h.taskService.SetTaskFocus(c.Request().Context(), taskApp.SetFocusInput{
		TaskID:      taskId,
		WorkspaceID: workspaceId,
		FocusOrder:  req.FocusOrder,
	})
	if err != nil {
		return respondTaskDomainError(c, err)
	}

	return c.JSON(http.StatusOK, mapDomainTaskToGenerated(t))
}

func (h *Handler) ListWorkspaceTags(c echo.Context, workspaceId openapi_types.UUID) error {
	tags, err := h.taskService.ListWorkspaceTags(c.Request().Context(), workspaceId)
	if err != nil {
		return respondTaskDomainError(c, err)
	}

	resp := make([]generated.Tag, 0, len(tags))
	for _, tag := range tags {
		resp = append(resp, mapDomainTagToGenerated(tag))
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateWorkspaceTag(c echo.Context, workspaceId openapi_types.UUID) error {
	var req generated.CreateWorkspaceTagJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid request payload",
		})
	}

	color := ""
	if req.ColorHex != nil {
		color = *req.ColorHex
	}

	tag, err := h.taskService.CreateTag(c.Request().Context(), taskApp.CreateTagInput{
		WorkspaceID: workspaceId,
		Name:        req.Name,
		ColorHex:    color,
	})
	if err != nil {
		return respondTaskDomainError(c, err)
	}

	return c.JSON(http.StatusCreated, mapDomainTagToGenerated(tag))
}

func (h *Handler) DeleteWorkspaceTag(c echo.Context, workspaceId openapi_types.UUID, tagId openapi_types.UUID) error {
	if err := h.taskService.DeleteTag(c.Request().Context(), tagId, workspaceId); err != nil {
		return respondTaskDomainError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddTagToTask(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID, tagId openapi_types.UUID) error {
	if err := h.taskService.AddTagToTask(c.Request().Context(), taskId, tagId, workspaceId); err != nil {
		return respondTaskDomainError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) RemoveTagFromTask(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID, tagId openapi_types.UUID) error {
	if err := h.taskService.RemoveTagFromTask(c.Request().Context(), taskId, tagId, workspaceId); err != nil {
		return respondTaskDomainError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
