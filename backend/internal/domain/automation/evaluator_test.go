package automation_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"takt/internal/domain/automation"
)

func TestRuleEvaluator_Evaluate(t *testing.T) {
	evaluator := automation.NewRuleEvaluator()
	projID := uuid.New()
	colID := uuid.New()
	now := time.Now().UTC()

	testCases := []struct {
		name           string
		conditionsJSON json.RawMessage
		ctx            automation.TaskContext
		expectedMatch  bool
		expectError    bool
	}{
		{
			name:           "empty conditions always match (Done & Archive)",
			conditionsJSON: json.RawMessage(`{}`),
			ctx: automation.TaskContext{
				IsCompleted: true,
			},
			expectedMatch: true,
		},
		{
			name:           "null conditions match",
			conditionsJSON: json.RawMessage(`null`),
			ctx: automation.TaskContext{
				IsCompleted: false,
			},
			expectedMatch: true,
		},
		{
			name:           "empty byte slice matches",
			conditionsJSON: nil,
			ctx: automation.TaskContext{
				IsCompleted: true,
			},
			expectedMatch: true,
		},
		{
			name:           "Auto-Focus preset: has_due_date = true matches task with due date",
			conditionsJSON: json.RawMessage(`{"has_due_date": true}`),
			ctx: automation.TaskContext{
				HasDueDate: true,
				DueDate:    &now,
			},
			expectedMatch: true,
		},
		{
			name:           "Auto-Focus preset: has_due_date = true rejects task without due date",
			conditionsJSON: json.RawMessage(`{"has_due_date": true}`),
			ctx: automation.TaskContext{
				HasDueDate: false,
				DueDate:    nil,
			},
			expectedMatch: false,
		},
		{
			name:           "Inbox Zero preset: project_assigned = true matches task with project",
			conditionsJSON: json.RawMessage(`{"project_assigned": true}`),
			ctx: automation.TaskContext{
				ProjectID: &projID,
			},
			expectedMatch: true,
		},
		{
			name:           "Inbox Zero preset: project_assigned = true rejects task without project",
			conditionsJSON: json.RawMessage(`{"project_assigned": true}`),
			ctx: automation.TaskContext{
				ProjectID: nil,
			},
			expectedMatch: false,
		},
		{
			name:           "Inbox Zero combo: project_assigned = true AND inbox_only = true",
			conditionsJSON: json.RawMessage(`{"project_assigned": true, "inbox_only": true}`),
			ctx: automation.TaskContext{
				ProjectID: &projID,
				ColumnID:  nil, // in inbox
			},
			expectedMatch: true,
		},
		{
			name:           "Inbox Zero combo: project_assigned = true AND inbox_only = true rejects task already in column",
			conditionsJSON: json.RawMessage(`{"project_assigned": true, "inbox_only": true}`),
			ctx: automation.TaskContext{
				ProjectID: &projID,
				ColumnID:  &colID, // already in a column
			},
			expectedMatch: false,
		},
		{
			name:           "is_completed = true condition filter",
			conditionsJSON: json.RawMessage(`{"is_completed": true}`),
			ctx: automation.TaskContext{
				IsCompleted: true,
			},
			expectedMatch: true,
		},
		{
			name:           "is_completed = true condition filter rejects incomplete task",
			conditionsJSON: json.RawMessage(`{"is_completed": true}`),
			ctx: automation.TaskContext{
				IsCompleted: false,
			},
			expectedMatch: false,
		},
		{
			name:           "malformed JSON returns error",
			conditionsJSON: json.RawMessage(`{"has_due_date": invalid}`),
			ctx:            automation.TaskContext{},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			match, err := evaluator.Evaluate(tc.conditionsJSON, tc.ctx)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedMatch, match)
			}
		})
	}
}

