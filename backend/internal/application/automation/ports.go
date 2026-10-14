package automation

import (
	"context"

	"github.com/google/uuid"
	domainAutomation "takt/internal/domain/automation"
)

// AutomationRepository handles persistence of automation rules.
type AutomationRepository interface {
	ListWorkspaceAutomations(ctx context.Context, workspaceID uuid.UUID) ([]*domainAutomation.Automation, error)
	GetAutomationByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainAutomation.Automation, error)
	ToggleAutomation(ctx context.Context, id, workspaceID uuid.UUID, isEnabled bool) (*domainAutomation.Automation, error)
	CreateAutomation(ctx context.Context, a *domainAutomation.Automation) error
	ListEnabledAutomationsByTrigger(ctx context.Context, workspaceID uuid.UUID, triggerType string) ([]*domainAutomation.Automation, error)
}

// TaskModifier modifies task properties in response to automation triggers.
type TaskModifier interface {
	MoveTaskToColumn(ctx context.Context, taskID, workspaceID uuid.UUID, columnID, projectID *uuid.UUID, position float64) error
	SetTaskFocus(ctx context.Context, taskID, workspaceID uuid.UUID, focusOrder *int) error
}

// ProjectReader inspects board and column structure to determine target columns.
type ProjectReader interface {
	GetFirstTodoColumn(ctx context.Context, projectID, workspaceID uuid.UUID) (*uuid.UUID, error)
	GetDoneColumn(ctx context.Context, projectID, workspaceID uuid.UUID) (*uuid.UUID, error)
}

