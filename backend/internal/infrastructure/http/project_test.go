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
)

type mockProjectService struct {
	createProjectFn func(ctx context.Context, input projectApp.CreateProjectInput) (*domainProject.Project, error)
	getProjectsFn   func(ctx context.Context, wsID uuid.UUID) ([]*domainProject.Project, error)
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

func (m *mockProjectService) GetProject(ctx context.Context, projectID, wsID uuid.UUID) (*domainProject.Project, error) {
	return nil, nil
}

func (m *mockProjectService) UpdateProject(ctx context.Context, input projectApp.UpdateProjectInput) (*domainProject.Project, error) {
	return nil, nil
}

func (m *mockProjectService) DeleteProject(ctx context.Context, projectID, wsID uuid.UUID) error {
	return nil
}

func (m *mockProjectService) GetProjectBoards(ctx context.Context, projectID, wsID uuid.UUID) ([]*domainProject.Board, error) {
	return nil, nil
}

func (m *mockProjectService) CreateBoard(ctx context.Context, projectID, wsID uuid.UUID, name string) (*domainProject.Board, error) {
	return nil, nil
}

func (m *mockProjectService) GetBoardColumns(ctx context.Context, boardID, wsID uuid.UUID) ([]*domainProject.Column, error) {
	return nil, nil
}

func (m *mockProjectService) CreateColumn(ctx context.Context, input projectApp.CreateColumnInput) (*domainProject.Column, error) {
	return nil, nil
}

func (m *mockProjectService) UpdateColumn(ctx context.Context, input projectApp.UpdateColumnInput) (*domainProject.Column, error) {
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

func (m *mockProjectService) GetProjectTasks(ctx context.Context, projectID, wsID uuid.UUID) ([]*domainTask.Task, error) {
	return nil, nil
}

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
		e := echo.New()
		body := `{"name":"Новый проект","description":"Описание","color":"#6366f1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/projects", bytes.NewReader([]byte(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("workspace_id")
		c.SetParamValues(wsID.String())

		err := handler.CreateProject(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp ProjectResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Новый проект", resp.Name)
		assert.Equal(t, projID, resp.ID)
	})

	t.Run("ListProjects success", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID.String()+"/projects", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("workspace_id")
		c.SetParamValues(wsID.String())

		err := handler.ListProjects(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var list []ProjectResponse
		err = json.Unmarshal(rec.Body.Bytes(), &list)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})
}

func TestHandler_DeleteColumn_Conflict(t *testing.T) {
	wsID := uuid.New()
	boardID := uuid.New()
	colID := uuid.New()

	mockSvc := &mockProjectService{
		deleteColumnFn: func(ctx context.Context, bID, cID, wID uuid.UUID) error {
			return domainProject.ErrColumnNotEmpty
		},
	}

	handler := NewHandler(nil, "0.3.1", false, nil, nil, mockSvc, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/"+wsID.String()+"/boards/"+boardID.String()+"/columns/"+colID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("workspace_id", "board_id", "column_id")
	c.SetParamValues(wsID.String(), boardID.String(), colID.String())

	err := handler.DeleteBoardColumn(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestHandler_MoveTask(t *testing.T) {
	wsID := uuid.New()
	colID := uuid.New()
	taskID := uuid.New()
	now := time.Now().UTC()

	mockSvc := &mockProjectService{
		moveTaskFn: func(ctx context.Context, input projectApp.MoveTaskInput) (*domainTask.Task, error) {
			return &domainTask.Task{
				ID:          input.TaskID,
				WorkspaceID: input.WorkspaceID,
				ColumnID:    &input.ColumnID,
				Title:       "Задача перемещена",
				Position:    int(input.Position),
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	handler := NewHandler(nil, "0.3.1", false, nil, nil, mockSvc, nil)

	e := echo.New()
	body := `{"columnId":"` + colID.String() + `","position":1500.0}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID.String()+"/tasks/"+taskID.String()+"/move", bytes.NewReader([]byte(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("workspace_id", "task_id")
	c.SetParamValues(wsID.String(), taskID.String())

	err := handler.MoveTaskToColumn(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp TaskResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1500, resp.Position)
	assert.Equal(t, &colID, resp.ColumnID)
}
