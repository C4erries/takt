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
	ListSubtasks(ctx context.Context, parentID uuid.UUID) ([]*Task, error)
	ListArchivedTasks(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID) ([]*Task, error)
	CountArchivedTasks(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	ArchiveCompletedTasksInColumn(ctx context.Context, columnID uuid.UUID) (int64, error)
	UnarchiveTask(ctx context.Context, taskID, workspaceID, columnID uuid.UUID, projectID *uuid.UUID, position float64) (*Task, error)
	ArchiveCompletedTasksBatchSQL(ctx context.Context, workspaceID, columnID uuid.UUID, olderThanHours float64) (int64, error)
	CreateActivity(ctx context.Context, activity *TaskActivity) error
	ListActivitiesByTaskID(ctx context.Context, taskID uuid.UUID) ([]*TaskActivity, error)
}
