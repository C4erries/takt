package project

import (
	"github.com/google/uuid"
	domainProject "takt/internal/domain/project"
)

type CreateProjectInput struct {
	WorkspaceID uuid.UUID
	Name        string
	Description string
	Color       string
}

type UpdateProjectInput struct {
	ProjectID   uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Description string
	Color       string
}

type CreateColumnInput struct {
	WorkspaceID    uuid.UUID
	BoardID        uuid.UUID
	Name           string
	Position       float64
	StatusCategory domainProject.StatusCategory
}

type UpdateColumnInput struct {
	WorkspaceID    uuid.UUID
	BoardID        *uuid.UUID
	ColumnID       uuid.UUID
	Name           string
	Position       float64
	StatusCategory domainProject.StatusCategory
	IsInbox        *bool
}

type MoveTaskInput struct {
	WorkspaceID uuid.UUID
	TaskID      uuid.UUID
	ColumnID    uuid.UUID
	Position    float64
	ProjectID   *uuid.UUID
}

