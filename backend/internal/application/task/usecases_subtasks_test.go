package task_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	taskapp "takt/internal/application/task"
	"takt/internal/application/task/mocks"
	domainTask "takt/internal/domain/task"
)

func TestService_CreateInboxTask_Subtask(t *testing.T) {
	wsID := uuid.New()
	parentID := uuid.New()
	projectID := uuid.New()
	columnID := uuid.New()

	t.Run("success creating subtask inherits project and column from parent", func(t *testing.T) {
		mockRepo := mocks.NewMockTaskRepository(t)
		mockTagRepo := mocks.NewMockTagRepository(t)
		mockEvents := mocks.NewMockEventPublisher(t)

		parentTask := &domainTask.Task{
			ID:          parentID,
			WorkspaceID: wsID,
			ProjectID:   &projectID,
			ColumnID:    &columnID,
			Title:       "Parent Task",
		}

		mockRepo.EXPECT().GetByID(mock.Anything, parentID, wsID).Return(parentTask, nil).Once()
		mockRepo.EXPECT().GetMaxInboxPosition(mock.Anything, wsID).Return(2, nil).Once()
		mockRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(task *domainTask.Task) bool {
			return task.WorkspaceID == wsID &&
				task.ParentID != nil && *task.ParentID == parentID &&
				task.ProjectID != nil && *task.ProjectID == projectID &&
				task.ColumnID != nil && *task.ColumnID == columnID &&
				task.Title == "Subtask 1"
		})).Return(nil).Once()

		mockEvents.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(evt any) bool {
			return true
		})).Return(nil).Once()

		svc := taskapp.NewService(mockRepo, mockTagRepo, nil, mockEvents)

		created, err := svc.CreateInboxTask(context.Background(), taskapp.CreateInboxTaskInput{
			WorkspaceID: wsID,
			Title:       "Subtask 1",
			ParentID:    &parentID,
		})

		require.NoError(t, err)
		assert.Equal(t, &parentID, created.ParentID)
		assert.Equal(t, &projectID, created.ProjectID)
		assert.Equal(t, &columnID, created.ColumnID)
	})

	t.Run("error when parent task not found", func(t *testing.T) {
		mockRepo := mocks.NewMockTaskRepository(t)
		mockRepo.EXPECT().GetByID(mock.Anything, parentID, wsID).Return(nil, domainTask.ErrTaskNotFound).Once()

		svc := taskapp.NewService(mockRepo, nil, nil)

		_, err := svc.CreateInboxTask(context.Background(), taskapp.CreateInboxTaskInput{
			WorkspaceID: wsID,
			Title:       "Subtask 1",
			ParentID:    &parentID,
		})

		require.ErrorIs(t, err, domainTask.ErrTaskNotFound)
	})
}

func TestService_ListSubtasks(t *testing.T) {
	wsID := uuid.New()
	parentID := uuid.New()

	t.Run("success listing subtasks", func(t *testing.T) {
		mockRepo := mocks.NewMockTaskRepository(t)
		mockTagRepo := mocks.NewMockTagRepository(t)

		parentTask := &domainTask.Task{
			ID:          parentID,
			WorkspaceID: wsID,
			Title:       "Parent",
		}
		subtask1 := &domainTask.Task{
			ID:          uuid.New(),
			WorkspaceID: wsID,
			ParentID:    &parentID,
			Title:       "Sub 1",
		}
		subtask2 := &domainTask.Task{
			ID:          uuid.New(),
			WorkspaceID: wsID,
			ParentID:    &parentID,
			Title:       "Sub 2",
		}

		mockRepo.EXPECT().GetByID(mock.Anything, parentID, wsID).Return(parentTask, nil).Once()
		mockRepo.EXPECT().ListSubtasks(mock.Anything, parentID).Return([]*domainTask.Task{subtask1, subtask2}, nil).Once()
		mockTagRepo.EXPECT().GetTaskTags(mock.Anything, subtask1.ID).Return(nil, nil).Once()
		mockTagRepo.EXPECT().GetTaskTags(mock.Anything, subtask2.ID).Return(nil, nil).Once()

		svc := taskapp.NewService(mockRepo, mockTagRepo, nil)

		subtasks, err := svc.ListSubtasks(context.Background(), wsID, parentID)
		require.NoError(t, err)
		assert.Len(t, subtasks, 2)
		assert.Equal(t, "Sub 1", subtasks[0].Title)
		assert.Equal(t, "Sub 2", subtasks[1].Title)
	})

	t.Run("parent not found returns error", func(t *testing.T) {
		mockRepo := mocks.NewMockTaskRepository(t)
		mockRepo.EXPECT().GetByID(mock.Anything, parentID, wsID).Return(nil, domainTask.ErrTaskNotFound).Once()

		svc := taskapp.NewService(mockRepo, nil, nil)

		_, err := svc.ListSubtasks(context.Background(), wsID, parentID)
		require.ErrorIs(t, err, domainTask.ErrTaskNotFound)
	})
}

func TestService_UpdateTask_Subtask(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	parentID := uuid.New()
	projectID := uuid.New()

	t.Run("setting parent on existing task inherits project", func(t *testing.T) {
		mockRepo := mocks.NewMockTaskRepository(t)
		mockTagRepo := mocks.NewMockTagRepository(t)

		existingTask := &domainTask.Task{
			ID:          taskID,
			WorkspaceID: wsID,
			Title:       "Task",
		}
		parentTask := &domainTask.Task{
			ID:          parentID,
			WorkspaceID: wsID,
			ProjectID:   &projectID,
			Title:       "Parent",
		}

		mockRepo.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(existingTask, nil).Once()
		mockRepo.EXPECT().GetByID(mock.Anything, parentID, wsID).Return(parentTask, nil).Once()
		mockRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(task *domainTask.Task) bool {
			return task.ID == taskID && task.ParentID != nil && *task.ParentID == parentID &&
				task.ProjectID != nil && *task.ProjectID == projectID
		})).Return(nil).Once()
		mockTagRepo.EXPECT().GetTaskTags(mock.Anything, taskID).Return(nil, nil).Once()

		svc := taskapp.NewService(mockRepo, mockTagRepo, nil)

		updated, err := svc.UpdateTask(context.Background(), taskapp.UpdateTaskInput{
			TaskID:      taskID,
			WorkspaceID: wsID,
			ParentID:    &parentID,
		})

		require.NoError(t, err)
		assert.Equal(t, &parentID, updated.ParentID)
		assert.Equal(t, &projectID, updated.ProjectID)
	})

	t.Run("clearing parent", func(t *testing.T) {
		mockRepo := mocks.NewMockTaskRepository(t)
		mockTagRepo := mocks.NewMockTagRepository(t)

		existingTask := &domainTask.Task{
			ID:          taskID,
			WorkspaceID: wsID,
			ParentID:    &parentID,
			Title:       "Task",
		}

		mockRepo.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(existingTask, nil).Once()
		mockRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(task *domainTask.Task) bool {
			return task.ID == taskID && task.ParentID == nil
		})).Return(nil).Once()
		mockTagRepo.EXPECT().GetTaskTags(mock.Anything, taskID).Return(nil, nil).Once()

		svc := taskapp.NewService(mockRepo, mockTagRepo, nil)

		updated, err := svc.UpdateTask(context.Background(), taskapp.UpdateTaskInput{
			TaskID:        taskID,
			WorkspaceID:   wsID,
			ClearParentID: true,
		})

		require.NoError(t, err)
		assert.Nil(t, updated.ParentID)
	})
}
