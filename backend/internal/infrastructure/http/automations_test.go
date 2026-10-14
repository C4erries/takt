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

	domainAutomation "takt/internal/domain/automation"
	"takt/internal/infrastructure/http/generated"
)

type mockAutomationService struct {
	listFn   func(ctx context.Context, wsID uuid.UUID) ([]*domainAutomation.Automation, error)
	toggleFn func(ctx context.Context, wsID, autoID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error)
	seedFn   func(ctx context.Context, wsID uuid.UUID) error
}

func (m *mockAutomationService) ListAutomations(ctx context.Context, wsID uuid.UUID) ([]*domainAutomation.Automation, error) {
	if m.listFn != nil {
		return m.listFn(ctx, wsID)
	}
	return nil, nil
}

func (m *mockAutomationService) ToggleAutomation(ctx context.Context, wsID, autoID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error) {
	if m.toggleFn != nil {
		return m.toggleFn(ctx, wsID, autoID, isEnabled)
	}
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
				assert.Equal(t, generated.TaskCompleted, resp[0].TriggerType)
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

