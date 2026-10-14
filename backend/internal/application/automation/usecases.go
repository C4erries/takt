package automation

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	domainAutomation "takt/internal/domain/automation"
)

// UseCases defines the interface for managing automation rules.
type UseCases interface {
	ListAutomations(ctx context.Context, workspaceID uuid.UUID) ([]*domainAutomation.Automation, error)
	ToggleAutomation(ctx context.Context, workspaceID, automationID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error)
	SeedDefaultAutomations(ctx context.Context, workspaceID uuid.UUID) error
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

// ToggleAutomation toggles activation status of an automation rule.
func (s *Service) ToggleAutomation(ctx context.Context, workspaceID, automationID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error) {
	return s.repo.ToggleAutomation(ctx, automationID, workspaceID, isEnabled)
}

// SeedDefaultAutomations creates the 3 preset automation rules for a new workspace.
func (s *Service) SeedDefaultAutomations(ctx context.Context, workspaceID uuid.UUID) error {
	presets := []struct {
		name        string
		trigger     string
		conditions  string
		actions     string
		isEnabled   bool
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

