package automation_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appAutomation "takt/internal/application/automation"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/domain/events"
)

type mockTaskModifier struct {
	mock.Mock
}

func (m *mockTaskModifier) MoveTaskToColumn(ctx context.Context, taskID, wsID uuid.UUID, colID, prjID *uuid.UUID, pos float64) error {
	args := m.Called(ctx, taskID, wsID, colID, prjID, pos)
	return args.Error(0)
}

func (m *mockTaskModifier) SetTaskFocus(ctx context.Context, taskID, wsID uuid.UUID, focusOrder *int) error {
	args := m.Called(ctx, taskID, wsID, focusOrder)
	return args.Error(0)
}

func (m *mockTaskModifier) SetTaskPriority(ctx context.Context, taskID, wsID uuid.UUID, priority string) error {
	args := m.Called(ctx, taskID, wsID, priority)
	return args.Error(0)
}

func (m *mockTaskModifier) AddTagToTask(ctx context.Context, taskID, wsID uuid.UUID, tagID *uuid.UUID, tagName string) error {
	args := m.Called(ctx, taskID, wsID, tagID, tagName)
	return args.Error(0)
}

func (m *mockTaskModifier) RemoveTagFromTask(ctx context.Context, taskID, wsID uuid.UUID, tagID *uuid.UUID, tagName string) error {
	args := m.Called(ctx, taskID, wsID, tagID, tagName)
	return args.Error(0)
}

func (m *mockTaskModifier) SetTaskDueDate(ctx context.Context, taskID, wsID uuid.UUID, dueDate *time.Time) error {
	args := m.Called(ctx, taskID, wsID, dueDate)
	return args.Error(0)
}

func (m *mockTaskModifier) SetTaskCompleted(ctx context.Context, taskID, wsID uuid.UUID, isCompleted bool) error {
	args := m.Called(ctx, taskID, wsID, isCompleted)
	return args.Error(0)
}

func (m *mockTaskModifier) ArchiveTask(ctx context.Context, taskID, wsID uuid.UUID) error {
	args := m.Called(ctx, taskID, wsID)
	return args.Error(0)
}

