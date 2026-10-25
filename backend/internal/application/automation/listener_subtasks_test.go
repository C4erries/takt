package automation_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appAutomation "takt/internal/application/automation"
	"takt/internal/application/automation/mocks"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/domain/events"
)

func TestEventListener_HandleTaskUpdated_TriggerTaskUncompleted(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()

	repo := mocks.NewMockAutomationRepository(t)
	modifier := mocks.NewMockTaskModifier(t)
	reader := mocks.NewMockProjectReader(t)

	rule, err := domainAutomation.NewAutomation(
		wsID,
		nil,
		"On Uncompleted Rule",
		domainAutomation.TriggerTaskUncompleted,
		json.RawMessage(`{}`),
		json.RawMessage(`{"action": "set_priority", "priority": "high"}`),
		true,
	)
	require.NoError(t, err)

	repo.EXPECT().ListEnabledAutomationsByTrigger(
		mock.Anything,
		wsID,
		domainAutomation.TriggerTaskUpdated,
	).Return(nil, nil).Once()

	repo.EXPECT().ListEnabledAutomationsByTrigger(
		mock.Anything,
		wsID,
		domainAutomation.TriggerTaskUncompleted,
	).Return([]*domainAutomation.Automation{rule}, nil).Once()

	modifier.EXPECT().SetTaskPriority(mock.Anything, taskID, wsID, "high").Return(nil).Once()

	listener := appAutomation.NewEventListener(repo, modifier, reader, nil)

	evt := events.TaskUpdatedEvent{
		TaskID:         taskID,
		WorkspaceIDVal: wsID,
		IsCompleted:    false,
	}

	err = listener.HandleTaskUpdated(context.Background(), evt)
	require.NoError(t, err)
}

func TestEventListener_SubtaskTriggers(t *testing.T) {
	wsID := uuid.New()
	parentID := uuid.New()
	subtask1ID := uuid.New()
	subtask2ID := uuid.New()

	parentTask := &appAutomation.TaskInfo{
		ID:          parentID,
		WorkspaceID: wsID,
		Priority:    "medium",
	}

	t.Run("HandleTaskCompleted triggers AnySubtaskCompleted and AllSubtasksCompleted", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		modifier := mocks.NewMockTaskModifier(t)
		reader := mocks.NewMockProjectReader(t)
		taskReader := new(mockTaskReader)

		ruleAny, err := domainAutomation.NewAutomation(
			wsID,
			nil,
			"Any Subtask Done",
			domainAutomation.TriggerAnySubtaskCompleted,
			json.RawMessage(`{}`),
			json.RawMessage(`{"action": "set_priority", "priority": "high"}`),
			true,
		)
		require.NoError(t, err)

		ruleAll, err := domainAutomation.NewAutomation(
			wsID,
			nil,
			"All Subtasks Done",
			domainAutomation.TriggerAllSubtasksCompleted,
			json.RawMessage(`{}`),
			json.RawMessage(`{"action": "set_priority", "priority": "urgent"}`),
			true,
		)
		require.NoError(t, err)

		repo.EXPECT().ListEnabledAutomationsByTrigger(mock.Anything, wsID, domainAutomation.TriggerTaskCompleted).
			Return(nil, nil).Once()
		repo.EXPECT().ListWorkspaceAutomations(mock.Anything, wsID).
			Return(nil, nil).Once()

		taskReader.On("GetTaskByID", mock.Anything, parentID, wsID).Return(parentTask, nil).Once()
		repo.EXPECT().ListEnabledAutomationsByTrigger(mock.Anything, wsID, domainAutomation.TriggerAnySubtaskCompleted).
			Return([]*domainAutomation.Automation{ruleAny}, nil).Once()

		// Siblings: subtask1 is being completed, subtask2 was already completed
		taskReader.On("ListSubtasks", mock.Anything, parentID).Return([]*appAutomation.TaskInfo{
			{ID: subtask1ID, IsCompleted: true},
			{ID: subtask2ID, IsCompleted: true},
		}, nil).Once()

		repo.EXPECT().ListEnabledAutomationsByTrigger(mock.Anything, wsID, domainAutomation.TriggerAllSubtasksCompleted).
			Return([]*domainAutomation.Automation{ruleAll}, nil).Once()

		modifier.EXPECT().SetTaskPriority(mock.Anything, parentID, wsID, "high").Return(nil).Once()
		modifier.EXPECT().SetTaskPriority(mock.Anything, parentID, wsID, "urgent").Return(nil).Once()

		listener := appAutomation.NewEventListener(repo, modifier, reader, nil, taskReader)

		evt := events.TaskCompletedEvent{
			TaskID:         subtask1ID,
			ParentID:       &parentID,
			WorkspaceIDVal: wsID,
		}

		err = listener.HandleTaskCompleted(context.Background(), evt)
		require.NoError(t, err)
		modifier.AssertExpectations(t)
	})

	t.Run("HandleTaskCreated triggers TriggerSubtaskCreated for parent", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		modifier := mocks.NewMockTaskModifier(t)
		reader := mocks.NewMockProjectReader(t)
		taskReader := new(mockTaskReader)

		ruleCreated, err := domainAutomation.NewAutomation(
			wsID,
			nil,
			"Subtask Created",
			domainAutomation.TriggerSubtaskCreated,
			json.RawMessage(`{}`),
			json.RawMessage(`{"action": "set_priority", "priority": "low"}`),
			true,
		)
		require.NoError(t, err)

		repo.EXPECT().ListEnabledAutomationsByTrigger(mock.Anything, wsID, domainAutomation.TriggerTaskUpdated).
			Return(nil, nil).Once()

		taskReader.On("GetTaskByID", mock.Anything, parentID, wsID).Return(parentTask, nil).Once()
		repo.EXPECT().ListEnabledAutomationsByTrigger(mock.Anything, wsID, domainAutomation.TriggerSubtaskCreated).
			Return([]*domainAutomation.Automation{ruleCreated}, nil).Once()

		modifier.EXPECT().SetTaskPriority(mock.Anything, parentID, wsID, "low").Return(nil).Once()

		listener := appAutomation.NewEventListener(repo, modifier, reader, nil, taskReader)

		evt := events.TaskCreatedEvent{
			TaskID:         subtask1ID,
			ParentID:       &parentID,
			WorkspaceIDVal: wsID,
		}

		err = listener.HandleTaskCreated(context.Background(), evt)
		require.NoError(t, err)
		modifier.AssertExpectations(t)
	})
}
