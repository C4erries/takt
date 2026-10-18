package repository

import (
	"context"
	"errors"
	"strings"
	"time"

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
	_ appAutomation.TaskReader           = (*AutomationRepository)(nil)
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
	var ruleDef *domainAutomation.RuleDefinition
	if len(m.RuleDefinition) > 0 {
		ruleDef, _ = domainAutomation.RuleDefinitionFromJSON(m.RuleDefinition)
	}
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
		RuleDefinition: ruleDef,
		CreatedAt:      pgToTime(m.CreatedAt),
		UpdatedAt:      pgToTime(m.UpdatedAt),
	}
}

func getRuleBytes(r *domainAutomation.RuleDefinition) []byte {
	if r != nil {
		if b, err := r.ToJSON(); err == nil {
			return b
		}
	}
	return []byte("{}")
}

func toDomainAutomations(rows []sqlc.Automation) []*domainAutomation.Automation {
	res := make([]*domainAutomation.Automation, len(rows))
	for i, row := range rows {
		res[i] = toDomainAutomation(row)
	}
	return res
}

func (r *AutomationRepository) ListWorkspaceAutomations(ctx context.Context, wsID uuid.UUID) ([]*domainAutomation.Automation, error) {
	rows, err := getQuerier(ctx, r.queries).ListWorkspaceAutomations(ctx, uuidToPg(wsID))
	if err != nil {
		return nil, err
	}
	return toDomainAutomations(rows), nil
}

func (r *AutomationRepository) GetAutomationByID(ctx context.Context, id, wsID uuid.UUID) (*domainAutomation.Automation, error) {
	row, err := getQuerier(ctx, r.queries).GetAutomationByID(ctx, sqlc.GetAutomationByIDParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(wsID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainAutomation.ErrAutomationNotFound
		}
		return nil, err
	}
	return toDomainAutomation(row), nil
}