func TestEvaluator_EvaluateConditions(t *testing.T) {
	t.Parallel()
	eval := appAutomation.NewEvaluator(nil, nil)

	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	baseTask := &appAutomation.TaskInfo{
		ID:                uuid.New(),
		WorkspaceID:       uuid.New(),
		Priority:          "urgent",
		IsCompleted:       false,
		DueDate:           &now,
		Tags:              []string{"backend", "bug"},
		TimeInColumnHours: 25.5,
	}

	tests := []struct {
		name       string
		logicalOp  string
		conditions []domainAutomation.Condition
		task       *appAutomation.TaskInfo
		wantMatch  bool
	}{
		{name: "empty conditions always match (AND logic)", conditions: nil, task: baseTask, wantMatch: true},
		{name: "empty conditions always match (OR logic)", logicalOp: domainAutomation.LogicOr, conditions: nil, task: baseTask, wantMatch: true},
		{name: "priority equals match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "urgent"}}, task: baseTask, wantMatch: true},
		{name: "priority equals case-insensitive match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "URGENT"}}, task: baseTask, wantMatch: true},
		{name: "priority not_equals match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorNotEquals, Value: "low"}}, task: baseTask, wantMatch: true},
		{name: "priority in list match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorIn, Value: []string{"urgent", "high"}}}, task: baseTask, wantMatch: true},
		{name: "priority not_in list match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorNotIn, Value: "low,medium"}}, task: baseTask, wantMatch: true},
		{name: "is_completed equals false match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldIsCompleted, Operator: domainAutomation.OperatorEquals, Value: false}}, task: baseTask, wantMatch: true},
		{name: "is_completed is_empty match for false", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldIsCompleted, Operator: domainAutomation.OperatorIsEmpty}}, task: baseTask, wantMatch: true},
		{name: "time_in_column_hours greater_than match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldTimeInColumnHours, Operator: domainAutomation.OperatorGreaterThan, Value: 24}}, task: baseTask, wantMatch: true},
		{name: "time_in_column_hours less_than reject", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldTimeInColumnHours, Operator: domainAutomation.OperatorLessThan, Value: 20}}, task: baseTask, wantMatch: false},
		{name: "tags contains match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldTags, Operator: domainAutomation.OperatorContains, Value: "backend"}}, task: baseTask, wantMatch: true},
		{name: "tags not_contains reject", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldTags, Operator: domainAutomation.OperatorNotContains, Value: "bug"}}, task: baseTask, wantMatch: false},
		{name: "tags in list match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldTags, Operator: domainAutomation.OperatorIn, Value: []string{"frontend", "backend"}}}, task: baseTask, wantMatch: true},
		{name: "tags not_in list match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldTags, Operator: domainAutomation.OperatorNotIn, Value: "infra,ops"}}, task: baseTask, wantMatch: true},
		{name: "tags is_empty reject for non-empty tags", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldTags, Operator: domainAutomation.OperatorIsEmpty}}, task: baseTask, wantMatch: false},
		{name: "due_date is_not_empty match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldDueDate, Operator: domainAutomation.OperatorIsNotEmpty}}, task: baseTask, wantMatch: true},
		{name: "due_date greater_than yesterday match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldDueDate, Operator: domainAutomation.OperatorGreaterThan, Value: yesterday.Format("2006-01-02")}}, task: baseTask, wantMatch: true},
		{name: "due_date less_than tomorrow match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldDueDate, Operator: domainAutomation.OperatorLessThan, Value: tomorrow.Format("2006-01-02")}}, task: baseTask, wantMatch: true},
		{name: "due_date equals today preset match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldDueDate, Operator: domainAutomation.OperatorEquals, Value: "today"}}, task: baseTask, wantMatch: true},
		{name: "due_date less_than tomorrow preset match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldDueDate, Operator: domainAutomation.OperatorLessThan, Value: "tomorrow"}}, task: baseTask, wantMatch: true},
		{name: "due_date_status due_soon match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldDueDateStatus, Operator: domainAutomation.OperatorEquals, Value: "due_soon"}}, task: baseTask, wantMatch: true},
		{name: "due_date_status overdue match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldDueDateStatus, Operator: domainAutomation.OperatorEquals, Value: "overdue"}}, task: &appAutomation.TaskInfo{DueDate: &yesterday}, wantMatch: true},
		{name: "due_date_status no_due_date match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldDueDateStatus, Operator: domainAutomation.OperatorEquals, Value: "no_due_date"}}, task: &appAutomation.TaskInfo{DueDate: nil}, wantMatch: true},
		{name: "due_date_status on_track match", conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldDueDateStatus, Operator: domainAutomation.OperatorEquals, Value: "on_track"}}, task: &appAutomation.TaskInfo{DueDate: func() *time.Time { f := now.AddDate(0, 0, 5); return &f }()}, wantMatch: true},
		{
			name: "multiple conditions with AND logic: all pass",
			conditions: []domainAutomation.Condition{
				{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "urgent"},
				{Field: domainAutomation.ConditionFieldTimeInColumnHours, Operator: domainAutomation.OperatorGreaterThan, Value: 24},
				{Field: domainAutomation.ConditionFieldTags, Operator: domainAutomation.OperatorContains, Value: "bug"},
			},
			task:      baseTask,
			wantMatch: true,
		},
		{
			name: "multiple conditions with AND logic: one fails -> reject",
			conditions: []domainAutomation.Condition{
				{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "urgent"},
				{Field: domainAutomation.ConditionFieldTimeInColumnHours, Operator: domainAutomation.OperatorGreaterThan, Value: 48},
				{Field: domainAutomation.ConditionFieldTags, Operator: domainAutomation.OperatorContains, Value: "bug"},
			},
			task:      baseTask,
			wantMatch: false,
		},
		{
			name:      "multiple conditions with OR logic: first passes -> match",
			logicalOp: domainAutomation.LogicOr,
			conditions: []domainAutomation.Condition{
				{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "urgent"},
				{Field: domainAutomation.ConditionFieldTimeInColumnHours, Operator: domainAutomation.OperatorGreaterThan, Value: 48},
			},
			task:      baseTask,
			wantMatch: true,
		},
		{
			name:      "multiple conditions with OR logic: last passes -> match",
			logicalOp: domainAutomation.LogicOr,
			conditions: []domainAutomation.Condition{
				{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "low"},
				{Field: domainAutomation.ConditionFieldTags, Operator: domainAutomation.OperatorContains, Value: "bug"},
			},
			task:      baseTask,
			wantMatch: true,
		},
		{
			name:      "multiple conditions with OR logic: all fail -> reject",
			logicalOp: domainAutomation.LogicOr,
			conditions: []domainAutomation.Condition{
				{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "low"},
				{Field: domainAutomation.ConditionFieldTimeInColumnHours, Operator: domainAutomation.OperatorGreaterThan, Value: 48},
				{Field: domainAutomation.ConditionFieldTags, Operator: domainAutomation.OperatorContains, Value: "frontend"},
			},
			task:      baseTask,
			wantMatch: false,
		},
		{
			name:       "subtasks_all_completed true when all done",
			conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldSubtasksAllCompleted, Operator: domainAutomation.OperatorEquals, Value: true}},
			task:       &appAutomation.TaskInfo{SubtasksCount: 3, SubtasksDoneCount: 3},
			wantMatch:  true,
		},
		{
			name:       "subtasks_all_completed false when some not done",
			conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldSubtasksAllCompleted, Operator: domainAutomation.OperatorEquals, Value: true}},
			task:       &appAutomation.TaskInfo{SubtasksCount: 3, SubtasksDoneCount: 2},
			wantMatch:  false,
		},
		{
			name:       "subtasks_all_completed false when 0 subtasks",
			conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldSubtasksAllCompleted, Operator: domainAutomation.OperatorEquals, Value: true}},
			task:       &appAutomation.TaskInfo{SubtasksCount: 0, SubtasksDoneCount: 0},
			wantMatch:  false,
		},
		{
			name:       "subtasks_count equals match",
			conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldSubtasksCount, Operator: domainAutomation.OperatorEquals, Value: 5}},
			task:       &appAutomation.TaskInfo{SubtasksCount: 5},
			wantMatch:  true,
		},
		{
			name:       "subtasks_count greater_than match",
			conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldSubtasksCount, Operator: domainAutomation.OperatorGreaterThan, Value: 2}},
			task:       &appAutomation.TaskInfo{SubtasksCount: 3},
			wantMatch:  true,
		},
		{
			name:       "subtasks_count less_than match",
			conditions: []domainAutomation.Condition{{Field: domainAutomation.ConditionFieldSubtasksCount, Operator: domainAutomation.OperatorLessThan, Value: 5}},
			task:       &appAutomation.TaskInfo{SubtasksCount: 3},
			wantMatch:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := eval.EvaluateConditions(tt.logicalOp, tt.conditions, tt.task)
			assert.Equal(t, tt.wantMatch, got)
		})
	}
}

