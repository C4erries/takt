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

	appAutomation "takt/internal/application/automation"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/infrastructure/http/generated"
)

type mockAutomationService struct {
	listFn     func(ctx context.Context, wsID uuid.UUID) ([]*domainAutomation.Automation, error)
	listTaskFn func(ctx context.Context, wsID, taskID uuid.UUID) ([]*domainAutomation.Automation, error)
	toggleFn   func(ctx context.Context, wsID, autoID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error)
	seedFn     func(ctx context.Context, wsID uuid.UUID) error
	createFn   func(ctx context.Context, input appAutomation.CreateAutomationInput) (*domainAutomation.Automation, error)
	updateFn   func(ctx context.Context, input appAutomation.UpdateAutomationInput) (*domainAutomation.Automation, error)
	deleteFn   func(ctx context.Context, wsID, autoID uuid.UUID) error
	getFn      func(ctx context.Context, wsID, autoID uuid.UUID) (*domainAutomation.Automation, error)
}

func (m *mockAutomationService) ListAutomations(ctx context.Context, wsID uuid.UUID) ([]*domainAutomation.Automation, error) {
	if m.listFn != nil {
		return m.listFn(ctx, wsID)
	}
	return nil, nil
}

func (m *mockAutomationService) ListTaskAutomations(ctx context.Context, wsID, taskID uuid.UUID) ([]*domainAutomation.Automation, error) {
	if m.listTaskFn != nil {
		return m.listTaskFn(ctx, wsID, taskID)
	}
	return nil, nil
}

func (m *mockAutomationService) ExecuteTaskAutomation(ctx context.Context, wsID, taskID, autoID uuid.UUID) error {
	return nil
}

func (m *mockAutomationService) ListColumnAutomations(ctx context.Context, wsID, colID uuid.UUID) ([]*domainAutomation.Automation, error) {
	return nil, nil
}

func (m *mockAutomationService) GetAutomation(ctx context.Context, wsID, autoID uuid.UUID) (*domainAutomation.Automation, error) {
	if m.getFn != nil {
		return m.getFn(ctx, wsID, autoID)
	}
	return nil, nil
}

func (m *mockAutomationService) ToggleAutomation(ctx context.Context, wsID, autoID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error) {
	if m.toggleFn != nil {
		return m.toggleFn(ctx, wsID, autoID, isEnabled)
	}
	return nil, nil
}

func (m *mockAutomationService) CreateAutomation(ctx context.Context, input appAutomation.CreateAutomationInput) (*domainAutomation.Automation, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return nil, nil
}

func (m *mockAutomationService) UpdateAutomation(ctx context.Context, input appAutomation.UpdateAutomationInput) (*domainAutomation.Automation, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, input)
	}
	return nil, nil
}

func (m *mockAutomationService) DeleteAutomation(ctx context.Context, wsID, autoID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, wsID, autoID)
	}
	return nil
}

func (m *mockAutomationService) CreateColumnAutomation(ctx context.Context, input appAutomation.CreateColumnAutomationInput) (*domainAutomation.Automation, error) {
	return nil, nil
}

func (m *mockAutomationService) SeedDefaultAutomations(ctx context.Context, wsID uuid.UUID) error {
	if m.seedFn != nil {
		return m.seedFn(ctx, wsID)
	}
	return nil
}

