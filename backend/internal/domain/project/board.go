package project

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultBoardName = "Основная"
)

type Board struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewBoard(projectID uuid.UUID, name string) (*Board, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > 255 {
		return nil, ErrInvalidBoardName
	}

	now := time.Now().UTC()
	return &Board{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      trimmedName,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func NewDefaultBoard(projectID uuid.UUID) (*Board, error) {
	return NewBoard(projectID, DefaultBoardName)
}

func (b *Board) Update(name string) error {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > 255 {
		return ErrInvalidBoardName
	}
	b.Name = trimmedName
	b.UpdatedAt = time.Now().UTC()
	return nil
}

