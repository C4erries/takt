package task

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	domainTask "takt/internal/domain/task"
)

type mockTaskRepo struct {
	tasks map[uuid.UUID]*domainTask.Task
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{tasks: make(map[uuid.UUID]*domainTask.Task)}
}

func (m *mockTaskRepo) Create(ctx context.Context, t *domainTask.Task) error {
	m.tasks[t.ID] = t
	return nil
}

func (m *mockTaskRepo) GetInboxTasks(ctx context.Context, wsID uuid.UUID) ([]*domainTask.Task, error) {
	var list []*domainTask.Task
	for _, t := range m.tasks {
		if t.WorkspaceID == wsID && t.ProjectID == nil && t.ColumnID == nil {
			list = append(list, t)
		}
	}
	return list, nil
}

func (m *mockTaskRepo) GetByID(ctx context.Context, id, wsID uuid.UUID) (*domainTask.Task, error) {
	t, ok := m.tasks[id]
	if !ok || t.WorkspaceID != wsID {
		return nil, domainTask.ErrTaskNotFound
	}
	return t, nil
}

func (m *mockTaskRepo) Toggle(ctx context.Context, id, wsID uuid.UUID) (*domainTask.Task, error) {
	t, err := m.GetByID(ctx, id, wsID)
	if err != nil {
		return nil, err
	}
	t.Toggle()
	return t, nil
}

func (m *mockTaskRepo) Delete(ctx context.Context, id, wsID uuid.UUID) error {
	t, err := m.GetByID(ctx, id, wsID)
	if err != nil {
		return err
	}
	delete(m.tasks, t.ID)
	return nil
}

func (m *mockTaskRepo) GetMaxInboxPosition(ctx context.Context, wsID uuid.UUID) (int, error) {
	maxPos := -1
	for _, t := range m.tasks {
		if t.WorkspaceID == wsID && t.ProjectID == nil && t.ColumnID == nil {
			if t.Position > maxPos {
				maxPos = t.Position
			}
		}
	}
	return maxPos, nil
}

func TestService_CreateInboxTask(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewService(repo, nil)
	wsID := uuid.New()

	task1, err := svc.CreateInboxTask(context.Background(), CreateInboxTaskInput{
		WorkspaceID: wsID,
		Title:       "Task 1",
		Priority:    domainTask.PriorityHigh,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task1.Position != 0 {
		t.Errorf("expected position 0, got %d", task1.Position)
	}

	task2, err := svc.CreateInboxTask(context.Background(), CreateInboxTaskInput{
		WorkspaceID: wsID,
		Title:       "Task 2",
		Priority:    domainTask.PriorityLow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task2.Position != 1 {
		t.Errorf("expected position 1, got %d", task2.Position)
	}
}

func TestService_ToggleTask(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewService(repo, nil)
	wsID := uuid.New()

	task, _ := svc.CreateInboxTask(context.Background(), CreateInboxTaskInput{
		WorkspaceID: wsID,
		Title:       "Task to toggle",
		Priority:    domainTask.PriorityNone,
	})

	toggled, err := svc.ToggleTask(context.Background(), task.ID, wsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !toggled.IsCompleted {
		t.Errorf("expected task to be completed")
	}

	// Toggle for wrong workspace (tenant isolation check)
	otherWsID := uuid.New()
	_, err = svc.ToggleTask(context.Background(), task.ID, otherWsID)
	if !errors.Is(err, domainTask.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound when accessing from another workspace, got %v", err)
	}
}

func TestService_DeleteTask(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewService(repo, nil)
	wsID := uuid.New()

	task, _ := svc.CreateInboxTask(context.Background(), CreateInboxTaskInput{
		WorkspaceID: wsID,
		Title:       "Task to delete",
		Priority:    domainTask.PriorityNone,
	})

	err := svc.DeleteTask(context.Background(), task.ID, wsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.GetByID(context.Background(), task.ID, wsID)
	if !errors.Is(err, domainTask.ErrTaskNotFound) {
		t.Errorf("expected task to be deleted, got %v", err)
	}
}

func TestService_CreateInitialInboxTasks(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewService(repo, nil)
	wsID := uuid.New()

	err := svc.CreateInitialInboxTasks(context.Background(), wsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inboxTasks, err := svc.GetInboxTasks(context.Background(), wsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inboxTasks) != 3 {
		t.Fatalf("expected exactly 3 onboarding tasks, got %d", len(inboxTasks))
	}
}
