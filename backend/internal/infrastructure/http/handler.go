package http

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	automationApp "takt/internal/application/automation"
	projectApp "takt/internal/application/project"
	taskApp "takt/internal/application/task"
	workspaceApp "takt/internal/application/workspace"
	"takt/internal/infrastructure/http/generated"
)

var _ generated.ServerInterface = (*Handler)(nil)

var startTime = time.Now()

// DBChecker defines an interface for pinging the database.
type DBChecker interface {
	Ping(ctx context.Context) error
}

// Handler holds HTTP dependencies and handlers.
type Handler struct {
	dbChecker         DBChecker
	version           string
	isProduction      bool
	workspaceService  workspaceApp.UseCases
	taskService       taskApp.UseCases
	projectService    projectApp.UseCases
	automationService automationApp.UseCases
	jwtManager        *JWTManager
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
	automationService ...automationApp.UseCases,
) *Handler {
	var autoSvc automationApp.UseCases
	if len(automationService) > 0 {
		autoSvc = automationService[0]
	}
	return &Handler{
		dbChecker:         dbChecker,
		version:           version,
		isProduction:      isProduction,
		workspaceService:  workspaceService,
		taskService:       taskService,
		projectService:    projectService,
		automationService: autoSvc,
		jwtManager:        jwtManager,
	}
}

// JWTManager returns configured JWT manager.
func (h *Handler) JWTManager() *JWTManager {
	return h.jwtManager
}

// GetHealth handles GET /health and GET /api/v1/health.
func (h *Handler) GetHealth(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	dbStatus := generated.Connected
	overallStatus := generated.HealthStatusOk

	if h.dbChecker != nil {
		if err := h.dbChecker.Ping(ctx); err != nil {
			dbStatus = generated.Disconnected
			overallStatus = generated.HealthStatusDegraded
		}
	} else {
		dbStatus = generated.Disconnected
		overallStatus = generated.HealthStatusDegraded
	}

	uptime := int(time.Since(startTime).Seconds())
	resp := generated.Health{
		Status:        overallStatus,
		Version:       h.version,
		UptimeSeconds: &uptime,
		Database: struct {
			Status generated.HealthDatabaseStatus `json:"status"`
		}{
			Status: dbStatus,
		},
	}

	return c.JSON(http.StatusOK, resp)
}

// HealthCheck provides backward compatibility for health checking.
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.GetHealth(c)
}
