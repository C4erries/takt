package search

import (
	"context"

	"github.com/google/uuid"

	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
)

// TaskSearcher queries tasks matching the text criteria within a workspace.
type TaskSearcher interface {
	SearchTasks(ctx context.Context, workspaceID uuid.UUID, query string, limit int32) ([]*domainTask.Task, error)
}

// ProjectSearcher queries projects matching the text criteria within a workspace.
type ProjectSearcher interface {
	SearchProjects(ctx context.Context, workspaceID uuid.UUID, query string, limit int32) ([]*domainProject.Project, error)
}

// WorkspaceMembershipChecker verifies that the given user has access to the workspace.
type WorkspaceMembershipChecker interface {
	CheckAccess(ctx context.Context, workspaceID, userID uuid.UUID) error
}

// SearchResult aggregates matched tasks and projects.
type SearchResult struct {
	Tasks    []*domainTask.Task
	Projects []*domainProject.Project
}

// UseCases defines the search service contract.
type UseCases interface {
	Search(ctx context.Context, workspaceID, userID uuid.UUID, query string) (*SearchResult, error)
}
