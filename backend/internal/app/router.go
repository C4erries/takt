package app

import (
	"github.com/labstack/echo/v4"

	httpDelivery "takt/internal/delivery/http"
)

// setupRouter registers all HTTP routes and groups.
func setupRouter(e *echo.Echo, h *httpDelivery.Handler) {
	// Standard middlewares
	httpDelivery.SetupMiddlewares(e)

	// Direct root health check
	e.GET("/health", h.HealthCheck)

	// API v1 group
	v1 := e.Group("/api/v1")
	{
		v1.GET("/health", h.HealthCheck)
	}
}

