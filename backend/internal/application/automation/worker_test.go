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
