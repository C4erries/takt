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

	projectApp "takt/internal/application/project"
	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
	"takt/internal/infrastructure/http/generated"
)

type mockProjectService struct {
	createProjectFn func(ctx context.Context, input projectApp.CreateProjectInput) (*domainProject.Project, error)
	getProjectsFn   func(ctx context.Context, wsID uuid.UUID) ([]*domainProject.Project, error)
	updateBoardFn   func(ctx context.Context, input projectApp.UpdateBoardInput) (*domainProject.Board, error)
	deleteBoardFn   func(ctx context.Context, projectID, boardID, wsID uuid.UUID) error
	updateColumnFn  func(ctx context.Context, input projectApp.UpdateColumnInput) (*domainProject.Column, error)
	deleteColumnFn  func(ctx context.Context, boardID, colID, wsID uuid.UUID) error
	moveTaskFn      func(ctx context.Context, input projectApp.MoveTaskInput) (*domainTask.Task, error)
}

func (m *mockProjectService) CreateProject(ctx context.Context, input projectApp.CreateProjectInput) (*domainProject.Project, error) {
	if m.createProjectFn != nil {
		return m.createProjectFn(ctx, input)
	}
	return nil, nil
}

func (m *mockProjectService) GetProjects(ctx context.Context, wsID uuid.UUID) ([]*domainProject.Project, error) {
	if m.getProjectsFn != nil {
		return m.getProjectsFn(ctx, wsID)
	}
	return nil, nil
}

func (m *mockProjectService) GetProject(ctx context.Context, projectID, wsID uuid.UUID) (*domainProject.Project, error) { return nil, nil }
func (m *mockProjectService) UpdateProject(ctx context.Context, input projectApp.UpdateProjectInput) (*domainProject.Project, error) { return nil, nil }
func (m *mockProjectService) DeleteProject(ctx context.Context, projectID, wsID uuid.UUID) error { return nil }
func (m *mockProjectService) GetProjectBoards(ctx context.Context, projectID, wsID uuid.UUID) ([]*domainProject.Board, error) { return nil, nil }
func (m *mockProjectService) CreateBoard(ctx context.Context, projectID, wsID uuid.UUID, name string) (*domainProject.Board, error) { return nil, nil }

func (m *mockProjectService) UpdateBoard(ctx context.Context, input projectApp.UpdateBoardInput) (*domainProject.Board, error) {
	if m.updateBoardFn != nil {
		return m.updateBoardFn(ctx, input)
	}
	return nil, nil
}

func (m *mockProjectService) DeleteBoard(ctx context.Context, projectID, boardID, wsID uuid.UUID) error {
	if m.deleteBoardFn != nil {
		return m.deleteBoardFn(ctx, projectID, boardID, wsID)
	}
	return nil
}

func (m *mockProjectService) GetBoardColumns(ctx context.Context, boardID, wsID uuid.UUID) ([]*domainProject.Column, error) { return nil, nil }
func (m *mockProjectService) CreateColumn(ctx context.Context, input projectApp.CreateColumnInput) (*domainProject.Column, error) { return nil, nil }

func (m *mockProjectService) UpdateColumn(ctx context.Context, input projectApp.UpdateColumnInput) (*domainProject.Column, error) {
	if m.updateColumnFn != nil {
		return m.updateColumnFn(ctx, input)
	}
	return nil, nil
}

func (m *mockProjectService) DeleteColumn(ctx context.Context, boardID, colID, wsID uuid.UUID) error {
	if m.deleteColumnFn != nil {
		return m.deleteColumnFn(ctx, boardID, colID, wsID)
	}
	return nil
}

func (m *mockProjectService) MoveTask(ctx context.Context, input projectApp.MoveTaskInput) (*domainTask.Task, error) {
	if m.moveTaskFn != nil {
		return m.moveTaskFn(ctx, input)
	}
	return nil, nil
}

func (m *mockProjectService) GetProjectTasks(ctx context.Context, projectID, wsID uuid.UUID) ([]*domainTask.Task, error) { return nil, nil }

