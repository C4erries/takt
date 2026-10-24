package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	domainTask "takt/internal/domain/task"
	"takt/internal/infrastructure/db/sqlc"
)

type TaskRepository struct {
	queries *sqlc.Queries
}

func NewTaskRepository(queries *sqlc.Queries) *TaskRepository {
	return &TaskRepository{queries: queries}
}

func (r *TaskRepository) Create(ctx context.Context, t *domainTask.Task) error {
	q := getQuerier(ctx, r.queries)

	row, err := q.CreateTask(ctx, sqlc.CreateTaskParams{
		WorkspaceID:         uuidToPg(t.WorkspaceID),
		ProjectID:           uuidPtrToPg(t.ProjectID),
		ColumnID:            uuidPtrToPg(t.ColumnID),
		Title:               t.Title,
		DescriptionMarkdown: t.DescriptionMarkdown,
		IsCompleted:         t.IsCompleted,
		CompletedAt:         timePtrToPg(t.CompletedAt),
		DueDate:             dateToPg(t.DueDate),
		Priority:            string(t.Priority),
		FocusOrder:          intPtrToPg(t.FocusOrder),
		Position:            float64(t.Position),
		ParentID:            uuidPtrToPg(t.ParentID),
	})
	if err != nil {
		return err
	}

	t.ID = pgToUUID(row.ID)
	t.CreatedAt = pgToTime(row.CreatedAt)
	t.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

func (r *TaskRepository) GetInboxTasks(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.GetInboxTasks(ctx, uuidToPg(workspaceID))
	if err != nil {
		return nil, err
	}
	return mapTaskRowsToDomain(ctx, q, rows)
}

func (r *TaskRepository) GetByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetTaskByID(ctx, sqlc.GetTaskByIDParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainTask.ErrTaskNotFound
		}
		return nil, err
	}
	return mapSingleTaskRow(ctx, q, row), nil
}

func (r *TaskRepository) ListSubtasks(ctx context.Context, parentID uuid.UUID) ([]*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListSubtasksByParent(ctx, uuidToPg(parentID))
	if err != nil {
		return nil, err
	}
	return mapTaskRowsToDomain(ctx, q, rows)
}

func (r *TaskRepository) Toggle(ctx context.Context, id, workspaceID uuid.UUID) (*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.ToggleTaskCompletion(ctx, sqlc.ToggleTaskCompletionParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainTask.ErrTaskNotFound
		}
		return nil, err
	}
	return mapSingleTaskRow(ctx, q, row), nil
}

func (r *TaskRepository) Delete(ctx context.Context, id, workspaceID uuid.UUID) error {
	if _, err := r.GetByID(ctx, id, workspaceID); err != nil {
		return err
	}
	q := getQuerier(ctx, r.queries)
	return q.DeleteTask(ctx, sqlc.DeleteTaskParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
	})
}

func (r *TaskRepository) GetMaxInboxPosition(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	q := getQuerier(ctx, r.queries)
	maxPos, err := q.GetMaxInboxPosition(ctx, uuidToPg(workspaceID))
	if err != nil {
		return -1, err
	}
	return int(maxPos), nil
}

func (r *TaskRepository) GetTodayTasks(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.GetTodayTasks(ctx, uuidToPg(workspaceID))
	if err != nil {
		return nil, err
	}
	return mapTaskRowsToDomain(ctx, q, rows)
}

func (r *TaskRepository) Update(ctx context.Context, t *domainTask.Task) error {
	q := getQuerier(ctx, r.queries)

	row, err := q.UpdateTask(ctx, sqlc.UpdateTaskParams{
		ID:                  uuidToPg(t.ID),
		WorkspaceID:         uuidToPg(t.WorkspaceID),
		Title:               t.Title,
		DescriptionMarkdown: t.DescriptionMarkdown,
		DueDate:             dateToPg(t.DueDate),
		Priority:            string(t.Priority),
		FocusOrder:          intPtrToPg(t.FocusOrder),
		ProjectID:           uuidPtrToPg(t.ProjectID),
		ColumnID:            uuidPtrToPg(t.ColumnID),
		ParentID:            uuidPtrToPg(t.ParentID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainTask.ErrTaskNotFound
		}
		return err
	}

	t.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

func (r *TaskRepository) SetFocusOrder(ctx context.Context, id, workspaceID uuid.UUID, focusOrder *int) (*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)

	row, err := q.SetTaskFocus(ctx, sqlc.SetTaskFocusParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
		FocusOrder:  intPtrToPg(focusOrder),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainTask.ErrTaskNotFound
		}
		return nil, err
	}

	return mapSingleTaskRow(ctx, q, row), nil
}

func (r *TaskRepository) CreateTag(ctx context.Context, tag *domainTask.Tag) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.CreateTag(ctx, sqlc.CreateTagParams{
		WorkspaceID: uuidToPg(tag.WorkspaceID),
		Name:        tag.Name,
		ColorHex:    tag.ColorHex,
	})
	if err != nil {
		return err
	}

	tag.ID = pgToUUID(row.ID)
	tag.CreatedAt = pgToTime(row.CreatedAt)
	return nil
}

func (r *TaskRepository) GetTagByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainTask.Tag, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetTagByID(ctx, sqlc.GetTagByIDParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainTask.ErrTagNotFound
		}
		return nil, err
	}

	return mapTagRowToDomain(row), nil
}

func (r *TaskRepository) ListWorkspaceTags(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.Tag, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListWorkspaceTags(ctx, uuidToPg(workspaceID))
	if err != nil {
		return nil, err
	}
	return mapTagRowsToDomain(rows), nil
}

