package task

import (
	"context"

	"github.com/google/uuid"
)

type TaskRepository interface {
	Create(ctx context.Context, t *Task) error
	GetInboxTasks(ctx context.Context, workspaceID uuid.UUID) ([]*Task, error)
	GetTodayTasks(ctx context.Context, workspaceID uuid.UUID) ([]*Task, error)
	GetByID(ctx context.Context, id, workspaceID uuid.UUID) (*Task, error)
	Update(ctx context.Context, t *Task) error
	SetFocusOrder(ctx context.Context, id, workspaceID uuid.UUID, focusOrder *int) (*Task, error)
	Toggle(ctx context.Context, id, workspaceID uuid.UUID) (*Task, error)
	Delete(ctx context.Context, id, workspaceID uuid.UUID) error
	GetMaxInboxPosition(ctx context.Context, workspaceID uuid.UUID) (int, error)
}
