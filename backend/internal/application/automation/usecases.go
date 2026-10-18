package automation

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	domainAutomation "takt/internal/domain/automation"
)

// UseCases defines the interface for managing automation rules.
type UseCases interface {
	ListAutomations(ctx context.Context, workspaceID uuid.UUID) ([]*domainAutomation.Automation, error)
	ListColumnAutomations(ctx context.Context, workspaceID, columnID uuid.UUID) ([]*domainAutomation.Automation, error)
	GetAutomation(ctx context.Context, workspaceID, automationID uuid.UUID) (*domainAutomation.Automation, error)
	ToggleAutomation(ctx context.Context, workspaceID, automationID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error)
	CreateAutomation(ctx context.Context, input CreateAutomationInput) (*domainAutomation.Automation, error)
	UpdateAutomation(ctx context.Context, input UpdateAutomationInput) (*domainAutomation.Automation, error)
	DeleteAutomation(ctx context.Context, workspaceID, automationID uuid.UUID) error
	CreateColumnAutomation(ctx context.Context, input CreateColumnAutomationInput) (*domainAutomation.Automation, error)
	SeedDefaultAutomations(ctx context.Context, workspaceID uuid.UUID) error
}

// CreateAutomationInput defines payload to create a workspace automation rule.
type CreateAutomationInput struct {
	WorkspaceID    uuid.UUID
	BoardID        *uuid.UUID
	ColumnID       *uuid.UUID
	Name           string
	TriggerType    string
	ConditionsJSON json.RawMessage
	ActionsJSON    json.RawMessage
	RuleDefinition *domainAutomation.RuleDefinition
	IsEnabled      bool
}

// UpdateAutomationInput defines payload to update an existing automation rule.
type UpdateAutomationInput struct {
	AutomationID   uuid.UUID
	WorkspaceID    uuid.UUID
	BoardID        *uuid.UUID
	ColumnID       *uuid.UUID
	Name           *string
	TriggerType    *string
	ConditionsJSON json.RawMessage
	ActionsJSON    json.RawMessage
	RuleDefinition *domainAutomation.RuleDefinition
	IsEnabled      *bool
}

// CreateColumnAutomationInput defines payload to create a column-scoped automation rule.
type CreateColumnAutomationInput struct {
	WorkspaceID    uuid.UUID
	BoardID        *uuid.UUID
	ColumnID       uuid.UUID
	Name           string
	TriggerType    string
	ConditionsJSON json.RawMessage
	ActionsJSON    json.RawMessage
	IsEnabled      bool
}

// Service implements UseCases.
type Service struct {
	repo AutomationRepository
}

// NewService constructs a new automation Service.
func NewService(repo AutomationRepository) *Service {
	return &Service{repo: repo}
}

// ListAutomations returns all automations configured for the given workspace.
func (s *Service) ListAutomations(ctx context.Context, workspaceID uuid.UUID) ([]*domainAutomation.Automation, error) {
	return s.repo.ListWorkspaceAutomations(ctx, workspaceID)
}

// ListColumnAutomations returns all automations configured for a specific column.
func (s *Service) ListColumnAutomations(ctx context.Context, workspaceID, columnID uuid.UUID) ([]*domainAutomation.Automation, error) {
	return s.repo.ListAutomationsByColumn(ctx, workspaceID, columnID)
}

// GetAutomation returns a specific automation rule.
func (s *Service) GetAutomation(ctx context.Context, workspaceID, automationID uuid.UUID) (*domainAutomation.Automation, error) {
	return s.repo.GetAutomationByID(ctx, automationID, workspaceID)
}

// ToggleAutomation toggles activation status of an automation rule.
func (s *Service) ToggleAutomation(ctx context.Context, workspaceID, automationID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error) {
	return s.repo.ToggleAutomation(ctx, automationID, workspaceID, isEnabled)
}

