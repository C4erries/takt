package task

import (
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

const (
	DefaultTagColor = "matcha"
	MaxTagNameLen   = 64
	MaxTagColorLen  = 32
)

// Tag represents a context tag assigned to tasks within a workspace.
type Tag struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	ColorHex    string
	CreatedAt   time.Time
}

// ValidateTagName checks if the tag name satisfies domain invariants.
func ValidateTagName(name string) error {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) == 0 || len([]rune(trimmed)) > MaxTagNameLen {
		return ErrInvalidTagName
	}

	for _, r := range trimmed {
		if unicode.IsControl(r) || r == '\n' || r == '\r' || r == '\t' || r == '/' || r == '\\' || r == ';' {
			return ErrInvalidTagName
		}
	}
	return nil
}

// NewTag creates a new Tag instance enforcing domain invariants.
func NewTag(workspaceID uuid.UUID, name, colorHex string) (*Tag, error) {
	trimmedName := strings.TrimSpace(name)
	if err := ValidateTagName(trimmedName); err != nil {
		return nil, err
	}

	trimmedColor := strings.TrimSpace(colorHex)
	if trimmedColor == "" {
		trimmedColor = DefaultTagColor
	}
	if len(trimmedColor) > MaxTagColorLen {
		return nil, ErrInvalidTagColor
	}

	return &Tag{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        trimmedName,
		ColorHex:    trimmedColor,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

