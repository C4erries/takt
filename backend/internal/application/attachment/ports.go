package attachment

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"

	domainAttachment "takt/internal/domain/attachment"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

// AttachmentRepository defines persistence operations for attachment metadata.
type AttachmentRepository interface {
	Create(ctx context.Context, att *domainAttachment.Attachment) (*domainAttachment.Attachment, error)
	GetByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainAttachment.Attachment, error)
	ListByTaskID(ctx context.Context, taskID, workspaceID uuid.UUID) ([]*domainAttachment.Attachment, error)
	Delete(ctx context.Context, id, workspaceID uuid.UUID) error
	DeleteByTaskID(ctx context.Context, taskID, workspaceID uuid.UUID) error
}

// StorageService defines binary storage operations in S3/MinIO.
type StorageService interface {
	PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, key string) (io.ReadCloser, string, int64, error)
	DeleteObject(ctx context.Context, key string) error
}

// WorkspaceAccessChecker ensures the requesting user belongs to the target workspace.
type WorkspaceAccessChecker interface {
	CheckAccess(ctx context.Context, workspaceID, userID uuid.UUID) error
}

// TaskChecker validates that a task exists and belongs to the given workspace.
type TaskChecker interface {
	TaskBelongsToWorkspace(ctx context.Context, taskID, workspaceID uuid.UUID) (bool, error)
}

// UploadAttachmentInput parameters for uploading an attachment.
type UploadAttachmentInput struct {
	WorkspaceID  uuid.UUID
	TaskID       uuid.UUID
	UserID       uuid.UUID
	FileName     string
	FileSize     int64
	ContentType  string
	Body         io.Reader
	MaxSizeBytes int64
}

// AttachmentContentOutput encapsulates file streaming details.
type AttachmentContentOutput struct {
	Body        io.ReadCloser
	FileName    string
	ContentType string
	FileSize    int64
}

// UseCases defines application use case operations for attachments.
type UseCases interface {
	UploadAttachment(ctx context.Context, input UploadAttachmentInput) (*domainAttachment.Attachment, error)
	ListTaskAttachments(ctx context.Context, workspaceID, taskID, userID uuid.UUID) ([]*domainAttachment.Attachment, error)
	GetAttachmentContent(ctx context.Context, workspaceID, attachmentID, userID uuid.UUID) (*AttachmentContentOutput, error)
	DeleteAttachment(ctx context.Context, workspaceID, attachmentID, userID uuid.UUID) error
}

