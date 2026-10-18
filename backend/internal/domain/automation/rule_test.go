package automation_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"takt/internal/domain/automation"
)

func TestRuleDefinition_Validate(t *testing.T) {
	t.Parallel()

	validTrigger := automation.Trigger{
		Type:   automation.TriggerTaskEnteredColumn,
		Params: map[string]interface{}{"column_id": uuid.New().String()},
	}

	validAction := automation.Action{
		Type:   automation.ActionMoveToColumn,
		Params: map[string]interface{}{"target_column_id": uuid.New().String()},
	}

	validCondition := automation.Condition{
		Field:    automation.ConditionFieldPriority,
		Operator: automation.OperatorEquals,
		Value:    "urgent",
	}

	tests := []struct {
		name    string
		rule    *automation.RuleDefinition
		wantErr error
	}{
		{
			name:    "nil rule definition",
			rule:    nil,
			wantErr: automation.ErrRuleDefinitionNil,
		},
		{
			name: "valid unconditional rule (0 conditions, 1 action)",
			rule: &automation.RuleDefinition{
				Trigger:    validTrigger,
				Conditions: []automation.Condition{},
				Actions:    []automation.Action{validAction},
			},
			wantErr: nil,
		},
		{
			name: "valid rule with all supported triggers",
			rule: &automation.RuleDefinition{
				Trigger: automation.Trigger{
					Type: automation.TriggerColumnSLATimer,
					Params: map[string]interface{}{
						"column_id":      uuid.New().String(),
						"duration_hours": 24,
					},
				},
				Conditions: []automation.Condition{validCondition},
				Actions:    []automation.Action{validAction},
			},
			wantErr: nil,
		},
		{
			name: "valid boundary rule: exactly 5 conditions and 5 actions",
			rule: &automation.RuleDefinition{
				Trigger: validTrigger,
				Conditions: []automation.Condition{
					{Field: automation.ConditionFieldPriority, Operator: automation.OperatorEquals, Value: "urgent"},
					{Field: automation.ConditionFieldIsCompleted, Operator: automation.OperatorEquals, Value: false},
					{Field: automation.ConditionFieldDueDate, Operator: automation.OperatorIsNotEmpty},
					{Field: automation.ConditionFieldTimeInColumnHours, Operator: automation.OperatorGreaterThan, Value: 48},
					{Field: automation.ConditionFieldTags, Operator: automation.OperatorContains, Value: "bug"},
				},
				Actions: []automation.Action{
					{Type: automation.ActionSetPriority, Params: map[string]interface{}{"priority": "urgent"}},
					{Type: automation.ActionAddTag, Params: map[string]interface{}{"name": "sla-breach"}},
					{Type: automation.ActionRemoveTag, Params: map[string]interface{}{"name": "triage"}},
					{Type: automation.ActionSetDueDate, Params: map[string]interface{}{"preset": "today"}},
					{Type: automation.ActionMoveToColumn, Params: map[string]interface{}{"target_column_id": uuid.New().String()}},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid trigger type: unknown string",
			rule: &automation.RuleDefinition{
				Trigger:    automation.Trigger{Type: "unknown_trigger"},
				Conditions: []automation.Condition{validCondition},
				Actions:    []automation.Action{validAction},
			},
			wantErr: automation.ErrInvalidTriggerType,
		},
		{
			name: "invalid trigger type: legacy trigger rejected in DSL",
			rule: &automation.RuleDefinition{
				Trigger:    automation.Trigger{Type: "task_completed"},
				Conditions: []automation.Condition{validCondition},
				Actions:    []automation.Action{validAction},
			},
			wantErr: automation.ErrInvalidTriggerType,
		},
		{
			name: "exceeds condition limit (> 5 conditions)",
			rule: &automation.RuleDefinition{
				Trigger: validTrigger,
				Conditions: []automation.Condition{
					validCondition, validCondition, validCondition,
					validCondition, validCondition, validCondition,
				},
				Actions: []automation.Action{validAction},
			},
			wantErr: automation.ErrTooManyConditions,
		},
		{
			name: "invalid condition field",
			rule: &automation.RuleDefinition{
				Trigger: validTrigger,
				Conditions: []automation.Condition{
					{Field: "invalid_field", Operator: automation.OperatorEquals, Value: "test"},
				},
				Actions: []automation.Action{validAction},
			},
			wantErr: automation.ErrInvalidConditionField,
		},
		{
			name: "invalid condition operator",
			rule: &automation.RuleDefinition{
				Trigger: validTrigger,
				Conditions: []automation.Condition{
					{Field: automation.ConditionFieldPriority, Operator: "invalid_op", Value: "test"},
				},
				Actions: []automation.Action{validAction},
			},
			wantErr: automation.ErrInvalidConditionOperator,
		},
		{
			name: "empty actions array (0 actions)",
			rule: &automation.RuleDefinition{
				Trigger:    validTrigger,
				Conditions: []automation.Condition{validCondition},
				Actions:    []automation.Action{},
			},
			wantErr: automation.ErrEmptyActions,
		},
		{
			name: "exceeds actions limit (> 5 actions)",
			rule: &automation.RuleDefinition{
				Trigger:    validTrigger,
				Conditions: []automation.Condition{validCondition},
				Actions: []automation.Action{
					validAction, validAction, validAction,
					validAction, validAction, validAction,
				},
			},
			wantErr: automation.ErrTooManyActions,
		},
		{
			name: "invalid action type",
			rule: &automation.RuleDefinition{
				Trigger:    validTrigger,
				Conditions: []automation.Condition{validCondition},
				Actions: []automation.Action{
					{Type: "invalid_action_type"},
				},
			},
			wantErr: automation.ErrInvalidActionType,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := automation.ValidateRuleDefinition(tt.rule)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRuleDefinition_JSONSerialization(t *testing.T) {
	t.Parallel()

	rule := &automation.RuleDefinition{
		Trigger: automation.Trigger{
			Type:   automation.TriggerTaskEnteredColumn,
			Params: map[string]interface{}{"column_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"},
		},
		Conditions: []automation.Condition{
			{
				Field:    automation.ConditionFieldPriority,
				Operator: automation.OperatorEquals,
				Value:    "urgent",
			},
		},
		Actions: []automation.Action{
			{
				Type:   automation.ActionMoveToColumn,
				Params: map[string]interface{}{"target_column_id": "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f"},
			},
		},
	}

	raw, err := rule.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, raw)

	parsed, err := automation.RuleDefinitionFromJSON(raw)
	require.NoError(t, err)
	require.NotNil(t, parsed)
	assert.Equal(t, rule.Trigger.Type, parsed.Trigger.Type)
	assert.Len(t, parsed.Conditions, 1)
	assert.Len(t, parsed.Actions, 1)
	assert.Equal(t, automation.ConditionFieldPriority, parsed.Conditions[0].Field)
	assert.Equal(t, automation.ActionMoveToColumn, parsed.Actions[0].Type)

	// Test empty or null inputs
	emptyParsed, err := automation.RuleDefinitionFromJSON([]byte("{}"))
	require.NoError(t, err)
	assert.Nil(t, emptyParsed)

	nullParsed, err := automation.RuleDefinitionFromJSON([]byte("null"))
	require.NoError(t, err)
	assert.Nil(t, nullParsed)

	// Test malformed JSON
	_, err = automation.RuleDefinitionFromJSON([]byte("{invalid-json"))
	require.Error(t, err)

	// Test valid JSON syntax with invalid domain rule
	invalidRuleJSON := []byte(`{"trigger":{"type":"unknown"},"actions":[{"type":"move_to_column"}]}`)
	_, err = automation.RuleDefinitionFromJSON(invalidRuleJSON)
	require.ErrorIs(t, err, automation.ErrInvalidTriggerType)
}

func TestNewRuleAutomation(t *testing.T) {
	t.Parallel()

	wsID := uuid.New()
	boardID := uuid.New()
	colID := uuid.New()

	validRule := &automation.RuleDefinition{
		Trigger: automation.Trigger{
			Type:   automation.TriggerTaskEnteredColumn,
			Params: map[string]interface{}{"column_id": colID.String()},
		},
		Conditions: []automation.Condition{
			{Field: automation.ConditionFieldPriority, Operator: automation.OperatorEquals, Value: "urgent"},
		},
		Actions: []automation.Action{
			{Type: automation.ActionSetPriority, Params: map[string]interface{}{"priority": "high"}},
		},
	}

	t.Run("successful creation", func(t *testing.T) {
		a, err := automation.NewRuleAutomation(wsID, &boardID, &colID, "Auto Urgent", validRule, true)
		require.NoError(t, err)
		require.NotNil(t, a)

		assert.NotEqual(t, uuid.Nil, a.ID)
		assert.Equal(t, wsID, a.WorkspaceID)
		assert.Equal(t, &boardID, a.BoardID)
		assert.Equal(t, &colID, a.ColumnID)
		assert.Equal(t, "Auto Urgent", a.Name)
		assert.Equal(t, automation.TriggerTaskEnteredColumn, a.TriggerType)
		assert.True(t, a.IsEnabled)
		assert.Equal(t, validRule, a.RuleDefinition)
		assert.False(t, a.CreatedAt.IsZero())
		assert.False(t, a.UpdatedAt.IsZero())
	})

	t.Run("nil rule definition", func(t *testing.T) {
		_, err := automation.NewRuleAutomation(wsID, &boardID, &colID, "Auto Urgent", nil, true)
		require.ErrorIs(t, err, automation.ErrRuleDefinitionNil)
	})

	t.Run("invalid rule definition inside", func(t *testing.T) {
		badRule := &automation.RuleDefinition{
			Trigger: automation.Trigger{Type: "invalid"},
			Actions: []automation.Action{{Type: automation.ActionArchiveTask}},
		}
		_, err := automation.NewRuleAutomation(wsID, &boardID, &colID, "Auto Urgent", badRule, true)
		require.ErrorIs(t, err, automation.ErrInvalidTriggerType)
	})

	t.Run("invalid rule name: empty", func(t *testing.T) {
		_, err := automation.NewRuleAutomation(wsID, &boardID, &colID, "   ", validRule, true)
		require.ErrorIs(t, err, automation.ErrInvalidAutomationName)
	})

	t.Run("invalid rule name: exceeds max length", func(t *testing.T) {
		_, err := automation.NewRuleAutomation(wsID, &boardID, &colID, strings.Repeat("x", 256), validRule, true)
		require.ErrorIs(t, err, automation.ErrInvalidAutomationName)
	})
}
