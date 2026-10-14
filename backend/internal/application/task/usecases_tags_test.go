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

func TestService_CreateTag(t *testing.T) {
	wsID := uuid.New()
	color := "#3B82F6"
	invalidColor := "123456789012345678901234567890123" // 33 chars > MaxTagColorLen

	testCases := []struct {
		name         string
		input        taskapp.CreateTagInput
		setupTagRepo func(m *mocks.MockTagRepository)
		checkOut     func(t *testing.T, tag *domainTask.Tag)
		wantErr      error
	}{
		{
			name: "успешное создание тега",
			input: taskapp.CreateTagInput{
				WorkspaceID: wsID,
				Name:        "Urgent Bug",
				ColorHex:    color,
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().CreateTag(mock.Anything, mock.MatchedBy(func(tag *domainTask.Tag) bool {
					return tag.WorkspaceID == wsID && tag.Name == "Urgent Bug" && tag.ColorHex == "#3B82F6"
				})).Return(nil).Once()
			},
			checkOut: func(t *testing.T, tag *domainTask.Tag) {
				assert.Equal(t, wsID, tag.WorkspaceID)
				assert.Equal(t, "Urgent Bug", tag.Name)
				assert.Equal(t, "#3B82F6", tag.ColorHex)
			},
		},
		{
			name: "ошибка валидации: пустое имя тега",
			input: taskapp.CreateTagInput{
				WorkspaceID: wsID,
				Name:        "   ",
				ColorHex:    color,
			},
			wantErr: domainTask.ErrInvalidTagName,
		},
		{
			name: "ошибка валидации: неверный HEX цвет",
			input: taskapp.CreateTagInput{
				WorkspaceID: wsID,
				Name:        "Feature",
				ColorHex:    invalidColor,
			},
			wantErr: domainTask.ErrInvalidTagColor,
		},
		{
			name: "ошибка репозитория при сохранении тега",
			input: taskapp.CreateTagInput{
				WorkspaceID: wsID,
				Name:        "DB Fail Tag",
				ColorHex:    color,
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().CreateTag(mock.Anything, mock.Anything).Return(errTestDB).Once()
			},
			wantErr: errTestDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tagRepo := mocks.NewMockTagRepository(t)
			if tc.setupTagRepo != nil {
				tc.setupTagRepo(tagRepo)
			}

			svc := taskapp.NewService(nil, tagRepo, nil)
			tag, err := svc.CreateTag(context.Background(), tc.input)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, tag)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tag)
				if tc.checkOut != nil {
					tc.checkOut(t, tag)
				}
			}
		})
	}
}

func TestService_ListWorkspaceTags(t *testing.T) {
	wsID := uuid.New()
	tag1, _ := domainTask.NewTag(wsID, "Tag 1", "#3B82F6")
	tag2, _ := domainTask.NewTag(wsID, "Tag 2", "#10B981")

	testCases := []struct {
		name         string
		workspaceID  uuid.UUID
		setupTagRepo func(m *mocks.MockTagRepository)
		checkOut     func(t *testing.T, tags []*domainTask.Tag)
		wantErr      error
	}{
		{
			name:        "успешный список тегов воркспейса",
			workspaceID: wsID,
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().ListWorkspaceTags(mock.Anything, wsID).Return([]*domainTask.Tag{tag1, tag2}, nil).Once()
			},
			checkOut: func(t *testing.T, tags []*domainTask.Tag) {
				require.Len(t, tags, 2)
				assert.Equal(t, "Tag 1", tags[0].Name)
			},
		},
		{
			name:        "ошибка репозитория при получении тегов",
			workspaceID: wsID,
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().ListWorkspaceTags(mock.Anything, wsID).Return(nil, errTestDB).Once()
			},
			wantErr: errTestDB,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tagRepo := mocks.NewMockTagRepository(t)
			if tc.setupTagRepo != nil {
				tc.setupTagRepo(tagRepo)
			}

			svc := taskapp.NewService(nil, tagRepo, nil)
			tags, err := svc.ListWorkspaceTags(context.Background(), tc.workspaceID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, tags)
			} else {
				require.NoError(t, err)
				if tc.checkOut != nil {
					tc.checkOut(t, tags)
				}
			}
		})
	}
}

