package task_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	taskApp "takt/internal/application/task"
	"takt/internal/application/task/mocks"
	"takt/internal/domain/events"
	domainTask "takt/internal/domain/task"
	"takt/internal/infrastructure/eventbus"
)

func TestActivityLogger_Register(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewInMemoryEventBus()
	taskRepo := mocks.NewMockTaskRepository(t)
	logger := taskApp.NewActivityLogger(taskRepo)

	err := logger.Register(bus)
	require.NoError(t, err)

	// Register with nil bus should return nil
	assert.NoError(t, logger.Register(nil))
}

func TestActivityLogger_HandleTaskMoved(t *testing.T) {
	t.Parallel()

	taskID := uuid.New()
	wsID := uuid.New()
	prevCol := uuid.New()
	toCol := uuid.New()
	userID := uuid.New()

	testCases := []struct {
		name       string
		event      events.Event
		ctx        context.Context
		setupRepo  func(m *mocks.MockTaskRepository)
		wantErr    bool
	}{
		{
			name: "successful column moved with actor",
			event: events.TaskMovedEvent{
				TaskID:           taskID,
				WorkspaceIDVal:   wsID,
				ColumnID:         &toCol,
				PreviousColumnID: &prevCol,
				OccurredAtVal:    time.Now().UTC(),
			},
			ctx: context.WithValue(context.Background(), "userID", userID),
			setupRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().CreateActivity(mock.Anything, mock.MatchedBy(func(act *domainTask.TaskActivity) bool {
					return act.TaskID == taskID &&
						act.UserID != nil && *act.UserID == userID &&
						act.ActivityType == domainTask.ActivityColumnMoved &&
						act.Payload["from_column_id"] == prevCol.String() &&
						act.Payload["to_column_id"] == toCol.String()
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "successful column moved without actor (system action)",
			event: events.TaskMovedEvent{
				TaskID:           taskID,
				WorkspaceIDVal:   wsID,
				ColumnID:         &toCol,
				PreviousColumnID: &prevCol,
				OccurredAtVal:    time.Now().UTC(),
			},
			ctx: context.Background(),
			setupRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().CreateActivity(mock.Anything, mock.MatchedBy(func(act *domainTask.TaskActivity) bool {
					return act.TaskID == taskID &&
						act.UserID == nil &&
						act.ActivityType == domainTask.ActivityColumnMoved
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "repository error propagates",
			event: events.TaskMovedEvent{
				TaskID:           taskID,
				WorkspaceIDVal:   wsID,
				ColumnID:         &toCol,
				PreviousColumnID: &prevCol,
				OccurredAtVal:    time.Now().UTC(),
			},
			ctx: context.Background(),
			setupRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().CreateActivity(mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
		{
			name:      "ignored for mismatched event type",
			event:     events.TaskCreatedEvent{},
			ctx:       context.Background(),
			setupRepo: func(m *mocks.MockTaskRepository) {},
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			taskRepo := mocks.NewMockTaskRepository(t)
			tc.setupRepo(taskRepo)

			logger := taskApp.NewActivityLogger(taskRepo)
			err := logger.HandleTaskMoved(tc.ctx, tc.event)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestActivityLogger_HandleTaskCompleted(t *testing.T) {
	t.Parallel()

	taskID := uuid.New()
	wsID := uuid.New()
	userID := uuid.New()

	testCases := []struct {
		name       string
		event      events.Event
		ctx        context.Context
		setupRepo  func(m *mocks.MockTaskRepository)
		wantErr    bool
	}{
		{
			name: "successful task completed logging",
			event: events.TaskCompletedEvent{
				TaskID:         taskID,
				WorkspaceIDVal: wsID,
				CompletedAt:    time.Now().UTC(),
				OccurredAtVal:  time.Now().UTC(),
			},
			ctx: context.WithValue(context.Background(), "userID", userID),
			setupRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().CreateActivity(mock.Anything, mock.MatchedBy(func(act *domainTask.TaskActivity) bool {
					return act.TaskID == taskID &&
						act.UserID != nil && *act.UserID == userID &&
						act.ActivityType == domainTask.ActivityStatusChanged &&
						act.Payload["is_completed"] == true
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "repository error propagates",
			event: events.TaskCompletedEvent{
				TaskID:         taskID,
				WorkspaceIDVal: wsID,
				CompletedAt:    time.Now().UTC(),
				OccurredAtVal:  time.Now().UTC(),
			},
			ctx: context.Background(),
			setupRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().CreateActivity(mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
		{
			name:      "ignored for mismatched event type",
			event:     events.TaskCreatedEvent{},
			ctx:       context.Background(),
			setupRepo: func(m *mocks.MockTaskRepository) {},
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			taskRepo := mocks.NewMockTaskRepository(t)
			tc.setupRepo(taskRepo)

			logger := taskApp.NewActivityLogger(taskRepo)
			err := logger.HandleTaskCompleted(tc.ctx, tc.event)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestActivityLogger_HandleTaskUpdated(t *testing.T) {
	t.Parallel()

	taskID := uuid.New()
	wsID := uuid.New()
	prevCol := uuid.New()
	toCol := uuid.New()

	testCases := []struct {
		name      string
		event     events.Event
		ctx       context.Context
		setupRepo func(m *mocks.MockTaskRepository)
	}{
		{
			name: "column changed logs column_moved",
			event: events.TaskUpdatedEvent{
				TaskID:            taskID,
				WorkspaceIDVal:    wsID,
				ColumnID:          &toCol,
				PreviousColumnID:  &prevCol,
				OccurredAtVal:     time.Now().UTC(),
			},
			ctx: context.Background(),
			setupRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().CreateActivity(mock.Anything, mock.MatchedBy(func(act *domainTask.TaskActivity) bool {
					return act.TaskID == taskID &&
						act.ActivityType == domainTask.ActivityColumnMoved &&
						act.Payload["from_column_id"] == prevCol.String() &&
						act.Payload["to_column_id"] == toCol.String()
				})).Return(nil).Once()
			},
		},
		{
			name: "column unchanged does not log activity",
			event: events.TaskUpdatedEvent{
				TaskID:            taskID,
				WorkspaceIDVal:    wsID,
				ColumnID:          &toCol,
				PreviousColumnID:  &toCol,
				OccurredAtVal:     time.Now().UTC(),
			},
			ctx:       context.Background(),
			setupRepo: func(m *mocks.MockTaskRepository) {},
		},
		{
			name:      "ignored for mismatched event type",
			event:     events.TaskCreatedEvent{},
			ctx:       context.Background(),
			setupRepo: func(m *mocks.MockTaskRepository) {},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			taskRepo := mocks.NewMockTaskRepository(t)
			tc.setupRepo(taskRepo)

			logger := taskApp.NewActivityLogger(taskRepo)
			err := logger.HandleTaskUpdated(tc.ctx, tc.event)
			require.NoError(t, err)
		})
	}
}
