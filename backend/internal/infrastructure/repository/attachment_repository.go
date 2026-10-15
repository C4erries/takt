package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domainAttachment "takt/internal/domain/attachment"
	"takt/internal/infrastructure/db/sqlc"
)

// AttachmentRepository implements attachment persistence via sqlc.
type AttachmentRepository struct {
	queries *sqlc.Queries
}

// NewAttachmentRepository creates a new AttachmentRepository.
func NewAttachmentRepository(queries *sqlc.Queries) *AttachmentRepository {
	return &AttachmentRepository{queries: queries}
}

// Create persists a new attachment metadata record and updates ID and CreatedAt.
func (r *AttachmentRepository) Create(ctx context.Context, att *domainAttachment.Attachment) (*domainAttachment.Attachment, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.CreateAttachment(ctx, sqlc.CreateAttachmentParams{
		WorkspaceID: uuidToPg(att.WorkspaceID),
		TaskID:      uuidToPg(att.TaskID),
		FileName:    att.FileName,
		FileSize:    att.FileSize,
		ContentType: att.ContentType,
		S3Key:       att.S3Key,
		CreatedByID: uuidToPg(att.CreatedByID),
	})
	if err != nil {
		return nil, err
	}

	return &domainAttachment.Attachment{
		ID:          pgToUUID(row.ID),
		WorkspaceID: pgToUUID(row.WorkspaceID),
		TaskID:      pgToUUID(row.TaskID),
		FileName:    row.FileName,
		FileSize:    row.FileSize,
		ContentType: row.ContentType,
		S3Key:       row.S3Key,
		CreatedByID: pgToUUID(row.CreatedByID),
		CreatedAt:   pgToTime(row.CreatedAt),
	}, nil
}

// GetByID retrieves attachment metadata for the given workspace.
func (r *AttachmentRepository) GetByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainAttachment.Attachment, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetAttachmentByID(ctx, sqlc.GetAttachmentByIDParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainAttachment.ErrAttachmentNotFound
		}
		return nil, err
	}

	return &domainAttachment.Attachment{
		ID:          pgToUUID(row.ID),
		WorkspaceID: pgToUUID(row.WorkspaceID),
		TaskID:      pgToUUID(row.TaskID),
		FileName:    row.FileName,
		FileSize:    row.FileSize,
		ContentType: row.ContentType,
		S3Key:       row.S3Key,
		CreatedByID: pgToUUID(row.CreatedByID),
		CreatedAt:   pgToTime(row.CreatedAt),
	}, nil
}

// ListByTaskID returns all attachments associated with a task in the given workspace.
func (r *AttachmentRepository) ListByTaskID(ctx context.Context, taskID, workspaceID uuid.UUID) ([]*domainAttachment.Attachment, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListAttachmentsByTaskID(ctx, sqlc.ListAttachmentsByTaskIDParams{
		TaskID:      uuidToPg(taskID),
		WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		return nil, err
	}

	res := make([]*domainAttachment.Attachment, 0, len(rows))
	for _, row := range rows {
		res = append(res, &domainAttachment.Attachment{
			ID:          pgToUUID(row.ID),
			WorkspaceID: pgToUUID(row.WorkspaceID),
			TaskID:      pgToUUID(row.TaskID),
			FileName:    row.FileName,
			FileSize:    row.FileSize,
			ContentType: row.ContentType,
			S3Key:       row.S3Key,
			CreatedByID: pgToUUID(row.CreatedByID),
			CreatedAt:   pgToTime(row.CreatedAt),
		})
	}
	return res, nil
}

// Delete removes an attachment record by ID and workspaceID.
func (r *AttachmentRepository) Delete(ctx context.Context, id, workspaceID uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.DeleteAttachment(ctx, sqlc.DeleteAttachmentParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
	})
}

// DeleteByTaskID removes all attachment records for a task in the workspace.
func (r *AttachmentRepository) DeleteByTaskID(ctx context.Context, taskID, workspaceID uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.DeleteAttachmentsByTaskID(ctx, sqlc.DeleteAttachmentsByTaskIDParams{
		TaskID:      uuidToPg(taskID),
		WorkspaceID: uuidToPg(workspaceID),
	})
}

// TaskBelongsToWorkspace validates task existence in the given workspace.
func (r *AttachmentRepository) TaskBelongsToWorkspace(ctx context.Context, taskID, workspaceID uuid.UUID) (bool, error) {
	q := getQuerier(ctx, r.queries)
	_, err := q.GetTaskByID(ctx, sqlc.GetTaskByIDParams{
		ID:          uuidToPg(taskID),
		WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