func (r *AutomationRepository) ToggleAutomation(ctx context.Context, id, wsID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error) {
	row, err := getQuerier(ctx, r.queries).ToggleAutomation(ctx, sqlc.ToggleAutomationParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(wsID),
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
	row, err := getQuerier(ctx, r.queries).CreateAutomation(ctx, sqlc.CreateAutomationParams{
		WorkspaceID:    uuidToPg(a.WorkspaceID),
		BoardID:        uuidPtrToPg(a.BoardID),
		ColumnID:       uuidPtrToPg(a.ColumnID),
		Name:           a.Name,
		TriggerType:    a.TriggerType,
		ConditionsJson: a.ConditionsJSON,
		ActionsJson:    a.ActionsJSON,
		IsEnabled:      a.IsEnabled,
		RuleDefinition: getRuleBytes(a.RuleDefinition),
	})
	if err != nil {
		return err
	}
	a.ID = pgToUUID(row.ID)
	a.CreatedAt = pgToTime(row.CreatedAt)
	a.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

func (r *AutomationRepository) UpdateAutomation(ctx context.Context, a *domainAutomation.Automation) error {
	row, err := getQuerier(ctx, r.queries).UpdateAutomation(ctx, sqlc.UpdateAutomationParams{
		ID:             uuidToPg(a.ID),
		WorkspaceID:    uuidToPg(a.WorkspaceID),
		Name:           a.Name,
		TriggerType:    a.TriggerType,
		ConditionsJson: a.ConditionsJSON,
		ActionsJson:    a.ActionsJSON,
		IsEnabled:      a.IsEnabled,
		RuleDefinition: getRuleBytes(a.RuleDefinition),
		BoardID:        uuidPtrToPg(a.BoardID),
		ColumnID:       uuidPtrToPg(a.ColumnID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainAutomation.ErrAutomationNotFound
		}
		return err
	}
	a.BoardID = pgToUUIDPtr(row.BoardID)
	a.ColumnID = pgToUUIDPtr(row.ColumnID)
	a.UpdatedAt = pgToTime(row.UpdatedAt)
	return nil
}

func (r *AutomationRepository) DeleteAutomation(ctx context.Context, id, wsID uuid.UUID) error {
	return getQuerier(ctx, r.queries).DeleteAutomation(ctx, sqlc.DeleteAutomationParams{
		ID:          uuidToPg(id),
		WorkspaceID: uuidToPg(wsID),
	})
}

func (r *AutomationRepository) ListEnabledAutomationsByTrigger(ctx context.Context, wsID uuid.UUID, trig string) ([]*domainAutomation.Automation, error) {
	rows, err := getQuerier(ctx, r.queries).ListEnabledAutomationsByTrigger(ctx, sqlc.ListEnabledAutomationsByTriggerParams{
		WorkspaceID: uuidToPg(wsID),
		TriggerType: trig,
	})
	if err != nil {
		return nil, err
	}
	return toDomainAutomations(rows), nil
}

func (r *AutomationRepository) ListAutomationsByColumn(ctx context.Context, wsID, colID uuid.UUID) ([]*domainAutomation.Automation, error) {
	rows, err := getQuerier(ctx, r.queries).ListAutomationsByColumn(ctx, sqlc.ListAutomationsByColumnParams{
		ColumnID:    uuidToPg(colID),
		WorkspaceID: uuidToPg(wsID),
	})
	if err != nil {
		return nil, err
	}
	return toDomainAutomations(rows), nil
}

func (r *AutomationRepository) ListAllEnabledTimerAutomations(ctx context.Context) ([]*domainAutomation.Automation, error) {
	rows, err := getQuerier(ctx, r.queries).ListAllEnabledTimerAutomations(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainAutomations(rows), nil
}

func (r *AutomationRepository) CreateColumnAutomation(ctx context.Context, a *domainAutomation.Automation) error {
	var colPg pgtype.UUID
	if a.ColumnID != nil {
		colPg = uuidToPg(*a.ColumnID)
	}
	row, err := getQuerier(ctx, r.queries).CreateColumnAutomation(ctx, sqlc.CreateColumnAutomationParams{
		WorkspaceID:    uuidToPg(a.WorkspaceID),
		BoardID:        uuidPtrToPg(a.BoardID),
		ColumnID:       colPg,
		Name:           a.Name,
		TriggerType:    a.TriggerType,
		ConditionsJson: a.ConditionsJSON,
		ActionsJson:    a.ActionsJSON,
		IsEnabled:      a.IsEnabled,
		RuleDefinition: getRuleBytes(a.RuleDefinition),
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

func (r *AutomationRepository) MoveTaskToColumn(ctx context.Context, taskID, wsID uuid.UUID, colID, prjID *uuid.UUID, pos float64) error {
	_, err := getQuerier(ctx, r.queries).MoveTaskToColumnAndPosition(ctx, sqlc.MoveTaskToColumnAndPositionParams{
		ID:          uuidToPg(taskID),
		WorkspaceID: uuidToPg(wsID),
		ColumnID:    uuidPtrToPg(colID),
		ProjectID:   uuidPtrToPg(prjID),
		Position:    pos,
	})
	return err
}

func (r *AutomationRepository) SetTaskFocus(ctx context.Context, taskID, wsID uuid.UUID, focusOrder *int) error {
	var fo pgtype.Int4
	if focusOrder != nil {
		fo = pgtype.Int4{Int32: int32(*focusOrder), Valid: true}
	}
	_, err := getQuerier(ctx, r.queries).SetTaskFocus(ctx, sqlc.SetTaskFocusParams{
		ID:          uuidToPg(taskID),
		WorkspaceID: uuidToPg(wsID),
		FocusOrder:  fo,
	})
	return err
}

func (r *AutomationRepository) SetTaskPriority(ctx context.Context, taskID, wsID uuid.UUID, priority string) error {
	_, err := getQuerier(ctx, r.queries).SetTaskPriority(ctx, sqlc.SetTaskPriorityParams{
		ID:          uuidToPg(taskID),
		WorkspaceID: uuidToPg(wsID),
		Priority:    priority,
	})
	return err
}

func (r *AutomationRepository) resolveTagID(ctx context.Context, wsID uuid.UUID, tagID *uuid.UUID, tagName string, create bool) (pgtype.UUID, error) {
	if tagID != nil {
		return uuidToPg(*tagID), nil
	}
	if tagName == "" {
		return pgtype.UUID{}, nil
	}
	q := getQuerier(ctx, r.queries)
	tagRow, err := q.GetTagByName(ctx, sqlc.GetTagByNameParams{
		WorkspaceID: uuidToPg(wsID),
		Lower:       strings.ToLower(tagName),
	})
	if err == nil {
		return tagRow.ID, nil
	}
	if !create {
		return pgtype.UUID{}, err
	}
	newTag, err := q.CreateTag(ctx, sqlc.CreateTagParams{
		WorkspaceID: uuidToPg(wsID),
		Name:        tagName,
		ColorHex:    "#6366f1",
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	return newTag.ID, nil
}

func (r *AutomationRepository) AddTagToTask(ctx context.Context, taskID, wsID uuid.UUID, tagID *uuid.UUID, tagName string) error {
	tid, err := r.resolveTagID(ctx, wsID, tagID, tagName, true)
	if err != nil || !tid.Valid {
		return err
	}
	return getQuerier(ctx, r.queries).AddTagToTask(ctx, sqlc.AddTagToTaskParams{TaskID: uuidToPg(taskID), TagID: tid})
}

func (r *AutomationRepository) RemoveTagFromTask(ctx context.Context, taskID, wsID uuid.UUID, tagID *uuid.UUID, tagName string) error {
	tid, err := r.resolveTagID(ctx, wsID, tagID, tagName, false)
	if err != nil || !tid.Valid {
		return nil
	}
	return getQuerier(ctx, r.queries).RemoveTagFromTask(ctx, sqlc.RemoveTagFromTaskParams{TaskID: uuidToPg(taskID), TagID: tid})
}

func (r *AutomationRepository) SetTaskDueDate(ctx context.Context, taskID, wsID uuid.UUID, dueDate *time.Time) error {
	var duePg pgtype.Date
	if dueDate != nil {
		duePg = pgtype.Date{Time: *dueDate, Valid: true}
	}
	_, err := getQuerier(ctx, r.queries).SetTaskDueDate(ctx, sqlc.SetTaskDueDateParams{
		ID:          uuidToPg(taskID),
		WorkspaceID: uuidToPg(wsID),
		DueDate:     duePg,
	})
	return err
}

func (r *AutomationRepository) SetTaskCompleted(ctx context.Context, taskID, wsID uuid.UUID, isCompleted bool) error {
	_, err := getQuerier(ctx, r.queries).SetTaskCompletion(ctx, sqlc.SetTaskCompletionParams{
		ID:          uuidToPg(taskID),
		WorkspaceID: uuidToPg(wsID),
		IsCompleted: isCompleted,
	})
	return err
}

func (r *AutomationRepository) ArchiveTask(ctx context.Context, taskID, wsID uuid.UUID) error {
	return getQuerier(ctx, r.queries).DeleteTask(ctx, sqlc.DeleteTaskParams{
		ID:          uuidToPg(taskID),
		WorkspaceID: uuidToPg(wsID),
	})
}

// TaskReader implementation

func (r *AutomationRepository) ListTasksByColumn(ctx context.Context, wsID, colID uuid.UUID) ([]*appAutomation.TaskInfo, error) {
	rows, err := getQuerier(ctx, r.queries).ListTasksByColumn(ctx, sqlc.ListTasksByColumnParams{
		ColumnID:    uuidToPg(colID),
		WorkspaceID: uuidToPg(wsID),
	})
	if err != nil {
		return nil, err
	}
	result := make([]*appAutomation.TaskInfo, len(rows))
	for i, row := range rows {
		var due *time.Time
		if row.DueDate.Valid {
			due = &row.DueDate.Time
		}
		result[i] = &appAutomation.TaskInfo{
			ID:          pgToUUID(row.ID),
			WorkspaceID: pgToUUID(row.WorkspaceID),
			ProjectID:   pgToUUIDPtr(row.ProjectID),
			ColumnID:    pgToUUIDPtr(row.ColumnID),
			Title:       row.Title,
			Priority:    row.Priority,
			IsCompleted: row.IsCompleted,
			DueDate:     due,
			CreatedAt:   pgToTime(row.CreatedAt),
			UpdatedAt:   pgToTime(row.UpdatedAt),
		}
	}
	return result, nil
}

// ProjectReader implementation

func (r *AutomationRepository) getBoardColumns(ctx context.Context, prjID uuid.UUID) ([]sqlc.Column, error) {
	boards, err := getQuerier(ctx, r.queries).ListBoardsByProject(ctx, uuidToPg(prjID))
	if err != nil || len(boards) == 0 {
		return nil, err
	}
	return getQuerier(ctx, r.queries).ListColumnsByBoard(ctx, boards[0].ID)
}

func (r *AutomationRepository) findColumnByCategory(ctx context.Context, prjID uuid.UUID, cat string) (*uuid.UUID, []sqlc.Column, error) {
	cols, err := r.getBoardColumns(ctx, prjID)
	if err != nil || len(cols) == 0 {
		return nil, nil, err
	}
	for _, c := range cols {
		if c.StatusCategory == cat {
			id := pgToUUID(c.ID)
			return &id, cols, nil
		}
	}
	return nil, cols, nil
}

func (r *AutomationRepository) GetFirstTodoColumn(ctx context.Context, prjID, wsID uuid.UUID) (*uuid.UUID, error) {
	id, cols, err := r.findColumnByCategory(ctx, prjID, "todo")
	if id != nil || err != nil || len(cols) == 0 {
		return id, err
	}
	firstID := pgToUUID(cols[0].ID)
	return &firstID, nil
}

func (r *AutomationRepository) GetDoneColumn(ctx context.Context, prjID, wsID uuid.UUID) (*uuid.UUID, error) {
	id, _, err := r.findColumnByCategory(ctx, prjID, "done")
	return id, err
}
