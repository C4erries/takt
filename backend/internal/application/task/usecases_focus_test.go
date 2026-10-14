package task_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	taskapp "takt/internal/application/task"
	"takt/internal/application/task/mocks"
	domainTask "takt/internal/domain/task"
)

func TestService_GetTodayTasks(t *testing.T) {
	wsID := uuid.New()
	task1, err := domainTask.NewInboxTask(wsID, "Today Task 1", "", domainTask.PriorityHigh, nil, 0)
	require.NoError(t, err)
	task2, err := domainTask.NewInboxTask(wsID, "Today Task 2", "", domainTask.PriorityMedium, nil, 1)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		workspaceID   uuid.UUID
		setupTaskRepo func(m *mocks.MockTaskRepository)
		setupTagRepo  func(m *mocks.MockTagRepository)
		checkOut      func(t *testing.T, tasks []*domainTask.Task)
		wantErr       error
	}{
		{
			name:        "успешное получение списка задач Today с тегами",
			workspaceID: wsID,
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetTodayTasks(mock.Anything, wsID).Return([]*domainTask.Task{task1, task2}, nil).Once()
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().GetTaskTags(mock.Anything, task1.ID).Return([]*domainTask.Tag{}, nil).Once()
				m.EXPECT().GetTaskTags(mock.Anything, task2.ID).Return([]*domainTask.Tag{}, nil).Once()
			},
			checkOut: func(t *testing.T, tasks []*domainTask.Task) {
				require.Len(t, tasks, 2)
				assert.Equal(t, "Today Task 1", tasks[0].Title)
				assert.Equal(t, "Today Task 2", tasks[1].Title)
			},
		},
		{
			name:        "ошибка репозитория при получении Today задач",
			workspaceID: wsID,
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetTodayTasks(mock.Anything, wsID).Return(nil, errTestDB).Once()
			},
			wantErr: errTestDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			taskRepo := mocks.NewMockTaskRepository(t)
			tagRepo := mocks.NewMockTagRepository(t)
			if tc.setupTaskRepo != nil {
				tc.setupTaskRepo(taskRepo)
			}
			if tc.setupTagRepo != nil {
				tc.setupTagRepo(tagRepo)
			}

			svc := taskapp.NewService(taskRepo, tagRepo, nil)
			tasks, err := svc.GetTodayTasks(context.Background(), tc.workspaceID)

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

