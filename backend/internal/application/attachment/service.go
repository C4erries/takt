package attachment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainAttachment "takt/internal/domain/attachment"
)

// Service implements the attachment UseCases.
type Service struct {
	repo          AttachmentRepository
	storage       StorageService
	accessChecker WorkspaceAccessChecker
	taskChecker   TaskChecker
}

// NewService creates a new attachment application service.
func NewService(
	repo AttachmentRepository,
	storage StorageService,
	accessChecker WorkspaceAccessChecker,
	taskChecker TaskChecker,
) *Service {
	return &Service{
		repo:          repo,
		storage:       storage,
		accessChecker: accessChecker,
		taskChecker:   taskChecker,
	}
}

// UploadAttachment checks access, task existence, validates file invariants, streams to S3 and persists metadata.
func (s *Service) UploadAttachment(ctx context.Context, input UploadAttachmentInput) (*domainAttachment.Attachment, error) {
	if s.accessChecker != nil {
		if err := s.accessChecker.CheckAccess(ctx, input.WorkspaceID, input.UserID); err != nil {
			return nil, err
		}
	}

	if s.taskChecker != nil {
		exists, err := s.taskChecker.TaskBelongsToWorkspace(ctx, input.TaskID, input.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify task: %w", err)
		}
		if !exists {
			return nil, ErrTaskNotFound
		}
	}

	att, err := domainAttachment.NewAttachment(
		input.WorkspaceID,
		input.TaskID,
		input.UserID,
		input.FileName,
		input.FileSize,
		input.ContentType,
		input.MaxSizeBytes,
	)
	if err != nil {
		return nil, err
	}

	if err := s.storage.PutObject(ctx, att.S3Key, input.Body, att.FileSize, att.ContentType); err != nil {
		return nil, fmt.Errorf("failed to store object in s3: %w", err)
	}

	created, err := s.repo.Create(ctx, att)
	if err != nil {
		// Attempt compensating cleanup in S3 if DB insert fails
		_ = s.storage.DeleteObject(ctx, att.S3Key)
		return nil, fmt.Errorf("failed to save attachment metadata: %w", err)
	}

	return created, nil
}

// ListTaskAttachments returns all attachments for the specified task.
func (s *Service) ListTaskAttachments(ctx context.Context, workspaceID, taskID, userID uuid.UUID) ([]*domainAttachment.Attachment, error) {
	if s.accessChecker != nil {
		if err := s.accessChecker.CheckAccess(ctx, workspaceID, userID); err != nil {
			return nil, err
		}
	}

	if s.taskChecker != nil {
		exists, err := s.taskChecker.TaskBelongsToWorkspace(ctx, taskID, workspaceID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify task: %w", err)
		}
		if !exists {
			return nil, ErrTaskNotFound
		}
	}

	return s.repo.ListByTaskID(ctx, taskID, workspaceID)
}

// GetAttachmentContent streams the file binary from S3 after checking workspace access.
func (s *Service) GetAttachmentContent(ctx context.Context, workspaceID, attachmentID, userID uuid.UUID) (*AttachmentContentOutput, error) {
	if s.accessChecker != nil {
		if err := s.accessChecker.CheckAccess(ctx, workspaceID, userID); err != nil {
			return nil, err
		}
	}

	att, err := s.repo.GetByID(ctx, attachmentID, workspaceID)
	if err != nil {
		return nil, err
	}
	if att == nil {
		return nil, domainAttachment.ErrAttachmentNotFound
	}

	body, contentType, size, err := s.storage.GetObject(ctx, att.S3Key)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve object content: %w", err)
	}

	if contentType == "" || contentType == "application/octet-stream" {
		contentType = att.ContentType
	}
	if size <= 0 {
		size = att.FileSize
	}

	return &AttachmentContentOutput{
		Body:        body,
		FileName:    att.FileName,
		ContentType: contentType,
		FileSize:    size,
	}, nil
}

// DeleteAttachment removes the attachment object from S3 and deletes its record in DB.
func (s *Service) DeleteAttachment(ctx context.Context, workspaceID, attachmentID, userID uuid.UUID) error {
	if s.accessChecker != nil {
		if err := s.accessChecker.CheckAccess(ctx, workspaceID, userID); err != nil {
			return err
		}
	}

	att, err := s.repo.GetByID(ctx, attachmentID, workspaceID)
	if err != nil {
		return err
	}
	if att == nil {
		return domainAttachment.ErrAttachmentNotFound
	}

	// Remove from S3 first (or best-effort delete)
	_ = s.storage.DeleteObject(ctx, att.S3Key)

	return s.repo.Delete(ctx, attachmentID, workspaceID)
}

