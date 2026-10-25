package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		WorkspaceID: uuidToPg(p.WorkspaceID), Name: p.Name, Description: p.Description, Color: p.Color,
	})
	if err != nil {
		return err
	}

	p.ID, p.CreatedAt, p.UpdatedAt = pgToUUID(row.ID), pgToTime(row.CreatedAt), pgToTime(row.UpdatedAt)
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainProject.Project, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetProjectByID(ctx, sqlc.GetProjectByIDParams{ID: uuidToPg(id), WorkspaceID: uuidToPg(workspaceID)})
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
	res := make([]*domainProject.Project, len(rows))
	for i, row := range rows {
		res[i] = mapProjectRowToDomain(row)
	}
	return res, nil
}

// SearchProjects finds projects matching the query in name or description.
func (r *ProjectRepository) SearchProjects(ctx context.Context, workspaceID uuid.UUID, query string, limit int32) ([]*domainProject.Project, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.SearchProjectsInWorkspace(ctx, sqlc.SearchProjectsInWorkspaceParams{
		WorkspaceID: uuidToPg(workspaceID), Query: query, LimitCount: limit,
	})
	if err != nil {
		return nil, err
	}
	res := make([]*domainProject.Project, len(rows))
	for i, row := range rows {
		res[i] = mapProjectRowToDomain(row)
	}
	return res, nil
}

