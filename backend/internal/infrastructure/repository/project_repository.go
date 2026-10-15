package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
	"takt/internal/infrastructure/db/sqlc"
)

// ProjectRepository handles project persistence.
type ProjectRepository struct {
	queries *sqlc.Queries
}

func NewProjectRepository(queries *sqlc.Queries) *ProjectRepository {
	return &ProjectRepository{queries: queries}
}

func (r *ProjectRepository) Create(ctx context.Context, p *domainProject.Project) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.CreateProject(ctx, sqlc.CreateProjectParams{
		WorkspaceID: uuidToPg(p.WorkspaceID),
		Name:        p.Name,
		Description: p.Description,
		Color:       p.Color,
	})
	if err != nil {
		return err
	}

	p.ID = pgToUUID(row.ID)
	p.CreatedAt = pgToTime(row.CreatedAt)
	p.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainProject.Project, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetProjectByID(ctx, sqlc.GetProjectByIDParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainProject.ErrProjectNotFound
		}
		return nil, err
	}

	return mapProjectRowToDomain(row), nil
}

func (r *ProjectRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*domainProject.Project, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListProjectsByWorkspace(ctx, uuidToPg(workspaceID))
	if err != nil {
		return nil, err
	}

	res := make([]*domainProject.Project, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapProjectRowToDomain(row))
	}
	return res, nil
}

// SearchProjects finds projects matching the query in name or description.
func (r *ProjectRepository) SearchProjects(ctx context.Context, workspaceID uuid.UUID, query string, limit int32) ([]*domainProject.Project, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.SearchProjectsInWorkspace(ctx, sqlc.SearchProjectsInWorkspaceParams{
		WorkspaceID: uuidToPg(workspaceID),
		Query:       query,
		LimitCount:  limit,
	})
	if err != nil {
		return nil, err
	}

	res := make([]*domainProject.Project, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapProjectRowToDomain(row))
	}
	return res, nil
}

func (r *ProjectRepository) Update(ctx context.Context, p *domainProject.Project) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.UpdateProject(ctx, sqlc.UpdateProjectParams{
		ID:          uuidToPg(p.ID),
		WorkspaceID: uuidToPg(p.WorkspaceID),
		Name:        p.Name,
		Description: p.Description,
		Color:       p.Color,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainProject.ErrProjectNotFound
		}
		return err
	}

	p.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id, workspaceID uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.DeleteProject(ctx, sqlc.DeleteProjectParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
	})
}

// BoardRepository handles board persistence.
type BoardRepository struct {
	queries *sqlc.Queries
}

func NewBoardRepository(queries *sqlc.Queries) *BoardRepository {
	return &BoardRepository{queries: queries}
}

func (r *BoardRepository) Create(ctx context.Context, b *domainProject.Board) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ProjectID: uuidToPg(b.ProjectID),
		Name:      b.Name,
	})
	if err != nil {
		return err
	}

	b.ID = pgToUUID(row.ID)
	b.CreatedAt = pgToTime(row.CreatedAt)
	b.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

func (r *BoardRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainProject.Board, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetBoardByID(ctx, uuidToPg(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainProject.ErrBoardNotFound
		}
		return nil, err
	}

	return mapBoardRowToDomain(row), nil
}

func (r *BoardRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*domainProject.Board, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListBoardsByProject(ctx, uuidToPg(projectID))
	if err != nil {
		return nil, err
	}

	res := make([]*domainProject.Board, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapBoardRowToDomain(row))
	}
	return res, nil
}

func (r *BoardRepository) Update(ctx context.Context, b *domainProject.Board) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.UpdateBoard(ctx, sqlc.UpdateBoardParams{
		ID:   uuidToPg(b.ID),
		Name: b.Name,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainProject.ErrBoardNotFound
		}
		return err
	}

	b.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

func (r *BoardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.DeleteBoard(ctx, uuidToPg(id))
}

// ColumnRepository handles column persistence.
type ColumnRepository struct {
	queries *sqlc.Queries
}

func NewColumnRepository(queries *sqlc.Queries) *ColumnRepository {
	return &ColumnRepository{queries: queries}
}

func (r *ColumnRepository) Create(ctx context.Context, c *domainProject.Column) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.CreateColumn(ctx, sqlc.CreateColumnParams{
		BoardID:        uuidToPg(c.BoardID),
		Name:           c.Name,
		Position:       c.Position,
		StatusCategory: string(c.StatusCategory),
	})
	if err != nil {
		return err
	}

	c.ID = pgToUUID(row.ID)
	c.CreatedAt = pgToTime(row.CreatedAt)
	c.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

