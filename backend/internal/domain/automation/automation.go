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
	TriggerTaskCompleted         = "task_completed"
	TriggerTaskUpdated           = "task_updated"
	TriggerTaskCreated           = "task_created"
	TriggerDueDateReached        = "due_date_reached"
	TriggerTaskEnteredColumn     = "task_entered_column"
	TriggerTaskCompletedInColumn = "task_completed_in_column"
	TriggerTaskFieldChanged      = "task_field_changed"
	TriggerColumnSLATimer        = "column_sla_timer"
	TriggerScheduledSweep        = "scheduled_sweep"
)

var validTriggers = map[string]bool{
	TriggerTaskCompleted:         true,
	TriggerTaskUpdated:           true,
	TriggerTaskCreated:           true,
	TriggerDueDateReached:        true,
	TriggerTaskEnteredColumn:     true,
	TriggerTaskCompletedInColumn: true,
	TriggerTaskFieldChanged:      true,
	TriggerColumnSLATimer:        true,
	TriggerScheduledSweep:        true,
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
	ColumnID       *uuid.UUID
	Name           string
	TriggerType    string
	ConditionsJSON json.RawMessage
	ActionsJSON    json.RawMessage
	IsEnabled      bool
	RuleDefinition *RuleDefinition
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

// SetColumnID updates the associated column and timestamp.
func (a *Automation) SetColumnID(columnID *uuid.UUID) {
	a.ColumnID = columnID
	a.UpdatedAt = time.Now().UTC()
}

// SetRuleDefinition updates the rule definition and timestamp.
func (a *Automation) SetRuleDefinition(def *RuleDefinition) {
	a.RuleDefinition = def
	a.UpdatedAt = time.Now().UTC()
}

// NewColumnAutomation creates a new valid column-scoped Automation entity.
func NewColumnAutomation(
	workspaceID uuid.UUID,
	boardID *uuid.UUID,
	columnID uuid.UUID,
	name string,
	triggerType string,
	conditionsJSON json.RawMessage,
	actionsJSON json.RawMessage,
	isEnabled bool,
) (*Automation, error) {
	if columnID == uuid.Nil {
		return nil, ErrInvalidColumnID
	}
	a, err := NewAutomation(workspaceID, boardID, name, triggerType, conditionsJSON, actionsJSON, isEnabled)
	if err != nil {
		return nil, err
	}
	a.ColumnID = &columnID
	return a, nil
}

// NewRuleAutomation creates a new valid Automation entity configured with a RuleDefinition.
func NewRuleAutomation(
	workspaceID uuid.UUID,
	boardID *uuid.UUID,
	columnID *uuid.UUID,
	name string,
	ruleDef *RuleDefinition,
	isEnabled bool,
) (*Automation, error) {
	if ruleDef == nil {
		return nil, ErrRuleDefinitionNil
	}
	if err := ruleDef.Validate(); err != nil {
		return nil, err
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len([]rune(trimmedName)) > MaxAutomationNameLen {
		return nil, ErrInvalidAutomationName
	}

	now := time.Now().UTC()
	return &Automation{
		ID:             uuid.New(),
		WorkspaceID:    workspaceID,
		BoardID:        boardID,
		ColumnID:       columnID,
		Name:           trimmedName,
		TriggerType:    ruleDef.Trigger.Type,
		ConditionsJSON: json.RawMessage("{}"),
		ActionsJSON:    json.RawMessage("{}"),
		RuleDefinition: ruleDef,
		IsEnabled:      isEnabled,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}
