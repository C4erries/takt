package task

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInboxTask_Valid(t *testing.T) {
	wsID := uuid.New()
	task, err := NewInboxTask(wsID, "Buy Matcha tea", "desc", PriorityHigh, nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task.Title != "Buy Matcha tea" {
		t.Errorf("expected title 'Buy Matcha tea', got %s", task.Title)
	}
	if task.Priority != PriorityHigh {
		t.Errorf("expected priority High, got %s", task.Priority)
	}
	if task.IsCompleted {
		t.Errorf("expected isCompleted false")
	}
	if task.CompletedAt != nil {
		t.Errorf("expected completedAt nil")
	}
	if task.WorkspaceID != wsID {
		t.Errorf("expected workspaceID %v, got %v", wsID, task.WorkspaceID)
	}
}

func TestNewInboxTask_EmptyTitle(t *testing.T) {
	wsID := uuid.New()
	_, err := NewInboxTask(wsID, "   ", "", PriorityNone, nil, 0)
	if err != ErrInvalidTaskTitle {
		t.Errorf("expected ErrInvalidTaskTitle, got %v", err)
	}
}

func TestNewInboxTask_InvalidPriority(t *testing.T) {
	wsID := uuid.New()
	_, err := NewInboxTask(wsID, "Valid title", "", Priority("invalid"), nil, 0)
	if err != ErrInvalidPriority {
		t.Errorf("expected ErrInvalidPriority, got %v", err)
	}
}

func TestTask_Toggle(t *testing.T) {
	wsID := uuid.New()
	task, err := NewInboxTask(wsID, "Task to complete", "", PriorityLow, nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First toggle: complete
	task.Toggle()
	if !task.IsCompleted {
		t.Errorf("expected task to be completed")
	}
	if task.CompletedAt == nil {
		t.Errorf("expected completedAt to be set")
	}

	// Second toggle: uncomplete
	time.Sleep(2 * time.Millisecond)
	task.Toggle()
	if task.IsCompleted {
		t.Errorf("expected task to be uncompleted")
	}
	if task.CompletedAt != nil {
		t.Errorf("expected completedAt to be cleared")
	}
}

func TestNewInboxTask_PendingPriority(t *testing.T) {
	wsID := uuid.New()
	task, err := NewInboxTask(wsID, "Pending task", "", PriorityPending, nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Priority != PriorityPending {
		t.Errorf("expected priority %s, got %s", PriorityPending, task.Priority)
	}
}

func TestValidatePriority(t *testing.T) {
	tests := []struct {
		priority Priority
		expected bool
	}{
		{PriorityNone, true},
		{PriorityLow, true},
		{PriorityMedium, true},
		{PriorityHigh, true},
		{PriorityPending, true},
		{PriorityUrgent, true},
		{Priority(""), false},
		{Priority("invalid"), false},
		{Priority("PENDING"), false},
		{Priority("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			if got := ValidatePriority(tt.priority); got != tt.expected {
				t.Errorf("ValidatePriority(%q) = %v; want %v", tt.priority, got, tt.expected)
			}
			if got := tt.priority.IsValid(); got != tt.expected {
				t.Errorf("Priority(%q).IsValid() = %v; want %v", tt.priority, got, tt.expected)
			}
		})
	}
}

func TestTask_SetParent(t *testing.T) {
	wsID := uuid.New()
	task, err := NewInboxTask(wsID, "My Task", "", PriorityLow, nil, 0)
	require.NoError(t, err)

	parentID := uuid.New()
	anotherParentID := uuid.New()

	tests := []struct {
		name        string
		parentID    *uuid.UUID
		expectedErr error
		checkParent *uuid.UUID
	}{
		{
			name:        "set valid parent",
			parentID:    &parentID,
			expectedErr: nil,
			checkParent: &parentID,
		},
		{
			name:        "change parent to another",
			parentID:    &anotherParentID,
			expectedErr: nil,
			checkParent: &anotherParentID,
		},
		{
			name:        "clear parent",
			parentID:    nil,
			expectedErr: nil,
			checkParent: nil,
		},
		{
			name:        "self as parent returns error",
			parentID:    &task.ID,
			expectedErr: ErrSelfParenting,
			checkParent: nil, // unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := task.SetParent(tt.parentID)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.checkParent, task.ParentID)
		})
	}
}

func TestTask_SubtasksCount(t *testing.T) {
	wsID := uuid.New()
	task, err := NewInboxTask(wsID, "Parent Task", "", PriorityLow, nil, 0)
	require.NoError(t, err)

	assert.Equal(t, 0, task.SubtasksCount)
	assert.Equal(t, 0, task.SubtasksCompletedCount)
	assert.Nil(t, task.ParentID)

	tests := []struct {
		name              string
		total             int
		completed         int
		expectedTotal     int
		expectedCompleted int
	}{
		{
			name:              "valid positive counts",
			total:             5,
			completed:         2,
			expectedTotal:     5,
			expectedCompleted: 2,
		},
		{
			name:              "all completed",
			total:             3,
			completed:         3,
			expectedTotal:     3,
			expectedCompleted: 3,
		},
		{
			name:              "negative counts normalized to zero",
			total:             -1,
			completed:         -5,
			expectedTotal:     0,
			expectedCompleted: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task.SetSubtaskCounts(tt.total, tt.completed)
			assert.Equal(t, tt.expectedTotal, task.SubtasksCount)
			assert.Equal(t, tt.expectedCompleted, task.SubtasksCompletedCount)
		})
	}
}

