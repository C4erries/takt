package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	appAutomation "takt/internal/application/automation"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/infrastructure/db/sqlc"
)

var (
	_ appAutomation.AutomationRepository = (*AutomationRepository)(nil)
	_ appAutomation.TaskModifier         = (*AutomationRepository)(nil)
	_ appAutomation.ProjectReader        = (*AutomationRepository)(nil)
)

// AutomationRepository handles database access for automation rules and task automation actions.
type AutomationRepository struct {
	queries *sqlc.Queries
}

// NewAutomationRepository constructs a new AutomationRepository.
func NewAutomationRepository(queries *sqlc.Queries) *AutomationRepository {
	return &AutomationRepository{queries: queries}
}

func toDomainAutomation(m sqlc.Automation) *domainAutomation.Automation {
	return &domainAutomation.Automation{
		ID:             pgToUUID(m.ID),
		WorkspaceID:    pgToUUID(m.WorkspaceID),
		BoardID:        pgToUUIDPtr(m.BoardID),
		ColumnID:       pgToUUIDPtr(m.ColumnID),
		Name:           m.Name,
		TriggerType:    m.TriggerType,
		ConditionsJSON: m.ConditionsJson,
		ActionsJSON:    m.ActionsJson,
		IsEnabled:      m.IsEnabled,
		CreatedAt:      pgToTime(m.CreatedAt),
		UpdatedAt:      pgToTime(m.UpdatedAt),
	}
}

func (r *AutomationRepository) ListWorkspaceAutomations(ctx context.Context, workspaceID uuid.UUID) ([]*domainAutomation.Automation, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListWorkspaceAutomations(ctx, uuidToPg(workspaceID))
	if err != nil {
		return nil, err
	}

	result := make([]*domainAutomation.Automation, len(rows))
	for i, row := range rows {
		result[i] = toDomainAutomation(row)
	}
	return result, nil
}

func (r *AutomationRepository) GetAutomationByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainAutomation.Automation, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.GetAutomationByID(ctx, sqlc.GetAutomationByIDParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainAutomation.ErrAutomationNotFound
		}
		return nil, err
	}

	return toDomainAutomation(row), nil
}

func (r *AutomationRepository) ToggleAutomation(ctx context.Context, id, workspaceID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error) {
	q := getQuerier(ctx, r.queries)
	row, err := q.ToggleAutomation(ctx, sqlc.ToggleAutomationParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(workspaceID),
		IsEnabled:   isEnabled,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainAutomation.ErrAutomationNotFound
		}
		return nil, err
	}

	return toDomainAutomation(row), nil
}