func (r *TaskRepository) ListWorkspaceTagsWithMetrics(ctx context.Context, workspaceID uuid.UUID) ([]*domainTask.TagWithMetrics, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListWorkspaceTagsWithMetrics(ctx, uuidToPg(workspaceID))
	if err != nil {
		return nil, err
	}
	return mapTagMetricsRowsToDomain(rows), nil
}

func (r *TaskRepository) UpdateTag(ctx context.Context, tag *domainTask.Tag) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.UpdateTag(ctx, sqlc.UpdateTagParams{
		ID:          uuidToPg(tag.ID),
		WorkspaceID: uuidToPg(tag.WorkspaceID),
		Name:        tag.Name,
		ColorHex:    tag.ColorHex,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainTask.ErrTagNotFound
		}
		return err
	}
	tag.Name = row.Name
	tag.ColorHex = row.ColorHex
	return nil
}

func (r *TaskRepository) DeleteTag(ctx context.Context, id, workspaceID uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.DeleteTag(ctx, sqlc.DeleteTagParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
	})
}

func (r *TaskRepository) DeleteUnusedTags(ctx context.Context, workspaceID uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.DeleteUnusedTags(ctx, uuidToPg(workspaceID))
}

func (r *TaskRepository) AddTagToTask(ctx context.Context, taskID, tagID uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.AddTagToTask(ctx, sqlc.AddTagToTaskParams{
		TaskID: uuidToPg(taskID),
		TagID:  uuidToPg(tagID),
	})
}

func (r *TaskRepository) RemoveTagFromTask(ctx context.Context, taskID, tagID uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.RemoveTagFromTask(ctx, sqlc.RemoveTagFromTaskParams{
		TaskID: uuidToPg(taskID),
		TagID:  uuidToPg(tagID),
	})
}

func (r *TaskRepository) GetTaskTags(ctx context.Context, taskID uuid.UUID) ([]*domainTask.Tag, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.GetTaskTags(ctx, uuidToPg(taskID))
	if err != nil {
		return nil, err
	}
	return mapTagRowsToDomain(rows), nil
}

func (r *TaskRepository) ClearTaskTags(ctx context.Context, taskID uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.ClearTaskTags(ctx, uuidToPg(taskID))
}

func (r *TaskRepository) SearchTasks(ctx context.Context, workspaceID uuid.UUID, query string, limit int32) ([]*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.SearchTasksInWorkspace(ctx, sqlc.SearchTasksInWorkspaceParams{
		WorkspaceID: uuidToPg(workspaceID),
		Query:       query,
		LimitCount:  limit,
	})
	if err != nil {
		return nil, err
	}
	return mapTaskRowsToDomain(ctx, q, rows)
}

func (r *TaskRepository) ListArchivedTasks(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID) ([]*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListArchivedTasks(ctx, sqlc.ListArchivedTasksParams{
		WorkspaceID: uuidToPg(workspaceID),
		ProjectID:   uuidPtrToPg(projectID),
	})
	if err != nil {
		return nil, err
	}
	return mapTaskRowsToDomain(ctx, q, rows)
}

func (r *TaskRepository) CountArchivedTasks(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	q := getQuerier(ctx, r.queries)
	return q.CountArchivedTasks(ctx, uuidToPg(workspaceID))
}

func (r *TaskRepository) ArchiveCompletedTasksInColumn(ctx context.Context, columnID uuid.UUID) (int64, error) {
	q := getQuerier(ctx, r.queries)
	return q.ArchiveCompletedTasksInColumn(ctx, uuidToPg(columnID))
}

func (r *TaskRepository) UnarchiveTask(ctx context.Context, taskID, workspaceID, columnID uuid.UUID, projectID *uuid.UUID, position float64) (*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.UnarchiveTask(ctx, sqlc.UnarchiveTaskParams{
		ID:          uuidToPg(taskID),
		WorkspaceID: uuidToPg(workspaceID),
		ColumnID:    uuidToPg(columnID),
		ProjectID:   uuidPtrToPg(projectID),
		Position:    position,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainTask.ErrTaskNotFound
		}
		return nil, err
	}
	return mapSingleTaskRow(ctx, q, row), nil
}

func (r *TaskRepository) ArchiveCompletedTasksBatchSQL(ctx context.Context, workspaceID, columnID uuid.UUID, olderThanHours float64) (int64, error) {
	q := getQuerier(ctx, r.queries)
	return q.ArchiveCompletedTasksBatchSQL(ctx, sqlc.ArchiveCompletedTasksBatchSQLParams{
		WorkspaceID: uuidToPg(workspaceID),
		ColumnID:    uuidToPg(columnID),
		Column3:     fmt.Sprintf("%f", olderThanHours),
	})
}

func (r *TaskRepository) CreateActivity(ctx context.Context, activity *domainTask.TaskActivity) error {
	q := getQuerier(ctx, r.queries)
	payloadBytes, err := json.Marshal(activity.Payload)
	if err != nil {
		return err
	}
	_, err = q.CreateTaskActivity(ctx, sqlc.CreateTaskActivityParams{
		ID:           uuidToPg(activity.ID),
		TaskID:       uuidToPg(activity.TaskID),
		UserID:       uuidPtrToPg(activity.UserID),
		ActivityType: activity.ActivityType,
		Payload:      payloadBytes,
		CreatedAt:    timeToPg(activity.CreatedAt),
	})
	return err
}

func (r *TaskRepository) ListActivitiesByTaskID(ctx context.Context, taskID uuid.UUID) ([]*domainTask.TaskActivity, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListTaskActivitiesByTaskID(ctx, uuidToPg(taskID))
	if err != nil {
		return nil, err
	}
	activities := make([]*domainTask.TaskActivity, len(rows))
	for i, row := range rows {
		activities[i] = mapActivityRowToDomain(row)
	}
	return activities, nil
}

