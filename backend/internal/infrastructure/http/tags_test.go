package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func TestHandler_ListWorkspaceTags_Success(t *testing.T) {
	wsID := uuid.New()
	tagID := uuid.New()
	projectID := uuid.New()
	now := time.Now().UTC()

	mockSvc := &mockTaskService{
		listTagsFn: func(ctx context.Context, wID uuid.UUID, pID *uuid.UUID) ([]*domainTask.Tag, error) {
			assert.Equal(t, &projectID, pID)
			return []*domainTask.Tag{
				{
					ID:          tagID,
					WorkspaceID: wID,
					ProjectID:   pID,
					Name:        "Bug",
					ColorHex:    "#EF4444",
					CreatedAt:   now,
				},
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID.String()+"/tags?project_id="+projectID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.ListWorkspaceTags(c, wsID, generated.ListWorkspaceTagsParams{ProjectId: &projectID}))
	assert.Equal(t, http.StatusOK, rec.Code)

	var tags []generated.Tag
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tags))
	require.Len(t, tags, 1)
	assert.Equal(t, "Bug", tags[0].Name)
	assert.Equal(t, "#EF4444", tags[0].ColorHex)
	require.NotNil(t, tags[0].ProjectId)
	assert.Equal(t, projectID, *tags[0].ProjectId)
}

