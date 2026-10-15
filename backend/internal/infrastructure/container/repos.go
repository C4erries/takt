package container

import (
	"context"
	"log"

	"takt/internal/infrastructure/config"
	"takt/internal/infrastructure/db"
	"takt/internal/infrastructure/db/sqlc"
	"takt/internal/infrastructure/repository"
)

// Repositories container holding all database repositories.
type Repositories struct {
	DB            *db.DB
	UserRepo      *repository.UserRepository
	WorkspaceRepo *repository.WorkspaceRepository
	TaskRepo      *repository.TaskRepository
	ProjectRepo   *repository.ProjectRepository
	BoardRepo     *repository.BoardRepository
	ColumnRepo     *repository.ColumnRepository
	TaskMoveRepo   *repository.TaskMoveRepository
	AutomationRepo *repository.AutomationRepository
	AttachmentRepo *repository.AttachmentRepository
}

// initRepositories initializes database connection pool and repositories.
func initRepositories(ctx context.Context, cfg *config.Config) (*Repositories, error) {
	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("[WARN] Failed to connect to database at startup: %v (service will run in degraded mode)", err)
	}

	var userRepo *repository.UserRepository
	var workspaceRepo *repository.WorkspaceRepository
	var taskRepo *repository.TaskRepository
	var projectRepo *repository.ProjectRepository
	var boardRepo *repository.BoardRepository
	var columnRepo *repository.ColumnRepository
	var taskMoveRepo *repository.TaskMoveRepository
	var automationRepo *repository.AutomationRepository
	var attachmentRepo *repository.AttachmentRepository

	if database != nil && database.Pool != nil {
		queries := sqlc.New(database.Pool)
		userRepo = repository.NewUserRepository(queries)
		workspaceRepo = repository.NewWorkspaceRepository(queries)
		taskRepo = repository.NewTaskRepository(queries)
		projectRepo = repository.NewProjectRepository(queries)
		boardRepo = repository.NewBoardRepository(queries)
		columnRepo = repository.NewColumnRepository(queries)
		taskMoveRepo = repository.NewTaskMoveRepository(queries)
		automationRepo = repository.NewAutomationRepository(queries)
		attachmentRepo = repository.NewAttachmentRepository(queries)
	}

	return &Repositories{
		DB:             database,
		UserRepo:       userRepo,
		WorkspaceRepo:  workspaceRepo,
		TaskRepo:       taskRepo,
		ProjectRepo:    projectRepo,
		BoardRepo:      boardRepo,
		ColumnRepo:     columnRepo,
		TaskMoveRepo:   taskMoveRepo,
		AutomationRepo: automationRepo,
		AttachmentRepo: attachmentRepo,
	}, nil
}
