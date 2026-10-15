package attachment_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appAttachment "takt/internal/application/attachment"
	domainAttachment "takt/internal/domain/attachment"
)

type mockRepo struct {
	createFunc       func(ctx context.Context, att *domainAttachment.Attachment) (*domainAttachment.Attachment, error)
	getByIDFunc      func(ctx context.Context, id, workspaceID uuid.UUID) (*domainAttachment.Attachment, error)
	listByTaskIDFunc func(ctx context.Context, taskID, workspaceID uuid.UUID) ([]*domainAttachment.Attachment, error)
	deleteFunc       func(ctx context.Context, id, workspaceID uuid.UUID) error
	deleteByTaskFunc func(ctx context.Context, taskID, workspaceID uuid.UUID) error
}

func (m *mockRepo) Create(ctx context.Context, att *domainAttachment.Attachment) (*domainAttachment.Attachment, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, att)
	}
	return att, nil
}

func (m *mockRepo) GetByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainAttachment.Attachment, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id, workspaceID)
	}
	return nil, nil
}

func (m *mockRepo) ListByTaskID(ctx context.Context, taskID, workspaceID uuid.UUID) ([]*domainAttachment.Attachment, error) {
	if m.listByTaskIDFunc != nil {
		return m.listByTaskIDFunc(ctx, taskID, workspaceID)
	}
	return []*domainAttachment.Attachment{}, nil
}

func (m *mockRepo) Delete(ctx context.Context, id, workspaceID uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id, workspaceID)
	}
	return nil
}

func (m *mockRepo) DeleteByTaskID(ctx context.Context, taskID, workspaceID uuid.UUID) error {
	if m.deleteByTaskFunc != nil {
		return m.deleteByTaskFunc(ctx, taskID, workspaceID)
	}
	return nil
}

type mockStorage struct {
	putFunc    func(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	getFunc    func(ctx context.Context, key string) (io.ReadCloser, string, int64, error)
	deleteFunc func(ctx context.Context, key string) error
}

func (m *mockStorage) PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if m.putFunc != nil {
		return m.putFunc(ctx, key, body, size, contentType)
	}
	return nil
}

func (m *mockStorage) GetObject(ctx context.Context, key string) (io.ReadCloser, string, int64, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, key)
	}
	return io.NopCloser(bytes.NewReader([]byte{})), "application/octet-stream", 0, nil
}

func (m *mockStorage) DeleteObject(ctx context.Context, key string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, key)
	}
	return nil
}

type mockAccessChecker struct {
	checkAccessFunc func(ctx context.Context, workspaceID, userID uuid.UUID) error
}

func (m *mockAccessChecker) CheckAccess(ctx context.Context, workspaceID, userID uuid.UUID) error {
	if m.checkAccessFunc != nil {
		return m.checkAccessFunc(ctx, workspaceID, userID)
	}
	return nil
}

type mockTaskChecker struct {
	belongsFunc func(ctx context.Context, taskID, workspaceID uuid.UUID) (bool, error)
}

func (m *mockTaskChecker) TaskBelongsToWorkspace(ctx context.Context, taskID, workspaceID uuid.UUID) (bool, error) {
	if m.belongsFunc != nil {
		return m.belongsFunc(ctx, taskID, workspaceID)
	}
	return true, nil
}