func TestHandler_CreateWorkspaceTag_Success(t *testing.T) {
	wsID := uuid.New()
	tagID := uuid.New()
	now := time.Now().UTC()

	mockSvc := &mockTaskService{
		createTagFn: func(ctx context.Context, input taskApp.CreateTagInput) (*domainTask.Tag, error) {
			return &domainTask.Tag{
				ID:          tagID,
				WorkspaceID: input.WorkspaceID,
				Name:        input.Name,
				ColorHex:    input.ColorHex,
				CreatedAt:   now,
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	payload := `{"name":"Frontend","colorHex":"#3B82F6"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/tags", bytes.NewReader([]byte(payload)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.CreateWorkspaceTag(c, wsID))
	assert.Equal(t, http.StatusCreated, rec.Code)

	var tag generated.Tag
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tag))
	assert.Equal(t, "Frontend", tag.Name)
	assert.Equal(t, "#3B82F6", tag.ColorHex)
}

func TestHandler_CreateWorkspaceTag_ValidationError(t *testing.T) {
	wsID := uuid.New()

	mockSvc := &mockTaskService{
		createTagFn: func(ctx context.Context, input taskApp.CreateTagInput) (*domainTask.Tag, error) {
			return nil, domainTask.ErrInvalidTagName
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	payload := `{"name":"","colorHex":"#3B82F6"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/tags", bytes.NewReader([]byte(payload)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.CreateWorkspaceTag(c, wsID))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_DeleteWorkspaceTag_Success(t *testing.T) {
	wsID := uuid.New()
	tagID := uuid.New()

	mockSvc := &mockTaskService{
		deleteTagFn: func(ctx context.Context, tID, wID uuid.UUID) error {
			return nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/"+wsID.String()+"/tags/"+tagID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.DeleteWorkspaceTag(c, wsID, tagID))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_DeleteWorkspaceTag_NotFound(t *testing.T) {
	wsID := uuid.New()
	tagID := uuid.New()

	mockSvc := &mockTaskService{
		deleteTagFn: func(ctx context.Context, tID, wID uuid.UUID) error {
			return domainTask.ErrTagNotFound
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/"+wsID.String()+"/tags/"+tagID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.DeleteWorkspaceTag(c, wsID, tagID))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_AddTagToTask_Success(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	tagID := uuid.New()

	mockSvc := &mockTaskService{
		addTagToTaskFn: func(ctx context.Context, tskID, tgID, wID uuid.UUID) error {
			return nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+taskID.String()+"/tags/"+tagID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.AddTagToTask(c, wsID, taskID, tagID))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_RemoveTagFromTask_Success(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	tagID := uuid.New()

	mockSvc := &mockTaskService{
		removeTagFromTaskFn: func(ctx context.Context, tskID, tgID, wID uuid.UUID) error {
			return nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+taskID.String()+"/tags/"+tagID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.RemoveTagFromTask(c, wsID, taskID, tagID))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_UpdateTag(t *testing.T) {
	wsID := uuid.New()
	tagID := uuid.New()
	now := time.Now().UTC()

	t.Run("success update tag", func(t *testing.T) {
		mockSvc := &mockTaskService{
			updateTagFn: func(ctx context.Context, input taskApp.UpdateTagInput) (*domainTask.Tag, error) {
				return &domainTask.Tag{
					ID:          tagID,
					WorkspaceID: wsID,
					Name:        *input.Name,
					ColorHex:    *input.ColorHex,
					CreatedAt:   now,
				}, nil
			},
		}

		handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)
		payload := `{"name":"Backend","colorHex":"#10B981"}`
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wsID.String()+"/tags/"+tagID.String(), bytes.NewReader([]byte(payload)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		require.NoError(t, handler.UpdateTag(c, wsID, tagID))
		assert.Equal(t, http.StatusOK, rec.Code)

		var tag generated.Tag
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tag))
		assert.Equal(t, "Backend", tag.Name)
		assert.Equal(t, "#10B981", tag.ColorHex)
	})

	t.Run("invalid payload", func(t *testing.T) {
		handler := NewHandler(nil, "0.2.1", false, nil, &mockTaskService{}, nil, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wsID.String()+"/tags/"+tagID.String(), bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		require.NoError(t, handler.UpdateTag(c, wsID, tagID))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		mockSvc := &mockTaskService{
			updateTagFn: func(ctx context.Context, input taskApp.UpdateTagInput) (*domainTask.Tag, error) {
				return nil, domainTask.ErrInvalidTagColor
			},
		}
		handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)
		payload := `{"colorHex":"invalid"}`
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wsID.String()+"/tags/"+tagID.String(), bytes.NewReader([]byte(payload)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		require.NoError(t, handler.UpdateTag(c, wsID, tagID))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("tag not found", func(t *testing.T) {
		mockSvc := &mockTaskService{
			updateTagFn: func(ctx context.Context, input taskApp.UpdateTagInput) (*domainTask.Tag, error) {
				return nil, domainTask.ErrTagNotFound
			},
		}
		handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)
		payload := `{"name":"NewName"}`
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wsID.String()+"/tags/"+tagID.String(), bytes.NewReader([]byte(payload)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		require.NoError(t, handler.UpdateTag(c, wsID, tagID))
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestHandler_ListWorkspaceTagsMetrics_Success(t *testing.T) {
	wsID := uuid.New()
	tagID := uuid.New()
	now := time.Now().UTC()

	mockSvc := &mockTaskService{
		listTagsMetricsFn: func(ctx context.Context, wID uuid.UUID) ([]*domainTask.TagWithMetrics, error) {
			m := domainTask.NewTagWithMetrics(domainTask.Tag{
				ID:          tagID,
				WorkspaceID: wID,
				Name:        "Feature",
				ColorHex:    "#3B82F6",
				CreatedAt:   now,
			}, 3)
			return []*domainTask.TagWithMetrics{&m}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID.String()+"/tags/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.ListWorkspaceTagsMetrics(c, wsID))
	assert.Equal(t, http.StatusOK, rec.Code)

	var tags []generated.TagWithMetrics
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tags))
	require.Len(t, tags, 1)
	assert.Equal(t, "Feature", tags[0].Name)
	assert.Equal(t, "#3B82F6", tags[0].ColorHex)
	assert.Equal(t, 3, tags[0].TaskCount)
}

func TestHandler_DeleteUnusedWorkspaceTags(t *testing.T) {
	wsID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := &mockTaskService{
			deleteUnusedTagsFn: func(ctx context.Context, wID uuid.UUID) error {
				return nil
			},
		}

		handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/"+wsID.String()+"/tags/unused", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		require.NoError(t, handler.DeleteUnusedWorkspaceTags(c, wsID))
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		mockSvc := &mockTaskService{
			deleteUnusedTagsFn: func(ctx context.Context, wID uuid.UUID) error {
				return errors.New("db failure")
			},
		}

		handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/"+wsID.String()+"/tags/unused", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		require.NoError(t, handler.DeleteUnusedWorkspaceTags(c, wsID))
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}



