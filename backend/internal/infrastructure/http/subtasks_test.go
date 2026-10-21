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

func TestHandler_ListSubtasks(t *testing.T) {
	wsID := uuid.New()
	parentID := uuid.New()
	subID := uuid.New()
	now := time.Now().UTC()

	mockSvc := &mockTaskService{
		listSubtasksFn: func(ctx context.Context, wID, pID uuid.UUID) ([]*domainTask.Task, error) {
			assert.Equal(t, wsID, wID)
			assert.Equal(t, parentID, pID)
			sub := &domainTask.Task{
				ID:          subID,
				WorkspaceID: wID,
				ParentID:    &parentID,
				Title:       "Subtask 1",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			sub.SetSubtaskCounts(2, 1)
			return []*domainTask.Task{sub}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+parentID.String()+"/subtasks", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.ListSubtasks(c, wsID, parentID))
	assert.Equal(t, http.StatusOK, rec.Code)

	var subtasks []generated.Task
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &subtasks))
	require.Len(t, subtasks, 1)
	assert.Equal(t, subID, subtasks[0].Id)
	require.NotNil(t, subtasks[0].ParentId)
	assert.Equal(t, parentID, *subtasks[0].ParentId)
	require.NotNil(t, subtasks[0].SubtasksCount)
	assert.Equal(t, 2, *subtasks[0].SubtasksCount)
	require.NotNil(t, subtasks[0].SubtasksCompletedCount)
	assert.Equal(t, 1, *subtasks[0].SubtasksCompletedCount)
}

func TestHandler_CreateInboxTask_WithParentId(t *testing.T) {
	wsID := uuid.New()
	parentID := uuid.New()
	var capturedInput taskApp.CreateInboxTaskInput

	mockSvc := &mockTaskService{
		createInboxFn: func(ctx context.Context, input taskApp.CreateInboxTaskInput) (*domainTask.Task, error) {
			capturedInput = input
			now := time.Now().UTC()
			return &domainTask.Task{
				ID:          uuid.New(),
				WorkspaceID: input.WorkspaceID,
				ParentID:    input.ParentID,
				Title:       input.Title,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	payload := `{"title":"Subtask Title","parentId":"` + parentID.String() + `"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/inbox", bytes.NewReader([]byte(payload)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(WorkspaceIDContextKey), wsID)

	require.NoError(t, handler.CreateInboxTask(c))
	assert.Equal(t, http.StatusCreated, rec.Code)
	require.NotNil(t, capturedInput.ParentID)
	assert.Equal(t, parentID, *capturedInput.ParentID)
}

func TestHandler_UpdateTask_WithParentId(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	parentID := uuid.New()
	var capturedInput taskApp.UpdateTaskInput

	mockSvc := &mockTaskService{
		updateFn: func(ctx context.Context, input taskApp.UpdateTaskInput) (*domainTask.Task, error) {
			capturedInput = input
			now := time.Now().UTC()
			return &domainTask.Task{
				ID:          input.TaskID,
				WorkspaceID: input.WorkspaceID,
				ParentID:    input.ParentID,
				Title:       "Updated",
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	t.Run("set parentId", func(t *testing.T) {
		payload := `{"parentId":"` + parentID.String() + `"}`
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+taskID.String(), bytes.NewReader([]byte(payload)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		require.NoError(t, handler.UpdateTask(c, wsID, taskID))
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, capturedInput.ParentID)
		assert.Equal(t, parentID, *capturedInput.ParentID)
		assert.False(t, capturedInput.ClearParentID)
	})

	t.Run("clear parentId with null", func(t *testing.T) {
		payload := `{"parentId":null}`
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+taskID.String(), bytes.NewReader([]byte(payload)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		require.NoError(t, handler.UpdateTask(c, wsID, taskID))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, capturedInput.ClearParentID)
	})
}
