package task_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	taskapp "takt/internal/application/task"
	"takt/internal/application/task/mocks"
	domainTask "takt/internal/domain/task"
)

var errTestDB = errors.New("database connection failed")

func TestService_CreateInboxTask(t *testing.T) {
	wsID := uuid.New()

	testCases := []struct {
		name      string
		input     taskapp.CreateInboxTaskInput
		setupMock func(m *mocks.MockTaskRepository)
		checkOut  func(t *testing.T, task *domainTask.Task)
		wantErr   error
	}{
		{
			name: "успешное создание первой задачи (позиция 0)",
			input: taskapp.CreateInboxTaskInput{
				WorkspaceID: wsID, Title: "Task 1", DescriptionMarkdown: "Desc",
				Priority: domainTask.PriorityHigh,
			},
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetMaxInboxPosition(mock.Anything, wsID).Return(-1, nil).Once()
				m.EXPECT().Create(mock.Anything, mock.MatchedBy(func(t *domainTask.Task) bool {
					return t.WorkspaceID == wsID && t.Title == "Task 1" && t.Position == 0 && t.Priority == domainTask.PriorityHigh
				})).Return(nil).Once()
			},
			checkOut: func(t *testing.T, task *domainTask.Task) {
				assert.Equal(t, 0, task.Position)
				assert.Equal(t, "Task 1", task.Title)
				assert.Equal(t, domainTask.PriorityHigh, task.Priority)
				assert.False(t, task.IsCompleted)
			},
		},
		{
			name: "успешное создание второй задачи (позиция 1)",
			input: taskapp.CreateInboxTaskInput{
				WorkspaceID: wsID, Title: "Task 2", Priority: domainTask.PriorityLow,
			},
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetMaxInboxPosition(mock.Anything, wsID).Return(0, nil).Once()
				m.EXPECT().Create(mock.Anything, mock.MatchedBy(func(t *domainTask.Task) bool {
					return t.WorkspaceID == wsID && t.Title == "Task 2" && t.Position == 1
				})).Return(nil).Once()
			},
			checkOut: func(t *testing.T, task *domainTask.Task) {
				assert.Equal(t, 1, task.Position)
				assert.Equal(t, "Task 2", task.Title)
			},
		},
		{
			name: "ошибка валидации: пустой заголовок задачи",
			input: taskapp.CreateInboxTaskInput{
				WorkspaceID: wsID, Title: "", Priority: domainTask.PriorityMedium,
			},
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetMaxInboxPosition(mock.Anything, wsID).Return(-1, nil).Once()
			},
			wantErr: domainTask.ErrInvalidTaskTitle,
		},
		{
			name: "ошибка репозитория при сохранении задачи",
			input: taskapp.CreateInboxTaskInput{
				WorkspaceID: wsID, Title: "Failing Task", Priority: domainTask.PriorityNone,
			},
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetMaxInboxPosition(mock.Anything, wsID).Return(-1, nil).Once()
				m.EXPECT().Create(mock.Anything, mock.Anything).Return(errTestDB).Once()
			},
			wantErr: errTestDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := mocks.NewMockTaskRepository(t)
			tc.setupMock(repo)

			svc := taskapp.NewService(repo, nil)
			task, err := svc.CreateInboxTask(context.Background(), tc.input)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				if tc.checkOut != nil {
					tc.checkOut(t, task)
				}
			}
		})
	}
}

func TestService_ToggleTask(t *testing.T) {
	wsID := uuid.New()
	wrongWsID := uuid.New()
	taskID := uuid.New()

	completedTask, err := domainTask.NewInboxTask(wsID, "Task to toggle", "", domainTask.PriorityNone, nil, 0)
	require.NoError(t, err)
	completedTask.Toggle()

	testCases := []struct {
		name        string
		taskID      uuid.UUID
		workspaceID uuid.UUID
		setupMock   func(m *mocks.MockTaskRepository)
		checkOut    func(t *testing.T, task *domainTask.Task)
		wantErr     error
	}{
		{
			name:        "успешное переключение статуса задачи",
			taskID:      taskID,
			workspaceID: wsID,
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().Toggle(mock.Anything, taskID, wsID).Return(completedTask, nil).Once()
			},
			checkOut: func(t *testing.T, task *domainTask.Task) {
				assert.True(t, task.IsCompleted)
			},
		},
		{
			name:        "задача не найдена в текущем воркспейсе (tenant isolation)",
			taskID:      taskID,
			workspaceID: wrongWsID,
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().Toggle(mock.Anything, taskID, wrongWsID).Return(nil, domainTask.ErrTaskNotFound).Once()
			},
			wantErr: domainTask.ErrTaskNotFound,
		},
		{
			name:        "ошибка репозитория при переключении статуса",
			taskID:      taskID,
			workspaceID: wsID,
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().Toggle(mock.Anything, taskID, wsID).Return(nil, errTestDB).Once()
			},
			wantErr: errTestDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := mocks.NewMockTaskRepository(t)
			tc.setupMock(repo)

			svc := taskapp.NewService(repo, nil)
			task, err := svc.ToggleTask(context.Background(), tc.taskID, tc.workspaceID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				if tc.checkOut != nil {
					tc.checkOut(t, task)
				}
			}
		})
	}
}

