package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
	"takt/internal/infrastructure/db"
	"takt/internal/infrastructure/db/sqlc"
)

func getQuerier(ctx context.Context, defaultQueries *sqlc.Queries) *sqlc.Queries {
	if tx, ok := db.TxFromContext(ctx); ok {
		return defaultQueries.WithTx(tx)
	}
	return defaultQueries
}

func uuidToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: id != uuid.Nil}
}

func pgToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return id.Bytes
}

func uuidPtrToPg(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func pgToUUIDPtr(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	res := uuid.UUID(id.Bytes)
	return &res
}


func timeToPg(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func timePtrToPg(t *time.Time) pgtype.Timestamptz {
	if t == nil || t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func pgToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func pgToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	res := t.Time
	return &res
}

func dateToPg(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func pgToDatePtr(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	res := d.Time
	return &res
}

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
		Name: row.Name, Position: row.Position, CreatedAt: pgToTime(row.CreatedAt), UpdatedAt: pgToTime(row.UpdatedAt),
	}
}

func mapColumnRowToDomain(row sqlc.Column) *domainProject.Column {
	return &domainProject.Column{
		ID:             pgToUUID(row.ID),
		BoardID:        pgToUUID(row.BoardID),
		Name:           row.Name,
		Position:       row.Position,
		StatusCategory: domainProject.StatusCategory(row.StatusCategory),
		IsInbox:        row.IsInbox,
		CreatedAt:      pgToTime(row.CreatedAt),
		UpdatedAt:      pgToTime(row.UpdatedAt),
	}
}

func loadTagsForTasks(ctx context.Context, q *sqlc.Queries, tasks []*domainTask.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	taskIDs := make([]pgtype.UUID, len(tasks))
	taskMap := make(map[uuid.UUID]*domainTask.Task, len(tasks))
	for i, t := range tasks {
		taskIDs[i] = uuidToPg(t.ID)
		taskMap[t.ID] = t
	}
	rows, err := q.ListTagsForTasks(ctx, taskIDs)
	if err != nil {
		return err
	}
	for _, row := range rows {
		tID := pgToUUID(row.TaskID)
		if t, ok := taskMap[tID]; ok {
			t.Tags = append(t.Tags, &domainTask.Tag{
				ID:          pgToUUID(row.ID),
				WorkspaceID: pgToUUID(row.WorkspaceID),
				Name:        row.Name,
				ColorHex:    row.ColorHex,
				CreatedAt:   pgToTime(row.CreatedAt),
			})
		}
	}
	return nil
}

func pgToIntPtr(v pgtype.Int4) *int {
	if !v.Valid {
		return nil
	}
	res := int(v.Int32)
	return &res
}

func intPtrToPg(v *int) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*v), Valid: true}
}

func mapTagRowToDomain(row sqlc.Tag) *domainTask.Tag {
	return &domainTask.Tag{
		ID:          pgToUUID(row.ID),
		WorkspaceID: pgToUUID(row.WorkspaceID),
		ProjectID:   pgToUUIDPtr(row.ProjectID),
		Name:        row.Name,
		ColorHex:    row.ColorHex,
		CreatedAt:   pgToTime(row.CreatedAt),
	}
}

func mapTaskRowToDomain(row sqlc.Task) *domainTask.Task {
	return &domainTask.Task{
		ID:                  pgToUUID(row.ID),
		WorkspaceID:         pgToUUID(row.WorkspaceID),
		ParentID:            pgToUUIDPtr(row.ParentID),
		ProjectID:           pgToUUIDPtr(row.ProjectID),
		ColumnID:            pgToUUIDPtr(row.ColumnID),
		Title:               row.Title,
		DescriptionMarkdown: row.DescriptionMarkdown,
		IsCompleted:         row.IsCompleted,
		CompletedAt:         pgToTimePtr(row.CompletedAt),
		DueDate:             pgToDatePtr(row.DueDate),
		Priority:            domainTask.Priority(row.Priority),
		FocusOrder:          pgToIntPtr(row.FocusOrder),
		Position:            int(row.Position),
		Tags:                []*domainTask.Tag{},
		IsArchived:          row.IsArchived,
		ArchivedAt:          pgToTimePtr(row.ArchivedAt),
		CreatedAt:           pgToTime(row.CreatedAt),
		UpdatedAt:           pgToTime(row.UpdatedAt),
	}
}

func mapTaskRowsToDomain(ctx context.Context, q *sqlc.Queries, rows []sqlc.Task) ([]*domainTask.Task, error) {
	tasks := make([]*domainTask.Task, 0, len(rows))
	for _, row := range rows {
		tasks = append(tasks, mapTaskRowToDomain(row))
	}
	if err := loadSubtaskCountsForTasks(ctx, q, tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func loadSubtaskCountsForTasks(ctx context.Context, q *sqlc.Queries, tasks []*domainTask.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	taskIDs := make([]pgtype.UUID, len(tasks))
	taskMap := make(map[uuid.UUID]*domainTask.Task, len(tasks))
	for i, t := range tasks {
		taskIDs[i] = uuidToPg(t.ID)
		taskMap[t.ID] = t
	}
	rows, err := q.GetSubtaskCountsByParentIds(ctx, taskIDs)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if !row.ParentID.Valid {
			continue
		}
		pID := pgToUUID(row.ParentID)
		if t, ok := taskMap[pID]; ok {
			t.SetSubtaskCounts(int(row.TotalCount), int(row.CompletedCount))
		}
	}
	return nil
}

func mapSingleTaskRow(ctx context.Context, q *sqlc.Queries, row sqlc.Task) *domainTask.Task {
	task := mapTaskRowToDomain(row)
	_ = loadSubtaskCountsForTasks(ctx, q, []*domainTask.Task{task})
	return task
}

func mapTagRowsToDomain(rows []sqlc.Tag) []*domainTask.Tag {
	tags := make([]*domainTask.Tag, len(rows))
	for i, row := range rows {
		tags[i] = mapTagRowToDomain(row)
	}
	return tags
}

func mapTagMetricsRowsToDomain(rows []sqlc.ListWorkspaceTagsWithMetricsRow) []*domainTask.TagWithMetrics {
	tags := make([]*domainTask.TagWithMetrics, len(rows))
	for i, row := range rows {
		item := domainTask.NewTagWithMetrics(domainTask.Tag{
			ID:          pgToUUID(row.ID),
			WorkspaceID: pgToUUID(row.WorkspaceID),
			ProjectID:   pgToUUIDPtr(row.ProjectID),
			Name:        row.Name,
			ColorHex:    row.ColorHex,
			CreatedAt:   pgToTime(row.CreatedAt),
		}, row.TaskCount)
		tags[i] = &item
	}
	return tags
}

func mapActivityRowToDomain(row sqlc.ListTaskActivitiesByTaskIDRow) *domainTask.TaskActivity {
	var payload map[string]interface{}
	if len(row.Payload) > 0 {
		_ = json.Unmarshal(row.Payload, &payload)
	}
	if payload == nil {
		payload = make(map[string]interface{})
	}
	var userName *string
	if row.UserName.Valid && row.UserName.String != "" {
		userName = &row.UserName.String
	} else if row.UserEmail.Valid && row.UserEmail.String != "" {
		userName = &row.UserEmail.String
	}
	return &domainTask.TaskActivity{
		ID:           pgToUUID(row.ID),
		TaskID:       pgToUUID(row.TaskID),
		UserID:       pgToUUIDPtr(row.UserID),
		UserName:     userName,
		ActivityType: row.ActivityType,
		Payload:      payload,
		CreatedAt:    pgToTime(row.CreatedAt),
	}
}