func (r *ProjectRepository) Update(ctx context.Context, p *domainProject.Project) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.UpdateProject(ctx, sqlc.UpdateProjectParams{
		ID: uuidToPg(p.ID), WorkspaceID: uuidToPg(p.WorkspaceID), Name: p.Name, Description: p.Description, Color: p.Color,
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
	return getQuerier(ctx, r.queries).DeleteProject(ctx, sqlc.DeleteProjectParams{ID: uuidToPg(id), WorkspaceID: uuidToPg(workspaceID)})
}

func (r *ProjectRepository) ColumnExists(ctx context.Context, columnID uuid.UUID) (bool, error) {
	_, err := getQuerier(ctx, r.queries).GetColumnByID(ctx, uuidToPg(columnID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *ProjectRepository) GetInboxColumn(ctx context.Context, projectID, workspaceID uuid.UUID) (*uuid.UUID, error) {
	boards, err := getQuerier(ctx, r.queries).ListBoardsByProject(ctx, uuidToPg(projectID))
	if err != nil || len(boards) == 0 {
		return nil, err
	}
	cols, err := getQuerier(ctx, r.queries).ListColumnsByBoard(ctx, boards[0].ID)
	if err != nil || len(cols) == 0 {
		return nil, err
	}
	for _, c := range cols {
		if c.IsInbox {
			id := pgToUUID(c.ID)
			return &id, nil
		}
	}
	return nil, nil
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
		ProjectID: uuidToPg(b.ProjectID), Name: b.Name, Position: b.Position,
	})
	if err != nil {
		return err
	}

	b.ID, b.CreatedAt, b.UpdatedAt = pgToUUID(row.ID), pgToTime(row.CreatedAt), pgToTime(row.UpdatedAt)
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
	res := make([]*domainProject.Board, len(rows))
	for i, row := range rows {
		res[i] = mapBoardRowToDomain(row)
	}
	return res, nil
}

func (r *BoardRepository) Update(ctx context.Context, b *domainProject.Board) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.UpdateBoard(ctx, sqlc.UpdateBoardParams{
		ID: uuidToPg(b.ID), Name: b.Name, Position: b.Position,
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
	return getQuerier(ctx, r.queries).DeleteBoard(ctx, uuidToPg(id))
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
		IsInbox:        c.IsInbox,
	})
	if err != nil {
		return err
	}

	c.ID, c.CreatedAt, c.UpdatedAt = pgToUUID(row.ID), pgToTime(row.CreatedAt), pgToTime(row.UpdatedAt)
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

func (r *ColumnRepository) ColumnExists(ctx context.Context, columnID uuid.UUID) (bool, error) {
	_, err := getQuerier(ctx, r.queries).GetColumnByID(ctx, uuidToPg(columnID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *ColumnRepository) ListByBoard(ctx context.Context, boardID uuid.UUID) ([]*domainProject.Column, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListColumnsByBoard(ctx, uuidToPg(boardID))
	if err != nil {
		return nil, err
	}
	res := make([]*domainProject.Column, len(rows))
	for i, row := range rows {
		res[i] = mapColumnRowToDomain(row)
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
		IsInbox:        c.IsInbox,
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

func (r *ColumnRepository) ResetBoardInboxColumns(ctx context.Context, boardID uuid.UUID) error {
	return getQuerier(ctx, r.queries).ResetBoardInboxColumns(ctx, uuidToPg(boardID))
}

func (r *ColumnRepository) MoveToBoard(ctx context.Context, columnID, newBoardID uuid.UUID, newPosition float64) error {
	q := getQuerier(ctx, r.queries)
	_, err := q.MoveColumnToBoard(ctx, sqlc.MoveColumnToBoardParams{
		ID: uuidToPg(columnID), BoardID: uuidToPg(newBoardID), Position: newPosition,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainProject.ErrColumnNotFound
		}
		return err
	}
	return nil
}

func (r *ColumnRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return getQuerier(ctx, r.queries).DeleteColumn(ctx, uuidToPg(id))
}

func (r *ColumnRepository) CountTasksInColumn(ctx context.Context, columnID uuid.UUID) (int64, error) {
	return getQuerier(ctx, r.queries).CountTasksInColumn(ctx, uuidToPg(columnID))
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

	colPg := uuidPtrToPg(columnID)
	projPg := uuidPtrToPg(projectID)

	row, err := q.MoveTaskToColumnAndPosition(ctx, sqlc.MoveTaskToColumnAndPositionParams{
		ColumnID: colPg, ProjectID: projPg, Position: position,
		ID: uuidToPg(taskID), WorkspaceID: uuidToPg(workspaceID),
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
	return getQuerier(ctx, r.queries).ResetProjectTasksToInbox(ctx, sqlc.ResetProjectTasksToInboxParams{ProjectID: uuidToPg(projectID), WorkspaceID: uuidToPg(workspaceID)})
}

func (r *TaskMoveRepository) ResetBoardTasksToInbox(ctx context.Context, boardID, workspaceID uuid.UUID) error {
	return getQuerier(ctx, r.queries).ResetBoardTasksToInbox(ctx, sqlc.ResetBoardTasksToInboxParams{BoardID: uuidToPg(boardID), WorkspaceID: uuidToPg(workspaceID)})
}

func (r *TaskMoveRepository) ListTasksByProject(ctx context.Context, projectID, workspaceID uuid.UUID) ([]*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListTasksByProject(ctx, sqlc.ListTasksByProjectParams{
		ProjectID: uuidToPg(projectID), WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*domainTask.Task, len(rows))
	for i, row := range rows {
		res[i] = mapTaskRowToDomain(row)
	}
	if err := loadTagsForTasks(ctx, q, res); err != nil {
		return nil, err
	}
	if err := loadSubtaskCountsForTasks(ctx, q, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *TaskMoveRepository) ListTasksByColumn(ctx context.Context, columnID, workspaceID uuid.UUID) ([]*domainTask.Task, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListTasksByColumn(ctx, sqlc.ListTasksByColumnParams{
		ColumnID: uuidToPg(columnID), WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*domainTask.Task, len(rows))
	for i, row := range rows {
		res[i] = mapTaskRowToDomain(row)
	}
	if err := loadTagsForTasks(ctx, q, res); err != nil {
		return nil, err
	}
	if err := loadSubtaskCountsForTasks(ctx, q, res); err != nil {
		return nil, err
	}
	return res, nil
}

