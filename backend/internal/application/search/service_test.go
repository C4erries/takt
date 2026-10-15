package search_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appSearch "takt/internal/application/search"
	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
	domainWorkspace "takt/internal/domain/workspace"
)

type mockTaskSearcher struct {
	searchTasksFunc func(ctx context.Context, workspaceID uuid.UUID, query string, limit int32) ([]*domainTask.Task, error)
}

func (m *mockTaskSearcher) SearchTasks(ctx context.Context, workspaceID uuid.UUID, query string, limit int32) ([]*domainTask.Task, error) {
	if m.searchTasksFunc != nil {
		return m.searchTasksFunc(ctx, workspaceID, query, limit)
	}
	return nil, nil
}

type mockProjectSearcher struct {
	searchProjectsFunc func(ctx context.Context, workspaceID uuid.UUID, query string, limit int32) ([]*domainProject.Project, error)
}

func (m *mockProjectSearcher) SearchProjects(ctx context.Context, workspaceID uuid.UUID, query string, limit int32) ([]*domainProject.Project, error) {
	if m.searchProjectsFunc != nil {
		return m.searchProjectsFunc(ctx, workspaceID, query, limit)
	}
	return nil, nil
}

type mockMembershipChecker struct {
	checkAccessFunc func(ctx context.Context, workspaceID, userID uuid.UUID) error
}

func (m *mockMembershipChecker) CheckAccess(ctx context.Context, workspaceID, userID uuid.UUID) error {
	if m.checkAccessFunc != nil {
		return m.checkAccessFunc(ctx, workspaceID, userID)
	}
	return nil
}

func TestService_Search(t *testing.T) {
	wsID := uuid.New()
	userID := uuid.New()

	dummyTask := &domainTask.Task{
		ID:          uuid.New(),
		WorkspaceID: wsID,
		Title:       "Купить молоко",
	}

	dummyProject := &domainProject.Project{
		ID:          uuid.New(),
		WorkspaceID: wsID,
		Name:        "Ремонт кухни",
	}

	tests := []struct {
		name         string
		query        string
		membershipErr error
		taskErr      error
		projectErr   error
		tasks        []*domainTask.Task
		projects     []*domainProject.Project
		wantErr      bool
		expectedErr  error
		assertResult func(t *testing.T, res *appSearch.SearchResult)
	}{
		{
			name:     "successful search with results",
			query:    "молоко",
			tasks:    []*domainTask.Task{dummyTask},
			projects: []*domainProject.Project{dummyProject},
			wantErr:  false,
			assertResult: func(t *testing.T, res *appSearch.SearchResult) {
				require.NotNil(t, res)
				assert.Len(t, res.Tasks, 1)
				assert.Equal(t, "Купить молоко", res.Tasks[0].Title)
				assert.Len(t, res.Projects, 1)
				assert.Equal(t, "Ремонт кухни", res.Projects[0].Name)
			},
		},
		{
			name:    "empty query returns empty slices without db call",
			query:   "   ",
			wantErr: false,
			assertResult: func(t *testing.T, res *appSearch.SearchResult) {
				require.NotNil(t, res)
				assert.Empty(t, res.Tasks)
				assert.Empty(t, res.Projects)
			},
		},
		{
			name:          "workspace access denied",
			query:         "test",
			membershipErr: domainWorkspace.ErrAccessDenied,
			wantErr:       true,
			expectedErr:   domainWorkspace.ErrAccessDenied,
		},
		{
			name:        "task search failure",
			query:       "test",
			taskErr:     errors.New("db task search error"),
			wantErr:     true,
			expectedErr: nil,
		},
		{
			name:        "project search failure",
			query:       "test",
			projectErr:  errors.New("db project search error"),
			wantErr:     true,
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			taskSearcher := &mockTaskSearcher{
				searchTasksFunc: func(ctx context.Context, workspaceID uuid.UUID, query string, limit int32) ([]*domainTask.Task, error) {
					if tc.taskErr != nil {
						return nil, tc.taskErr
					}
					return tc.tasks, nil
				},
			}

			projectSearcher := &mockProjectSearcher{
				searchProjectsFunc: func(ctx context.Context, workspaceID uuid.UUID, query string, limit int32) ([]*domainProject.Project, error) {
					if tc.projectErr != nil {
						return nil, tc.projectErr
					}
					return tc.projects, nil
				},
			}

			checker := &mockMembershipChecker{
				checkAccessFunc: func(ctx context.Context, workspaceID, userID uuid.UUID) error {
					return tc.membershipErr
				},
			}

			svc := appSearch.NewService(taskSearcher, projectSearcher, checker)
			res, err := svc.Search(context.Background(), wsID, userID, tc.query)

			if tc.wantErr {
				require.Error(t, err)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
				}
			} else {
				require.NoError(t, err)
				if tc.assertResult != nil {
					tc.assertResult(t, res)
				}
			}
		})
	}
}
