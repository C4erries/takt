package task

import (
	"time"

	"github.com/google/uuid"
)

// Supported activity types for task history log.
const (
	ActivityColumnMoved    = "column_moved"
	ActivityStatusChanged  = "status_changed"
	ActivityPriorityChanged = "priority_changed"
	ActivityDueDateSet     = "due_date_set"
	ActivityDueDateCleared = "due_date_cleared"
	ActivityTagAdded       = "tag_added"
	ActivityTagRemoved     = "tag_removed"
	ActivityTitleChanged   = "title_changed"
	ActivitySubtaskAdded   = "subtask_added"
	ActivityTaskArchived   = "task_archived"
	ActivityTaskUnarchived = "task_unarchived"
)

var validActivityTypes = map[string]bool{
	ActivityColumnMoved:     true,
	ActivityStatusChanged:   true,
	ActivityPriorityChanged: true,
	ActivityDueDateSet:      true,
	ActivityDueDateCleared:  true,
	ActivityTagAdded:        true,
	ActivityTagRemoved:      true,
	ActivityTitleChanged:    true,
	ActivitySubtaskAdded:    true,
	ActivityTaskArchived:    true,
	ActivityTaskUnarchived:  true,
}

// IsValidActivityType verifies if the given activity type string is recognized.
func IsValidActivityType(t string) bool {
	return validActivityTypes[t]
}

// TaskActivity represents an immutable audit log entry for changes to a task.
type TaskActivity struct {
	ID           uuid.UUID
	TaskID       uuid.UUID
	UserID       *uuid.UUID
	UserName     *string
	ActivityType string
	Payload      map[string]interface{}
	CreatedAt    time.Time
}

// NewTaskActivity creates a valid TaskActivity domain entity enforcing business invariants.
func NewTaskActivity(
	taskID uuid.UUID,
	userID *uuid.UUID,
	activityType string,
	payload map[string]interface{},
) (*TaskActivity, error) {
	if taskID == uuid.Nil {
		return nil, ErrInvalidTaskID
	}
	if !IsValidActivityType(activityType) {
		return nil, ErrInvalidActivityType
	}
	if payload == nil {
		payload = make(map[string]interface{})
	}

	return &TaskActivity{
		ID:           uuid.New(),
		TaskID:       taskID,
		UserID:       userID,
		ActivityType: activityType,
		Payload:      payload,
		CreatedAt:    time.Now().UTC(),
	}, nil
}
