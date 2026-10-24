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
	domainTask "takt/internal/domain/task"
)

func TestService_ListTaskActivities(t *testing.T) {
	t.Parallel()

	wsID := uuid.New()
	taskID := uuid.New()
	userID := uuid.New()
	userName := "Alice"

	sampleTask, _ := domainTask.NewInboxTask(wsID, "My Task", "", domainTask.PriorityNone, nil, 0)
	sampleTask.ID = taskID

	sampleActivities := []*domainTask.TaskActivity{
		{
			ID:           uuid.New(),
			TaskID:       taskID,
			UserID:       &userID,
			UserName:     &userName,
			ActivityType: domainTask.ActivityColumnMoved,
			Payload:      map[string]interface{}{"from_column_id": uuid.New().String(), "to_column_id": uuid.New().String()},
			CreatedAt:    time.Now().UTC(),
		},
		{
			ID:           uuid.New(),
			TaskID:       taskID,
			UserID:       &userID,
			UserName:     &userName,
			ActivityType: domainTask.ActivityStatusChanged,
			Payload:      map[string]interface{}{"is_completed": true},
			CreatedAt:    time.Now().UTC().Add(time.Minute),
		},
	}

	testCases := []struct {
		name          string
		workspaceID   uuid.UUID
		taskID        uuid.UUID
		setupAccess   func(m *mocks.MockWorkspaceAccessChecker)
		setupTaskRepo func(m *mocks.MockTaskRepository)
		wantCount     int
		wantErr       error
	}{
		{
			name:        "успешное получение списка активностей задачи",
			workspaceID: wsID,
			taskID:      taskID,
			setupAccess: func(m *mocks.MockWorkspaceAccessChecker) {
				m.EXPECT().CheckAccess(mock.Anything, wsID, userID).Return(nil).Once()
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(sampleTask, nil).Once()
				m.EXPECT().ListActivitiesByTaskID(mock.Anything, taskID).Return(sampleActivities, nil).Once()
			},
			wantCount: 2,
			wantErr:   nil,
		},
		{
			name:        "ошибка доступа к воркспейсу",
			workspaceID: wsID,
			taskID:      taskID,
			setupAccess: func(m *mocks.MockWorkspaceAccessChecker) {
				m.EXPECT().CheckAccess(mock.Anything, wsID, userID).Return(errors.New("access denied")).Once()
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {},
			wantCount:     0,
			wantErr:       errors.New("access denied"),
		},
		{
			name:        "задача не найдена в воркспейсе (изоляция тенанта)",
			workspaceID: wsID,
			taskID:      taskID,
			setupAccess: func(m *mocks.MockWorkspaceAccessChecker) {
				m.EXPECT().CheckAccess(mock.Anything, wsID, userID).Return(nil).Once()
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(nil, domainTask.ErrTaskNotFound).Once()
			},
			wantCount: 0,
			wantErr:   domainTask.ErrTaskNotFound,
		},
		{
			name:        "ошибка БД при выборке активностей",
			workspaceID: wsID,
			taskID:      taskID,
			setupAccess: func(m *mocks.MockWorkspaceAccessChecker) {
				m.EXPECT().CheckAccess(mock.Anything, wsID, userID).Return(nil).Once()
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(sampleTask, nil).Once()
				m.EXPECT().ListActivitiesByTaskID(mock.Anything, taskID).Return(nil, errors.New("query failed")).Once()
			},
			wantCount: 0,
			wantErr:   errors.New("query failed"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			taskRepo := mocks.NewMockTaskRepository(t)
			tagRepo := mocks.NewMockTagRepository(t)
			accessChecker := mocks.NewMockWorkspaceAccessChecker(t)

			if tc.setupAccess != nil {
				tc.setupAccess(accessChecker)
			}
			if tc.setupTaskRepo != nil {
				tc.setupTaskRepo(taskRepo)
			}

			svc := taskApp.NewService(taskRepo, tagRepo, accessChecker)
			ctx := context.WithValue(context.Background(), "userID", userID)

			res, err := svc.ListTaskActivities(ctx, tc.workspaceID, tc.taskID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr.Error())
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				assert.Len(t, res, tc.wantCount)
				assert.Equal(t, sampleActivities[0].ID, res[0].ID)
				assert.Equal(t, sampleActivities[1].ID, res[1].ID)
			}
		})
	}
}
