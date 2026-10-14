package http

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

var startTime = time.Now()

// DBStatus represents the database connectivity status.
type DBStatus struct {
	Status string `json:"status"`
}

// HealthResponse represents the payload returned by healthcheck endpoints.
type HealthResponse struct {
	Status        string   `json:"status"`
	Version       string   `json:"version"`
	UptimeSeconds int64    `json:"uptime_seconds"`
	Database      DBStatus `json:"database"`
}

// HealthCheck handles GET /health and GET /api/v1/health.
func (h *Handler) HealthCheck(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	dbStatus := "connected"
	overallStatus := "ok"

	if h.dbChecker != nil {
		if err := h.dbChecker.Ping(ctx); err != nil {
			dbStatus = "disconnected"
			overallStatus = "degraded"
		}
	} else {
		dbStatus = "disconnected"
		overallStatus = "degraded"
	}

	resp := HealthResponse{
		Status:        overallStatus,
		Version:       h.version,
		UptimeSeconds: int64(time.Since(startTime).Seconds()),
		Database: DBStatus{
			Status: dbStatus,
		},
	}

	return c.JSON(http.StatusOK, resp)
}
