package http

import (
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

func TestHandler_DuplicateTask_Table(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name           string
		duplicateFn    func(ctx context.Context, wID, tID uuid.UUID) (*domainTask.Task, error)
		expectedStatus int
		checkBody      func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "Success 201 Created",
			duplicateFn: func(ctx context.Context, wID, tID uuid.UUID) (*domainTask.Task, error) {
				return &domainTask.Task{
					ID:          uuid.New(),
					WorkspaceID: wID,
					Title:       "Копия — Исходная задача",
					Position:    2,
					Tags:        []*domainTask.Tag{},
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil
			},
			expectedStatus: http.StatusCreated,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var res generated.Task
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
				assert.Equal(t, "Копия — Исходная задача", res.Title)
				assert.Equal(t, float64(2), res.Position)
			},
		},
		{
			name: "Task Not Found 404",
			duplicateFn: func(ctx context.Context, wID, tID uuid.UUID) (*domainTask.Task, error) {
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
			name: "Invalid Title Error 400",
			duplicateFn: func(ctx context.Context, wID, tID uuid.UUID) (*domainTask.Task, error) {
				return nil, domainTask.ErrInvalidTaskTitle
			},
			expectedStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var errResp generated.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
				assert.Equal(t, "BAD_REQUEST", errResp.Code)
			},
		},
		{
			name: "Internal Error 500",
			duplicateFn: func(ctx context.Context, wID, tID uuid.UUID) (*domainTask.Task, error) {
				return nil, errors.New("db failure")
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
			mockSvc := &mockTaskService{duplicateTaskFn: tc.duplicateFn}
			handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+taskID.String()+"/duplicate", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.DuplicateTask(c, wsID, taskID)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, rec)
			}
		})
	}
}
