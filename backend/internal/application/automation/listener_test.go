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

func TestEventListener_RecursionLoopGuard(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()

	repo := mocks.NewMockAutomationRepository(t)
	modifier := mocks.NewMockTaskModifier(t)
	reader := mocks.NewMockProjectReader(t)

	listener := appAutomation.NewEventListener(repo, modifier, reader, nil)

	// Case 1: context reached MaxAutomationDepth -> Cycle Guard prevents further execution
	ctx := events.WithAutomationDepth(context.Background(), events.MaxAutomationDepth)
	evt := events.TaskCompletedEvent{
		TaskID:         taskID,
		WorkspaceIDVal: wsID,
	}
	err := listener.HandleTaskCompleted(ctx, evt)
	require.NoError(t, err)

	// Case 2: nil event -> safe no-op
	err = listener.HandleTaskCompleted(context.Background(), nil)
	require.NoError(t, err)
}

func TestEventListener_DoneAndArchive(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	projID := uuid.New()
	doneColID := uuid.New()

	repo := mocks.NewMockAutomationRepository(t)
	modifier := mocks.NewMockTaskModifier(t)
	reader := mocks.NewMockProjectReader(t)

	rule, err := domainAutomation.NewAutomation(
		wsID,
		nil,
		"Done & Archive",
		domainAutomation.TriggerTaskCompleted,
		json.RawMessage(`{}`),
		json.RawMessage(`{"action": "move_to_done"}`),
		true,
	)
	require.NoError(t, err)

	repo.EXPECT().ListEnabledAutomationsByTrigger(
		mock.MatchedBy(func(ctx context.Context) bool {
			return events.IsAutomationTriggered(ctx)
		}),
		wsID,
		domainAutomation.TriggerTaskCompleted,
	).Return([]*domainAutomation.Automation{rule}, nil).Once()

	reader.EXPECT().GetDoneColumn(
		mock.MatchedBy(func(ctx context.Context) bool {
			return events.IsAutomationTriggered(ctx)
		}),
		projID,
		wsID,
	).Return(&doneColID, nil).Once()

	modifier.EXPECT().MoveTaskToColumn(
		mock.MatchedBy(func(ctx context.Context) bool {
			return events.IsAutomationTriggered(ctx)
		}),
		taskID,
		wsID,
		&doneColID,
		&projID,
		1000.0,
	).Return(nil).Once()

	listener := appAutomation.NewEventListener(repo, modifier, reader, nil)

	evt := events.TaskCompletedEvent{
		TaskID:         taskID,
		WorkspaceIDVal: wsID,
		ProjectID:      &projID,
		FromAutomation: false,
	}

	err = listener.HandleTaskCompleted(context.Background(), evt)
	require.NoError(t, err)
}

func TestEventListener_AutoFocus(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	now := time.Now().UTC()

	repo := mocks.NewMockAutomationRepository(t)
	modifier := mocks.NewMockTaskModifier(t)
	reader := mocks.NewMockProjectReader(t)

	rule, err := domainAutomation.NewAutomation(
		wsID,
		nil,
		"Auto-Focus",
		domainAutomation.TriggerTaskUpdated,
		json.RawMessage(`{"has_due_date": true}`),
		json.RawMessage(`{"action": "set_focus"}`),
		true,
	)
	require.NoError(t, err)

	repo.EXPECT().ListEnabledAutomationsByTrigger(
		mock.Anything,
		wsID,
		domainAutomation.TriggerTaskUpdated,
	).Return([]*domainAutomation.Automation{rule}, nil).Once()

	modifier.EXPECT().SetTaskFocus(
		mock.MatchedBy(func(ctx context.Context) bool {
			return events.IsAutomationTriggered(ctx)
		}),
		taskID,
		wsID,
		mock.MatchedBy(func(order *int) bool {
			return order != nil && *order == 1
		}),
	).Return(nil).Once()

	listener := appAutomation.NewEventListener(repo, modifier, reader, nil)

	evt := events.TaskUpdatedEvent{
		TaskID:         taskID,
		WorkspaceIDVal: wsID,
		DueDate:        &now,
		FromAutomation: false,
	}

	err = listener.HandleTaskUpdated(context.Background(), evt)
	require.NoError(t, err)
}

