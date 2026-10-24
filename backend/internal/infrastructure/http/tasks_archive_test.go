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

	domainTask "takt/internal/domain/task"
	"takt/internal/infrastructure/http/generated"
)

func TestHandler_ListArchivedTasks_Table(t *testing.T) {
	wsID := uuid.New()
	projID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name           string
		listFn         func(ctx context.Context, wID uuid.UUID, pID *uuid.UUID) ([]*domainTask.Task, error)
		params         generated.ListArchivedTasksParams
		expectedStatus int
		checkBody      func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "Success 200 returns archived tasks",
			listFn: func(ctx context.Context, wID uuid.UUID, pID *uuid.UUID) ([]*domainTask.Task, error) {
				return []*domainTask.Task{
					{
						ID:          uuid.New(),
						WorkspaceID: wID,
						ProjectID:   pID,
						Title:       "Archived item",
						IsArchived:  true,
						ArchivedAt:  &now,
						CreatedAt:   now,
						UpdatedAt:   now,
					},
				}, nil
			},
			params:         generated.ListArchivedTasksParams{ProjectId: &projID},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var tasks []generated.Task
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tasks))
				assert.Len(t, tasks, 1)
				assert.Equal(t, "Archived item", tasks[0].Title)
				assert.True(t, tasks[0].IsArchived)
			},
		},
		{
			name: "Internal Error 500",
			listFn: func(ctx context.Context, wID uuid.UUID, pID *uuid.UUID) ([]*domainTask.Task, error) {
				return nil, errors.New("db error")
			},
			params:         generated.ListArchivedTasksParams{},
			expectedStatus: http.StatusInternalServerError,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var errResp generated.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
				assert.Equal(t, "INTERNAL_ERROR", errResp.Code)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockTaskService{listArchivedFn: tc.listFn}
			handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID.String()+"/tasks/archived", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.ListArchivedTasks(c, wsID, tc.params)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, rec)
			}
		})
	}
}

func TestHandler_UnarchiveTask_Table(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	colID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name           string
		body           string
		unarchiveFn    func(ctx context.Context, wID, tID, cID uuid.UUID, bID *uuid.UUID) (*domainTask.Task, error)
		expectedStatus int
		checkBody      func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "Success 200 restored task",
			body: `{"columnId":"` + colID.String() + `"}`,
			unarchiveFn: func(ctx context.Context, wID, tID, cID uuid.UUID, bID *uuid.UUID) (*domainTask.Task, error) {
				return &domainTask.Task{
					ID:          tID,
					WorkspaceID: wID,
					ColumnID:    &cID,
					Title:       "Restored item",
					IsArchived:  false,
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var task generated.Task
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &task))
				assert.Equal(t, taskID, task.Id)
				assert.False(t, task.IsArchived)
			},
		},
		{
			name:           "Bad Request on malformed body",
			body:           `{not-valid-json`,
			unarchiveFn:    nil,
			expectedStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var errResp generated.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
				assert.Equal(t, "BAD_REQUEST", errResp.Code)
			},
		},
		{
			name: "Not Found 404",
			body: `{"columnId":"` + colID.String() + `"}`,
			unarchiveFn: func(ctx context.Context, wID, tID, cID uuid.UUID, bID *uuid.UUID) (*domainTask.Task, error) {
				return nil, domainTask.ErrTaskNotFound
			},
			expectedStatus: http.StatusNotFound,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var errResp generated.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
				assert.Equal(t, "NOT_FOUND", errResp.Code)
			},
		},
		{
			name: "Bad Request on invalid column id",
			body: `{"columnId":"` + colID.String() + `"}`,
			unarchiveFn: func(ctx context.Context, wID, tID, cID uuid.UUID, bID *uuid.UUID) (*domainTask.Task, error) {
				return nil, domainTask.ErrInvalidColumnID
			},
			expectedStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var errResp generated.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
				assert.Equal(t, "BAD_REQUEST", errResp.Code)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockTaskService{unarchiveFn: tc.unarchiveFn}
			handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+taskID.String()+"/unarchive", bytes.NewBufferString(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.UnarchiveTask(c, wsID, taskID)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, rec)
			}
		})
	}
}

func TestHandler_ArchiveCompletedColumnTasks_Table(t *testing.T) {
	wsID := uuid.New()
	colID := uuid.New()

	tests := []struct {
		name           string
		archiveInColFn func(ctx context.Context, wID, cID uuid.UUID) (int64, error)
		expectedStatus int
		checkBody      func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "Success 200 returns archived count",
			archiveInColFn: func(ctx context.Context, wID, cID uuid.UUID) (int64, error) {
				return 5, nil
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.Equal(t, float64(5), resp["archivedCount"])
			},
		},
		{
			name: "Internal Error 500",
			archiveInColFn: func(ctx context.Context, wID, cID uuid.UUID) (int64, error) {
				return 0, errors.New("db error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var errResp generated.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
				assert.Equal(t, "INTERNAL_ERROR", errResp.Code)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockTaskService{archiveInColFn: tc.archiveInColFn}
			handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/columns/"+colID.String()+"/archive-completed", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.ArchiveCompletedColumnTasks(c, wsID, colID)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, rec)
			}
		})
	}
}
