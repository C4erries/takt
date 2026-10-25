package container

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	httpDelivery "takt/internal/infrastructure/http"
	"takt/internal/infrastructure/http/generated"
)

// setupRouter registers all HTTP routes and groups.
func setupRouter(e *echo.Echo, h *httpDelivery.Handler) {
	// Standard middlewares
	httpDelivery.SetupMiddlewares(e)

	// Direct root health check
	e.GET("/health", h.GetHealth)

	// API v1 group
	v1 := e.Group("/api/v1")

	// Apply AuthMiddleware with skipper for public endpoints
	v1.Use(httpDelivery.AuthMiddlewareWithSkipper(h.JWTManager(), func(c echo.Context) bool {
		p := strings.TrimSuffix(c.Path(), "/")
		reqPath := strings.TrimSuffix(c.Request().URL.Path, "/")
		return strings.HasSuffix(p, "/auth/login") ||
			strings.HasSuffix(p, "/auth/register") ||
			strings.HasSuffix(p, "/auth/logout") ||
			strings.HasSuffix(p, "/health") ||
			strings.HasSuffix(reqPath, "/auth/login") ||
			strings.HasSuffix(reqPath, "/auth/register") ||
			strings.HasSuffix(reqPath, "/auth/logout") ||
			strings.HasSuffix(reqPath, "/health") ||
			(c.Request().Method == http.MethodGet && (strings.Contains(p, "/invites/") || strings.Contains(reqPath, "/invites/"))) ||
			((c.Request().Method == http.MethodGet || c.Request().Method == http.MethodHead) && (strings.Contains(p, "/avatar") || strings.Contains(reqPath, "/avatar")))
	}))

	generated.RegisterHandlers(v1, h)

	// User avatar binary download routes
	v1.Match([]string{http.MethodGet, http.MethodHead}, "/users/me/avatar/*", h.GetUserAvatar)
	v1.Match([]string{http.MethodGet, http.MethodHead}, "/users/avatar/*", h.GetUserAvatar)
}

