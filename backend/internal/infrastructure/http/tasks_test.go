package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	taskApp "takt/internal/application/task"
	domainTask "takt/internal/domain/task"
	"takt/internal/infrastructure/http/generated"
)

type mockTaskService struct {
	getInboxFn          func(ctx context.Context, wsID uuid.UUID) ([]*domainTask.Task, error)
	createInboxFn       func(ctx context.Context, input taskApp.CreateInboxTaskInput) (*domainTask.Task, error)
	toggleFn            func(ctx context.Context, id, wsID uuid.UUID) (*domainTask.Task, error)
	deleteFn            func(ctx context.Context, id, wsID uuid.UUID) error
	getTodayFn          func(ctx context.Context, wsID uuid.UUID) ([]*domainTask.Task, error)
	updateFn            func(ctx context.Context, input taskApp.UpdateTaskInput) (*domainTask.Task, error)
	setFocusFn          func(ctx context.Context, input taskApp.SetFocusInput) (*domainTask.Task, error)
	listSubtasksFn      func(ctx context.Context, wsID, parentID uuid.UUID) ([]*domainTask.Task, error)
	createTagFn         func(ctx context.Context, input taskApp.CreateTagInput) (*domainTask.Tag, error)
	listTagsFn          func(ctx context.Context, wsID uuid.UUID) ([]*domainTask.Tag, error)
	listTagsMetricsFn   func(ctx context.Context, wsID uuid.UUID) ([]*domainTask.TagWithMetrics, error)
	updateTagFn         func(ctx context.Context, input taskApp.UpdateTagInput) (*domainTask.Tag, error)
	deleteTagFn         func(ctx context.Context, tagID, wsID uuid.UUID) error
	deleteUnusedTagsFn  func(ctx context.Context, wsID uuid.UUID) error
	addTagToTaskFn      func(ctx context.Context, taskID, tagID, wsID uuid.UUID) error
	removeTagFromTaskFn func(ctx context.Context, taskID, tagID, wsID uuid.UUID) error
	duplicateTaskFn     func(ctx context.Context, wsID, taskID uuid.UUID) (*domainTask.Task, error)
	listArchivedFn      func(ctx context.Context, wsID uuid.UUID, projectID *uuid.UUID) ([]*domainTask.Task, error)
	archiveInColFn      func(ctx context.Context, wsID, columnID uuid.UUID) (int64, error)
	unarchiveFn         func(ctx context.Context, wsID, taskID, columnID uuid.UUID, boardID *uuid.UUID) (*domainTask.Task, error)
}

func (m *mockTaskService) GetInboxTasks(ctx context.Context, wsID uuid.UUID) ([]*domainTask.Task, error) {
	if m.getInboxFn != nil { return m.getInboxFn(ctx, wsID) }
	return nil, nil
}
func (m *mockTaskService) CreateInboxTask(ctx context.Context, input taskApp.CreateInboxTaskInput) (*domainTask.Task, error) {
	if m.createInboxFn != nil { return m.createInboxFn(ctx, input) }
	return nil, nil
}
func (m *mockTaskService) ToggleTask(ctx context.Context, id, wsID uuid.UUID) (*domainTask.Task, error) {
	if m.toggleFn != nil { return m.toggleFn(ctx, id, wsID) }
	return nil, nil
}
func (m *mockTaskService) DeleteTask(ctx context.Context, id, wsID uuid.UUID) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, id, wsID) }
	return nil
}
func (m *mockTaskService) CreateInitialInboxTasks(ctx context.Context, wsID uuid.UUID) error { return nil }
func (m *mockTaskService) GetTodayTasks(ctx context.Context, wsID uuid.UUID) ([]*domainTask.Task, error) {
	if m.getTodayFn != nil { return m.getTodayFn(ctx, wsID) }
	return nil, nil
}
func (m *mockTaskService) UpdateTask(ctx context.Context, input taskApp.UpdateTaskInput) (*domainTask.Task, error) {
	if m.updateFn != nil { return m.updateFn(ctx, input) }
	return nil, nil
}
func (m *mockTaskService) SetTaskFocus(ctx context.Context, input taskApp.SetFocusInput) (*domainTask.Task, error) {
	if m.setFocusFn != nil { return m.setFocusFn(ctx, input) }
	return nil, nil
}
func (m *mockTaskService) ListSubtasks(ctx context.Context, wsID, parentID uuid.UUID) ([]*domainTask.Task, error) {
	if m.listSubtasksFn != nil { return m.listSubtasksFn(ctx, wsID, parentID) }
	return nil, nil
}
func (m *mockTaskService) CreateTag(ctx context.Context, input taskApp.CreateTagInput) (*domainTask.Tag, error) {
	if m.createTagFn != nil { return m.createTagFn(ctx, input) }
	return nil, nil
}
func (m *mockTaskService) ListWorkspaceTags(ctx context.Context, wsID uuid.UUID) ([]*domainTask.Tag, error) {
	if m.listTagsFn != nil { return m.listTagsFn(ctx, wsID) }
	return nil, nil
}
func (m *mockTaskService) ListWorkspaceTagsWithMetrics(ctx context.Context, wsID uuid.UUID) ([]*domainTask.TagWithMetrics, error) {
	if m.listTagsMetricsFn != nil { return m.listTagsMetricsFn(ctx, wsID) }
	return nil, nil
}
func (m *mockTaskService) UpdateTag(ctx context.Context, input taskApp.UpdateTagInput) (*domainTask.Tag, error) {
	if m.updateTagFn != nil { return m.updateTagFn(ctx, input) }
	return nil, nil
}
func (m *mockTaskService) DeleteTag(ctx context.Context, tagID, wsID uuid.UUID) error {
	if m.deleteTagFn != nil { return m.deleteTagFn(ctx, tagID, wsID) }
	return nil
}
func (m *mockTaskService) DeleteUnusedTags(ctx context.Context, wsID uuid.UUID) error {
	if m.deleteUnusedTagsFn != nil { return m.deleteUnusedTagsFn(ctx, wsID) }
	return nil
}
func (m *mockTaskService) AddTagToTask(ctx context.Context, taskID, tagID, wsID uuid.UUID) error {
	if m.addTagToTaskFn != nil { return m.addTagToTaskFn(ctx, taskID, tagID, wsID) }
	return nil
}

