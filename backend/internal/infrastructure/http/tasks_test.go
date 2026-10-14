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
	taskApp "takt/internal/application/task"
	domainTask "takt/internal/domain/task"
)

type mockTaskService struct {
	getInboxFn    func(ctx context.Context, wsID uuid.UUID) ([]*domainTask.Task, error)
	createInboxFn func(ctx context.Context, input taskApp.CreateInboxTaskInput) (*domainTask.Task, error)
	toggleFn      func(ctx context.Context, id, wsID uuid.UUID) (*domainTask.Task, error)
	deleteFn      func(ctx context.Context, id, wsID uuid.UUID) error
}

func (m *mockTaskService) GetInboxTasks(ctx context.Context, wsID uuid.UUID) ([]*domainTask.Task, error) {
	if m.getInboxFn != nil {
		return m.getInboxFn(ctx, wsID)
	}
	return nil, nil
}

func (m *mockTaskService) CreateInboxTask(ctx context.Context, input taskApp.CreateInboxTaskInput) (*domainTask.Task, error) {
	if m.createInboxFn != nil {
		return m.createInboxFn(ctx, input)
	}
	return nil, nil
}

func (m *mockTaskService) ToggleTask(ctx context.Context, id, wsID uuid.UUID) (*domainTask.Task, error) {
	if m.toggleFn != nil {
		return m.toggleFn(ctx, id, wsID)
	}
	return nil, nil
}

func (m *mockTaskService) DeleteTask(ctx context.Context, id, wsID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, wsID)
	}
	return nil
}

func (m *mockTaskService) CreateInitialInboxTasks(ctx context.Context, wsID uuid.UUID) error {
	return nil
}

func TestHandler_GetInboxTasks_Success(t *testing.T) {
	wsID := uuid.New()
	mockSvc := &mockTaskService{
		getInboxFn: func(ctx context.Context, wID uuid.UUID) ([]*domainTask.Task, error) {
			now := time.Now().UTC()
			return []*domainTask.Task{
				{
					ID:                  uuid.New(),
					WorkspaceID:         wID,
					Title:               "Task 1",
					DescriptionMarkdown: "",
					IsCompleted:         false,
					Priority:            domainTask.PriorityHigh,
					Position:            0,
					CreatedAt:           now,
					UpdatedAt:           now,
				},
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/inbox", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(WorkspaceIDContextKey), wsID)

	if err := handler.GetInboxTasks(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var tasks []TaskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(tasks) != 1 || tasks[0].Title != "Task 1" {
		t.Errorf("unexpected task payload: %v", tasks)
	}
}

func TestHandler_CreateInboxTask_Success(t *testing.T) {
	wsID := uuid.New()
	mockSvc := &mockTaskService{
		createInboxFn: func(ctx context.Context, input taskApp.CreateInboxTaskInput) (*domainTask.Task, error) {
			now := time.Now().UTC()
			return &domainTask.Task{
				ID:                  uuid.New(),
				WorkspaceID:         input.WorkspaceID,
				Title:               input.Title,
				DescriptionMarkdown: input.DescriptionMarkdown,
				Priority:            input.Priority,
				Position:            0,
				CreatedAt:           now,
				UpdatedAt:           now,
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	payload := `{"title":"Quick Task","priority":"medium"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/inbox", bytes.NewReader([]byte(payload)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(WorkspaceIDContextKey), wsID)

	if err := handler.CreateInboxTask(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", rec.Code)
	}

	var task TaskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse task: %v", err)
	}
	if task.Title != "Quick Task" || task.Priority != domainTask.PriorityMedium {
		t.Errorf("unexpected task fields: %v", task)
	}
}

func TestHandler_ToggleTask_Success(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	now := time.Now().UTC()

	mockSvc := &mockTaskService{
		toggleFn: func(ctx context.Context, id, wID uuid.UUID) (*domainTask.Task, error) {
			return &domainTask.Task{
				ID:          id,
				WorkspaceID: wID,
				Title:       "Toggled",
				IsCompleted: true,
				CompletedAt: &now,
				Position:    0,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+taskID.String()+"/toggle", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(taskID.String())
	c.Set(string(WorkspaceIDContextKey), wsID)

	if err := handler.ToggleTask(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var task TaskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse task: %v", err)
	}
	if !task.IsCompleted || task.CompletedAt == nil {
		t.Errorf("expected completed task with completedAt set")
	}
}

func TestHandler_DeleteTask_Success(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()

	mockSvc := &mockTaskService{
		deleteFn: func(ctx context.Context, id, wID uuid.UUID) error {
			return nil
		},
	}

	handler := NewHandler(nil, "0.2.1", false, nil, mockSvc, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+taskID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(taskID.String())
	c.Set(string(WorkspaceIDContextKey), wsID)

	if err := handler.DeleteTask(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 NoContent, got %d", rec.Code)
	}
}