func TestService_DeleteTag(t *testing.T) {
	wsID := uuid.New()
	tagID := uuid.New()
	tag, _ := domainTask.NewTag(wsID, "Tag", "#3B82F6")
	tag.ID = tagID

	testCases := []struct {
		name         string
		tagID        uuid.UUID
		workspaceID  uuid.UUID
		setupTagRepo func(m *mocks.MockTagRepository)
		wantErr      error
	}{
		{
			name:        "успешное удаление тега",
			tagID:       tagID,
			workspaceID: wsID,
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().GetTagByID(mock.Anything, tagID, wsID).Return(tag, nil).Once()
				m.EXPECT().DeleteTag(mock.Anything, tagID, wsID).Return(nil).Once()
			},
		},
		{
			name:        "тег не найден при удалении",
			tagID:       tagID,
			workspaceID: wsID,
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().GetTagByID(mock.Anything, tagID, wsID).Return(nil, domainTask.ErrTagNotFound).Once()
			},
			wantErr: domainTask.ErrTagNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tagRepo := mocks.NewMockTagRepository(t)
			if tc.setupTagRepo != nil {
				tc.setupTagRepo(tagRepo)
			}

			svc := taskapp.NewService(nil, tagRepo, nil)
			err := svc.DeleteTag(context.Background(), tc.tagID, tc.workspaceID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_AddTagToTask(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	tagID := uuid.New()
	task, _ := domainTask.NewInboxTask(wsID, "Task", "", domainTask.PriorityNone, nil, 0)
	task.ID = taskID
	tag, _ := domainTask.NewTag(wsID, "Tag", "#3B82F6")
	tag.ID = tagID

	testCases := []struct {
		name          string
		taskID        uuid.UUID
		tagID         uuid.UUID
		workspaceID   uuid.UUID
		setupTaskRepo func(m *mocks.MockTaskRepository)
		setupTagRepo  func(m *mocks.MockTagRepository)
		wantErr       error
	}{
		{
			name:        "успешное добавление тега к задаче",
			taskID:      taskID,
			tagID:       tagID,
			workspaceID: wsID,
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(task, nil).Once()
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().GetTagByID(mock.Anything, tagID, wsID).Return(tag, nil).Once()
				m.EXPECT().AddTagToTask(mock.Anything, taskID, tagID).Return(nil).Once()
			},
		},
		{
			name:        "задача не найдена",
			taskID:      taskID,
			tagID:       tagID,
			workspaceID: wsID,
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(nil, domainTask.ErrTaskNotFound).Once()
			},
			wantErr: domainTask.ErrTaskNotFound,
		},
		{
			name:        "тег не найден",
			taskID:      taskID,
			tagID:       tagID,
			workspaceID: wsID,
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(task, nil).Once()
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().GetTagByID(mock.Anything, tagID, wsID).Return(nil, domainTask.ErrTagNotFound).Once()
			},
			wantErr: domainTask.ErrTagNotFound,
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
			err := svc.AddTagToTask(context.Background(), tc.taskID, tc.tagID, tc.workspaceID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_RemoveTagFromTask(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	tagID := uuid.New()
	task, _ := domainTask.NewInboxTask(wsID, "Task", "", domainTask.PriorityNone, nil, 0)
	task.ID = taskID
	tag, _ := domainTask.NewTag(wsID, "Tag", "#3B82F6")
	tag.ID = tagID

	testCases := []struct {
		name          string
		taskID        uuid.UUID
		tagID         uuid.UUID
		workspaceID   uuid.UUID
		setupTaskRepo func(m *mocks.MockTaskRepository)
		setupTagRepo  func(m *mocks.MockTagRepository)
		wantErr       error
	}{
		{
			name:        "успешное удаление тега с задачи",
			taskID:      taskID,
			tagID:       tagID,
			workspaceID: wsID,
			setupTaskRepo: func(m *mocks.MockTaskRepository) {
				m.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(task, nil).Once()
			},
			setupTagRepo: func(m *mocks.MockTagRepository) {
				m.EXPECT().GetTagByID(mock.Anything, tagID, wsID).Return(tag, nil).Once()
				m.EXPECT().RemoveTagFromTask(mock.Anything, taskID, tagID).Return(nil).Once()
			},
		},
		{
			name:        "задача не найдена при удалении тега",
			taskID:      taskID,
			tagID:       tagID,
			workspaceID: wsID,
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
			err := svc.RemoveTagFromTask(context.Background(), tc.taskID, tc.tagID, tc.workspaceID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
