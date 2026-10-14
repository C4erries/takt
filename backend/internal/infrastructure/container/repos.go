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

	if database != nil && database.Pool != nil {
		queries := sqlc.New(database.Pool)
		userRepo = repository.NewUserRepository(queries)
		workspaceRepo = repository.NewWorkspaceRepository(queries)
		taskRepo = repository.NewTaskRepository(queries)
	}

	return &Repositories{
		DB:            database,
		UserRepo:      userRepo,
		WorkspaceRepo: workspaceRepo,
		TaskRepo:      taskRepo,
	}, nil
}
