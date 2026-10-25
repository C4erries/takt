package automation

import (
	"context"
	"time"

	"github.com/google/uuid"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/domain/events"
)

// TaskInfo provides task state for automation evaluation and timer processing.
type TaskInfo struct {
	ID                uuid.UUID
	WorkspaceID       uuid.UUID
	ParentID          *uuid.UUID
	ProjectID         *uuid.UUID
	ColumnID          *uuid.UUID
	PreviousColumnID  *uuid.UUID
	PreviousProjectID *uuid.UUID
	Title             string
	Priority          string
	IsCompleted       bool
	DueDate           *time.Time
	Tags              []string
	TimeInColumnHours float64
	SubtasksCount     int
	SubtasksDoneCount int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// AutomationRepository handles persistence of automation rules.
type AutomationRepository interface {
	ListWorkspaceAutomations(ctx context.Context, workspaceID uuid.UUID) ([]*domainAutomation.Automation, error)
	GetAutomationByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainAutomation.Automation, error)
	ToggleAutomation(ctx context.Context, id, workspaceID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error)
	CreateAutomation(ctx context.Context, a *domainAutomation.Automation) error
	UpdateAutomation(ctx context.Context, a *domainAutomation.Automation) error
	DeleteAutomation(ctx context.Context, id, workspaceID uuid.UUID) error
	ListEnabledAutomationsByTrigger(ctx context.Context, workspaceID uuid.UUID, triggerType string) ([]*domainAutomation.Automation, error)
	ListAutomationsByColumn(ctx context.Context, workspaceID, columnID uuid.UUID) ([]*domainAutomation.Automation, error)
	ListAutomationsByTask(ctx context.Context, taskID, workspaceID uuid.UUID) ([]*domainAutomation.Automation, error)
	CreateColumnAutomation(ctx context.Context, a *domainAutomation.Automation) error
	ListAllEnabledTimerAutomations(ctx context.Context) ([]*domainAutomation.Automation, error)
}

// TaskModifier modifies task properties in response to automation triggers.
type TaskModifier interface {
	MoveTaskToColumn(ctx context.Context, taskID, workspaceID uuid.UUID, columnID, projectID *uuid.UUID, position float64) error
	SetTaskFocus(ctx context.Context, taskID, workspaceID uuid.UUID, focusOrder *int) error
	SetTaskPriority(ctx context.Context, taskID, workspaceID uuid.UUID, priority string) error
	AddTagToTask(ctx context.Context, taskID, workspaceID uuid.UUID, tagID *uuid.UUID, tagName string) error
	RemoveTagFromTask(ctx context.Context, taskID, workspaceID uuid.UUID, tagID *uuid.UUID, tagName string) error
	SetTaskDueDate(ctx context.Context, taskID, workspaceID uuid.UUID, dueDate *time.Time) error
	SetTaskCompleted(ctx context.Context, taskID, workspaceID uuid.UUID, isCompleted bool) error
	ArchiveTask(ctx context.Context, taskID, workspaceID uuid.UUID) error
}

// TaskActivityInfo represents an audit log entry for task movements and changes.
type TaskActivityInfo struct {
	ActivityType string
	Payload      map[string]interface{}
	CreatedAt    time.Time
}

// TaskReader reads tasks for timer and schedule evaluation.
type TaskReader interface {
	ListTasksByColumn(ctx context.Context, workspaceID, columnID uuid.UUID) ([]*TaskInfo, error)
	GetTaskByID(ctx context.Context, taskID, workspaceID uuid.UUID) (*TaskInfo, error)
	ListSubtasks(ctx context.Context, parentID uuid.UUID) ([]*TaskInfo, error)
	ListTaskActivities(ctx context.Context, taskID uuid.UUID) ([]*TaskActivityInfo, error)
}

// ProjectReader inspects board and column structure to determine target columns.
type ProjectReader interface {
	GetFirstTodoColumn(ctx context.Context, projectID, workspaceID uuid.UUID) (*uuid.UUID, error)
	GetDoneColumn(ctx context.Context, projectID, workspaceID uuid.UUID) (*uuid.UUID, error)
	GetInboxColumn(ctx context.Context, projectID, workspaceID uuid.UUID) (*uuid.UUID, error)
	ColumnExists(ctx context.Context, columnID uuid.UUID) (bool, error)
	GetColumnName(ctx context.Context, columnID uuid.UUID) (string, error)
	ListColumnIDsByBoard(ctx context.Context, boardID uuid.UUID) ([]uuid.UUID, error)
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, event events.Event) error
}

// BatchTaskExecutor executes batch actions on groups of tasks.
type BatchTaskExecutor interface {
	ArchiveCompletedTasksBatchSQL(
		ctx context.Context,
		workspaceID, columnID uuid.UUID,
		olderThanHours float64,
	) (int64, error)
}
