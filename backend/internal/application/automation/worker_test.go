package automation_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appAutomation "takt/internal/application/automation"
	appMocks "takt/internal/application/automation/mocks"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/domain/events"
)

type mockTaskReader struct {
	mock.Mock
}

func (m *mockTaskReader) ListTasksByColumn(ctx context.Context, wsID, colID uuid.UUID) ([]*appAutomation.TaskInfo, error) {
	args := m.Called(ctx, wsID, colID)
	if tasks := args.Get(0); tasks != nil {
		return tasks.([]*appAutomation.TaskInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTaskReader) GetTaskByID(ctx context.Context, taskID, wsID uuid.UUID) (*appAutomation.TaskInfo, error) {
	args := m.Called(ctx, taskID, wsID)
	if task := args.Get(0); task != nil {
		return task.(*appAutomation.TaskInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTaskReader) ListSubtasks(ctx context.Context, parentID uuid.UUID) ([]*appAutomation.TaskInfo, error) {
	args := m.Called(ctx, parentID)
	if tasks := args.Get(0); tasks != nil {
		return tasks.([]*appAutomation.TaskInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTaskReader) ListTaskActivities(ctx context.Context, taskID uuid.UUID) ([]*appAutomation.TaskActivityInfo, error) {
	args := m.Called(ctx, taskID)
	if acts := args.Get(0); acts != nil {
		return acts.([]*appAutomation.TaskActivityInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestWorker_Tick(t *testing.T) {
	t.Parallel()

	wsID := uuid.New()
	colID := uuid.New()
	targetColID := uuid.New()

	t.Run("processes column_sla_timer rule when aging threshold exceeded", func(t *testing.T) {
		mockRepo := new(appMocks.MockAutomationRepository)
		mockReader := new(mockTaskReader)
		mockMod := new(mockTaskModifier)

		eval := appAutomation.NewEvaluator(mockMod, nil)
		worker := appAutomation.NewWorker(mockRepo, mockReader, eval, 1*time.Minute)

		ruleDef := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{
				Type: domainAutomation.TriggerColumnSLATimer,
				Params: map[string]interface{}{
					"column_id":      colID.String(),
					"duration_hours": 24,
				},
			},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionSetPriority, Params: map[string]interface{}{"priority": "urgent"}},
			},
		}

		autoRule, err := domainAutomation.NewRuleAutomation(wsID, nil, &colID, "Aging Rule", ruleDef, true)
		require.NoError(t, err)

		mockRepo.EXPECT().ListAllEnabledTimerAutomations(mock.Anything).Return([]*domainAutomation.Automation{autoRule}, nil).Once()

		oldTask := &appAutomation.TaskInfo{
			ID:          uuid.New(),
			WorkspaceID: wsID,
			ColumnID:    &colID,
			IsCompleted: false,
			UpdatedAt:   time.Now().UTC().Add(-25 * time.Hour),
		}

		mockReader.On("ListTasksByColumn", mock.Anything, wsID, colID).Return([]*appAutomation.TaskInfo{oldTask}, nil).Once()
		mockMod.On("SetTaskPriority", mock.Anything, oldTask.ID, wsID, "urgent").Return(nil).Once()

		err = worker.Tick(context.Background())
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockReader.AssertExpectations(t)
		mockMod.AssertExpectations(t)
	})

	t.Run("processes scheduled_sweep rule for all tasks in column", func(t *testing.T) {
		mockRepo := new(appMocks.MockAutomationRepository)
		mockReader := new(mockTaskReader)
		mockMod := new(mockTaskModifier)

		eval := appAutomation.NewEvaluator(mockMod, nil)
		worker := appAutomation.NewWorker(mockRepo, mockReader, eval, 1*time.Minute)

		ruleDef := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{
				Type: domainAutomation.TriggerScheduledSweep,
				Params: map[string]interface{}{
					"column_id": colID.String(),
				},
			},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionMoveToColumn, Params: map[string]interface{}{"target_column_id": targetColID.String()}},
			},
		}

		autoRule, err := domainAutomation.NewRuleAutomation(wsID, nil, &colID, "Daily Sweep", ruleDef, true)
		require.NoError(t, err)

		mockRepo.EXPECT().ListAllEnabledTimerAutomations(mock.Anything).Return([]*domainAutomation.Automation{autoRule}, nil).Once()

		task1 := &appAutomation.TaskInfo{
			ID:          uuid.New(),
			WorkspaceID: wsID,
			ColumnID:    &colID,
			IsCompleted: false,
		}

		mockReader.On("ListTasksByColumn", mock.Anything, wsID, colID).Return([]*appAutomation.TaskInfo{task1}, nil).Once()
		mockMod.On("MoveTaskToColumn", mock.Anything, task1.ID, wsID, &targetColID, (*uuid.UUID)(nil), 1000.0).Return(nil).Once()

		err = worker.Tick(context.Background())
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockReader.AssertExpectations(t)
		mockMod.AssertExpectations(t)
	})

	t.Run("processes task_timer rule when delay threshold exceeded", func(t *testing.T) {
		mockRepo := new(appMocks.MockAutomationRepository)
		mockReader := new(mockTaskReader)
		mockMod := new(mockTaskModifier)

		eval := appAutomation.NewEvaluator(mockMod, nil)
		worker := appAutomation.NewWorker(mockRepo, mockReader, eval, 1*time.Minute)

		taskID := uuid.New()
		ruleDef := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{
				Type: domainAutomation.TriggerTaskTimer,
				Params: map[string]interface{}{
					"delay_minutes": 30,
				},
			},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionSetPriority, Params: map[string]interface{}{"priority": "urgent"}},
			},
		}

		autoRule, err := domainAutomation.NewRuleAutomation(wsID, nil, nil, "Task Timer Rule", ruleDef, true)
		require.NoError(t, err)
		autoRule.TaskID = &taskID
		autoRule.CreatedAt = time.Now().UTC().Add(-35 * time.Minute)

		mockRepo.EXPECT().ListAllEnabledTimerAutomations(mock.Anything).Return([]*domainAutomation.Automation{autoRule}, nil).Once()

		task := &appAutomation.TaskInfo{
			ID:          taskID,
			WorkspaceID: wsID,
			IsCompleted: false,
		}

		mockReader.On("GetTaskByID", mock.Anything, taskID, wsID).Return(task, nil).Once()
		mockMod.On("SetTaskPriority", mock.Anything, taskID, wsID, "urgent").Return(nil).Once()

		err = worker.Tick(context.Background())
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockReader.AssertExpectations(t)
		mockMod.AssertExpectations(t)
	})

	t.Run("lifecycle start and stop cleanly", func(t *testing.T) {
		mockRepo := new(appMocks.MockAutomationRepository)
		worker := appAutomation.NewWorker(mockRepo, nil, nil, 10*time.Millisecond)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		worker.Start(ctx)
		time.Sleep(20 * time.Millisecond)
		worker.Stop()
	})
}

func TestCanUseBatchPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rule     *domainAutomation.RuleDefinition
		expected bool
	}{
		{name: "nil rule returns false", rule: nil, expected: false},
		{name: "no actions returns false", rule: &domainAutomation.RuleDefinition{Actions: []domainAutomation.Action{}}, expected: false},
		{
			name: "multiple actions returns false",
			rule: &domainAutomation.RuleDefinition{
				Actions: []domainAutomation.Action{
					{Type: domainAutomation.ActionArchiveTask},
					{Type: domainAutomation.ActionSetPriority},
				},
			},
			expected: false,
		},
		{
			name: "non-archive action returns false",
			rule: &domainAutomation.RuleDefinition{
				Actions: []domainAutomation.Action{{Type: domainAutomation.ActionMoveToColumn}},
			},
			expected: false,
		},
		{
			name: "archive action with no conditions returns true",
			rule: &domainAutomation.RuleDefinition{
				Actions: []domainAutomation.Action{{Type: domainAutomation.ActionArchiveTask}},
			},
			expected: true,
		},
		{
			name: "archive action with is_completed string true returns true",
			rule: &domainAutomation.RuleDefinition{
				Actions: []domainAutomation.Action{
					{Type: domainAutomation.ActionArchiveTask},
				},
				Conditions: []domainAutomation.Condition{
					{Field: domainAutomation.ConditionFieldIsCompleted, Operator: "==", Value: "true"},
				},
			},
			expected: true,
		},
		{
			name: "archive action with is_completed bool true returns true",
			rule: &domainAutomation.RuleDefinition{
				Actions: []domainAutomation.Action{
					{Type: domainAutomation.ActionArchiveTask},
				},
				Conditions: []domainAutomation.Condition{
					{Field: domainAutomation.ConditionFieldIsCompleted, Operator: "==", Value: true},
				},
			},
			expected: true,
		},
		{
			name: "archive action with is_completed false returns false",
			rule: &domainAutomation.RuleDefinition{
				Actions: []domainAutomation.Action{
					{Type: domainAutomation.ActionArchiveTask},
				},
				Conditions: []domainAutomation.Condition{
					{Field: domainAutomation.ConditionFieldIsCompleted, Operator: "==", Value: false},
				},
			},
			expected: false,
		},
		{
			name: "archive action with non-completion condition returns false",
			rule: &domainAutomation.RuleDefinition{
				Actions: []domainAutomation.Action{
					{Type: domainAutomation.ActionArchiveTask},
				},
				Conditions: []domainAutomation.Condition{
					{Field: "priority", Operator: "==", Value: "urgent"},
				},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual := appAutomation.CanUseBatchPath(tc.rule)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestWorker_Tick_FastPath(t *testing.T) {
	t.Parallel()

	wsID := uuid.New()
	colID := uuid.New()

	t.Run("column_sla_timer fast path executes batch SQL and publishes event", func(t *testing.T) {
		mockRepo := new(appMocks.MockAutomationRepository)
		mockBatch := new(appMocks.MockBatchTaskExecutor)
		mockPub := new(appMocks.MockEventPublisher)
		mockReader := new(mockTaskReader)
		eval := appAutomation.NewEvaluator(nil, nil)

		worker := appAutomation.NewWorker(mockRepo, mockReader, eval, 1*time.Minute)
		worker.SetBatchExecutor(mockBatch)
		worker.SetEventPublisher(mockPub)

		ruleDef := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{
				Type: domainAutomation.TriggerColumnSLATimer,
				Params: map[string]interface{}{
					"column_id":      colID.String(),
					"duration_hours": 48,
				},
			},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionArchiveTask},
			},
			Conditions: []domainAutomation.Condition{
				{Field: domainAutomation.ConditionFieldIsCompleted, Operator: domainAutomation.OperatorEquals, Value: "true"},
			},
		}

		autoRule, err := domainAutomation.NewRuleAutomation(wsID, nil, &colID, "Batch SLA", ruleDef, true)
		require.NoError(t, err)

		mockRepo.EXPECT().ListAllEnabledTimerAutomations(mock.Anything).Return([]*domainAutomation.Automation{autoRule}, nil).Once()
		mockBatch.EXPECT().ArchiveCompletedTasksBatchSQL(mock.Anything, wsID, colID, 48.0).Return(int64(7), nil).Once()
		mockPub.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(e events.Event) bool {
			ev, ok := e.(events.BatchSweepCompletedEvent)
			return ok && ev.WorkspaceID() == wsID && ev.ColumnID() == colID && ev.AffectedCount == 7
		})).Return(nil).Once()

		err = worker.Tick(context.Background())
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockBatch.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("scheduled_sweep fast path executes batch SQL with 0 hours", func(t *testing.T) {
		mockRepo := new(appMocks.MockAutomationRepository)
		mockBatch := new(appMocks.MockBatchTaskExecutor)
		mockPub := new(appMocks.MockEventPublisher)
		mockReader := new(mockTaskReader)
		eval := appAutomation.NewEvaluator(nil, nil)

		worker := appAutomation.NewWorker(mockRepo, mockReader, eval, 1*time.Minute)
		worker.SetBatchExecutor(mockBatch)
		worker.SetEventPublisher(mockPub)

		ruleDef := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{
				Type: domainAutomation.TriggerScheduledSweep,
				Params: map[string]interface{}{
					"column_id": colID.String(),
				},
			},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionArchiveTask},
			},
		}

		autoRule, err := domainAutomation.NewRuleAutomation(wsID, nil, &colID, "Batch Sweep", ruleDef, true)
		require.NoError(t, err)

		mockRepo.EXPECT().ListAllEnabledTimerAutomations(mock.Anything).Return([]*domainAutomation.Automation{autoRule}, nil).Once()
		mockBatch.EXPECT().ArchiveCompletedTasksBatchSQL(mock.Anything, wsID, colID, 0.0).Return(int64(3), nil).Once()
		mockPub.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(e events.Event) bool {
			ev, ok := e.(events.BatchSweepCompletedEvent)
			return ok && ev.WorkspaceID() == wsID && ev.ColumnID() == colID && ev.AffectedCount == 3
		})).Return(nil).Once()

		err = worker.Tick(context.Background())
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockBatch.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})
}