func (r *ColumnRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainProject.Column, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetColumnByID(ctx, uuidToPg(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainProject.ErrColumnNotFound
		}
		return nil, err
	}

	return mapColumnRowToDomain(row), nil
}

func (r *ColumnRepository) ListByBoard(ctx context.Context, boardID uuid.UUID) ([]*domainProject.Column, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListColumnsByBoard(ctx, uuidToPg(boardID))
	if err != nil {
		return nil, err
	}

	res := make([]*domainProject.Column, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapColumnRowToDomain(row))
	}
	return res, nil
}

func (r *ColumnRepository) Update(ctx context.Context, c *domainProject.Column) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.UpdateColumn(ctx, sqlc.UpdateColumnParams{
		ID:             uuidToPg(c.ID),
		Name:           c.Name,
		Position:       c.Position,
		StatusCategory: string(c.StatusCategory),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainProject.ErrColumnNotFound
		}
		return err
	}

	c.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

func (r *ColumnRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.DeleteColumn(ctx, uuidToPg(id))
}

func (r *ColumnRepository) CountTasksInColumn(ctx context.Context, columnID uuid.UUID) (int64, error) {
	q := getQuerier(ctx, r.queries)
	return q.CountTasksInColumn(ctx, uuidToPg(columnID))
}

// TaskMoveRepository handles task movement across columns and project boundaries.
type TaskMoveRepository struct {
	queries *sqlc.Queries
}

func NewTaskMoveRepository(queries *sqlc.Queries) *TaskMoveRepository {
	return &TaskMoveRepository{queries: queries}
}

func (r *TaskMoveRepository) MoveTaskToColumnAndPosition(
	ctx context.Context,
	taskID, workspaceID uuid.UUID,
	columnID, projectID *uuid.UUID,
	position float64,
) (*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)

	var colPg pgtype.UUID
	if columnID != nil {
		colPg = uuidToPg(*columnID)
	}

	var projPg pgtype.UUID
	if projectID != nil {
		projPg = uuidToPg(*projectID)
	}

	row, err := q.MoveTaskToColumnAndPosition(ctx, sqlc.MoveTaskToColumnAndPositionParams{
		ColumnID:    colPg,
		ProjectID:   projPg,
		Position:    position,
		ID:          uuidToPg(taskID),
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

func (r *TaskMoveRepository) ResetProjectTasksToInbox(ctx context.Context, projectID, workspaceID uuid.UUID) error {
	q := getQuerier(ctx, r.queries)
	return q.ResetProjectTasksToInbox(ctx, sqlc.ResetProjectTasksToInboxParams{
		ProjectID:   uuidToPg(projectID),
		WorkspaceID: uuidToPg(workspaceID),
	})
}

func (r *TaskMoveRepository) ListTasksByProject(ctx context.Context, projectID, workspaceID uuid.UUID) ([]*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListTasksByProject(ctx, sqlc.ListTasksByProjectParams{
		ProjectID:   uuidToPg(projectID),
		WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		return nil, err
	}

	res := make([]*domainTask.Task, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapTaskRowToDomain(row))
	}
	return res, nil
}

func (r *TaskMoveRepository) ListTasksByColumn(ctx context.Context, columnID, workspaceID uuid.UUID) ([]*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListTasksByColumn(ctx, sqlc.ListTasksByColumnParams{
		ColumnID:    uuidToPg(columnID),
		WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		return nil, err
	}

	res := make([]*domainTask.Task, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapTaskRowToDomain(row))
	}
	return res, nil
}

// mappers
func mapProjectRowToDomain(row sqlc.Project) *domainProject.Project {
	return &domainProject.Project{
		ID: pgToUUID(row.ID), WorkspaceID: pgToUUID(row.WorkspaceID),
		Name: row.Name, Description: row.Description, Color: row.Color,
		CreatedAt: pgToTime(row.CreatedAt), UpdatedAt: pgToTime(row.UpdatedAt),
	}
}

func mapBoardRowToDomain(row sqlc.Board) *domainProject.Board {
	return &domainProject.Board{
		ID: pgToUUID(row.ID), ProjectID: pgToUUID(row.ProjectID),
		Name: row.Name, CreatedAt: pgToTime(row.CreatedAt), UpdatedAt: pgToTime(row.UpdatedAt),
	}
}

func mapColumnRowToDomain(row sqlc.Column) *domainProject.Column {
	return &domainProject.Column{
		ID: pgToUUID(row.ID), BoardID: pgToUUID(row.BoardID),
		Name: row.Name, Position: row.Position,
		StatusCategory: domainProject.StatusCategory(row.StatusCategory),
		CreatedAt: pgToTime(row.CreatedAt), UpdatedAt: pgToTime(row.UpdatedAt),
	}
}

