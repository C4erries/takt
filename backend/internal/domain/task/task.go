package task

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Priority string

const (
	PriorityNone   Priority = "none"
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

func (p Priority) IsValid() bool {
	switch p {
	case PriorityNone, PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

type Task struct {
	ID                  uuid.UUID
	WorkspaceID         uuid.UUID
	ProjectID           *uuid.UUID
	ColumnID            *uuid.UUID
	Title               string
	DescriptionMarkdown string
	IsCompleted         bool
	CompletedAt         *time.Time
	DueDate             *time.Time
	Priority            Priority
	FocusOrder          *int
	Position            int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func NewInboxTask(
	workspaceID uuid.UUID,
	title string,
	descriptionMarkdown string,
	priority Priority,
	dueDate *time.Time,
	position int,
) (*Task, error) {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return nil, ErrInvalidTaskTitle
	}

	if priority == "" {
		priority = PriorityNone
	}
	if !priority.IsValid() {
		return nil, ErrInvalidPriority
	}

	now := time.Now().UTC()
	return &Task{
		ID:                  uuid.New(),
		WorkspaceID:         workspaceID,
		ProjectID:           nil,
		ColumnID:            nil,
		Title:               trimmedTitle,
		DescriptionMarkdown: descriptionMarkdown,
		IsCompleted:         false,
		CompletedAt:         nil,
		DueDate:             dueDate,
		Priority:            priority,
		FocusOrder:          nil,
		Position:            position,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

func (t *Task) Toggle() {
	t.IsCompleted = !t.IsCompleted
	now := time.Now().UTC()
	t.UpdatedAt = now
	if t.IsCompleted {
		t.CompletedAt = &now
	} else {
		t.CompletedAt = nil
	}
}
