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
	"takt/internal/domain/events"
)

func TestEvaluator_ReturnToPreviousColumn(t *testing.T) {
	wsID := uuid.New()
	projectID := uuid.New()
	taskID := uuid.New()
	prevColID := uuid.New()
	inboxColID := uuid.New()

	t.Run("successfully returns to previous column from activities and publishes event", func(t *testing.T) {
		task := &appAutomation.TaskInfo{
			ID:          taskID,
			WorkspaceID: wsID,
			ProjectID:   &projectID,
		}
		mockMod := new(mockTaskModifier)
		mockProj := mocks.NewMockProjectReader(t)
		mockReader := new(mockTaskReader)
		mockPub := mocks.NewMockEventPublisher(t)

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
		mockProj.EXPECT().GetColumnName(mock.Anything, prevColID).Return("Входящие", nil).Once()
		mockMod.On("MoveTaskToColumn", mock.Anything, taskID, wsID, &prevColID, &projectID, 1000.0).Return(nil).Once()

		ruleID := uuid.New()
		rule := &domainAutomation.RuleDefinition{
			ID:      &ruleID,
			Name:    "Возврат в предыдущую",
			Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
			Actions: []domainAutomation.Action{
				{Type: domainAutomation.ActionReturnToPreviousColumn},
			},
		}

		mockPub.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(evt events.Event) bool {
			e, ok := evt.(events.TaskMovedEvent)
			if !ok {
				return false
			}
			return e.TaskID == taskID &&
				e.FromAutomation == true &&
				e.AutomationID != nil && *e.AutomationID == ruleID &&
				e.AutomationName == "Возврат в предыдущую" &&
				e.ToColumnName == "Входящие" &&
				e.ColumnID != nil && *e.ColumnID == prevColID
		})).Return(nil).Once()

		eval := appAutomation.NewEvaluator(mockMod, mockProj, mockReader)
		eval.SetEventPublisher(mockPub)

		err := eval.ExecuteRule(context.Background(), rule, task)
		require.NoError(t, err)
		mockMod.AssertExpectations(t)
		mockReader.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("fallback 1: column does not exist -> moves to project inbox column", func(t *testing.T) {
		task := &appAutomation.TaskInfo{
			ID:          taskID,
			WorkspaceID: wsID,
			ProjectID:   &projectID,
		}
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
		mockProj.EXPECT().GetColumnName(mock.Anything, mock.Anything).Return("Входящие проекта", nil).Maybe()
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

	t.Run("fallback 1.5: no inbox column -> moves to first todo column", func(t *testing.T) {
		task := &appAutomation.TaskInfo{
			ID:          taskID,
			WorkspaceID: wsID,
			ProjectID:   &projectID,
		}
		mockMod := new(mockTaskModifier)
		mockProj := mocks.NewMockProjectReader(t)
		mockReader := new(mockTaskReader)

		firstTodoColID := uuid.New()
		mockReader.On("ListTaskActivities", mock.Anything, taskID).Return([]*appAutomation.TaskActivityInfo{}, nil).Once()
		mockProj.EXPECT().GetInboxColumn(mock.Anything, projectID, wsID).Return(nil, nil).Once()
		mockProj.EXPECT().GetFirstTodoColumn(mock.Anything, projectID, wsID).Return(&firstTodoColID, nil).Once()
		mockProj.EXPECT().GetColumnName(mock.Anything, mock.Anything).Return("To Do", nil).Maybe()
		mockMod.On("MoveTaskToColumn", mock.Anything, taskID, wsID, &firstTodoColID, &projectID, 1000.0).Return(nil).Once()

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

	t.Run("fallback 2: no activities, no inbox and no todo -> moves to workspace inbox (col nil)", func(t *testing.T) {
		task := &appAutomation.TaskInfo{
			ID:          taskID,
			WorkspaceID: wsID,
			ProjectID:   &projectID,
		}
		mockMod := new(mockTaskModifier)
		mockProj := mocks.NewMockProjectReader(t)
		mockReader := new(mockTaskReader)

		mockReader.On("ListTaskActivities", mock.Anything, taskID).Return([]*appAutomation.TaskActivityInfo{}, nil).Once()
		mockProj.EXPECT().GetInboxColumn(mock.Anything, projectID, wsID).Return(nil, nil).Once()
		mockProj.EXPECT().GetFirstTodoColumn(mock.Anything, projectID, wsID).Return(nil, nil).Once()
		mockProj.EXPECT().GetColumnName(mock.Anything, mock.Anything).Return("", nil).Maybe()
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

func TestEvaluator_MoveToColumn_TaskMovedEvent(t *testing.T) {
	wsID := uuid.New()
	projectID := uuid.New()
	taskID := uuid.New()
	fromColID := uuid.New()
	targetColID := uuid.New()
	ruleID := uuid.New()

	task := &appAutomation.TaskInfo{
		ID:          taskID,
		WorkspaceID: wsID,
		ProjectID:   &projectID,
		ColumnID:    &fromColID,
	}

	mockMod := new(mockTaskModifier)
	mockProj := mocks.NewMockProjectReader(t)
	mockPub := mocks.NewMockEventPublisher(t)

	mockMod.On("MoveTaskToColumn", mock.Anything, taskID, wsID, &targetColID, &projectID, 1000.0).Return(nil).Once()
	mockProj.EXPECT().GetColumnName(mock.Anything, fromColID).Return("В работе", nil).Once()
	mockProj.EXPECT().GetColumnName(mock.Anything, targetColID).Return("Готово", nil).Once()

	mockPub.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(evt events.Event) bool {
		e, ok := evt.(events.TaskMovedEvent)
		if !ok {
			return false
		}
		return e.TaskID == taskID &&
			e.FromAutomation == true &&
			e.AutomationID != nil && *e.AutomationID == ruleID &&
			e.AutomationName == "Автоперенос в Готово" &&
			e.FromColumnName == "В работе" &&
			e.ToColumnName == "Готово" &&
			e.PreviousColumnID != nil && *e.PreviousColumnID == fromColID &&
			e.ColumnID != nil && *e.ColumnID == targetColID
	})).Return(nil).Once()

	eval := appAutomation.NewEvaluator(mockMod, mockProj)
	eval.SetEventPublisher(mockPub)

	rule := &domainAutomation.RuleDefinition{
		ID:      &ruleID,
		Name:    "Автоперенос в Готово",
		Trigger: domainAutomation.Trigger{Type: domainAutomation.TriggerTaskEnteredColumn},
		Actions: []domainAutomation.Action{
			{
				Type: domainAutomation.ActionMoveToColumn,
				Params: map[string]interface{}{
					"target_column_id": targetColID.String(),
				},
			},
		},
	}

	err := eval.ExecuteRule(context.Background(), rule, task)
	require.NoError(t, err)
	mockMod.AssertExpectations(t)
	mockPub.AssertExpectations(t)
	require.Equal(t, &targetColID, task.ColumnID)
	require.Equal(t, &fromColID, task.PreviousColumnID)
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
