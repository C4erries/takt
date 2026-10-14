package automation

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxAutomationNameLen = 255
)

// Standard automation trigger types
const (
	TriggerTaskCompleted  = "task_completed"
	TriggerTaskUpdated    = "task_updated"
	TriggerTaskCreated    = "task_created"
	TriggerDueDateReached = "due_date_reached"
)

var validTriggers = map[string]bool{
	TriggerTaskCompleted:  true,
	TriggerTaskUpdated:    true,
	TriggerTaskCreated:    true,
	TriggerDueDateReached: true,
}

// IsValidTriggerType validates if the given trigger string is recognized.
func IsValidTriggerType(trigger string) bool {
	return validTriggers[trigger]
}

// Automation represents a declarative automation rule within a workspace.
type Automation struct {
	ID             uuid.UUID
	WorkspaceID    uuid.UUID
	BoardID        *uuid.UUID
	Name           string
	TriggerType    string
	ConditionsJSON json.RawMessage
	ActionsJSON    json.RawMessage
	IsEnabled      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewAutomation creates a new valid Automation entity.
func NewAutomation(
	workspaceID uuid.UUID,
	boardID *uuid.UUID,
	name string,
	triggerType string,
	conditionsJSON json.RawMessage,
	actionsJSON json.RawMessage,
	isEnabled bool,
) (*Automation, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len([]rune(trimmedName)) > MaxAutomationNameLen {
		return nil, ErrInvalidAutomationName
	}

	if !IsValidTriggerType(triggerType) {
		return nil, ErrInvalidTriggerType
	}

	if len(conditionsJSON) == 0 {
		conditionsJSON = json.RawMessage("{}")
	}
	if len(actionsJSON) == 0 {
		actionsJSON = json.RawMessage("{}")
	}

	now := time.Now().UTC()
	return &Automation{
		ID:             uuid.New(),
		WorkspaceID:    workspaceID,
		BoardID:        boardID,
		Name:           trimmedName,
		TriggerType:    triggerType,
		ConditionsJSON: conditionsJSON,
		ActionsJSON:    actionsJSON,
		IsEnabled:      isEnabled,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// SetEnabled updates the activation state and timestamp.
func (a *Automation) SetEnabled(enabled bool) {
	a.IsEnabled = enabled
	a.UpdatedAt = time.Now().UTC()
}