func TestListWorkspaceAutomations(t *testing.T) {
	wsID := uuid.New()
	autoID := uuid.New()
	boardID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name       string
		setupMock  func(*mockAutomationService)
		noService  bool
		wantStatus int
		verify     func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "Success with board and JSON",
			setupMock: func(m *mockAutomationService) {
				m.listFn = func(ctx context.Context, id uuid.UUID) ([]*domainAutomation.Automation, error) {
					assert.Equal(t, wsID, id)
					return []*domainAutomation.Automation{
						{
							ID:             autoID,
							WorkspaceID:    wsID,
							BoardID:        &boardID,
							Name:           "Test Auto",
							TriggerType:    domainAutomation.TriggerTaskCompleted,
							ConditionsJSON: []byte(`{"has_due_date":true}`),
							ActionsJSON:    []byte(`{"action":"move_to_done"}`),
							IsEnabled:      true,
							CreatedAt:      now,
							UpdatedAt:      now,
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			verify: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []generated.Automation
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.Len(t, resp, 1)
				assert.Equal(t, autoID, resp[0].Id)
				assert.Equal(t, "Test Auto", resp[0].Name)
				assert.Equal(t, generated.AutomationTriggerTypeTaskCompleted, resp[0].TriggerType)
				assert.True(t, resp[0].IsEnabled)
				require.NotNil(t, resp[0].BoardId)
				assert.Equal(t, boardID, *resp[0].BoardId)
			},
		},
		{
			name: "Internal Error from service",
			setupMock: func(m *mockAutomationService) {
				m.listFn = func(ctx context.Context, id uuid.UUID) ([]*domainAutomation.Automation, error) {
					return nil, errors.New("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			verify: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var errResp generated.Error
				err := json.Unmarshal(rec.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.Equal(t, "INTERNAL_ERROR", errResp.Code)
			},
		},
		{
			name:       "Nil service returns 500",
			noService:  true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID.String()+"/automations", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			var h *Handler
			if tt.noService {
				h = NewHandler(nil, "v1", false, nil, nil, nil, nil)
			} else {
				mockSvc := &mockAutomationService{}
				if tt.setupMock != nil {
					tt.setupMock(mockSvc)
				}
				h = NewHandler(nil, "v1", false, nil, nil, nil, nil, mockSvc)
			}

			err := h.ListWorkspaceAutomations(c, wsID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.verify != nil {
				tt.verify(t, rec)
			}
		})
	}
}

func TestToggleAutomation(t *testing.T) {
	wsID := uuid.New()
	autoID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name       string
		body       interface{}
		setupMock  func(*mockAutomationService)
		noService  bool
		wantStatus int
		verify     func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "Success toggle enable",
			body: generated.ToggleAutomationJSONRequestBody{IsEnabled: false},
			setupMock: func(m *mockAutomationService) {
				m.toggleFn = func(ctx context.Context, w, a uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error) {
					assert.Equal(t, wsID, w)
					assert.Equal(t, autoID, a)
					assert.False(t, isEnabled)
					return &domainAutomation.Automation{
						ID:             autoID,
						WorkspaceID:    wsID,
						Name:           "Test Auto",
						TriggerType:    domainAutomation.TriggerTaskUpdated,
						ConditionsJSON: []byte(`{}`),
						ActionsJSON:    []byte(`{}`),
						IsEnabled:      false,
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			verify: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp generated.Automation
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, autoID, resp.Id)
				assert.False(t, resp.IsEnabled)
			},
		},
		{
			name:       "Invalid JSON payload",
			body:       "invalid-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Not Found",
			body: generated.ToggleAutomationJSONRequestBody{IsEnabled: true},
			setupMock: func(m *mockAutomationService) {
				m.toggleFn = func(ctx context.Context, w, a uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error) {
					return nil, domainAutomation.ErrAutomationNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "Internal Error",
			body: generated.ToggleAutomationJSONRequestBody{IsEnabled: true},
			setupMock: func(m *mockAutomationService) {
				m.toggleFn = func(ctx context.Context, w, a uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error) {
					return nil, errors.New("database failure")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Nil service returns 500",
			body:       generated.ToggleAutomationJSONRequestBody{IsEnabled: true},
			noService:  true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			var reqBody []byte
			if str, ok := tt.body.(string); ok {
				reqBody = []byte(str)
			} else {
				var err error
				reqBody, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/"+wsID.String()+"/automations/"+autoID.String(), bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			var h *Handler
			if tt.noService {
				h = NewHandler(nil, "v1", false, nil, nil, nil, nil)
			} else {
				mockSvc := &mockAutomationService{}
				if tt.setupMock != nil {
					tt.setupMock(mockSvc)
				}
				h = NewHandler(nil, "v1", false, nil, nil, nil, nil, mockSvc)
			}

			err := h.ToggleAutomation(c, wsID, autoID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestListWorkspaceAutomations_WithTaskId(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	rule, _ := domainAutomation.NewAutomation(wsID, nil, "Task Rule", domainAutomation.TriggerTaskTimer, nil, nil, true)
	rule.TaskID = &taskID

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/workspaces/"+wsID.String()+"/automations?taskId="+taskID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("workspace_id", wsID.String())

	mockSvc := &mockAutomationService{
		listTaskFn: func(ctx context.Context, wid, tid uuid.UUID) ([]*domainAutomation.Automation, error) {
			assert.Equal(t, wsID, wid)
			assert.Equal(t, taskID, tid)
			return []*domainAutomation.Automation{rule}, nil
		},
	}

	h := NewHandler(nil, "v1", false, nil, nil, nil, nil, mockSvc)
	err := h.ListWorkspaceAutomations(c, wsID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []generated.Automation
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp, 1)
	assert.Equal(t, &taskID, resp[0].TaskId)
}


