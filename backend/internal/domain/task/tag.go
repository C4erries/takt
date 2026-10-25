package task

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

var hexColorRegex = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

const (
	DefaultTagColor = "matcha"
	MaxTagNameLen   = 64
	MaxTagColorLen  = 32
)

// Tag represents a context tag assigned to tasks within a workspace or project.
type Tag struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	ProjectID   *uuid.UUID
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
	return NewScopedTag(workspaceID, nil, name, colorHex)
}

// NewScopedTag creates a new Tag instance scoped optionally to a project, enforcing domain invariants.
func NewScopedTag(workspaceID uuid.UUID, projectID *uuid.UUID, name, colorHex string) (*Tag, error) {
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

	var pID *uuid.UUID
	if projectID != nil && *projectID != uuid.Nil {
		copyID := *projectID
		pID = &copyID
	}

	return &Tag{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		ProjectID:   pID,
		Name:        trimmedName,
		ColorHex:    trimmedColor,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// SetName validates and sets the tag's name.
func (t *Tag) SetName(name string) error {
	trimmed := strings.TrimSpace(name)
	if err := ValidateTagName(trimmed); err != nil {
		return err
	}
	t.Name = trimmed
	return nil
}

// SetColorHex validates that the color is a valid HEX format (#RGB or #RRGGBB) and sets it.
func (t *Tag) SetColorHex(colorHex string) error {
	trimmed := strings.TrimSpace(colorHex)
	if !hexColorRegex.MatchString(trimmed) {
		return ErrInvalidTagColor
	}
	t.ColorHex = trimmed
	return nil
}

// TagWithMetrics represents a tag along with its usage metrics (count of associated tasks).
type TagWithMetrics struct {
	Tag
	TaskCount int64
}

// NewTagWithMetrics constructs a TagWithMetrics domain object ensuring non-negative task count.
func NewTagWithMetrics(tag Tag, taskCount int64) TagWithMetrics {
	if taskCount < 0 {
		taskCount = 0
	}
	return TagWithMetrics{
		Tag:       tag,
		TaskCount: taskCount,
	}
}


