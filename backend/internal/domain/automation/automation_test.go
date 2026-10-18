package automation_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"takt/internal/domain/automation"
)

func TestNewAutomation(t *testing.T) {
	wsID := uuid.New()
	boardID := uuid.New()

	testCases := []struct {
		name           string
		boardID        *uuid.UUID
		ruleName       string
		triggerType    string
		conditionsJSON json.RawMessage
		actionsJSON    json.RawMessage
		isEnabled      bool
		wantErr        error
		checkOut       func(t *testing.T, a *automation.Automation)
	}{
		{
			name:           "valid global automation",
			boardID:        nil,
			ruleName:       "Done & Archive",
			triggerType:    automation.TriggerTaskCompleted,
			conditionsJSON: json.RawMessage(`{}`),
			actionsJSON:    json.RawMessage(`{"action":"move_to_done"}`),
			isEnabled:      true,
			wantErr:        nil,
			checkOut: func(t *testing.T, a *automation.Automation) {
				assert.Equal(t, wsID, a.WorkspaceID)
				assert.Nil(t, a.BoardID)
				assert.Equal(t, "Done & Archive", a.Name)
				assert.Equal(t, automation.TriggerTaskCompleted, a.TriggerType)
				assert.True(t, a.IsEnabled)
			},
		},
		{
			name:           "valid board-scoped automation with empty JSON defaulting",
			boardID:        &boardID,
			ruleName:       "Auto-Focus",
			triggerType:    automation.TriggerTaskUpdated,
			conditionsJSON: nil,
			actionsJSON:    nil,
			isEnabled:      false,
			wantErr:        nil,
			checkOut: func(t *testing.T, a *automation.Automation) {
				assert.Equal(t, &boardID, a.BoardID)
				assert.Equal(t, json.RawMessage("{}"), a.ConditionsJSON)
				assert.Equal(t, json.RawMessage("{}"), a.ActionsJSON)
				assert.False(t, a.IsEnabled)
			},
		},
		{
			name:           "valid trigger task entered column",
			boardID:        &boardID,
			ruleName:       "Set Urgent on Enter",
			triggerType:    automation.TriggerTaskEnteredColumn,
			conditionsJSON: json.RawMessage(`{}`),
			actionsJSON:    json.RawMessage(`{"action":"set_priority","priority":"urgent"}`),
			isEnabled:      true,
			wantErr:        nil,
			checkOut: func(t *testing.T, a *automation.Automation) {
				assert.Equal(t, automation.TriggerTaskEnteredColumn, a.TriggerType)
				assert.True(t, a.IsEnabled)
			},
		},
		{
			name:           "valid trigger task completed in column",
			boardID:        &boardID,
			ruleName:       "Move to Done on Complete",
			triggerType:    automation.TriggerTaskCompletedInColumn,
			conditionsJSON: json.RawMessage(`{}`),
			actionsJSON:    json.RawMessage(`{"action":"move_to_column"}`),
			isEnabled:      true,
			wantErr:        nil,
			checkOut: func(t *testing.T, a *automation.Automation) {
				assert.Equal(t, automation.TriggerTaskCompletedInColumn, a.TriggerType)
				assert.True(t, a.IsEnabled)
			},
		},
		{
			name:        "error: empty rule name",
			ruleName:    "   ",
			triggerType: automation.TriggerTaskCreated,
			wantErr:     automation.ErrInvalidAutomationName,
		},
		{
			name:        "error: rule name exceeds max length",
			ruleName:    strings.Repeat("a", 256),
			triggerType: automation.TriggerTaskCreated,
			wantErr:     automation.ErrInvalidAutomationName,
		},
		{
			name:        "error: invalid trigger type",
			ruleName:    "Valid Name",
			triggerType: "unknown_trigger",
			wantErr:     automation.ErrInvalidTriggerType,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a, err := automation.NewAutomation(
				wsID,
				tc.boardID,
				tc.ruleName,
				tc.triggerType,
				tc.conditionsJSON,
				tc.actionsJSON,
				tc.isEnabled,
			)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, a)
			} else {
				require.NoError(t, err)
				require.NotNil(t, a)
				if tc.checkOut != nil {
					tc.checkOut(t, a)
				}
			}
		})
	}
}

func TestAutomation_SetEnabled(t *testing.T) {
	wsID := uuid.New()
	a, err := automation.NewAutomation(
		wsID,
		nil,
		"Test Rule",
		automation.TriggerTaskCompleted,
		nil,
		nil,
		true,
	)
	require.NoError(t, err)

	initialUpdated := a.UpdatedAt

	a.SetEnabled(false)
	assert.False(t, a.IsEnabled)
	assert.True(t, a.UpdatedAt.After(initialUpdated) || a.UpdatedAt.Equal(initialUpdated))
}

func TestNewColumnAutomation(t *testing.T) {
	wsID := uuid.New()
	boardID := uuid.New()
	columnID := uuid.New()

	t.Run("valid column automation", func(t *testing.T) {
		a, err := automation.NewColumnAutomation(
			wsID,
			&boardID,
			columnID,
			"Column rule",
			automation.TriggerTaskEnteredColumn,
			nil,
			nil,
			true,
		)
		require.NoError(t, err)
		require.NotNil(t, a)
		assert.Equal(t, &columnID, a.ColumnID)
		assert.Equal(t, &boardID, a.BoardID)
		assert.Equal(t, wsID, a.WorkspaceID)
		assert.Equal(t, automation.TriggerTaskEnteredColumn, a.TriggerType)
		assert.True(t, a.IsEnabled)
	})

	t.Run("error: nil column ID", func(t *testing.T) {
		a, err := automation.NewColumnAutomation(
			wsID,
			&boardID,
			uuid.Nil,
			"Column rule",
			automation.TriggerTaskEnteredColumn,
			nil,
			nil,
			true,
		)
		require.ErrorIs(t, err, automation.ErrInvalidColumnID)
		assert.Nil(t, a)
	})

	t.Run("error: invalid name", func(t *testing.T) {
		a, err := automation.NewColumnAutomation(
			wsID,
			&boardID,
			columnID,
			"   ",
			automation.TriggerTaskEnteredColumn,
			nil,
			nil,
			true,
		)
		require.ErrorIs(t, err, automation.ErrInvalidAutomationName)
		assert.Nil(t, a)
	})
}

func TestAutomation_SetColumnID(t *testing.T) {
	wsID := uuid.New()
	colID := uuid.New()
	a, err := automation.NewAutomation(
		wsID,
		nil,
		"Test Rule",
		automation.TriggerTaskCompleted,
		nil,
		nil,
		true,
	)
	require.NoError(t, err)
	assert.Nil(t, a.ColumnID)

	initialUpdated := a.UpdatedAt

	a.SetColumnID(&colID)
	assert.Equal(t, &colID, a.ColumnID)
	assert.True(t, a.UpdatedAt.After(initialUpdated) || a.UpdatedAt.Equal(initialUpdated))

	a.SetColumnID(nil)
	assert.Nil(t, a.ColumnID)
}


