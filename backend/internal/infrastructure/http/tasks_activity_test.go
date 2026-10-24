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

func TestHandler_ListTaskActivities_Table(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	userID := uuid.New()
	userName := "Alice"
	now := time.Now().UTC()

	tests := []struct {
		name           string
		listFn         func(ctx context.Context, wID, tID uuid.UUID) ([]*domainTask.TaskActivity, error)
		expectedStatus int
		checkBody      func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "Success 200 with activities",
			listFn: func(ctx context.Context, wID, tID uuid.UUID) ([]*domainTask.TaskActivity, error) {
				return []*domainTask.TaskActivity{
					{
						ID:           uuid.New(),
						TaskID:       taskID,
						UserID:       &userID,
						UserName:     &userName,
						ActivityType: domainTask.ActivityColumnMoved,
						Payload:      map[string]interface{}{"from_column_id": "col-1", "to_column_id": "col-2"},
						CreatedAt:    now,
					},
					{
						ID:           uuid.New(),
						TaskID:       taskID,
						UserID:       nil,
						UserName:     nil,
						ActivityType: domainTask.ActivityStatusChanged,
						Payload:      map[string]interface{}{"is_completed": true},
						CreatedAt:    now.Add(time.Minute),
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []generated.TaskActivity
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.Len(t, resp, 2)
				assert.Equal(t, taskID, resp[0].TaskId)
				assert.Equal(t, domainTask.ActivityColumnMoved, resp[0].ActivityType)
				assert.Equal(t, &userName, resp[0].UserName)
				assert.Equal(t, domainTask.ActivityStatusChanged, resp[1].ActivityType)
				assert.Nil(t, resp[1].UserId)
			},
		},
		{
			name: "Task Not Found 404",
			listFn: func(ctx context.Context, wID, tID uuid.UUID) ([]*domainTask.TaskActivity, error) {
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
			name: "Internal Server Error 500",
			listFn: func(ctx context.Context, wID, tID uuid.UUID) ([]*domainTask.TaskActivity, error) {
				return nil, errors.New("database connection failed")
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockTaskService{listActivitiesFn: tc.listFn}
			handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/workspaces/"+wsID.String()+"/tasks/"+taskID.String()+"/activities", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.ListTaskActivities(c, wsID, taskID)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, rec)
			}
		})
	}
}