func TestService_UpdateTask(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	newTitle := "Updated Title"
	emptyTitle := "   "
	newDesc := "Updated Description"
	now := time.Now().UTC()
	priority := domainTask.PriorityUrgent
	invalidPriority := domainTask.Priority("INVALID")
	focusOrder := 1
	invalidFocusOrder := -1

	testCases := []struct {
		name          string
		input         taskapp.UpdateTaskInput
		setupTaskRepo func(m *mocks.MockTaskRepository)
		setupTagRepo  func(m *mocks.MockTagRepository)
		checkOut      func(t *testing.T, task *domainTask.Task)
		wantErr       error
	}{
		{
			name: "успешное обновление полей задачи",
			input: taskapp.UpdateTaskInput{
				TaskID:              taskID,
				WorkspaceID:         wsID,
				Title:               &newTitle,
				DescriptionMarkdown: &newDesc,
				DueDate:             &now,
				Priority:            &priority,
				FocusOrder:          &focusOrder,
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				existing, _ := domainTask.NewInboxTask(wsID, "Original Title", "", domainTask.PriorityNone, nil, 0)
				existing.ID = taskID
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(existing, nil).Once()
				m.EXPECT().Update(mock.Anything, mock.MatchedBy(func(t *domainTask.Task) bool {
					return t.ID == taskID && t.Title == newTitle && t.DescriptionMarkdown == newDesc &&
						t.Priority == priority && t.FocusOrder != nil && *t.FocusOrder == 1
				})).Return(nil).Once()
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().GetTaskTags(mock.Anything, taskID).Return([]*domainTask.Tag{}, nil).Once()
			},
			checkOut: func(t *testing.T, task *domainTask.Task) {
				assert.Equal(t, newTitle, task.Title)
				assert.Equal(t, newDesc, task.DescriptionMarkdown)
				assert.Equal(t, priority, task.Priority)
				require.NotNil(t, task.FocusOrder)
				assert.Equal(t, 1, *task.FocusOrder)
			},
		},
		{
			name: "ошибка: пустой заголовок при обновлении",
			input: taskapp.UpdateTaskInput{
				TaskID:      taskID,
				WorkspaceID: wsID,
				Title:       &emptyTitle,
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				existing, _ := domainTask.NewInboxTask(wsID, "Original Title", "", domainTask.PriorityNone, nil, 0)
				existing.ID = taskID
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(existing, nil).Once()
			},
			wantErr: domainTask.ErrInvalidTaskTitle,
		},
		{
			name: "ошибка: некорректный приоритет",
			input: taskapp.UpdateTaskInput{
				TaskID:      taskID,
				WorkspaceID: wsID,
				Priority:    &invalidPriority,
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				existing, _ := domainTask.NewInboxTask(wsID, "Original Title", "", domainTask.PriorityNone, nil, 0)
				existing.ID = taskID
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(existing, nil).Once()
			},
			wantErr: domainTask.ErrInvalidPriority,
		},
		{
			name: "ошибка: некорректный фокус-ордер (>3)",
			input: taskapp.UpdateTaskInput{
				TaskID:      taskID,
				WorkspaceID: wsID,
				FocusOrder:  &invalidFocusOrder,
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				existing, _ := domainTask.NewInboxTask(wsID, "Original Title", "", domainTask.PriorityNone, nil, 0)
				existing.ID = taskID
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(existing, nil).Once()
			},
			wantErr: domainTask.ErrInvalidFocusOrder,
		},
		{
			name: "ошибка: задача не найдена",
			input: taskapp.UpdateTaskInput{
				TaskID:      taskID,
				WorkspaceID: wsID,
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(nil, domainTask.ErrTaskNotFound).Once()
			},
			wantErr: domainTask.ErrTaskNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			taskRepo := mocks.NewMockTaskRepository(t)
			tagRepo := mocks.NewMockTagRepository(t)
			if tc.setupTaskRepo != nil {
				tc.setupTaskRepo(taskRepo)
			}
			if tc.setupTagRepo != nil {
				tc.setupTagRepo(tagRepo)
			}

			svc := taskapp.NewService(taskRepo, tagRepo, nil)
			task, err := svc.UpdateTask(context.Background(), tc.input)

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

func TestService_SetTaskFocus(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	validOrder := 2
	invalidOrder := -1

	testCases := []struct {
		name          string
		input         taskapp.SetFocusInput
		setupTaskRepo func(m *mocks.MockTaskRepository)
		setupTagRepo  func(m *mocks.MockTagRepository)
		checkOut      func(t *testing.T, task *domainTask.Task)
		wantErr       error
	}{
		{
			name: "успешная установка фокуса",
			input: taskapp.SetFocusInput{
				TaskID:      taskID,
				WorkspaceID: wsID,
				FocusOrder:  &validOrder,
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				existing, _ := domainTask.NewInboxTask(wsID, "Task Focus", "", domainTask.PriorityNone, nil, 0)
				existing.ID = taskID
				updatedTask := *existing
				updatedTask.FocusOrder = &validOrder

				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(existing, nil).Once()
				m.EXPECT().SetFocusOrder(mock.Anything, taskID, wsID, &validOrder).Return(&updatedTask, nil).Once()
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().GetTaskTags(mock.Anything, taskID).Return([]*domainTask.Tag{}, nil).Once()
			},
			checkOut: func(t *testing.T, task *domainTask.Task) {
				require.NotNil(t, task.FocusOrder)
				assert.Equal(t, 2, *task.FocusOrder)
			},
		},
		{
			name: "ошибка: некорректный focus_order (>3)",
			input: taskapp.SetFocusInput{
				TaskID:      taskID,
				WorkspaceID: wsID,
				FocusOrder:  &invalidOrder,
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				existing, _ := domainTask.NewInboxTask(wsID, "Task Focus", "", domainTask.PriorityNone, nil, 0)
				existing.ID = taskID
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(existing, nil).Once()
			},
			wantErr: domainTask.ErrInvalidFocusOrder,
		},
		{
			name: "ошибка: задача не найдена при установке фокуса",
			input: taskapp.SetFocusInput{
				TaskID:      taskID,
				WorkspaceID: wsID,
				FocusOrder:  &validOrder,
			},
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(nil, domainTask.ErrTaskNotFound).Once()
			},
			wantErr: domainTask.ErrTaskNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			taskRepo := mocks.NewMockTaskRepository(t)
			tagRepo := mocks.NewMockTagRepository(t)
			if tc.setupTaskRepo != nil {
				tc.setupTaskRepo(taskRepo)
			}
			if tc.setupTagRepo != nil {
				tc.setupTagRepo(tagRepo)
			}

			svc := taskapp.NewService(taskRepo, tagRepo, nil)
			task, err := svc.SetTaskFocus(context.Background(), tc.input)

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
