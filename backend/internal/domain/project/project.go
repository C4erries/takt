package project

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultColor = "#6366f1"
)

type Project struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Description string
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewProject(
	workspaceID uuid.UUID,
	name string,
	description string,
	color string,
) (*Project, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > 255 {
		return nil, ErrInvalidProjectName
	}

	trimmedColor := strings.TrimSpace(color)
	if trimmedColor == "" {
		trimmedColor = DefaultColor
	} else if len(trimmedColor) > 32 {
		return nil, ErrInvalidProjectColor
	}

	now := time.Now().UTC()
	return &Project{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        trimmedName,
		Description: strings.TrimSpace(description),
		Color:       trimmedColor,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (p *Project) Update(name, description, color string) error {
	trimmedName := strings.TrimSpace(name)
	if trimmedName != "" {
		if len(trimmedName) > 255 {
			return ErrInvalidProjectName
		}
		p.Name = trimmedName
	}

	trimmedColor := strings.TrimSpace(color)
	if trimmedColor != "" {
		if len(trimmedColor) > 32 {
			return ErrInvalidProjectColor
		}
		p.Color = trimmedColor
	}

	p.Description = strings.TrimSpace(description)
	p.UpdatedAt = time.Now().UTC()
	return nil
}

