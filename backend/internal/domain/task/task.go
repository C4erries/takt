package task

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Priority string

const (
	PriorityNone    Priority = "none"
	PriorityLow     Priority = "low"
	PriorityMedium  Priority = "medium"
	PriorityHigh    Priority = "high"
	PriorityPending Priority = "pending"
	PriorityUrgent  Priority = "urgent"
)

func (p Priority) IsValid() bool {
	return ValidatePriority(p)
}

func ValidatePriority(p Priority) bool {
	switch p {
	case PriorityNone, PriorityLow, PriorityMedium, PriorityHigh, PriorityPending, PriorityUrgent:
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
	Tags                []*Tag
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
		Tags:                []*Tag{},
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

// AddTag attaches a tag to the task if not already present.
func (t *Task) AddTag(tag *Tag) error {
	if tag == nil {
		return ErrTagNotFound
	}
	for _, existing := range t.Tags {
		if existing.ID == tag.ID {
			return ErrTagAlreadyAttached
		}
	}
	t.Tags = append(t.Tags, tag)
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// RemoveTag detaches a tag from the task by its ID. Returns true if removed.
func (t *Task) RemoveTag(tagID uuid.UUID) bool {
	for i, existing := range t.Tags {
		if existing.ID == tagID {
			t.Tags = append(t.Tags[:i], t.Tags[i+1:]...)
			t.UpdatedAt = time.Now().UTC()
			return true
		}
	}
	return false
}

// SetFocusOrder updates or clears the daily focus order.
func (t *Task) SetFocusOrder(order *int) error {
	if order != nil && *order < 0 {
		return ErrInvalidFocusOrder
	}
	t.FocusOrder = order
	t.UpdatedAt = time.Now().UTC()
	return nil
}
