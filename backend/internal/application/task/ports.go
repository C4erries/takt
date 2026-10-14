package task

import (
	"context"

	"github.com/google/uuid"
	domainTask "takt/internal/domain/task"
)

// TaskRepository alias to domain repository.
type TaskRepository = domainTask.TaskRepository

// WorkspaceAccessChecker ensures the user has access to the workspace.
type WorkspaceAccessChecker interface {
	CheckAccess(ctx context.Context, workspaceID, userID uuid.UUID) error
}
