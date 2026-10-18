package task

import (
	"testing"
	"time"

	"github.com/google/uuid"
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