func TestService_UploadAttachment(t *testing.T) {
	workspaceID := uuid.New()
	taskID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name          string
		accessErr     error
		taskBelongs   bool
		taskErr       error
		s3Err         error
		dbErr         error
		fileSize      int64
		maxSize       int64
		wantErr       bool
		expectedErrIs error
	}{
		{
			name:        "successful upload",
			taskBelongs: true,
			fileSize:    1024,
			maxSize:     2048,
			wantErr:     false,
		},
		{
			name:          "workspace access denied",
			accessErr:     errors.New("forbidden"),
			taskBelongs:   true,
			fileSize:      1024,
			maxSize:       2048,
			wantErr:       true,
			expectedErrIs: nil,
		},
		{
			name:          "task not found in workspace",
			taskBelongs:   false,
			fileSize:      1024,
			maxSize:       2048,
			wantErr:       true,
			expectedErrIs: appAttachment.ErrTaskNotFound,
		},
		{
			name:          "file size exceeds limit",
			taskBelongs:   true,
			fileSize:      3000,
			maxSize:       2048,
			wantErr:       true,
			expectedErrIs: domainAttachment.ErrFileTooLarge,
		},
		{
			name:        "s3 put error",
			taskBelongs: true,
			s3Err:       errors.New("s3 write failed"),
			fileSize:    1024,
			maxSize:     2048,
			wantErr:     true,
		},
		{
			name:        "db error triggers cleanup",
			taskBelongs: true,
			dbErr:       errors.New("db insert failed"),
			fileSize:    1024,
			maxSize:     2048,
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deletedKey := ""
			storage := &mockStorage{
				putFunc: func(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
					return tc.s3Err
				},
				deleteFunc: func(ctx context.Context, key string) error {
					deletedKey = key
					return nil
				},
			}
			repo := &mockRepo{
				createFunc: func(ctx context.Context, att *domainAttachment.Attachment) (*domainAttachment.Attachment, error) {
					if tc.dbErr != nil {
						return nil, tc.dbErr
					}
					return att, nil
				},
			}
			accessChecker := &mockAccessChecker{
				checkAccessFunc: func(ctx context.Context, wID, uID uuid.UUID) error {
					return tc.accessErr
				},
			}
			taskChecker := &mockTaskChecker{
				belongsFunc: func(ctx context.Context, tID, wID uuid.UUID) (bool, error) {
					return tc.taskBelongs, tc.taskErr
				},
			}

			svc := appAttachment.NewService(repo, storage, accessChecker, taskChecker)
			input := appAttachment.UploadAttachmentInput{
				WorkspaceID:  workspaceID,
				TaskID:       taskID,
				UserID:       userID,
				FileName:     "test.png",
				FileSize:     tc.fileSize,
				ContentType:  "image/png",
				Body:         strings.NewReader("sample content"),
				MaxSizeBytes: tc.maxSize,
			}

			att, err := svc.UploadAttachment(context.Background(), input)

			if tc.wantErr {
				require.Error(t, err)
				if tc.expectedErrIs != nil {
					assert.ErrorIs(t, err, tc.expectedErrIs)
				}
				if tc.dbErr != nil {
					assert.NotEmpty(t, deletedKey, "compensating s3 delete should be invoked")
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, att)
				assert.Equal(t, "test.png", att.FileName)
			}
		})
	}
}

func TestService_ListTaskAttachments(t *testing.T) {
	workspaceID := uuid.New()
	taskID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		taskBelongs bool
		repoList    []*domainAttachment.Attachment
		wantErr     bool
	}{
		{
			name:        "successful list",
			taskBelongs: true,
			repoList: []*domainAttachment.Attachment{
				{ID: uuid.New(), FileName: "doc1.pdf"},
				{ID: uuid.New(), FileName: "doc2.pdf"},
			},
			wantErr: false,
		},
		{
			name:        "task not in workspace",
			taskBelongs: false,
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockRepo{
				listByTaskIDFunc: func(ctx context.Context, tID, wID uuid.UUID) ([]*domainAttachment.Attachment, error) {
					return tc.repoList, nil
				},
			}
			taskChecker := &mockTaskChecker{
				belongsFunc: func(ctx context.Context, tID, wID uuid.UUID) (bool, error) {
					return tc.taskBelongs, nil
				},
			}

			svc := appAttachment.NewService(repo, &mockStorage{}, &mockAccessChecker{}, taskChecker)
			res, err := svc.ListTaskAttachments(context.Background(), workspaceID, taskID, userID)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, res, len(tc.repoList))
			}
		})
	}
}