func TestEventListener_InboxZero(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	projID := uuid.New()
	todoColID := uuid.New()

	repo := mocks.NewMockAutomationRepository(t)
	modifier := mocks.NewMockTaskModifier(t)
	reader := mocks.NewMockProjectReader(t)

	rule, err := domainAutomation.NewAutomation(
		wsID,
		nil,
		"Inbox Zero",
		domainAutomation.TriggerTaskUpdated,
		json.RawMessage(`{"project_assigned": true}`),
		json.RawMessage(`{"action": "move_to_first_column"}`),
		true,
	)
	require.NoError(t, err)

	repo.EXPECT().ListEnabledAutomationsByTrigger(
		mock.Anything,
		wsID,
		domainAutomation.TriggerTaskUpdated,
	).Return([]*domainAutomation.Automation{rule}, nil).Once()

	reader.EXPECT().GetFirstTodoColumn(
		mock.MatchedBy(func(ctx context.Context) bool {
			return events.IsAutomationTriggered(ctx)
		}),
		projID,
		wsID,
	).Return(&todoColID, nil).Once()

	modifier.EXPECT().MoveTaskToColumn(
		mock.MatchedBy(func(ctx context.Context) bool {
			return events.IsAutomationTriggered(ctx)
		}),
		taskID,
		wsID,
		&todoColID,
		&projID,
		1000.0,
	).Return(nil).Once()

	listener := appAutomation.NewEventListener(repo, modifier, reader, nil)

	evt := events.TaskUpdatedEvent{
		TaskID:            taskID,
		WorkspaceIDVal:    wsID,
		ProjectID:         &projID,
		ColumnID:          nil, // from inbox
		PreviousProjectID: nil,
		FromAutomation:    false,
	}

	err = listener.HandleTaskUpdated(context.Background(), evt)
	require.NoError(t, err)
}

func TestEventListener_TaskEnteredColumn_SetPriority(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	projID := uuid.New()
	colID := uuid.New()

	repo := mocks.NewMockAutomationRepository(t)
	modifier := mocks.NewMockTaskModifier(t)
	reader := mocks.NewMockProjectReader(t)

	rule, err := domainAutomation.NewColumnAutomation(
		wsID,
		nil,
		colID,
		"Auto-Urgent on Column Enter",
		domainAutomation.TriggerTaskEnteredColumn,
		json.RawMessage(`{}`),
		json.RawMessage(`{"action": "set_priority", "priority": "urgent"}`),
		true,
	)
	require.NoError(t, err)

	repo.EXPECT().ListEnabledAutomationsByTrigger(
		mock.MatchedBy(func(ctx context.Context) bool {
			return events.IsAutomationTriggered(ctx) && events.GetAutomationDepth(ctx) == 1
		}),
		wsID,
		domainAutomation.TriggerTaskEnteredColumn,
	).Return([]*domainAutomation.Automation{rule}, nil).Once()

	modifier.EXPECT().SetTaskPriority(
		mock.MatchedBy(func(ctx context.Context) bool {
			return events.IsAutomationTriggered(ctx) && events.GetAutomationDepth(ctx) == 1
		}),
		taskID,
		wsID,
		"urgent",
	).Return(nil).Once()

	listener := appAutomation.NewEventListener(repo, modifier, reader, nil)

	evt := events.TaskMovedEvent{
		TaskID:         taskID,
		WorkspaceIDVal: wsID,
		ProjectID:      &projID,
		ColumnID:       &colID,
		Position:       1500.0,
	}

	err = listener.HandleTaskMoved(context.Background(), evt)
	require.NoError(t, err)
}

