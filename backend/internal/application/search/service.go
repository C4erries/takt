package search

import (
	"context"
	"strings"

	"github.com/google/uuid"

	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
)

const (
	// DefaultTaskLimit limits search results for tasks.
	DefaultTaskLimit int32 = 20
	// DefaultProjectLimit limits search results for projects.
	DefaultProjectLimit int32 = 10
)

// Service implements search use cases across tasks and projects.
type Service struct {
	taskSearcher      TaskSearcher
	projectSearcher   ProjectSearcher
	membershipChecker WorkspaceMembershipChecker
}

// NewService constructs a new search application service.
func NewService(
	taskSearcher TaskSearcher,
	projectSearcher ProjectSearcher,
	membershipChecker WorkspaceMembershipChecker,
) *Service {
	return &Service{
		taskSearcher:      taskSearcher,
		projectSearcher:   projectSearcher,
		membershipChecker: membershipChecker,
	}
}

// Search executes scoped search in tasks and projects within the target workspace.
func (s *Service) Search(ctx context.Context, workspaceID, userID uuid.UUID, query string) (*SearchResult, error) {
	if s.membershipChecker != nil && userID != uuid.Nil {
		if err := s.membershipChecker.CheckAccess(ctx, workspaceID, userID); err != nil {
			return nil, err
		}
	}

	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return &SearchResult{
			Tasks:    []*domainTask.Task{},
			Projects: []*domainProject.Project{},
		}, nil
	}

	var tasks []*domainTask.Task
	if s.taskSearcher != nil {
		t, err := s.taskSearcher.SearchTasks(ctx, workspaceID, trimmedQuery, DefaultTaskLimit)
		if err != nil {
			return nil, err
		}
		tasks = t
	}
	if tasks == nil {
		tasks = []*domainTask.Task{}
	}

	var projects []*domainProject.Project
	if s.projectSearcher != nil {
		p, err := s.projectSearcher.SearchProjects(ctx, workspaceID, trimmedQuery, DefaultProjectLimit)
		if err != nil {
			return nil, err
		}
		projects = p
	}
	if projects == nil {
		projects = []*domainProject.Project{}
	}

	return &SearchResult{
		Tasks:    tasks,
		Projects: projects,
	}, nil
}
