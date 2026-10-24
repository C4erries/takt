package task

import (
	"context"
	"time"

	"github.com/google/uuid"
	"takt/internal/domain/events"
	domainTask "takt/internal/domain/task"
)

// ListArchivedTasks returns archived tasks for a workspace with optional project filtering.
func (s *Service) ListArchivedTasks(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID) ([]*domainTask.Task, error) {
	if err := s.checkWorkspaceAccess(ctx, workspaceID); err != nil {
		return nil, err
	}

	tasks, err := s.taskRepo.ListArchivedTasks(ctx, workspaceID, projectID)
	if err != nil {
		return nil, err
	}

	s.loadTaskTags(ctx, tasks)
	return tasks, nil
}

// ArchiveCompletedTasksInColumn flushes all completed tasks in a column to archive.
func (s *Service) ArchiveCompletedTasksInColumn(ctx context.Context, workspaceID, columnID uuid.UUID) (int64, error) {
	if err := s.checkWorkspaceAccess(ctx, workspaceID); err != nil {
		return 0, err
	}

	affected, err := s.taskRepo.ArchiveCompletedTasksInColumn(ctx, columnID)
	if err != nil {
		return 0, err
	}

	if affected > 0 && s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.BatchSweepCompletedEvent{
			WorkspaceIDVal: workspaceID,
			ColumnIDVal:    columnID,
			ActionVal:      "archive_task",
			AffectedCount:  affected,
			OccurredAtVal:  time.Now().UTC(),
		})
	}

	return affected, nil
}

// UnarchiveTask restores an archived task to the specified column and project.
func (s *Service) UnarchiveTask(ctx context.Context, workspaceID, taskID, columnID uuid.UUID, boardID *uuid.UUID) (*domainTask.Task, error) {
	if err := s.checkWorkspaceAccess(ctx, workspaceID); err != nil {
		return nil, err
	}
	if columnID == uuid.Nil {
		return nil, domainTask.ErrInvalidColumnID
	}

	t, err := s.taskRepo.GetByID(ctx, taskID, workspaceID)
	if err != nil {
		return nil, err
	}

	updated, err := s.taskRepo.UnarchiveTask(ctx, taskID, workspaceID, columnID, boardID, float64(t.Position))
	if err != nil {
		return nil, err
	}

	s.loadSingleTaskTags(ctx, updated)

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.TaskUpdatedEvent{
			TaskID:            updated.ID,
			WorkspaceIDVal:    updated.WorkspaceID,
			ParentID:          updated.ParentID,
			ProjectID:         updated.ProjectID,
			ColumnID:          updated.ColumnID,
			Title:             updated.Title,
			DueDate:           updated.DueDate,
			IsCompleted:       updated.IsCompleted,
			PreviousProjectID: t.ProjectID,
			PreviousColumnID:  t.ColumnID,
			OccurredAtVal:     time.Now().UTC(),
			FromAutomation:    events.IsAutomationTriggered(ctx),
		})
	}

	return updated, nil
}
