package task

import (
	"context"

	"github.com/google/uuid"
	domainAttachment "takt/internal/domain/attachment"
	"takt/internal/domain/events"
	domainTask "takt/internal/domain/task"
)

// Transactor defines transaction execution port.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

// TaskRepository alias to domain repository.
type TaskRepository = domainTask.TaskRepository

// TagRepository defines storage operations for context tags and task-tag associations.
type TagRepository interface {
	CreateTag(ctx context.Context, tag *domainTask.Tag) error
	GetTagByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainTask.Tag, error)
	ListWorkspaceTags(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Tag, error)
	ListWorkspaceTagsWithMetrics(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.TagWithMetrics, error)
	UpdateTag(ctx context.Context, tag *domainTask.Tag) error
	DeleteTag(ctx context.Context, id, workspaceID uuid.UUID) error
	DeleteUnusedTags(ctx context.Context, workspaceID uuid.UUID) error
	AddTagToTask(ctx context.Context, taskID, tagID uuid.UUID) error
	RemoveTagFromTask(ctx context.Context, taskID, tagID uuid.UUID) error
	GetTaskTags(ctx context.Context, taskID uuid.UUID) ([]*domainTask.Tag, error)
	ClearTaskTags(ctx context.Context, taskID uuid.UUID) error
}

// AttachmentRepository defines attachment operations needed for task duplication.
type AttachmentRepository interface {
	ListByTaskID(ctx context.Context, taskID, workspaceID uuid.UUID) ([]*domainAttachment.Attachment, error)
	Create(ctx context.Context, att *domainAttachment.Attachment) (*domainAttachment.Attachment, error)
}

// WorkspaceAccessChecker ensures the user has access to the workspace.
type WorkspaceAccessChecker interface {
	CheckAccess(ctx context.Context, workspaceID, userID uuid.UUID) error
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, event events.Event) error
}

// UseCases defines task business use cases.
type UseCases interface {
	GetInboxTasks(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Task, error)
	CreateInboxTask(ctx context.Context, input CreateInboxTaskInput) (*domainTask.Task, error)
	ToggleTask(ctx context.Context, id, workspaceID uuid.UUID) (*domainTask.Task, error)
	DeleteTask(ctx context.Context, id, workspaceID uuid.UUID) error
	CreateInitialInboxTasks(ctx context.Context, workspaceID uuid.UUID) error
	GetTodayTasks(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Task, error)
	UpdateTask(ctx context.Context, input UpdateTaskInput) (*domainTask.Task, error)
	SetTaskFocus(ctx context.Context, input SetFocusInput) (*domainTask.Task, error)
	ListSubtasks(ctx context.Context, workspaceID, parentID uuid.UUID) ([]*domainTask.Task, error)
	DuplicateTask(ctx context.Context, workspaceID, taskID uuid.UUID) (*domainTask.Task, error)
	CreateTag(ctx context.Context, input CreateTagInput) (*domainTask.Tag, error)
	ListWorkspaceTags(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Tag, error)
	ListWorkspaceTagsWithMetrics(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.TagWithMetrics, error)
	UpdateTag(ctx context.Context, input UpdateTagInput) (*domainTask.Tag, error)
	DeleteTag(ctx context.Context, tagID, workspaceID uuid.UUID) error
	DeleteUnusedTags(ctx context.Context, workspaceID uuid.UUID) error
	AddTagToTask(ctx context.Context, taskID, tagID, workspaceID uuid.UUID) error
	RemoveTagFromTask(ctx context.Context, taskID, tagID, workspaceID uuid.UUID) error
	ListArchivedTasks(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID) ([]*domainTask.Task, error)
	ArchiveCompletedTasksInColumn(ctx context.Context, workspaceID, columnID uuid.UUID) (int64, error)
	UnarchiveTask(ctx context.Context, workspaceID, taskID, columnID uuid.UUID, boardID *uuid.UUID) (*domainTask.Task, error)
}
