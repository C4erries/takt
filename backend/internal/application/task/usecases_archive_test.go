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

	taskapp "takt/internal/application/task"
	"takt/internal/application/task/mocks"
	"takt/internal/domain/events"
	domainTask "takt/internal/domain/task"
)

func TestService_ListArchivedTasks_Table(t *testing.T) {
	t.Parallel()

	wsID := uuid.New()
	projID := uuid.New()
	now := time.Now().UTC()

	archivedTask := &domainTask.Task{
		ID:          uuid.New(),
		WorkspaceID: wsID,
		ProjectID:   &projID,
		Title:       "Archived Task",
		IsArchived:  true,
		ArchivedAt:  &now,
	}

	tests := []struct {
		name       string
		projectID  *uuid.UUID
		setupMocks func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository)
		wantCount  int
		wantErr    bool
	}{
		{
			name:      "success with project id filter",
			projectID: &projID,
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository) {
				repo := mocks.NewMockTaskRepository(t)
				repo.EXPECT().ListArchivedTasks(mock.Anything, wsID, &projID).
					Return([]*domainTask.Task{archivedTask}, nil).Once()
				svc := taskapp.NewService(repo, nil, nil, nil)
				return svc, repo
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "success without project id filter (all archived)",
			projectID: nil,
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository) {
				repo := mocks.NewMockTaskRepository(t)
				repo.EXPECT().ListArchivedTasks(mock.Anything, wsID, (*uuid.UUID)(nil)).
					Return([]*domainTask.Task{archivedTask}, nil).Once()
				svc := taskapp.NewService(repo, nil, nil, nil)
				return svc, repo
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "repository returns error",
			projectID: &projID,
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository) {
				repo := mocks.NewMockTaskRepository(t)
				repo.EXPECT().ListArchivedTasks(mock.Anything, wsID, &projID).
					Return(nil, errors.New("db error")).Once()
				svc := taskapp.NewService(repo, nil, nil, nil)
				return svc, repo
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := tc.setupMocks(t)
			tasks, err := svc.ListArchivedTasks(context.Background(), wsID, tc.projectID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, tasks)
			} else {
				require.NoError(t, err)
				assert.Len(t, tasks, tc.wantCount)
			}
		})
	}
}

func TestService_ArchiveCompletedTasksInColumn_Table(t *testing.T) {
	t.Parallel()

	wsID := uuid.New()
	colID := uuid.New()

	tests := []struct {
		name         string
		setupMocks   func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockEventPublisher)
		wantAffected int64
		wantErr      bool
	}{
		{
			name: "success archives tasks and publishes batch sweep event",
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockEventPublisher) {
				repo := mocks.NewMockTaskRepository(t)
				publisher := mocks.NewMockEventPublisher(t)

				repo.EXPECT().ArchiveCompletedTasksInColumn(mock.Anything, colID).
					Return(int64(4), nil).Once()

				publisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(e events.Event) bool {
					ev, ok := e.(events.BatchSweepCompletedEvent)
					return ok && ev.WorkspaceID() == wsID && ev.ColumnID() == colID && ev.AffectedCount == 4
				})).Return(nil).Once()

				svc := taskapp.NewService(repo, nil, nil, publisher)
				return svc, repo, publisher
			},
			wantAffected: 4,
			wantErr:      false,
		},
		{
			name: "success when no tasks affected does not publish event",
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockEventPublisher) {
				repo := mocks.NewMockTaskRepository(t)
				publisher := mocks.NewMockEventPublisher(t)

				repo.EXPECT().ArchiveCompletedTasksInColumn(mock.Anything, colID).
					Return(int64(0), nil).Once()

				svc := taskapp.NewService(repo, nil, nil, publisher)
				return svc, repo, publisher
			},
			wantAffected: 0,
			wantErr:      false,
		},
		{
			name: "repository error propagates",
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockEventPublisher) {
				repo := mocks.NewMockTaskRepository(t)
				publisher := mocks.NewMockEventPublisher(t)

				repo.EXPECT().ArchiveCompletedTasksInColumn(mock.Anything, colID).
					Return(int64(0), errors.New("db failure")).Once()

				svc := taskapp.NewService(repo, nil, nil, publisher)
				return svc, repo, publisher
			},
			wantAffected: 0,
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc, _, _ := tc.setupMocks(t)
			affected, err := svc.ArchiveCompletedTasksInColumn(context.Background(), wsID, colID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantAffected, affected)
			}
		})
	}
}

func TestService_UnarchiveTask_Table(t *testing.T) {
	t.Parallel()

	wsID := uuid.New()
	taskID := uuid.New()
	colID := uuid.New()
	boardID := uuid.New()
	now := time.Now().UTC()

	unarchivedTask := &domainTask.Task{
		ID:          taskID,
		WorkspaceID: wsID,
		ColumnID:    &colID,
		Title:       "Restored Task",
		IsArchived:  false,
		ArchivedAt:  nil,
		Position:    1000.0,
		UpdatedAt:   now,
	}

	tests := []struct {
		name       string
		columnID   uuid.UUID
		boardID    *uuid.UUID
		setupMocks func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockEventPublisher)
		wantErr    error
	}{
		{
			name:     "validation error on nil column id",
			columnID: uuid.Nil,
			boardID:  &boardID,
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockEventPublisher) {
				repo := mocks.NewMockTaskRepository(t)
				svc := taskapp.NewService(repo, nil, nil, nil)
				return svc, repo, nil
			},
			wantErr: domainTask.ErrInvalidColumnID,
		},
		{
			name:     "success unarchives task and restores state",
			columnID: colID,
			boardID:  &boardID,
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockEventPublisher) {
				repo := mocks.NewMockTaskRepository(t)
				publisher := mocks.NewMockEventPublisher(t)

				archivedInRepo := &domainTask.Task{
					ID: taskID, WorkspaceID: wsID, Position: 1000,
				}
				repo.EXPECT().GetByID(mock.Anything, taskID, wsID).
					Return(archivedInRepo, nil).Once()

				repo.EXPECT().UnarchiveTask(mock.Anything, taskID, wsID, colID, &boardID, 1000.0).
					Return(unarchivedTask, nil).Once()

				publisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(e events.Event) bool {
					ev, ok := e.(events.TaskUpdatedEvent)
					return ok && ev.WorkspaceID() == wsID && ev.TaskID == taskID
				})).Return(nil).Once()

				svc := taskapp.NewService(repo, nil, nil, publisher)
				return svc, repo, publisher
			},
			wantErr: nil,
		},
		{
			name:     "repo returns error when task not found",
			columnID: colID,
			boardID:  nil,
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockEventPublisher) {
				repo := mocks.NewMockTaskRepository(t)

				repo.EXPECT().GetByID(mock.Anything, taskID, wsID).
					Return(nil, domainTask.ErrTaskNotFound).Once()

				svc := taskapp.NewService(repo, nil, nil, nil)
				return svc, repo, nil
			},
			wantErr: domainTask.ErrTaskNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc, _, _ := tc.setupMocks(t)
			task, err := svc.UnarchiveTask(context.Background(), wsID, taskID, tc.columnID, tc.boardID)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				assert.Equal(t, taskID, task.ID)
				assert.False(t, task.IsArchived)
			}
		})
	}
}
