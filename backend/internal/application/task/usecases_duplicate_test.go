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
	domainAttachment "takt/internal/domain/attachment"
	"takt/internal/domain/events"
	domainTask "takt/internal/domain/task"
)

func TestService_DuplicateTask_Table(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	projectID := uuid.New()
	columnID := uuid.New()
	userID := uuid.New()
	tagID1 := uuid.New()
	tagID2 := uuid.New()
	now := time.Now().UTC()
	dueDate := now.Add(24 * time.Hour)

	tests := []struct {
		name        string
		setupMocks  func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockTagRepository, *mocks.MockTransactor, *mocks.MockAttachmentRepository, *mocks.MockEventPublisher)
		ctx         context.Context
		expectError bool
		checkResult func(t *testing.T, res *domainTask.Task)
	}{
		{
			name: "Success deep duplicate with tags, subtasks, attachments and transaction",
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockTagRepository, *mocks.MockTransactor, *mocks.MockAttachmentRepository, *mocks.MockEventPublisher) {
				mockRepo := mocks.NewMockTaskRepository(t)
				mockTagRepo := mocks.NewMockTagRepository(t)
				mockTx := mocks.NewMockTransactor(t)
				mockAtt := mocks.NewMockAttachmentRepository(t)
				mockEvents := mocks.NewMockEventPublisher(t)

				mockTx.EXPECT().WithinTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(txCtx context.Context) error) error {
					return fn(ctx)
				}).Once()

				origTask := &domainTask.Task{
					ID:                  taskID,
					WorkspaceID:         wsID,
					ProjectID:           &projectID,
					ColumnID:            &columnID,
					Title:               "Original Task",
					DescriptionMarkdown: "Some notes",
					IsCompleted:         true,
					CompletedAt:         &now,
					DueDate:             &dueDate,
					Priority:            domainTask.PriorityHigh,
					Position:            5,
					Tags: []*domainTask.Tag{
						{ID: tagID1, WorkspaceID: wsID, Name: "Urgent", ColorHex: "#ff0000"},
						{ID: tagID2, WorkspaceID: wsID, Name: "Backend", ColorHex: "#00ff00"},
					},
					CreatedAt: now,
					UpdatedAt: now,
				}

				mockRepo.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(origTask, nil).Once()
				mockTagRepo.EXPECT().GetTaskTags(mock.Anything, taskID).Return(origTask.Tags, nil).Once()

				mockRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(t *domainTask.Task) bool {
					return t.Title == "Копия — Original Task" &&
						t.Position == 6 &&
						!t.IsCompleted &&
						t.CompletedAt == nil &&
						t.DueDate == nil &&
						t.WorkspaceID == wsID &&
						*t.ProjectID == projectID &&
						*t.ColumnID == columnID
				})).Return(nil).Once()

				// Expect tags added to duplicate
				mockTagRepo.EXPECT().AddTagToTask(mock.Anything, mock.MatchedBy(func(id uuid.UUID) bool {
					return id != uuid.Nil && id != taskID
				}), tagID1).Return(nil).Once()
				mockTagRepo.EXPECT().AddTagToTask(mock.Anything, mock.MatchedBy(func(id uuid.UUID) bool {
					return id != uuid.Nil && id != taskID
				}), tagID2).Return(nil).Once()

				// Subtasks
				subtaskID := uuid.New()
				subtask := &domainTask.Task{
					ID:          subtaskID,
					WorkspaceID: wsID,
					ParentID:    &taskID,
					Title:       "Subtask 1",
					Position:    1,
					Tags:        []*domainTask.Tag{},
				}
				mockRepo.EXPECT().ListSubtasks(mock.Anything, taskID).Return([]*domainTask.Task{subtask}, nil).Once()
				mockTagRepo.EXPECT().GetTaskTags(mock.Anything, subtaskID).Return([]*domainTask.Tag{}, nil).Once()
				mockRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(st *domainTask.Task) bool {
					return st.Title == "Subtask 1" &&
						st.ParentID != nil &&
						st.Position == 1
				})).Return(nil).Once()

				// Attachments
				attList := []*domainAttachment.Attachment{
					{
						ID:          uuid.New(),
						WorkspaceID: wsID,
						TaskID:      taskID,
						FileName:    "spec.pdf",
						FileSize:    1024,
						ContentType: "application/pdf",
						S3Key:       "s3-key-123",
						CreatedByID: userID,
					},
				}
				mockAtt.EXPECT().ListByTaskID(mock.Anything, taskID, wsID).Return(attList, nil).Once()
				mockAtt.EXPECT().Create(mock.Anything, mock.MatchedBy(func(att *domainAttachment.Attachment) bool {
					return att.TaskID != uuid.Nil && att.TaskID != taskID &&
						att.S3Key == "s3-key-123" &&
						att.FileName == "spec.pdf" &&
						att.FileSize == 1024
				})).Return(&domainAttachment.Attachment{}, nil).Once()

				// Reload duplicate tags
				mockTagRepo.EXPECT().GetTaskTags(mock.Anything, mock.MatchedBy(func(id uuid.UUID) bool {
					return id != uuid.Nil && id != taskID
				})).Return(origTask.Tags, nil).Once()

				// Event published
				mockEvents.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(evt events.Event) bool {
					te, ok := evt.(events.TaskCreatedEvent)
					return ok && te.Title == "Копия — Original Task" && te.TaskID != uuid.Nil && te.TaskID != taskID
				})).Return(nil).Once()

				svc := taskapp.NewService(mockRepo, mockTagRepo, nil, mockEvents)
				svc.SetTransactor(mockTx)
				svc.SetAttachmentRepository(mockAtt)
				return svc, mockRepo, mockTagRepo, mockTx, mockAtt, mockEvents
			},
			ctx:         context.Background(),
			expectError: false,
			checkResult: func(t *testing.T, res *domainTask.Task) {
				require.NotNil(t, res)
				assert.Equal(t, "Копия — Original Task", res.Title)
				assert.Equal(t, 6, res.Position)
				assert.False(t, res.IsCompleted)
				assert.Nil(t, res.CompletedAt)
				assert.Nil(t, res.DueDate)
				assert.Equal(t, 1, res.SubtasksCount)
				assert.Equal(t, 0, res.SubtasksCompletedCount)
			},
		},
		{
			name: "Fallback when transactor and attachment repo are nil",
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockTagRepository, *mocks.MockTransactor, *mocks.MockAttachmentRepository, *mocks.MockEventPublisher) {
				mockRepo := mocks.NewMockTaskRepository(t)

				origTask := &domainTask.Task{
					ID:          taskID,
					WorkspaceID: wsID,
					Title:       "Task Without Extras",
					Position:    1,
				}
				mockRepo.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(origTask, nil).Once()
				mockRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
				mockRepo.EXPECT().ListSubtasks(mock.Anything, taskID).Return([]*domainTask.Task{}, nil).Once()

				svc := taskapp.NewService(mockRepo, nil, nil)
				return svc, mockRepo, nil, nil, nil, nil
			},
			ctx:         context.Background(),
			expectError: false,
			checkResult: func(t *testing.T, res *domainTask.Task) {
				require.NotNil(t, res)
				assert.Equal(t, "Копия — Task Without Extras", res.Title)
				assert.Equal(t, 2, res.Position)
			},
		},
		{
			name: "Workspace access denied",
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockTagRepository, *mocks.MockTransactor, *mocks.MockAttachmentRepository, *mocks.MockEventPublisher) {
				mockChecker := mocks.NewMockWorkspaceAccessChecker(t)
				mockChecker.EXPECT().CheckAccess(mock.Anything, wsID, userID).Return(errors.New("forbidden")).Once()

				svc := taskapp.NewService(nil, nil, mockChecker)
				return svc, nil, nil, nil, nil, nil
			},
			ctx:         context.WithValue(context.Background(), "userID", userID),
			expectError: true,
			checkResult: nil,
		},
		{
			name: "Original task not found",
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockTagRepository, *mocks.MockTransactor, *mocks.MockAttachmentRepository, *mocks.MockEventPublisher) {
				mockRepo := mocks.NewMockTaskRepository(t)
				mockRepo.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(nil, domainTask.ErrTaskNotFound).Once()

				svc := taskapp.NewService(mockRepo, nil, nil)
				return svc, mockRepo, nil, nil, nil, nil
			},
			ctx:         context.Background(),
			expectError: true,
			checkResult: nil,
		},
		{
			name: "Create task returns error",
			setupMocks: func(t *testing.T) (*taskapp.Service, *mocks.MockTaskRepository, *mocks.MockTagRepository, *mocks.MockTransactor, *mocks.MockAttachmentRepository, *mocks.MockEventPublisher) {
				mockRepo := mocks.NewMockTaskRepository(t)
				origTask := &domainTask.Task{
					ID:          taskID,
					WorkspaceID: wsID,
					Title:       "Task",
				}
				mockRepo.EXPECT().GetByID(mock.Anything, taskID, wsID).Return(origTask, nil).Once()
				mockRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.New("db error")).Once()

				svc := taskapp.NewService(mockRepo, nil, nil)
				return svc, mockRepo, nil, nil, nil, nil
			},
			ctx:         context.Background(),
			expectError: true,
			checkResult: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, _, _, _, _ := tc.setupMocks(t)
			res, err := svc.DuplicateTask(tc.ctx, wsID, taskID)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tc.checkResult != nil {
					tc.checkResult(t, res)
				}
			}
		})
	}
}
