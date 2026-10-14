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
}

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
}

type SetFocusInput struct {
	TaskID      uuid.UUID
	WorkspaceID uuid.UUID
	FocusOrder  *int
}

type CreateTagInput struct {
	WorkspaceID uuid.UUID
	Name        string
	ColorHex    string
}
