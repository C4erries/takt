package project

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type StatusCategory string

const (
	StatusCategoryBacklog    StatusCategory = "backlog"
	StatusCategoryTodo       StatusCategory = "todo"
	StatusCategoryInProgress StatusCategory = "in_progress"
	StatusCategoryDone       StatusCategory = "done"
)

func (s StatusCategory) IsValid() bool {
	switch s {
	case StatusCategoryBacklog, StatusCategoryTodo, StatusCategoryInProgress, StatusCategoryDone:
		return true
	default:
		return false
	}
}

const (
	DefaultPositionTodo       = 1000.0
	DefaultPositionInProgress = 2000.0
	DefaultPositionDone       = 3000.0

	DefaultColumnNameTodo       = "К выполнению"
	DefaultColumnNameInProgress = "В работе"
	DefaultColumnNameDone       = "Готово"
)

type Column struct {
	ID             uuid.UUID
	BoardID        uuid.UUID
	Name           string
	Position       float64
	StatusCategory StatusCategory
	IsInbox        bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewColumn(
	boardID uuid.UUID,
	name string,
	position float64,
	statusCategory StatusCategory,
) (*Column, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > 255 {
		return nil, ErrInvalidColumnName
	}

	if statusCategory == "" {
		statusCategory = StatusCategoryTodo
	}
	if !statusCategory.IsValid() {
		return nil, ErrInvalidStatusCategory
	}

	if position <= 0 {
		return nil, ErrInvalidPosition
	}

	now := time.Now().UTC()
	return &Column{
		ID:             uuid.New(),
		BoardID:        boardID,
		Name:           trimmedName,
		Position:       position,
		StatusCategory: statusCategory,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func NewDefaultColumns(boardID uuid.UUID) []*Column {
	now := time.Now().UTC()
	return []*Column{
		{
			ID:             uuid.New(),
			BoardID:        boardID,
			Name:           DefaultColumnNameTodo,
			Position:       DefaultPositionTodo,
			StatusCategory: StatusCategoryTodo,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             uuid.New(),
			BoardID:        boardID,
			Name:           DefaultColumnNameInProgress,
			Position:       DefaultPositionInProgress,
			StatusCategory: StatusCategoryInProgress,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             uuid.New(),
			BoardID:        boardID,
			Name:           DefaultColumnNameDone,
			Position:       DefaultPositionDone,
			StatusCategory: StatusCategoryDone,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}
}

func (c *Column) Update(name string, position float64, statusCategory StatusCategory) error {
	trimmedName := strings.TrimSpace(name)
	if trimmedName != "" {
		if len(trimmedName) > 255 {
			return ErrInvalidColumnName
		}
		c.Name = trimmedName
	}

	if position > 0 {
		c.Position = position
	}

	if statusCategory != "" {
		if !statusCategory.IsValid() {
			return ErrInvalidStatusCategory
		}
		c.StatusCategory = statusCategory
	}

	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Column) MoveToBoard(newBoardID uuid.UUID, newPosition float64) error {
	if newBoardID == uuid.Nil {
		return ErrInvalidBoardID
	}
	if newPosition <= 0 {
		return ErrInvalidPosition
	}

	c.BoardID = newBoardID
	c.Position = newPosition
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// SetInbox sets whether this column serves as the designated inbox for the board.
func (c *Column) SetInbox(isInbox bool) {
	c.IsInbox = isInbox
	c.UpdatedAt = time.Now().UTC()
}

