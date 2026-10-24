package container

import (
	"context"
	"log"

	attachmentApp "takt/internal/application/attachment"
	automationApp "takt/internal/application/automation"
	projectApp "takt/internal/application/project"
	searchApp "takt/internal/application/search"
	taskApp "takt/internal/application/task"
	workspaceApp "takt/internal/application/workspace"
	"takt/internal/infrastructure/config"
	"takt/internal/infrastructure/eventbus"
	httpDelivery "takt/internal/infrastructure/http"
	"takt/internal/infrastructure/storage/s3"
)

// Services container holding all domain and application use cases.
type Services struct {
	WorkspaceService  workspaceApp.UseCases
	TaskService       taskApp.UseCases
	ProjectService    projectApp.UseCases
	AutomationService automationApp.UseCases
	AutomationWorker  *automationApp.Worker
	AttachmentService attachmentApp.UseCases
	SearchService     searchApp.UseCases
	EventBus          *eventbus.InMemoryEventBus
	JWTManager        *httpDelivery.JWTManager
}

// initServices initializes application use case services.
func initServices(repos *Repositories, cfg *config.Config) *Services {
	jwtManager := httpDelivery.NewJWTManager(cfg.JWTSecret)
	hasher := httpDelivery.NewBcryptHasher()

	eventBus := eventbus.NewInMemoryEventBus()

	var automationService *automationApp.Service
	var automationWorker *automationApp.Worker
	if repos.AutomationRepo != nil {
		evaluator := automationApp.NewEvaluator(repos.AutomationRepo, repos.AutomationRepo)
		automationService = automationApp.NewService(repos.AutomationRepo, repos.AutomationRepo, evaluator)
		listener := automationApp.NewEventListener(repos.AutomationRepo, repos.AutomationRepo, repos.AutomationRepo, nil)
		_ = listener.Register(eventBus)
		automationWorker = automationApp.NewWorker(repos.AutomationRepo, repos.AutomationRepo, evaluator, 0)
	}

	var taskService *taskApp.Service
	if repos.TaskRepo != nil {
		taskService = taskApp.NewService(repos.TaskRepo, repos.TaskRepo, nil, eventBus)
		if repos.DB != nil {
			taskService.SetTransactor(repos.DB)
		}
		if repos.AttachmentRepo != nil {
			taskService.SetAttachmentRepository(repos.AttachmentRepo)
		}
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

	var s3Storage *s3.Storage
	if cfg.S3Endpoint != "" && cfg.S3Bucket != "" {
		storageInstance, err := s3.NewStorage(s3.Config{
			Endpoint:  cfg.S3Endpoint,
			Region:    cfg.S3Region,
			Bucket:    cfg.S3Bucket,
			AccessKey: cfg.S3AccessKey,
			SecretKey: cfg.S3SecretKey,
			UseSSL:    cfg.S3UseSSL,
		})
		if err != nil {
			log.Printf("[WARN] Failed to initialize S3 storage adapter: %v", err)
		} else {
			s3Storage = storageInstance
			go func() {
				if err := s3Storage.EnsureBucket(context.Background()); err != nil {
					log.Printf("[WARN] Failed to ensure S3 bucket exists: %v", err)
				}
			}()
		}
	}

	var attachmentService *attachmentApp.Service
	if repos.AttachmentRepo != nil && s3Storage != nil {
		attachmentService = attachmentApp.NewService(
			repos.AttachmentRepo,
			s3Storage,
			repos.WorkspaceRepo,
			repos.AttachmentRepo,
		)
	}

	var searchService *searchApp.Service
	if repos.TaskRepo != nil && repos.ProjectRepo != nil && repos.WorkspaceRepo != nil {
		searchService = searchApp.NewService(
			repos.TaskRepo,
			repos.ProjectRepo,
			repos.WorkspaceRepo,
		)
	}

	return &Services{
		WorkspaceService:  workspaceService,
		TaskService:       taskService,
		ProjectService:    projectService,
		AutomationService: automationService,
		AutomationWorker:  automationWorker,
		AttachmentService: attachmentService,
		SearchService:     searchService,
		EventBus:          eventBus,
		JWTManager:        jwtManager,
	}
}
