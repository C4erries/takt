package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

	var focusOrder pgtype.Int4
	if t.FocusOrder != nil {
		focusOrder = pgtype.Int4{Int32: int32(*t.FocusOrder), Valid: true}
	}

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
		FocusOrder:          focusOrder,
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

	tasks := make([]*domainTask.Task, 0, len(rows))
	for _, row := range rows {
		tasks = append(tasks, mapTaskRowToDomain(row))
	}
	if err := loadSubtaskCountsForTasks(ctx, q, tasks); err != nil {
		return nil, err
	}
	return tasks, nil
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

	task := mapTaskRowToDomain(row)
	_ = loadSubtaskCountsForTasks(ctx, q, []*domainTask.Task{task})
	return task, nil
}

func (r *TaskRepository) ListSubtasks(ctx context.Context, parentID uuid.UUID) ([]*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListSubtasksByParent(ctx, uuidToPg(parentID))
	if err != nil {
		return nil, err
	}

	tasks := make([]*domainTask.Task, 0, len(rows))
	for _, row := range rows {
		tasks = append(tasks, mapTaskRowToDomain(row))
	}
	if err := loadSubtaskCountsForTasks(ctx, q, tasks); err != nil {
		return nil, err
	}
	return tasks, nil
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

	task := mapTaskRowToDomain(row)
	_ = loadSubtaskCountsForTasks(ctx, q, []*domainTask.Task{task})
	return task, nil
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

	tasks := make([]*domainTask.Task, 0, len(rows))
	for _, row := range rows {
		tasks = append(tasks, mapTaskRowToDomain(row))
	}
	if err := loadSubtaskCountsForTasks(ctx, q, tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, t *domainTask.Task) error {
	q := getQuerier(ctx, r.queries)

	var focusOrder pgtype.Int4
	if t.FocusOrder != nil {
		focusOrder = pgtype.Int4{Int32: int32(*t.FocusOrder), Valid: true}
	}

	row, err := q.UpdateTask(ctx, sqlc.UpdateTaskParams{
		ID:                  uuidToPg(t.ID),
		WorkspaceID:         uuidToPg(t.WorkspaceID),
		Title:               t.Title,
		DescriptionMarkdown: t.DescriptionMarkdown,
		DueDate:             dateToPg(t.DueDate),
		Priority:            string(t.Priority),
		FocusOrder:          focusOrder,
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

	var fo pgtype.Int4
	if focusOrder != nil {
		fo = pgtype.Int4{Int32: int32(*focusOrder), Valid: true}
	}

	row, err := q.SetTaskFocus(ctx, sqlc.SetTaskFocusParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
		FocusOrder:  fo,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainTask.ErrTaskNotFound
		}
		return nil, err
	}

	return mapTaskRowToDomain(row), nil
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

	tags := make([]*domainTask.Tag, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, mapTagRowToDomain(row))
	}
	return tags, nil
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

	tags := make([]*domainTask.Tag, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, mapTagRowToDomain(row))
	}
	return tags, nil
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

	tasks := make([]*domainTask.Task, 0, len(rows))
	for _, row := range rows {
		tasks = append(tasks, mapTaskRowToDomain(row))
	}
	return tasks, nil
}
