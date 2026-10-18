package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	domainProject "takt/internal/domain/project"
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
		Name: row.Name, CreatedAt: pgToTime(row.CreatedAt), UpdatedAt: pgToTime(row.UpdatedAt),
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

