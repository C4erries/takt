package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"takt/internal/domain/events"
)

func TestTaskEvents_Properties(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	projID := uuid.New()
	colID := uuid.New()
	prevColID := uuid.New()
	now := time.Now().UTC()

	t.Run("TaskCreatedEvent", func(t *testing.T) {
		e := events.TaskCreatedEvent{
			TaskID:         taskID,
			WorkspaceIDVal: wsID,
			ProjectID:      &projID,
			ColumnID:       &colID,
			Title:          "New Task",
			DueDate:        &now,
			OccurredAtVal:  now,
			FromAutomation: false,
		}

		assert.Equal(t, events.EventTaskCreated, e.EventName())
		assert.Equal(t, now, e.OccurredAt())
		assert.Equal(t, wsID, e.WorkspaceID())
		assert.False(t, e.IsFromAutomation())
	})

	t.Run("TaskUpdatedEvent", func(t *testing.T) {
		e := events.TaskUpdatedEvent{
			TaskID:            taskID,
			WorkspaceIDVal:    wsID,
			ProjectID:         &projID,
			ColumnID:          &colID,
			Title:             "Updated",
			IsCompleted:       false,
			PreviousProjectID: nil,
			PreviousColumnID:  &prevColID,
			OccurredAtVal:     now,
			FromAutomation:    true,
		}

		assert.Equal(t, events.EventTaskUpdated, e.EventName())
		assert.Equal(t, now, e.OccurredAt())
		assert.Equal(t, wsID, e.WorkspaceID())
		assert.True(t, e.IsFromAutomation())
	})

	t.Run("TaskCompletedEvent", func(t *testing.T) {
		e := events.TaskCompletedEvent{
			TaskID:         taskID,
			WorkspaceIDVal: wsID,
			CompletedAt:    now,
			OccurredAtVal:  now,
			FromAutomation: false,
		}

		assert.Equal(t, events.EventTaskCompleted, e.EventName())
		assert.Equal(t, now, e.OccurredAt())
		assert.Equal(t, wsID, e.WorkspaceID())
		assert.False(t, e.IsFromAutomation())
	})

	t.Run("TaskMovedEvent", func(t *testing.T) {
		e := events.TaskMovedEvent{
			TaskID:           taskID,
			WorkspaceIDVal:   wsID,
			ProjectID:        &projID,
			ColumnID:         &colID,
			PreviousColumnID: &prevColID,
			Position:         1500.0,
			OccurredAtVal:    now,
			FromAutomation:   true,
		}

		assert.Equal(t, events.EventTaskMoved, e.EventName())
		assert.Equal(t, now, e.OccurredAt())
		assert.Equal(t, wsID, e.WorkspaceID())
		assert.True(t, e.IsFromAutomation())
	})
}

func TestAutomationTriggerContext(t *testing.T) {
	testCases := []struct {
		name     string
		ctx      context.Context
		expected bool
	}{
		{
			name:     "nil context returns false",
			ctx:      nil,
			expected: false,
		},
		{
			name:     "clean background context returns false",
			ctx:      context.Background(),
			expected: false,
		},
		{
			name:     "context marked with automation trigger returns true",
			ctx:      events.WithAutomationTrigger(context.Background()),
			expected: true,
		},
		{
			name:     "marking nil context creates valid context with true",
			ctx:      events.WithAutomationTrigger(nil),
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual := events.IsAutomationTriggered(tc.ctx)
			require.Equal(t, tc.expected, actual)
		})
	}
}

