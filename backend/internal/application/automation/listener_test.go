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

	// Case 1: event flagged as FromAutomation
	evt := events.TaskCompletedEvent{
		TaskID:         taskID,
		WorkspaceIDVal: wsID,
		FromAutomation: true,
	}
	err := listener.HandleTaskCompleted(context.Background(), evt)
	require.NoError(t, err)

	// Case 2: context marked with WithAutomationTrigger
	ctx := events.WithAutomationTrigger(context.Background())
	evt2 := events.TaskCompletedEvent{
		TaskID:         taskID,
		WorkspaceIDVal: wsID,
		FromAutomation: false,
	}
	err = listener.HandleTaskCompleted(ctx, evt2)
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

