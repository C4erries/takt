package task

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	domainAttachment "takt/internal/domain/attachment"
	"takt/internal/domain/events"
	domainTask "takt/internal/domain/task"
)

// DuplicateTask performs a deep copy of a task including its tags, subtasks, and attachments.
func (s *Service) DuplicateTask(ctx context.Context, workspaceID, taskID uuid.UUID) (*domainTask.Task, error) {
	if err := s.checkWorkspaceAccess(ctx, workspaceID); err != nil {
		return nil, err
	}

	var duplicatedTask *domainTask.Task

	op := func(txCtx context.Context) error {
		orig, err := s.taskRepo.GetByID(txCtx, taskID, workspaceID)
		if err != nil {
			return err
		}

		s.loadSingleTaskTags(txCtx, orig)

		dup := orig.Duplicate("Копия — "+orig.Title, float64(orig.Position+1))

		if err := s.taskRepo.Create(txCtx, dup); err != nil {
			return fmt.Errorf("failed to create duplicated task: %w", err)
		}

		// Copy tags to duplicate
		if s.tagRepo != nil {
			for _, tag := range orig.Tags {
				if err := s.tagRepo.AddTagToTask(txCtx, dup.ID, tag.ID); err != nil {
					return fmt.Errorf("failed to copy tag to duplicate: %w", err)
				}
			}
		}

		// Copy subtasks
		subtasks, err := s.taskRepo.ListSubtasks(txCtx, orig.ID)
		if err != nil {
			return fmt.Errorf("failed to list subtasks: %w", err)
		}
		for _, sub := range subtasks {
			s.loadSingleTaskTags(txCtx, sub)
			subCopy := sub.Duplicate(sub.Title, float64(sub.Position))
			subCopy.ParentID = &dup.ID
			if err := s.taskRepo.Create(txCtx, subCopy); err != nil {
				return fmt.Errorf("failed to create duplicated subtask: %w", err)
			}
			if s.tagRepo != nil {
				for _, stag := range sub.Tags {
					_ = s.tagRepo.AddTagToTask(txCtx, subCopy.ID, stag.ID)
				}
			}
		}
		if len(subtasks) > 0 {
			dup.SetSubtaskCounts(len(subtasks), 0)
		}

		// Copy attachments (shallow copy of S3 files)
		if s.attachmentRepo != nil {
			attachments, err := s.attachmentRepo.ListByTaskID(txCtx, orig.ID, workspaceID)
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}
			for _, att := range attachments {
				newAtt := &domainAttachment.Attachment{
					ID:          uuid.New(),
					WorkspaceID: workspaceID,
					TaskID:      dup.ID,
					FileName:    att.FileName,
					FileSize:    att.FileSize,
					ContentType: att.ContentType,
					S3Key:       att.S3Key,
					CreatedByID: att.CreatedByID,
					CreatedAt:   time.Now().UTC(),
				}
				if _, err := s.attachmentRepo.Create(txCtx, newAtt); err != nil {
					return fmt.Errorf("failed to copy attachment: %w", err)
				}
			}
		}

		s.loadSingleTaskTags(txCtx, dup)
		duplicatedTask = dup
		return nil
	}

	if s.transactor != nil {
		if err := s.transactor.WithinTx(ctx, op); err != nil {
			return nil, err
		}
	} else {
		if err := op(ctx); err != nil {
			return nil, err
		}
	}

	if s.eventBus != nil && duplicatedTask != nil {
		_ = s.eventBus.Publish(ctx, events.TaskCreatedEvent{
			TaskID:         duplicatedTask.ID,
			WorkspaceIDVal: duplicatedTask.WorkspaceID,
			ParentID:       duplicatedTask.ParentID,
			ProjectID:      duplicatedTask.ProjectID,
			ColumnID:       duplicatedTask.ColumnID,
			Title:          duplicatedTask.Title,
			DueDate:        duplicatedTask.DueDate,
			OccurredAtVal:  time.Now().UTC(),
			FromAutomation: events.IsAutomationTriggered(ctx),
		})
	}

	return duplicatedTask, nil
}
