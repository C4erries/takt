package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	taskApp "takt/internal/application/task"
	domainTask "takt/internal/domain/task"
	"takt/internal/infrastructure/http/generated"
)

func (h *Handler) GetInboxTasks(c echo.Context) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: err.Error()})
	}
	tasks, err := h.taskService.GetInboxTasks(c.Request().Context(), workspaceID)
	if err != nil {
		return respondTaskDomainError(c, err)
	}
	return c.JSON(http.StatusOK, mapDomainTasksToGenerated(tasks))
}

func (h *Handler) CreateInboxTask(c echo.Context) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: err.Error()})
	}

	var req generated.CreateInboxTaskJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Title cannot be empty"})
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		dueDate = &req.DueDate.Time
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
		ParentID:            req.ParentId,
	})
	if err != nil {
		return respondTaskDomainError(c, err)
	}

	return c.JSON(http.StatusCreated, mapDomainTaskToGenerated(t))
}

func (h *Handler) ToggleTaskCompletion(c echo.Context, id openapi_types.UUID) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: err.Error()})
	}

	t, err := h.taskService.ToggleTask(c.Request().Context(), id, workspaceID)
	if err != nil {
		return respondTaskDomainError(c, err)
	}
	return c.JSON(http.StatusOK, mapDomainTaskToGenerated(t))
}

func (h *Handler) DeleteTask(c echo.Context, id openapi_types.UUID) error {
	workspaceID, err := GetWorkspaceIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, generated.Error{Code: "UNAUTHORIZED", Message: err.Error()})
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
	return c.JSON(http.StatusOK, mapDomainTasksToGenerated(tasks))
}

func (h *Handler) ListSubtasks(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID) error {
	tasks, err := h.taskService.ListSubtasks(c.Request().Context(), workspaceId, taskId)
	if err != nil {
		return respondTaskDomainError(c, err)
	}
	return c.JSON(http.StatusOK, mapDomainTasksToGenerated(tasks))
}

func (h *Handler) UpdateTask(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID) error {
	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request().Body)
		c.Request().Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	var req generated.UpdateTaskJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
	}

	input := taskApp.UpdateTaskInput{
		TaskID:              taskId,
		WorkspaceID:         workspaceId,
		Title:               req.Title,
		DescriptionMarkdown: req.DescriptionMarkdown,
		FocusOrder:          req.FocusOrder,
		ProjectID:           req.ProjectId,
		ColumnID:            req.ColumnId,
		ParentID:            req.ParentId,
	}

	if len(bodyBytes) > 0 {
		var rawMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &rawMap); err == nil {
			if v, ok := rawMap["dueDate"]; ok && v == nil { input.ClearDueDate = true }
			if v, ok := rawMap["due_date"]; ok && v == nil { input.ClearDueDate = true }
			if v, ok := rawMap["parentId"]; ok && v == nil { input.ClearParentID = true }
			if v, ok := rawMap["parent_id"]; ok && v == nil { input.ClearParentID = true }
		}
	}

	if req.DueDate != nil {
		input.DueDate = &req.DueDate.Time
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
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
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
	resp := make([]generated.Tag, len(tags))
	for i, tag := range tags {
		resp[i] = mapDomainTagToGenerated(tag)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListWorkspaceTagsMetrics(c echo.Context, workspaceId openapi_types.UUID) error {
	tags, err := h.taskService.ListWorkspaceTagsWithMetrics(c.Request().Context(), workspaceId)
	if err != nil {
		return respondTaskDomainError(c, err)
	}
	resp := make([]generated.TagWithMetrics, len(tags))
	for i, t := range tags {
		resp[i] = generated.TagWithMetrics{
			Id:          t.ID,
			WorkspaceId: t.WorkspaceID,
			Name:        t.Name,
			ColorHex:    t.ColorHex,
			CreatedAt:   t.CreatedAt,
			TaskCount:   int(t.TaskCount),
		}
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateWorkspaceTag(c echo.Context, workspaceId openapi_types.UUID) error {
	var req generated.CreateWorkspaceTagJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
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

func (h *Handler) UpdateTag(c echo.Context, workspaceId openapi_types.UUID, tagId openapi_types.UUID) error {
	var req generated.UpdateTagJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
	}

	tag, err := h.taskService.UpdateTag(c.Request().Context(), taskApp.UpdateTagInput{
		TagID:       tagId,
		WorkspaceID: workspaceId,
		Name:        req.Name,
		ColorHex:    req.ColorHex,
	})
	if err != nil {
		return respondTaskDomainError(c, err)
	}
	return c.JSON(http.StatusOK, mapDomainTagToGenerated(tag))
}

func (h *Handler) DeleteWorkspaceTag(c echo.Context, workspaceId openapi_types.UUID, tagId openapi_types.UUID) error {
	if err := h.taskService.DeleteTag(c.Request().Context(), tagId, workspaceId); err != nil { return respondTaskDomainError(c, err) }
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteUnusedWorkspaceTags(c echo.Context, workspaceId openapi_types.UUID) error {
	if err := h.taskService.DeleteUnusedTags(c.Request().Context(), workspaceId); err != nil { return respondTaskDomainError(c, err) }
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddTagToTask(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID, tagId openapi_types.UUID) error {
	if err := h.taskService.AddTagToTask(c.Request().Context(), taskId, tagId, workspaceId); err != nil { return respondTaskDomainError(c, err) }
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) RemoveTagFromTask(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID, tagId openapi_types.UUID) error {
	if err := h.taskService.RemoveTagFromTask(c.Request().Context(), taskId, tagId, workspaceId); err != nil { return respondTaskDomainError(c, err) }
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DuplicateTask(ctx echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID) error {
	t, err := h.taskService.DuplicateTask(ctx.Request().Context(), workspaceId, taskId)
	if err != nil {
		return respondTaskDomainError(ctx, err)
	}
	return ctx.JSON(http.StatusCreated, mapDomainTaskToGenerated(t))
}

func (h *Handler) ListArchivedTasks(c echo.Context, workspaceId openapi_types.UUID, params generated.ListArchivedTasksParams) error {
	tasks, err := h.taskService.ListArchivedTasks(c.Request().Context(), workspaceId, params.ProjectId)
	if err != nil {
		return respondTaskDomainError(c, err)
	}
	return c.JSON(http.StatusOK, mapDomainTasksToGenerated(tasks))
}

func (h *Handler) UnarchiveTask(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID) error {
	var req generated.UnarchiveTaskJSONRequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.Error{Code: "BAD_REQUEST", Message: "Invalid request payload"})
	}
	task, err := h.taskService.UnarchiveTask(c.Request().Context(), workspaceId, taskId, req.ColumnId, req.BoardId)
	if err != nil {
		return respondTaskDomainError(c, err)
	}
	return c.JSON(http.StatusOK, mapDomainTaskToGenerated(task))
}

func (h *Handler) ListTaskActivities(c echo.Context, workspaceId openapi_types.UUID, taskId openapi_types.UUID) error {
	activities, err := h.taskService.ListTaskActivities(c.Request().Context(), workspaceId, taskId)
	if err != nil {
		return respondTaskDomainError(c, err)
	}
	return c.JSON(http.StatusOK, mapDomainActivitiesToGenerated(activities))
}


