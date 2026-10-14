package http

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"takt/internal/infrastructure/http/generated"
)

var startTime = time.Now()

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