func (m *mockTaskService) RemoveTagFromTask(ctx context.Context, taskID, tagID, wsID uuid.UUID) error {
	if m.removeTagFromTaskFn != nil { return m.removeTagFromTaskFn(ctx, taskID, tagID, wsID) }
	return nil
}
func (m *mockTaskService) DuplicateTask(ctx context.Context, wsID, taskID uuid.UUID) (*domainTask.Task, error) {
	if m.duplicateTaskFn != nil { return m.duplicateTaskFn(ctx, wsID, taskID) }
	return nil, nil
}
func (m *mockTaskService) ListArchivedTasks(ctx context.Context, wsID uuid.UUID, projectID *uuid.UUID) ([]*domainTask.Task, error) {
	if m.listArchivedFn != nil { return m.listArchivedFn(ctx, wsID, projectID) }
	return nil, nil
}
func (m *mockTaskService) ArchiveCompletedTasksInColumn(ctx context.Context, wsID, columnID uuid.UUID) (int64, error) {
	if m.archiveInColFn != nil { return m.archiveInColFn(ctx, wsID, columnID) }
	return 0, nil
}
func (m *mockTaskService) UnarchiveTask(ctx context.Context, wsID, taskID, columnID uuid.UUID, boardID *uuid.UUID) (*domainTask.Task, error) {
	if m.unarchiveFn != nil { return m.unarchiveFn(ctx, wsID, taskID, columnID, boardID) }
	return nil, nil
}

