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

	var projectID pgtype.UUID
	if t.ProjectID != nil {
		projectID = uuidToPg(*t.ProjectID)
	}

	var columnID pgtype.UUID
	if t.ColumnID != nil {
		columnID = uuidToPg(*t.ColumnID)
	}

	var completedAt pgtype.Timestamptz
	if t.CompletedAt != nil {
		completedAt = timeToPg(*t.CompletedAt)
	}

	var focusOrder pgtype.Int4
	if t.FocusOrder != nil {
		focusOrder = pgtype.Int4{Int32: int32(*t.FocusOrder), Valid: true}
	}

	row, err := q.CreateTask(ctx, sqlc.CreateTaskParams{
		WorkspaceID:         uuidToPg(t.WorkspaceID),
		ProjectID:           projectID,
		ColumnID:            columnID,
		Title:               t.Title,
		DescriptionMarkdown: t.DescriptionMarkdown,
		IsCompleted:         t.IsCompleted,
		CompletedAt:         completedAt,
		DueDate:             dateToPg(t.DueDate),
		Priority:            string(t.Priority),
		FocusOrder:          focusOrder,
		Position:            float64(t.Position),
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

	return mapTaskRowToDomain(row), nil
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

	return mapTaskRowToDomain(row), nil
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

func mapTaskRowToDomain(row sqlc.Task) *domainTask.Task {
	var projectID *uuid.UUID
	if row.ProjectID.Valid {
		pid := uuid.UUID(row.ProjectID.Bytes)
		projectID = &pid
	}

	var columnID *uuid.UUID
	if row.ColumnID.Valid {
		cid := uuid.UUID(row.ColumnID.Bytes)
		columnID = &cid
	}

	var focusOrder *int
	if row.FocusOrder.Valid {
		fo := int(row.FocusOrder.Int32)
		focusOrder = &fo
	}

	return &domainTask.Task{
		ID:                  pgToUUID(row.ID),
		WorkspaceID:         pgToUUID(row.WorkspaceID),
		ProjectID:           projectID,
		ColumnID:            columnID,
		Title:               row.Title,
		DescriptionMarkdown: row.DescriptionMarkdown,
		IsCompleted:         row.IsCompleted,
		CompletedAt:         pgToTimePtr(row.CompletedAt),
		DueDate:             pgToDatePtr(row.DueDate),
		Priority:            domainTask.Priority(row.Priority),
		FocusOrder:          focusOrder,
		Position:            int(row.Position),
		CreatedAt:           pgToTime(row.CreatedAt),
		UpdatedAt:           pgToTime(row.UpdatedAt),
	}
}
