package task

import (
	"time"

	"github.com/google/uuid"
	domainTask "takt/internal/domain/task"
)

type CreateInboxTaskInput struct {
	WorkspaceID         uuid.UUID
	Title               string
	DescriptionMarkdown string
	Priority            domainTask.Priority
	DueDate             *time.Time
	ParentID            *uuid.UUID
}

// CreateTaskInput alias for consistency across general task creation endpoints.
type CreateTaskInput = CreateInboxTaskInput

type UpdateTaskInput struct {
	TaskID              uuid.UUID
	WorkspaceID         uuid.UUID
	Title               *string
	DescriptionMarkdown *string
	DueDate             *time.Time
	ClearDueDate        bool
	Priority            *domainTask.Priority
	FocusOrder          *int
	ClearFocusOrder     bool
	ProjectID           *uuid.UUID
	ClearProjectID      bool
	ColumnID            *uuid.UUID
	ClearColumnID       bool
	ParentID            *uuid.UUID
	ClearParentID       bool
}

type SetFocusInput struct {
	TaskID      uuid.UUID
	WorkspaceID uuid.UUID
	FocusOrder  *int
}

type CreateTagInput struct {
	WorkspaceID uuid.UUID
	ProjectID   *uuid.UUID
	Name        string
	ColorHex    string
}

type UpdateTagInput struct {
	TagID       uuid.UUID
	WorkspaceID uuid.UUID
	Name        *string
	ColorHex    *string
}