func TestHandler_Projects_CRUD(t *testing.T) {
	wsID := uuid.New()
	projID := uuid.New()
	now := time.Now().UTC()

	mockSvc := &mockProjectService{
		createProjectFn: func(ctx context.Context, input projectApp.CreateProjectInput) (*domainProject.Project, error) {
			return &domainProject.Project{
				ID:          projID,
				WorkspaceID: input.WorkspaceID,
				Name:        input.Name,
				Description: input.Description,
				Color:       input.Color,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
		getProjectsFn: func(ctx context.Context, wID uuid.UUID) ([]*domainProject.Project, error) {
			return []*domainProject.Project{
				{
					ID:          projID,
					WorkspaceID: wID,
					Name:        "Проект 1",
					Color:       "#6366f1",
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			}, nil
		},
	}

	handler := NewHandler(nil, "0.3.1", false, nil, nil, mockSvc, nil)

	t.Run("CreateProject success", func(t *testing.T) {
		body := `{"name":"Новый проект","description":"Описание","color":"#6366f1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/projects", bytes.NewReader([]byte(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		require.NoError(t, handler.CreateProject(echo.New().NewContext(req, rec), wsID))
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp generated.Project
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "Новый проект", resp.Name)
		assert.Equal(t, projID, resp.Id)
	})

	t.Run("ListProjects success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID.String()+"/projects", nil)
		rec := httptest.NewRecorder()
		require.NoError(t, handler.ListProjects(echo.New().NewContext(req, rec), wsID))
		assert.Equal(t, http.StatusOK, rec.Code)

		var list []generated.Project
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
		assert.Len(t, list, 1)
	})
}

func TestHandler_DeleteColumn_Conflict(t *testing.T) {
	wsID, boardID, colID := uuid.New(), uuid.New(), uuid.New()
	mockSvc := &mockProjectService{
		deleteColumnFn: func(ctx context.Context, bID, cID, wID uuid.UUID) error {
			return domainProject.ErrColumnNotEmpty
		},
	}
	handler := NewHandler(nil, "0.3.1", false, nil, nil, mockSvc, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/"+wsID.String()+"/boards/"+boardID.String()+"/columns/"+colID.String(), nil)
	rec := httptest.NewRecorder()
	require.NoError(t, handler.DeleteBoardColumn(echo.New().NewContext(req, rec), wsID, boardID, colID))
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestHandler_MoveTask(t *testing.T) {
	wsID, colID, taskID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	mockSvc := &mockProjectService{
		moveTaskFn: func(ctx context.Context, input projectApp.MoveTaskInput) (*domainTask.Task, error) {
			return &domainTask.Task{
				ID: input.TaskID, WorkspaceID: input.WorkspaceID, ColumnID: &input.ColumnID,
				Title: "Задача перемещена", Position: int(input.Position), CreatedAt: now, UpdatedAt: now,
			}, nil
		},
	}
	handler := NewHandler(nil, "0.3.1", false, nil, nil, mockSvc, nil)

	body := `{"columnId":"` + colID.String() + `","position":1500.0}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+taskID.String()+"/move", bytes.NewReader([]byte(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	require.NoError(t, handler.MoveTaskToColumn(echo.New().NewContext(req, rec), wsID, taskID))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp generated.Task
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, float64(1500), resp.Position)
	assert.Equal(t, &colID, resp.ColumnId)
}

func TestHandler_UpdateBoardColumn(t *testing.T) {
	wsID, boardID, targetBoardID, colID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()

	testCases := []struct {
		name       string
		body       string
		mockFn     func(ctx context.Context, input projectApp.UpdateColumnInput) (*domainProject.Column, error)
		wantStatus int
		wantCode   string
	}{
		{
			name: "successful update column with board move",
			body: `{"name":"Готово к релизу","boardId":"` + targetBoardID.String() + `","position":4000.0}`,
			mockFn: func(ctx context.Context, input projectApp.UpdateColumnInput) (*domainProject.Column, error) {
				assert.Equal(t, wsID, input.WorkspaceID)
				assert.Equal(t, colID, input.ColumnID)
				require.NotNil(t, input.BoardID)
				assert.Equal(t, targetBoardID, *input.BoardID)
				assert.Equal(t, "Готово к релизу", input.Name)
				assert.Equal(t, 4000.0, input.Position)
				return &domainProject.Column{
					ID: colID, BoardID: targetBoardID, Name: input.Name, Position: input.Position,
					StatusCategory: domainProject.StatusCategoryDone, CreatedAt: now, UpdatedAt: now,
				}, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "error invalid board ID",
			body: `{"boardId":"00000000-0000-0000-0000-000000000000"}`,
			mockFn: func(ctx context.Context, input projectApp.UpdateColumnInput) (*domainProject.Column, error) {
				return nil, domainProject.ErrInvalidBoardID
			},
			wantStatus: http.StatusBadRequest, wantCode: "BAD_REQUEST",
		},
		{
			name: "error target board not found",
			body: `{"boardId":"` + targetBoardID.String() + `"}`,
			mockFn: func(ctx context.Context, input projectApp.UpdateColumnInput) (*domainProject.Column, error) {
				return nil, domainProject.ErrBoardNotFound
			},
			wantStatus: http.StatusNotFound, wantCode: "BOARD_NOT_FOUND",
		},
		{
			name: "error access denied on target board",
			body: `{"boardId":"` + targetBoardID.String() + `"}`,
			mockFn: func(ctx context.Context, input projectApp.UpdateColumnInput) (*domainProject.Column, error) {
				return nil, domainProject.ErrAccessDenied
			},
			wantStatus: http.StatusForbidden, wantCode: "FORBIDDEN",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewHandler(nil, "0.3.1", false, nil, nil, &mockProjectService{updateColumnFn: tc.mockFn}, nil)
			req := httptest.NewRequest(
				http.MethodPatch, "/api/v1/workspaces/"+wsID.String()+"/boards/"+boardID.String()+"/columns/"+colID.String(),
				bytes.NewReader([]byte(tc.body)),
			)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			require.NoError(t, handler.UpdateBoardColumn(echo.New().NewContext(req, rec), wsID, boardID, colID))
			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantCode != "" {
				var errResp generated.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
				assert.Equal(t, tc.wantCode, errResp.Code)
			}
		})
	}
}

func TestHandler_UpdateProjectBoard(t *testing.T) {
	wsID, projID, boardID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()

	testCases := []struct {
		name       string
		body       string
		mockFn     func(ctx context.Context, input projectApp.UpdateBoardInput) (*domainProject.Board, error)
		wantStatus int
		wantPos    float64
		wantName   string
	}{
		{
			name: "update board name and position",
			body: `{"name":"Спринт 2","position":2500.0}`,
			mockFn: func(ctx context.Context, input projectApp.UpdateBoardInput) (*domainProject.Board, error) {
				assert.Equal(t, wsID, input.WorkspaceID)
				assert.Equal(t, projID, input.ProjectID)
				assert.Equal(t, boardID, input.BoardID)
				require.NotNil(t, input.Name)
				assert.Equal(t, "Спринт 2", *input.Name)
				require.NotNil(t, input.Position)
				assert.Equal(t, 2500.0, *input.Position)
				return &domainProject.Board{ID: boardID, ProjectID: projID, Name: *input.Name, Position: *input.Position, CreatedAt: now, UpdatedAt: now}, nil
			},
			wantStatus: http.StatusOK,
			wantPos:    2500.0,
			wantName:   "Спринт 2",
		},
		{
			name: "update only position",
			body: `{"position":3000.0}`,
			mockFn: func(ctx context.Context, input projectApp.UpdateBoardInput) (*domainProject.Board, error) {
				assert.Nil(t, input.Name)
				require.NotNil(t, input.Position)
				assert.Equal(t, 3000.0, *input.Position)
				return &domainProject.Board{ID: boardID, ProjectID: projID, Name: "Основная", Position: *input.Position, CreatedAt: now, UpdatedAt: now}, nil
			},
			wantStatus: http.StatusOK,
			wantPos:    3000.0,
			wantName:   "Основная",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewHandler(nil, "0.3.1", false, nil, nil, &mockProjectService{updateBoardFn: tc.mockFn}, nil)
			e := echo.New()
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wsID.String()+"/projects/"+projID.String()+"/boards/"+boardID.String(), bytes.NewReader([]byte(tc.body)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.UpdateProjectBoard(c, wsID, projID, boardID)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, rec.Code)

			var resp generated.Board
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, tc.wantName, resp.Name)
			assert.Equal(t, tc.wantPos, resp.Position)
		})
	}
}