func (r *AutomationRepository) CreateAutomation(ctx context.Context, a *domainAutomation.Automation) error {
	q := getQuerier(ctx, r.queries)
	row, err := q.CreateAutomation(ctx, sqlc.CreateAutomationParams{
		WorkspaceID:    uuidToPg(a.WorkspaceID),
		BoardID:        uuidPtrToPg(a.BoardID),
		Name:           a.Name,
		TriggerType:    a.TriggerType,
		ConditionsJson: a.ConditionsJSON,
		ActionsJson:    a.ActionsJSON,
		IsEnabled:      a.IsEnabled,
	})
	if err != nil {
		return err
	}

	a.ID = pgToUUID(row.ID)
	a.CreatedAt = pgToTime(row.CreatedAt)
	a.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

func (r *AutomationRepository) ListEnabledAutomationsByTrigger(ctx context.Context, workspaceID uuid.UUID, triggerType string) ([]*domainAutomation.Automation, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListEnabledAutomationsByTrigger(ctx, sqlc.ListEnabledAutomationsByTriggerParams{
		WorkspaceID: uuidToPg(workspaceID),
		TriggerType: triggerType,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*domainAutomation.Automation, len(rows))
	for i, row := range rows {
		result[i] = toDomainAutomation(row)
	}
	return result, nil
}

func (r *AutomationRepository) ListAutomationsByColumn(ctx context.Context, workspaceID, columnID uuid.UUID) ([]*domainAutomation.Automation, error) {
	q := getQuerier(ctx, r.queries)
	rows, err := q.ListAutomationsByColumn(ctx, sqlc.ListAutomationsByColumnParams{
		ColumnID:    uuidToPg(columnID),
		WorkspaceID: uuidToPg(workspaceID),
	})
	if err != nil {
		return nil, err
	}

	result := make([]*domainAutomation.Automation, len(rows))
	for i, row := range rows {
		result[i] = toDomainAutomation(row)
	}
	return result, nil
}

func (r *AutomationRepository) CreateColumnAutomation(ctx context.Context, a *domainAutomation.Automation) error {
	q := getQuerier(ctx, r.queries)
	var colPg pgtype.UUID
	if a.ColumnID != nil {
		colPg = uuidToPg(*a.ColumnID)
	}
	row, err := q.CreateColumnAutomation(ctx, sqlc.CreateColumnAutomationParams{
		WorkspaceID:    uuidToPg(a.WorkspaceID),
		BoardID:        uuidPtrToPg(a.BoardID),
		ColumnID:       colPg,
		Name:           a.Name,
		TriggerType:    a.TriggerType,
		ConditionsJson: a.ConditionsJSON,
		ActionsJson:    a.ActionsJSON,
		IsEnabled:      a.IsEnabled,
	})
	if err != nil {
		return err
	}

	a.ID = pgToUUID(row.ID)
	a.CreatedAt = pgToTime(row.CreatedAt)
	a.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

// TaskModifier implementation

func (r *AutomationRepository) MoveTaskToColumn(ctx context.Context, taskID, workspaceID uuid.UUID, columnID, projectID *uuid.UUID, position float64) error {
	q := getQuerier(ctx, r.queries)
	_, err := q.MoveTaskToColumnAndPosition(ctx, sqlc.MoveTaskToColumnAndPositionParams{
		ID:          uuidToPg(taskID),
		WorkspaceID: uuidToPg(workspaceID),
		ColumnID:    uuidPtrToPg(columnID),
		ProjectID:   uuidPtrToPg(projectID),
		Position:    position,
	})
	return err
}

func (r *AutomationRepository) SetTaskFocus(ctx context.Context, taskID, workspaceID uuid.UUID, focusOrder *int) error {
	q := getQuerier(ctx, r.queries)
	var fo pgtype.Int4
	if focusOrder != nil {
		fo = pgtype.Int4{Int32: int32(*focusOrder), Valid: true}
	}
	_, err := q.SetTaskFocus(ctx, sqlc.SetTaskFocusParams{
		ID:          uuidToPg(taskID),
		WorkspaceID: uuidToPg(workspaceID),
		FocusOrder:  fo,
	})
	return err
}

func (r *AutomationRepository) SetTaskPriority(ctx context.Context, taskID, workspaceID uuid.UUID, priority string) error {
	q := getQuerier(ctx, r.queries)
	_, err := q.SetTaskPriority(ctx, sqlc.SetTaskPriorityParams{
		ID:          uuidToPg(taskID),
		WorkspaceID: uuidToPg(workspaceID),
		Priority:    priority,
	})
	return err
}

// ProjectReader implementation

func (r *AutomationRepository) GetFirstTodoColumn(ctx context.Context, projectID, workspaceID uuid.UUID) (*uuid.UUID, error) {
	q := getQuerier(ctx, r.queries)
	boards, err := q.ListBoardsByProject(ctx, uuidToPg(projectID))
	if err != nil || len(boards) == 0 {
		return nil, err
	}

	cols, err := q.ListColumnsByBoard(ctx, boards[0].ID)
	if err != nil || len(cols) == 0 {
		return nil, err
	}

	for _, c := range cols {
		if c.StatusCategory == "todo" {
			id := pgToUUID(c.ID)
			return &id, nil
		}
	}

	firstID := pgToUUID(cols[0].ID)
	return &firstID, nil
}

func (r *AutomationRepository) GetDoneColumn(ctx context.Context, projectID, workspaceID uuid.UUID) (*uuid.UUID, error) {
	q := getQuerier(ctx, r.queries)
	boards, err := q.ListBoardsByProject(ctx, uuidToPg(projectID))
	if err != nil || len(boards) == 0 {
		return nil, err
	}

	cols, err := q.ListColumnsByBoard(ctx, boards[0].ID)
	if err != nil || len(cols) == 0 {
		return nil, err
	}

	for _, c := range cols {
		if c.StatusCategory == "done" {
			id := pgToUUID(c.ID)
			return &id, nil
		}
	}

	return nil, nil
}