func TestEventListener_TaskCompletedInColumn_MoveToColumn(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	projID := uuid.New()
	colID := uuid.New()
	targetColID := uuid.New()

	repo := mocks.NewMockAutomationRepository(t)
	modifier := mocks.NewMockTaskModifier(t)
	reader := mocks.NewMockProjectReader(t)

	rule, err := domainAutomation.NewColumnAutomation(
		wsID,
		nil,
		colID,
		"Move to Done on Complete",
		domainAutomation.TriggerTaskCompletedInColumn,
		json.RawMessage(`{}`),
		json.RawMessage(`{"action": "move_to_column", "target_column_id": "`+targetColID.String()+`"}`),
		true,
	)
	require.NoError(t, err)

	repo.EXPECT().ListEnabledAutomationsByTrigger(
		mock.Anything,
		wsID,
		domainAutomation.TriggerTaskCompleted,
	).Return([]*domainAutomation.Automation{}, nil).Once()

	repo.EXPECT().ListEnabledAutomationsByTrigger(
		mock.MatchedBy(func(ctx context.Context) bool {
			return events.IsAutomationTriggered(ctx) && events.GetAutomationDepth(ctx) == 1
		}),
		wsID,
		domainAutomation.TriggerTaskCompletedInColumn,
	).Return([]*domainAutomation.Automation{rule}, nil).Once()

	modifier.EXPECT().MoveTaskToColumn(
		mock.MatchedBy(func(ctx context.Context) bool {
			return events.IsAutomationTriggered(ctx) && events.GetAutomationDepth(ctx) == 1
		}),
		taskID,
		wsID,
		&targetColID,
		&projID,
		1000.0,
	).Return(nil).Once()

	listener := appAutomation.NewEventListener(repo, modifier, reader, nil)

	evt := events.TaskCompletedEvent{
		TaskID:         taskID,
		WorkspaceIDVal: wsID,
		ProjectID:      &projID,
		ColumnID:       &colID,
		CompletedAt:    time.Now().UTC(),
	}

	err = listener.HandleTaskCompleted(context.Background(), evt)
	require.NoError(t, err)
}

func TestEventListener_BoardScopeFilter(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	ruleBoardID := uuid.New()
	otherBoardID := uuid.New()

	repo := mocks.NewMockAutomationRepository(t)
	modifier := mocks.NewMockTaskModifier(t)
	reader := mocks.NewMockProjectReader(t)

	rule, err := domainAutomation.NewAutomation(
		wsID,
		&ruleBoardID,
		"Board-Scoped Rule",
		domainAutomation.TriggerTaskCompleted,
		json.RawMessage(`{}`),
		json.RawMessage(`{"action": "set_priority", "priority": "high"}`),
		true,
	)
	require.NoError(t, err)

	listener := appAutomation.NewEventListener(repo, modifier, reader, nil)

	// Scenario 1: task belongs to a different board -> rule should be skipped, modifier NOT called
	repo.EXPECT().ListEnabledAutomationsByTrigger(
		mock.Anything,
		wsID,
		domainAutomation.TriggerTaskCompleted,
	).Return([]*domainAutomation.Automation{rule}, nil).Once()

	evtOtherBoard := events.TaskCompletedEvent{
		TaskID:         taskID,
		WorkspaceIDVal: wsID,
		ProjectID:      &otherBoardID,
	}

	err = listener.HandleTaskCompleted(context.Background(), evtOtherBoard)
	require.NoError(t, err)

	// Scenario 2: task belongs to the matching board -> rule action executes
	repo.EXPECT().ListEnabledAutomationsByTrigger(
		mock.Anything,
		wsID,
		domainAutomation.TriggerTaskCompleted,
	).Return([]*domainAutomation.Automation{rule}, nil).Once()

	modifier.EXPECT().SetTaskPriority(
		mock.Anything,
		taskID,
		wsID,
		"high",
	).Return(nil).Once()

	evtMatchingBoard := events.TaskCompletedEvent{
		TaskID:         taskID,
		WorkspaceIDVal: wsID,
		ProjectID:      &ruleBoardID,
	}

	err = listener.HandleTaskCompleted(context.Background(), evtMatchingBoard)
	require.NoError(t, err)
}



