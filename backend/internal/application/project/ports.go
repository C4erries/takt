package project

import (
	"context"

	"github.com/google/uuid"
	domainProject "takt/internal/domain/project"
	domainTask "takt/internal/domain/task"
	"takt/internal/domain/events"
)

type Transactor interface {
	WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

type WorkspaceAccessChecker interface {
	CheckAccess(ctx context.Context, workspaceID, userID uuid.UUID) error
}

type ProjectRepository interface {
	Create(ctx context.Context, p *domainProject.Project) error
	GetByID(ctx context.Context, id, workspaceID uuid.UUID) (*domainProject.Project, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*domainProject.Project, error)
	Update(ctx context.Context, p *domainProject.Project) error
	Delete(ctx context.Context, id, workspaceID uuid.UUID) error
}

type BoardRepository interface {
	Create(ctx context.Context, b *domainProject.Board) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainProject.Board, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*domainProject.Board, error)
	Update(ctx context.Context, b *domainProject.Board) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ColumnRepository interface {
	Create(ctx context.Context, c *domainProject.Column) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainProject.Column, error)
	ListByBoard(ctx context.Context, boardID uuid.UUID) ([]*domainProject.Column, error)
	Update(ctx context.Context, c *domainProject.Column) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountTasksInColumn(ctx context.Context, columnID uuid.UUID) (int64, error)
}

type TaskMoveRepository interface {
	MoveTaskToColumnAndPosition(ctx context.Context, taskID, workspaceID uuid.UUID, columnID, projectID *uuid.UUID, position float64) (*domainTask.Task, error)
	ResetProjectTasksToInbox(ctx context.Context, projectID, workspaceID uuid.UUID) error
	ListTasksByProject(ctx context.Context, projectID, workspaceID uuid.UUID) ([]*domainTask.Task, error)
	ListTasksByColumn(ctx context.Context, columnID, workspaceID uuid.UUID) ([]*domainTask.Task, error)
}

type UseCases interface {
	CreateProject(ctx context.Context, input CreateProjectInput) (*domainProject.Project, error)
	GetProjects(ctx context.Context, workspaceID uuid.UUID) ([]*domainProject.Project, error)
	GetProject(ctx context.Context, projectID, workspaceID uuid.UUID) (*domainProject.Project, error)
	UpdateProject(ctx context.Context, input UpdateProjectInput) (*domainProject.Project, error)
	DeleteProject(ctx context.Context, projectID, workspaceID uuid.UUID) error
	GetProjectBoards(ctx context.Context, projectID, workspaceID uuid.UUID) ([]*domainProject.Board, error)
	CreateBoard(ctx context.Context, projectID, workspaceID uuid.UUID, name string) (*domainProject.Board, error)
	GetBoardColumns(ctx context.Context, boardID, workspaceID uuid.UUID) ([]*domainProject.Column, error)
	CreateColumn(ctx context.Context, input CreateColumnInput) (*domainProject.Column, error)
	UpdateColumn(ctx context.Context, input UpdateColumnInput) (*domainProject.Column, error)
	DeleteColumn(ctx context.Context, boardID, columnID, workspaceID uuid.UUID) error
	MoveTask(ctx context.Context, input MoveTaskInput) (*domainTask.Task, error)
	GetProjectTasks(ctx context.Context, projectID, workspaceID uuid.UUID) ([]*domainTask.Task, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, event events.Event) error
}

