package automation_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appAutomation "takt/internal/application/automation"
	"takt/internal/application/automation/mocks"
	domainAutomation "takt/internal/domain/automation"
)

func TestEvaluator_ReturnToPreviousColumn(t *testing.T) {
	wsID := uuid.New()
	projectID := uuid.New()
	taskID := uuid.New()
	prevColID := uuid.New()
	inboxColID := uuid.New()

	task := &appAutomation.TaskInfo{
		ID:          taskID,
		WorkspaceID: wsID,
		ProjectID:   &projectID,
	}

	t.Run("successfully returns to previous column from activities", func(t *testing.T) {
		mockMod := new(mockTaskModifier)
		mockProj := mocks.NewMockProjectReader(t)
		mockReader := new(mockTaskReader)

		now := time.Now().UTC()
		activities := []*appAutomation.TaskActivityInfo{
			{
				ActivityType: "comment_created",
				CreatedAt:    now.Add(-10 * time.Minute),
			},
			{
				ActivityType: "column_moved",
				Payload: map[string]interface{}{
					"from_column_id": prevColID.String(),
					"to_column_id":   uuid.New().String(),
				},
				CreatedAt: now.Add(-5 * time.Minute),
			},
		}

		mockReader.On("ListTaskActivities", mock.Anything, taskID).Return(activities, nil).Once()
		mockProj.EXPECT().ColumnExists(mock.Anything, prevColID).Return(true, nil).Once()
		mockMod.On("MoveTaskToColumn", mock.Anything, taskID, wsID, &prevColID, &projectID, 1000.0).Return(nil).Once()

		eval := appAutomation.NewEvaluator(mockMod, mockProj, mockReader)

		rule := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionReturnToPreviousColumn},
			},
		}

		err := eval.ExecuteRule(context.Background(), rule, task)
		require.NoError(t, err)
		mockMod.AssertExpectations(t)
		mockReader.AssertExpectations(t)
	})

	t.Run("fallback 1: column does not exist -> moves to project inbox column", func(t *testing.T) {
		mockMod := new(mockTaskModifier)
		mockProj := mocks.NewMockProjectReader(t)
		mockReader := new(mockTaskReader)

		activities := []*appAutomation.TaskActivityInfo{
			{
				ActivityType: "column_moved",
				Payload: map[string]interface{}{
					"from_column_id": prevColID.String(),
				},
				CreatedAt: time.Now().UTC(),
			},
		}

		mockReader.On("ListTaskActivities", mock.Anything, taskID).Return(activities, nil).Once()
		mockProj.EXPECT().ColumnExists(mock.Anything, prevColID).Return(false, nil).Once()
		mockProj.EXPECT().GetInboxColumn(mock.Anything, projectID, wsID).Return(&inboxColID, nil).Once()
		mockMod.On("MoveTaskToColumn", mock.Anything, taskID, wsID, &inboxColID, &projectID, 1000.0).Return(nil).Once()

		eval := appAutomation.NewEvaluator(mockMod, mockProj, mockReader)

		rule := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionReturnToPreviousColumn},
			},
		}

		err := eval.ExecuteRule(context.Background(), rule, task)
		require.NoError(t, err)
		mockMod.AssertExpectations(t)
	})

	t.Run("fallback 2: no activities and no project inbox -> moves to workspace inbox (col nil)", func(t *testing.T) {
		mockMod := new(mockTaskModifier)
		mockProj := mocks.NewMockProjectReader(t)
		mockReader := new(mockTaskReader)

		mockReader.On("ListTaskActivities", mock.Anything, taskID).Return([]*appAutomation.TaskActivityInfo{}, nil).Once()
		mockProj.EXPECT().GetInboxColumn(mock.Anything, projectID, wsID).Return(nil, nil).Once()
		mockMod.On("MoveTaskToColumn", mock.Anything, taskID, wsID, (*uuid.UUID)(nil), &projectID, 1000.0).Return(nil).Once()

		eval := appAutomation.NewEvaluator(mockMod, mockProj, mockReader)

		rule := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionReturnToPreviousColumn},
			},
		}

		err := eval.ExecuteRule(context.Background(), rule, task)
		require.NoError(t, err)
		mockMod.AssertExpectations(t)
	})
}

func TestEvaluator_SubtaskActions(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	subtask1 := &appAutomation.TaskInfo{ID: uuid.New(), IsCompleted: false}
	subtask2 := &appAutomation.TaskInfo{ID: uuid.New(), IsCompleted: true}

	task := &appAutomation.TaskInfo{
		ID:          taskID,
		WorkspaceID: wsID,
	}

	t.Run("complete_all_subtasks completes incomplete subtasks only", func(t *testing.T) {
		mockMod := new(mockTaskModifier)
		mockReader := new(mockTaskReader)

		mockReader.On("ListSubtasks", mock.Anything, taskID).Return([]*appAutomation.TaskInfo{subtask1, subtask2}, nil).Once()
		mockMod.On("SetTaskCompleted", mock.Anything, subtask1.ID, wsID, true).Return(nil).Once()

		eval := appAutomation.NewEvaluator(mockMod, nil, mockReader)

		rule := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionCompleteAllSubtasks},
			},
		}

		err := eval.ExecuteRule(context.Background(), rule, task)
		require.NoError(t, err)
		mockMod.AssertExpectations(t)
	})

	t.Run("reopen_all_subtasks reopens completed subtasks only", func(t *testing.T) {
		mockMod := new(mockTaskModifier)
		mockReader := new(mockTaskReader)

		mockReader.On("ListSubtasks", mock.Anything, taskID).Return([]*appAutomation.TaskInfo{subtask1, subtask2}, nil).Once()
		mockMod.On("SetTaskCompleted", mock.Anything, subtask2.ID, wsID, false).Return(nil).Once()

		eval := appAutomation.NewEvaluator(mockMod, nil, mockReader)

		rule := &domainAutomation.RuleDefinition{
			Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionReopenAllSubtasks},
			},
		}

		err := eval.ExecuteRule(context.Background(), rule, task)
		require.NoError(t, err)
		mockMod.AssertExpectations(t)
	})
}
