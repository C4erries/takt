package container

import (
	automationApp "takt/internal/application/automation"
	projectApp "takt/internal/application/project"
	taskApp "takt/internal/application/task"
	workspaceApp "takt/internal/application/workspace"
	"takt/internal/infrastructure/config"
	"takt/internal/infrastructure/eventbus"
	httpDelivery "takt/internal/infrastructure/http"
)

// Services container holding all domain and application use cases.
type Services struct {
	WorkspaceService  workspaceApp.UseCases
	TaskService       taskApp.UseCases
	ProjectService    projectApp.UseCases
	AutomationService automationApp.UseCases
	EventBus          *eventbus.InMemoryEventBus
	JWTManager        *httpDelivery.JWTManager
}

// initServices initializes application use case services.
func initServices(repos *Repositories, cfg *config.Config) *Services {
	jwtManager := httpDelivery.NewJWTManager(cfg.JWTSecret)
	hasher := httpDelivery.NewBcryptHasher()

	eventBus := eventbus.NewInMemoryEventBus()

	var automationService *automationApp.Service
	if repos.AutomationRepo != nil {
		automationService = automationApp.NewService(repos.AutomationRepo)
		listener := automationApp.NewEventListener(repos.AutomationRepo, repos.AutomationRepo, repos.AutomationRepo, nil)
		_ = listener.Register(eventBus)
	}

	var taskService *taskApp.Service
	if repos.TaskRepo != nil {
		taskService = taskApp.NewService(repos.TaskRepo, repos.TaskRepo, nil, eventBus)
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
			automationService,
		)
	}

	var projectService *projectApp.Service
	if repos.ProjectRepo != nil && repos.BoardRepo != nil && repos.ColumnRepo != nil && repos.TaskMoveRepo != nil && repos.DB != nil {
		projectService = projectApp.NewService(
			repos.ProjectRepo,
			repos.BoardRepo,
			repos.ColumnRepo,
			repos.TaskMoveRepo,
			nil,
			repos.DB,
		)
		projectService.SetEventPublisher(eventBus)
	}

	return &Services{
		WorkspaceService:  workspaceService,
		TaskService:       taskService,
		ProjectService:    projectService,
		AutomationService: automationService,
		EventBus:          eventBus,
		JWTManager:        jwtManager,
	}
}