func TestHandler_GetInboxTasks_Success(t *testing.T) {
	wsID := uuid.New()
	mockSvc := &mockTaskService{
		getInboxFn: func(ctx context.Context, wID uuid.UUID) ([]*domainTask.Task, error) {
			now := time.Now().UTC()
			return []*domainTask.Task{
				{
					ID:                  uuid.New(),
					WorkspaceID:         wID,
					Title:               "Task 1",
					DescriptionMarkdown: "",
					IsCompleted:         false,
					Priority:            domainTask.PriorityHigh,
					Position:            0,
					CreatedAt:           now,
					UpdatedAt:           now,
				},
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/inbox", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(WorkspaceIDContextKey), wsID)

	require.NoError(t, handler.GetInboxTasks(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var tasks []generated.Task
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tasks))
	require.Len(t, tasks, 1)
	assert.Equal(t, "Task 1", tasks[0].Title)
}

func TestHandler_CreateInboxTask_Success(t *testing.T) {
	wsID := uuid.New()
	mockSvc := &mockTaskService{
		createInboxFn: func(ctx context.Context, input taskApp.CreateInboxTaskInput) (*domainTask.Task, error) {
			now := time.Now().UTC()
			return &domainTask.Task{
				ID:                  uuid.New(),
				WorkspaceID:         input.WorkspaceID,
				Title:               input.Title,
				DescriptionMarkdown: input.DescriptionMarkdown,
				Priority:            input.Priority,
				Position:            0,
				CreatedAt:           now,
				UpdatedAt:           now,
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	payload := `{"title":"Quick Task","priority":"medium"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/inbox", bytes.NewReader([]byte(payload)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(WorkspaceIDContextKey), wsID)

	require.NoError(t, handler.CreateInboxTask(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	var task generated.Task
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &task))
	assert.Equal(t, "Quick Task", task.Title)
	assert.Equal(t, string(domainTask.PriorityMedium), string(task.Priority))
}

func TestHandler_ToggleTask_Success(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	now := time.Now().UTC()

	mockSvc := &mockTaskService{
		toggleFn: func(ctx context.Context, id, wID uuid.UUID) (*domainTask.Task, error) {
			return &domainTask.Task{
				ID:          id,
				WorkspaceID: wID,
				Title:       "Toggled",
				IsCompleted: true,
				CompletedAt: &now,
				Position:    0,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+taskID.String()+"/toggle", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(taskID.String())
	c.Set(string(WorkspaceIDContextKey), wsID)

	require.NoError(t, handler.ToggleTaskCompletion(c, taskID))
	assert.Equal(t, http.StatusOK, rec.Code)

	var task generated.Task
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &task))
	assert.True(t, task.IsCompleted)
	assert.NotNil(t, task.CompletedAt)
}

func TestHandler_DeleteTask_Success(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()

	mockSvc := &mockTaskService{
		deleteFn: func(ctx context.Context, id, wID uuid.UUID) error {
			return nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+taskID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(taskID.String())
	c.Set(string(WorkspaceIDContextKey), wsID)

	require.NoError(t, handler.DeleteTask(c, taskID))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_GetTodayTasks_Success(t *testing.T) {
	wsID := uuid.New()
	now := time.Now().UTC()

	mockSvc := &mockTaskService{
		getTodayFn: func(ctx context.Context, wID uuid.UUID) ([]*domainTask.Task, error) {
			return []*domainTask.Task{
				{
					ID:          uuid.New(),
					WorkspaceID: wID,
					Title:       "Focus Task Today",
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID.String()+"/tasks/today", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.GetTodayTasks(c, wsID))
	assert.Equal(t, http.StatusOK, rec.Code)

	var tasks []generated.Task
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tasks))
	require.Len(t, tasks, 1)
	assert.Equal(t, "Focus Task Today", tasks[0].Title)
}

func TestHandler_UpdateTask_Success(t *testing.T) {
	wsID, taskID, now := uuid.New(), uuid.New(), time.Now().UTC()
	var capturedInput taskApp.UpdateTaskInput

	mockSvc := &mockTaskService{
		updateFn: func(ctx context.Context, input taskApp.UpdateTaskInput) (*domainTask.Task, error) {
			capturedInput = input
			title := "Updated Task"
			if input.Title != nil {
				title = *input.Title
			}
			return &domainTask.Task{
				ID: input.TaskID, WorkspaceID: input.WorkspaceID, Title: title,
				Priority: domainTask.PriorityHigh, CreatedAt: now, UpdatedAt: now,
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	t.Run("basic update", func(t *testing.T) {
		payload := `{"title":"Updated Task","priority":"high"}`
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+taskID.String(), bytes.NewReader([]byte(payload)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		require.NoError(t, handler.UpdateTask(c, wsID, taskID))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.False(t, capturedInput.ClearDueDate)

		var task generated.Task
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &task))
		assert.Equal(t, "Updated Task", task.Title)
		assert.Equal(t, "high", string(task.Priority))
	})

	t.Run("clear due date with null", func(t *testing.T) {
		payload := `{"dueDate":null}`
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+taskID.String(), bytes.NewReader([]byte(payload)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		require.NoError(t, handler.UpdateTask(c, wsID, taskID))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, capturedInput.ClearDueDate)
	})
}

func TestHandler_SetTaskFocus_Success(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	now := time.Now().UTC()
	focusOrder := 1

	mockSvc := &mockTaskService{
		setFocusFn: func(ctx context.Context, input taskApp.SetFocusInput) (*domainTask.Task, error) {
			return &domainTask.Task{
				ID:          input.TaskID,
				WorkspaceID: input.WorkspaceID,
				Title:       "Focus Task",
				FocusOrder:  input.FocusOrder,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	payload := `{"focusOrder":1}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+taskID.String()+"/focus", bytes.NewReader([]byte(payload)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.SetTaskFocus(c, wsID, taskID))
	assert.Equal(t, http.StatusOK, rec.Code)

	var task generated.Task
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &task))
	require.NotNil(t, task.FocusOrder)
	assert.Equal(t, focusOrder, *task.FocusOrder)
}
