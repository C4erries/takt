package automation_test

import (
	"context"
	"encoding/json"
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

type mockTaskReaderForSubtasks struct {
	mock.Mock
}

func (m *mockTaskReaderForSubtasks) ListTasksByColumn(ctx context.Context, wsID, colID uuid.UUID) ([]*appAutomation.TaskInfo, error) {
	return nil, nil
}

func (m *mockTaskReaderForSubtasks) GetTaskByID(ctx context.Context, taskID, wsID uuid.UUID) (*appAutomation.TaskInfo, error) {
	args := m.Called(ctx, taskID, wsID)
	if ti := args.Get(0); ti != nil {
		return ti.(*appAutomation.TaskInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTaskReaderForSubtasks) ListSubtasks(ctx context.Context, parentID uuid.UUID) ([]*appAutomation.TaskInfo, error) {
	args := m.Called(ctx, parentID)
	if subtasks := args.Get(0); subtasks != nil {
		return subtasks.([]*appAutomation.TaskInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTaskReaderForSubtasks) ListTaskActivities(ctx context.Context, taskID uuid.UUID) ([]*appAutomation.TaskActivityInfo, error) {
	args := m.Called(ctx, taskID)
	if acts := args.Get(0); acts != nil {
		return acts.([]*appAutomation.TaskActivityInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestEventListener_HandleTaskCompleted_AutoCompleteParent(t *testing.T) {
	wsID := uuid.New()
	parentID := uuid.New()
	subtaskID := uuid.New()

	t.Run("all subtasks complete marks parent complete when rule enabled by name", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		modifier := mocks.NewMockTaskModifier(t)
		reader := mocks.NewMockProjectReader(t)
		taskReader := new(mockTaskReaderForSubtasks)

		rule, err := domainAutomation.NewAutomation(
			wsID,
			nil,
			"auto_complete_parent_on_subtasks_done",
			domainAutomation.TriggerTaskCompleted,
			json.RawMessage(`{}`),
			json.RawMessage(`{"action": "noop"}`),
			true,
		)
		require.NoError(t, err)

		repo.EXPECT().ListEnabledAutomationsByTrigger(mock.Anything, wsID, domainAutomation.TriggerTaskCompleted).
			Return([]*domainAutomation.Automation{}, nil).Once()
		repo.EXPECT().ListEnabledAutomationsByTrigger(mock.Anything, wsID, mock.Anything).
			Return([]*domainAutomation.Automation{}, nil).Maybe()
		repo.EXPECT().ListWorkspaceAutomations(mock.Anything, wsID).
			Return([]*domainAutomation.Automation{rule}, nil).Once()

		subtask1 := &appAutomation.TaskInfo{ID: subtaskID, IsCompleted: true, ParentID: &parentID}
		subtask2 := &appAutomation.TaskInfo{ID: uuid.New(), IsCompleted: true, ParentID: &parentID}
		taskReader.On("ListSubtasks", mock.Anything, parentID).Return([]*appAutomation.TaskInfo{subtask1, subtask2}, nil).Once()

		modifier.EXPECT().SetTaskCompleted(mock.Anything, parentID, wsID, true).Return(nil).Once()

		listener := appAutomation.NewEventListener(repo, modifier, reader, nil, taskReader)

		evt := events.TaskCompletedEvent{
			TaskID:         subtaskID,
			WorkspaceIDVal: wsID,
			ParentID:       &parentID,
			CompletedAt:    time.Now().UTC(),
		}

		err = listener.HandleTaskCompleted(context.Background(), evt)
		require.NoError(t, err)
		taskReader.AssertExpectations(t)
	})

	t.Run("sibling incomplete does not mark parent complete", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		modifier := mocks.NewMockTaskModifier(t)
		reader := mocks.NewMockProjectReader(t)
		taskReader := new(mockTaskReaderForSubtasks)

		rule, err := domainAutomation.NewAutomation(
			wsID,
			nil,
			"auto_complete_parent_on_subtasks_done",
			domainAutomation.TriggerTaskCompleted,
			json.RawMessage(`{}`),
			json.RawMessage(`{"action": "noop"}`),
			true,
		)
		require.NoError(t, err)

		repo.EXPECT().ListEnabledAutomationsByTrigger(mock.Anything, wsID, domainAutomation.TriggerTaskCompleted).
			Return([]*domainAutomation.Automation{}, nil).Once()
		repo.EXPECT().ListEnabledAutomationsByTrigger(mock.Anything, wsID, mock.Anything).
			Return([]*domainAutomation.Automation{}, nil).Maybe()
		repo.EXPECT().ListWorkspaceAutomations(mock.Anything, wsID).
			Return([]*domainAutomation.Automation{rule}, nil).Once()

		subtask1 := &appAutomation.TaskInfo{ID: subtaskID, IsCompleted: true, ParentID: &parentID}
		subtask2 := &appAutomation.TaskInfo{ID: uuid.New(), IsCompleted: false, ParentID: &parentID}
		taskReader.On("ListSubtasks", mock.Anything, parentID).Return([]*appAutomation.TaskInfo{subtask1, subtask2}, nil).Once()

		listener := appAutomation.NewEventListener(repo, modifier, reader, nil, taskReader)

		evt := events.TaskCompletedEvent{
			TaskID:         subtaskID,
			WorkspaceIDVal: wsID,
			ParentID:       &parentID,
			CompletedAt:    time.Now().UTC(),
		}

		err = listener.HandleTaskCompleted(context.Background(), evt)
		require.NoError(t, err)
		taskReader.AssertExpectations(t)
	})

	t.Run("rule enabled via action mark_parent_completed", func(t *testing.T) {
		repo := mocks.NewMockAutomationRepository(t)
		modifier := mocks.NewMockTaskModifier(t)
		reader := mocks.NewMockProjectReader(t)
		taskReader := new(mockTaskReaderForSubtasks)

		rule, err := domainAutomation.NewAutomation(
			wsID,
			nil,
			"Custom rule name",
			domainAutomation.TriggerTaskCompleted,
			json.RawMessage(`{}`),
			json.RawMessage(`{"action": "mark_parent_completed"}`),
			true,
		)
		require.NoError(t, err)

		repo.EXPECT().ListEnabledAutomationsByTrigger(mock.Anything, wsID, domainAutomation.TriggerTaskCompleted).
			Return([]*domainAutomation.Automation{}, nil).Once()
		repo.EXPECT().ListEnabledAutomationsByTrigger(mock.Anything, wsID, mock.Anything).
			Return([]*domainAutomation.Automation{}, nil).Maybe()
		repo.EXPECT().ListWorkspaceAutomations(mock.Anything, wsID).
			Return([]*domainAutomation.Automation{rule}, nil).Once()

		subtask1 := &appAutomation.TaskInfo{ID: subtaskID, IsCompleted: true, ParentID: &parentID}
		taskReader.On("ListSubtasks", mock.Anything, parentID).Return([]*appAutomation.TaskInfo{subtask1}, nil).Once()

		modifier.EXPECT().SetTaskCompleted(mock.Anything, parentID, wsID, true).Return(nil).Once()

		listener := appAutomation.NewEventListener(repo, modifier, reader, nil, taskReader)

		evt := events.TaskCompletedEvent{
			TaskID:         subtaskID,
			WorkspaceIDVal: wsID,
			ParentID:       &parentID,
			CompletedAt:    time.Now().UTC(),
		}

		err = listener.HandleTaskCompleted(context.Background(), evt)
		require.NoError(t, err)
		taskReader.AssertExpectations(t)
	})
}