func TestEvaluator_ExecuteRule(t *testing.T) {
	t.Parallel()

	wsID := uuid.New()
	taskID := uuid.New()
	targetColID := uuid.New()
	targetTagID := uuid.New()

	task := &appAutomation.TaskInfo{
		ID:          taskID,
		WorkspaceID: wsID,
		Priority:    "medium",
	}

	t.Run("executes all actions sequentially", func(t *testing.T) {
		mockMod := new(mockTaskModifier)
		eval := appAutomation.NewEvaluator(mockMod, nil)

		mockMod.On("SetTaskPriority", mock.Anything, taskID, wsID, "urgent").Return(nil).Once()
		mockMod.On("MoveTaskToColumn", mock.Anything, taskID, wsID, &targetColID, (*uuid.UUID)(nil), 1000.0).Return(nil).Once()
		mockMod.On("AddTagToTask", mock.Anything, taskID, wsID, &targetTagID, "").Return(nil).Once()
		mockMod.On("RemoveTagFromTask", mock.Anything, taskID, wsID, (*uuid.UUID)(nil), "old-tag").Return(nil).Once()
		mockMod.On("SetTaskDueDate", mock.Anything, taskID, wsID, mock.Anything).Return(nil).Once()
		mockMod.On("ArchiveTask", mock.Anything, taskID, wsID).Return(nil).Once()

		rule := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionSetPriority, Params: map[string]interface{}{"priority": "urgent"}},
				{Type: domainAutomation.ActionMoveToColumn, Params: map[string]interface{}{"target_column_id": targetColID.String()}},
				{Type: domainAutomation.ActionAddTag, Params: map[string]interface{}{"tag_id": targetTagID.String()}},
				{Type: domainAutomation.ActionRemoveTag, Params: map[string]interface{}{"name": "old-tag"}},
				{Type: domainAutomation.ActionSetDueDate, Params: map[string]interface{}{"preset": "tomorrow"}},
				{Type: domainAutomation.ActionArchiveTask},
			},
		}

		err := eval.ExecuteRule(context.Background(), rule, task)
		require.NoError(t, err)
		mockMod.AssertExpectations(t)
	})

	t.Run("Cycle Guard halts execution when depth limit exceeded", func(t *testing.T) {
		mockMod := new(mockTaskModifier)
		eval := appAutomation.NewEvaluator(mockMod, nil)

		ctx := events.WithAutomationTrigger(context.Background())
		ctx = events.WithAutomationTrigger(ctx)
		ctx = events.WithAutomationTrigger(ctx)
		// MaxAutomationDepth reached

		rule := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionSetPriority, Params: map[string]interface{}{"priority": "urgent"}},
			},
		}

		err := eval.ExecuteRule(ctx, rule, task)
		require.NoError(t, err)
		mockMod.AssertNotCalled(t, "SetTaskPriority")
	})

	t.Run("executes actions when OR condition matches at least one condition", func(t *testing.T) {
		mockMod := new(mockTaskModifier)
		eval := appAutomation.NewEvaluator(mockMod, nil)

		mockMod.On("SetTaskPriority", mock.Anything, taskID, wsID, "urgent").Return(nil).Once()

		rule := &domainAutomation.RuleDefinition{
			Trigger:         domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			LogicalOperator: domainAutomation.LogicOr,
			Conditions: []domainAutomation.Condition{
				{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "low"},    // fails
				{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "medium"}, // passes
			},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionSetPriority, Params: map[string]interface{}{"priority": "urgent"}},
			},
		}

		err := eval.ExecuteRule(context.Background(), rule, task)
		require.NoError(t, err)
		mockMod.AssertExpectations(t)
	})

	t.Run("skips actions when AND condition fails at least one condition", func(t *testing.T) {
		mockMod := new(mockTaskModifier)
		eval := appAutomation.NewEvaluator(mockMod, nil)

		rule := &domainAutomation.RuleDefinition{
			Trigger:         domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			LogicalOperator: domainAutomation.LogicAnd,
			Conditions: []domainAutomation.Condition{
				{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "medium"}, // passes
				{Field: domainAutomation.ConditionFieldPriority, Operator: domainAutomation.OperatorEquals, Value: "low"},    // fails
			},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionSetPriority, Params: map[string]interface{}{"priority": "urgent"}},
			},
		}

		err := eval.ExecuteRule(context.Background(), rule, task)
		require.NoError(t, err)
		mockMod.AssertNotCalled(t, "SetTaskPriority")
	})

	t.Run("executes Sprint 0.6.5 actions: mark_completed, mark_uncompleted, clear_due_date", func(t *testing.T) {
		mockMod := new(mockTaskModifier)
		eval := appAutomation.NewEvaluator(mockMod, nil)

		mockMod.On("SetTaskCompleted", mock.Anything, taskID, wsID, true).Return(nil).Once()
		mockMod.On("SetTaskCompleted", mock.Anything, taskID, wsID, false).Return(nil).Once()
		mockMod.On("SetTaskDueDate", mock.Anything, taskID, wsID, (*time.Time)(nil)).Return(nil).Once()

		rule := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionMarkCompleted},
				{Type: domainAutomation.ActionMarkUncompleted},
				{Type: domainAutomation.ActionClearDueDate},
			},
		}

		err := eval.ExecuteRule(context.Background(), rule, task)
		require.NoError(t, err)
		mockMod.AssertExpectations(t)
	})
}
