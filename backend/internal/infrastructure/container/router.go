package container

import (
	"github.com/labstack/echo/v4"

	httpDelivery "takt/internal/infrastructure/http"
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

		// Public Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
			auth.POST("/logout", h.Logout)
		}

		// Protected routes (Tenant Isolation & JWT Auth)
		protected := v1.Group("", httpDelivery.AuthMiddleware(h.JWTManager()))
		{
			protected.GET("/auth/me", h.GetMe)

			tasks := protected.Group("/tasks")
			{
				tasks.GET("/inbox", h.GetInboxTasks)
				tasks.POST("/inbox", h.CreateInboxTask)
				tasks.PATCH("/:id/toggle", h.ToggleTask)
				tasks.DELETE("/:id", h.DeleteTask)
			}
		}
	}
}
