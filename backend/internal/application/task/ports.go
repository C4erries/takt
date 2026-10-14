package task

import (
	"context"

	"github.com/google/uuid"
	"takt/internal/domain/events"
	domainTask "takt/internal/domain/task"
)

// TaskRepository alias to domain repository.
type TaskRepository = domainTask.TaskRepository

// TagRepository defines storage operations for context tags and task-tag associations.
type TagRepository interface {
	CreateTag(ctx context.Context, tag *domainTask.Tag) error
	GetTagByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainTask.Tag, error)
	ListWorkspaceTags(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Tag, error)
	DeleteTag(ctx context.Context, id, workspaceID uuid.UUID) error
	AddTagToTask(ctx context.Context, taskID, tagID uuid.UUID) error
	RemoveTagFromTask(ctx context.Context, taskID, tagID uuid.UUID) error
	GetTaskTags(ctx context.Context, taskID uuid.UUID) ([]*domainTask.Tag, error)
	ClearTaskTags(ctx context.Context, taskID uuid.UUID) error
}

// WorkspaceAccessChecker ensures the user has access to the workspace.
type WorkspaceAccessChecker interface {
	CheckAccess(ctx context.Context, workspaceID, userID uuid.UUID) error
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, event events.Event) error
}