// CreateAutomation creates and persists a new automation rule (legacy or Low-Code DSL).
func (s *Service) CreateAutomation(ctx context.Context, input CreateAutomationInput) (*domainAutomation.Automation, error) {
	if input.RuleDefinition != nil {
		rule, err := domainAutomation.NewRuleAutomation(
			input.WorkspaceID,
			input.BoardID,
			input.ColumnID,
			input.Name,
			input.RuleDefinition,
			input.IsEnabled,
		)
		if err != nil {
			return nil, err
		}
		if err := s.repo.CreateAutomation(ctx, rule); err != nil {
			return nil, err
		}
		return rule, nil
	}

	rule, err := domainAutomation.NewAutomation(
		input.WorkspaceID,
		input.BoardID,
		input.Name,
		input.TriggerType,
		input.ConditionsJSON,
		input.ActionsJSON,
		input.IsEnabled,
	)
	if err != nil {
		return nil, err
	}
	rule.ColumnID = input.ColumnID
	if err := s.repo.CreateAutomation(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// UpdateAutomation updates and persists an existing automation rule.
func (s *Service) UpdateAutomation(ctx context.Context, input UpdateAutomationInput) (*domainAutomation.Automation, error) {
	rule, err := s.repo.GetAutomationByID(ctx, input.AutomationID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" || len([]rune(trimmed)) > domainAutomation.MaxAutomationNameLen {
			return nil, domainAutomation.ErrInvalidAutomationName
		}
		rule.Name = trimmed
	}

	if input.IsEnabled != nil {
		rule.SetEnabled(*input.IsEnabled)
	}

	if input.RuleDefinition != nil {
		if err := domainAutomation.ValidateRuleDefinition(input.RuleDefinition); err != nil {
			return nil, err
		}
		rule.SetRuleDefinition(input.RuleDefinition)
		rule.TriggerType = input.RuleDefinition.Trigger.Type
	} else if input.TriggerType != nil {
		if !domainAutomation.IsValidTriggerType(*input.TriggerType) {
			return nil, domainAutomation.ErrInvalidTriggerType
		}
		rule.TriggerType = *input.TriggerType
	}

	if len(input.ConditionsJSON) > 0 {
		rule.ConditionsJSON = input.ConditionsJSON
	}
	if len(input.ActionsJSON) > 0 {
		rule.ActionsJSON = input.ActionsJSON
	}
	if input.BoardID != nil {
		rule.BoardID = input.BoardID
	}
	if input.ColumnID != nil {
		rule.ColumnID = input.ColumnID
	}

	if err := s.repo.UpdateAutomation(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// DeleteAutomation deletes an automation rule.
func (s *Service) DeleteAutomation(ctx context.Context, workspaceID, automationID uuid.UUID) error {
	_, err := s.repo.GetAutomationByID(ctx, automationID, workspaceID)
	if err != nil {
		return err
	}
	return s.repo.DeleteAutomation(ctx, automationID, workspaceID)
}

// CreateColumnAutomation creates and persists a column-scoped automation.
func (s *Service) CreateColumnAutomation(ctx context.Context, input CreateColumnAutomationInput) (*domainAutomation.Automation, error) {
	rule, err := domainAutomation.NewColumnAutomation(
		input.WorkspaceID,
		input.BoardID,
		input.ColumnID,
		input.Name,
		input.TriggerType,
		input.ConditionsJSON,
		input.ActionsJSON,
		input.IsEnabled,
	)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateColumnAutomation(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// SeedDefaultAutomations creates the 3 preset automation rules for a new workspace.
func (s *Service) SeedDefaultAutomations(ctx context.Context, workspaceID uuid.UUID) error {
	presets := []struct {
		name       string
		trigger    string
		conditions string
		actions    string
		isEnabled  bool
	}{
		{
			name:       "Done & Archive",
			trigger:    domainAutomation.TriggerTaskCompleted,
			conditions: "{}",
			actions:    `{"action": "move_to_done"}`,
			isEnabled:  true,
		},
		{
			name:       "Auto-Focus",
			trigger:    domainAutomation.TriggerTaskUpdated,
			conditions: `{"has_due_date": true, "due_date_le_today": true}`,
			actions:    `{"action": "set_focus"}`,
			isEnabled:  true,
		},
		{
			name:       "Inbox Zero",
			trigger:    domainAutomation.TriggerTaskUpdated,
			conditions: `{"project_assigned": true}`,
			actions:    `{"action": "move_to_first_column"}`,
			isEnabled:  false,
		},
	}

	for _, p := range presets {
		rule, err := domainAutomation.NewAutomation(
			workspaceID,
			nil,
			p.name,
			p.trigger,
			json.RawMessage(p.conditions),
			json.RawMessage(p.actions),
			p.isEnabled,
		)
		if err != nil {
			return err
		}
		if err := s.repo.CreateAutomation(ctx, rule); err != nil {
			return err
		}
	}

	return nil
}
