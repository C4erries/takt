package container

import (
	taskApp "takt/internal/application/task"
	workspaceApp "takt/internal/application/workspace"
	"takt/internal/infrastructure/config"
	httpDelivery "takt/internal/infrastructure/http"
)

// Services container holding all domain and application use cases.
type Services struct {
	WorkspaceService workspaceApp.UseCases
	TaskService      taskApp.UseCases
	JWTManager       *httpDelivery.JWTManager
}

// initServices initializes application use case services.
func initServices(repos *Repositories, cfg *config.Config) *Services {
	jwtManager := httpDelivery.NewJWTManager(cfg.JWTSecret)
	hasher := httpDelivery.NewBcryptHasher()

	var taskService *taskApp.Service
	if repos.TaskRepo != nil {
		taskService = taskApp.NewService(repos.TaskRepo, nil)
	}

	var workspaceService *workspaceApp.Service
	if repos.UserRepo != nil && repos.WorkspaceRepo != nil {
		workspaceService = workspaceApp.NewService(
			repos.UserRepo,
			repos.WorkspaceRepo,
			hasher,
			jwtManager,
			taskService,
			repos.DB,
		)
	}

	return &Services{
		WorkspaceService: workspaceService,
		TaskService:      taskService,
		JWTManager:       jwtManager,
	}
}
