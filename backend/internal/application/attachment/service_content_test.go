package attachment_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appAttachment "takt/internal/application/attachment"
	domainAttachment "takt/internal/domain/attachment"
)

func TestService_GetAttachmentContent(t *testing.T) {
	workspaceID := uuid.New()
	attachmentID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		repoAtt     *domainAttachment.Attachment
		repoErr     error
		s3Err       error
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful content fetch",
			repoAtt: &domainAttachment.Attachment{
				ID:          attachmentID,
				WorkspaceID: workspaceID,
				FileName:    "image.png",
				FileSize:    1234,
				ContentType: "image/png",
				S3Key:       "key1",
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name:        "attachment not found in db",
			repoAtt:     nil,
			wantErr:     true,
			expectedErr: domainAttachment.ErrAttachmentNotFound,
		},
		{
			name: "s3 read error",
			repoAtt: &domainAttachment.Attachment{
				ID:          attachmentID,
				WorkspaceID: workspaceID,
				FileName:    "image.png",
				S3Key:       "key1",
			},
			s3Err:   errors.New("s3 get failed"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockRepo{
				getByIDFunc: func(ctx context.Context, id, wID uuid.UUID) (*domainAttachment.Attachment, error) {
					return tc.repoAtt, tc.repoErr
				},
			}
			storage := &mockStorage{
				getFunc: func(ctx context.Context, key string) (io.ReadCloser, string, int64, error) {
					if tc.s3Err != nil {
						return nil, "", 0, tc.s3Err
					}
					return io.NopCloser(strings.NewReader("dummy binary")), "image/png", 12, nil
				},
			}

			svc := appAttachment.NewService(repo, storage, &mockAccessChecker{}, &mockTaskChecker{})
			out, err := svc.GetAttachmentContent(context.Background(), workspaceID, attachmentID, userID)

			if tc.wantErr {
				require.Error(t, err)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, out)
				assert.Equal(t, "image.png", out.FileName)
			}
		})
	}
}

func TestService_DeleteAttachment(t *testing.T) {
	workspaceID := uuid.New()
	attachmentID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		repoAtt     *domainAttachment.Attachment
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful delete",
			repoAtt: &domainAttachment.Attachment{
				ID:          attachmentID,
				WorkspaceID: workspaceID,
				S3Key:       "key1",
			},
			wantErr: false,
		},
		{
			name:        "not found",
			repoAtt:     nil,
			wantErr:     true,
			expectedErr: domainAttachment.ErrAttachmentNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deletedFromS3 := false
			repo := &mockRepo{
				getByIDFunc: func(ctx context.Context, id, wID uuid.UUID) (*domainAttachment.Attachment, error) {
					return tc.repoAtt, nil
				},
			}
			storage := &mockStorage{
				deleteFunc: func(ctx context.Context, key string) error {
					deletedFromS3 = true
					return nil
				},
			}

			svc := appAttachment.NewService(repo, storage, &mockAccessChecker{}, &mockTaskChecker{})
			err := svc.DeleteAttachment(context.Background(), workspaceID, attachmentID, userID)

			if tc.wantErr {
				require.Error(t, err)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.True(t, deletedFromS3)
			}
		})
	}
}

