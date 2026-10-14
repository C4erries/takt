package http

import (
	"context"

	projectApp "takt/internal/application/project"
	taskApp "takt/internal/application/task"
	workspaceApp "takt/internal/application/workspace"
	"takt/internal/infrastructure/http/generated"
)

var _ generated.ServerInterface = (*Handler)(nil)

// DBChecker defines an interface for pinging the database.
type DBChecker interface {
	Ping(ctx context.Context) error
}

// Handler holds HTTP dependencies and handlers.
type Handler struct {
	dbChecker        DBChecker
	version          string
	isProduction     bool
	workspaceService workspaceApp.UseCases
	taskService      taskApp.UseCases
	projectService   projectApp.UseCases
	jwtManager       *JWTManager
}

// NewHandler constructs a new HTTP Handler.
func NewHandler(
	dbChecker DBChecker,
	version string,
	isProduction bool,
	workspaceService workspaceApp.UseCases,
	taskService taskApp.UseCases,
	projectService projectApp.UseCases,
	jwtManager *JWTManager,
) *Handler {
	return &Handler{
		dbChecker:        dbChecker,
		version:          version,
		isProduction:     isProduction,
		workspaceService: workspaceService,
		taskService:      taskService,
		projectService:   projectService,
		jwtManager:       jwtManager,
	}
}

// JWTManager returns configured JWT manager.
func (h *Handler) JWTManager() *JWTManager {
	return h.jwtManager
}