func TestService_DeleteTask(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()

	testCases := []struct {
		name        string
		taskID      uuid.UUID
		workspaceID uuid.UUID
		setupMock   func(m *mocks.MockTaskRepository)
		wantErr     error
	}{
		{
			name:        "успешное удаление задачи",
			taskID:      taskID,
			workspaceID: wsID,
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().Delete(mock.Anything, taskID, wsID).Return(nil).Once()
			},
		},
		{
			name:        "задача не найдена при удалении",
			taskID:      taskID,
			workspaceID: wsID,
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().Delete(mock.Anything, taskID, wsID).Return(domainTask.ErrTaskNotFound).Once()
			},
			wantErr: domainTask.ErrTaskNotFound,
		},
		{
			name:        "ошибка репозитория при удалении задачи",
			taskID:      taskID,
			workspaceID: wsID,
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().Delete(mock.Anything, taskID, wsID).Return(errTestDB).Once()
			},
			wantErr: errTestDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := mocks.NewMockTaskRepository(t)
			tc.setupMock(repo)

			svc := taskapp.NewService(repo, nil)
			err := svc.DeleteTask(context.Background(), tc.taskID, tc.workspaceID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_CreateInitialInboxTasks(t *testing.T) {
	wsID := uuid.New()

	testCases := []struct {
		name        string
		workspaceID uuid.UUID
		setupMock   func(m *mocks.MockTaskRepository)
		wantErr     error
	}{
		{
			name:        "успешное создание трех стартовых задач онбординга",
			workspaceID: wsID,
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().Create(mock.Anything, mock.MatchedBy(func(t *domainTask.Task) bool {
					return t.WorkspaceID == wsID && t.Position == 0
				})).Return(nil).Once()
				m.EXPECT().Create(mock.Anything, mock.MatchedBy(func(t *domainTask.Task) bool {
					return t.WorkspaceID == wsID && t.Position == 1
				})).Return(nil).Once()
				m.EXPECT().Create(mock.Anything, mock.MatchedBy(func(t *domainTask.Task) bool {
					return t.WorkspaceID == wsID && t.Position == 2
				})).Return(nil).Once()
			},
		},
		{
			name:        "ошибка репозитория при сохранении первой онбординг-задачи",
			workspaceID: wsID,
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().Create(mock.Anything, mock.Anything).Return(errTestDB).Once()
			},
			wantErr: errTestDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := mocks.NewMockTaskRepository(t)
			tc.setupMock(repo)

			svc := taskapp.NewService(repo, nil)
			err := svc.CreateInitialInboxTasks(context.Background(), tc.workspaceID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_GetInboxTasks(t *testing.T) {
	wsID := uuid.New()
	task1, err := domainTask.NewInboxTask(wsID, "Task 1", "", domainTask.PriorityNone, nil, 0)
	require.NoError(t, err)
	task2, err := domainTask.NewInboxTask(wsID, "Task 2", "", domainTask.PriorityHigh, nil, 1)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		workspaceID uuid.UUID
		setupMock   func(m *mocks.MockTaskRepository)
		checkOut    func(t *testing.T, tasks []*domainTask.Task)
		wantErr     error
	}{
		{
			name:        "успешное получение списка задач Inbox",
			workspaceID: wsID,
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetInboxTasks(mock.Anything, wsID).Return([]*domainTask.Task{task1, task2}, nil).Once()
			},
			checkOut: func(t *testing.T, tasks []*domainTask.Task) {
				require.Len(t, tasks, 2)
				assert.Equal(t, "Task 1", tasks[0].Title)
				assert.Equal(t, "Task 2", tasks[1].Title)
			},
		},
		{
			name:        "ошибка репозитория при получении задач",
			workspaceID: wsID,
			setupMock: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetInboxTasks(mock.Anything, wsID).Return(nil, errTestDB).Once()
			},
			wantErr: errTestDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := mocks.NewMockTaskRepository(t)
			tc.setupMock(repo)

			svc := taskapp.NewService(repo, nil)
			tasks, err := svc.GetInboxTasks(context.Background(), tc.workspaceID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, tasks)
			} else {
				require.NoError(t, err)
				if tc.checkOut != nil {
					tc.checkOut(t, tasks)
				}
			}
		})
	}
}
