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

	appSearch "takt/internal/application/search"
	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
	domainWorkspace "takt/internal/domain/workspace"
	"takt/internal/infrastructure/http/generated"
)

type mockSearchService struct {
	searchFunc func(ctx context.Context, wsID, uID uuid.UUID, query string) (*appSearch.SearchResult, error)
}

func (m *mockSearchService) Search(ctx context.Context, wsID, uID uuid.UUID, query string) (*appSearch.SearchResult, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, wsID, uID, query)
	}
	return nil, nil
}

func newSearchTestContext(url string, userID, wsID uuid.UUID) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if userID != uuid.Nil {
		c.Set(string(UserIDContextKey), userID)
	}
	if wsID != uuid.Nil {
		c.Set(string(WorkspaceIDContextKey), wsID)
	}
	return c, rec
}

func TestHandler_SearchWorkspace(t *testing.T) {
	wsID := uuid.New()
	userID := uuid.New()

	dummyTask := &domainTask.Task{
		ID:          uuid.New(),
		WorkspaceID: wsID,
		Title:       "Срочная задача",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	dummyProject := &domainProject.Project{
		ID:          uuid.New(),
		WorkspaceID: wsID,
		Name:        "Проект Альфа",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	tests := []struct {
		name       string
		noAuth     bool
		noService  bool
		query      string
		mockErr    error
		tasks      []*domainTask.Task
		projects   []*domainProject.Project
		wantStatus int
	}{
		{
			name:       "successful search",
			query:      "альфа",
			tasks:      []*domainTask.Task{dummyTask},
			projects:   []*domainProject.Project{dummyProject},
			wantStatus: http.StatusOK,
		},
		{
			name:       "unauthorized if user not in context",
			noAuth:     true,
			query:      "альфа",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "service unavailable",
			noService:  true,
			query:      "альфа",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "forbidden access to workspace",
			query:      "альфа",
			mockErr:    domainWorkspace.ErrAccessDenied,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "internal error from service",
			query:      "альфа",
			mockErr:    errors.New("db error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockSearchService{
				searchFunc: func(ctx context.Context, wID, uID uuid.UUID, q string) (*appSearch.SearchResult, error) {
					if tc.mockErr != nil {
						return nil, tc.mockErr
					}
					return &appSearch.SearchResult{
						Tasks:    tc.tasks,
						Projects: tc.projects,
					}, nil
				},
			}

			u := userID
			if tc.noAuth {
				u = uuid.Nil
			}

			c, rec := newSearchTestContext("/workspaces/"+wsID.String()+"/search?q="+tc.query, u, wsID)
			h := NewHandler(nil, "0.4.3", false, nil, nil, nil, nil)
			if !tc.noService {
				h.SetSearchService(mockSvc)
			}

			params := generated.SearchWorkspaceParams{
				Q: tc.query,
			}

			err := h.SearchWorkspace(c, wsID, params)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				var resp generated.SearchResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.Len(t, resp.Tasks, len(tc.tasks))
				assert.Len(t, resp.Projects, len(tc.projects))
			}
		})
	}
}
