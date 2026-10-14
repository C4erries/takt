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

			workspaces := protected.Group("/workspaces/:workspace_id")
			{
				// Projects CRUD
				workspaces.GET("/projects", h.ListProjects)
				workspaces.POST("/projects", h.CreateProject)
				workspaces.GET("/projects/:project_id", h.GetProject)
				workspaces.PATCH("/projects/:project_id", h.UpdateProject)
				workspaces.DELETE("/projects/:project_id", h.DeleteProject)

				// Project Boards
				workspaces.GET("/projects/:project_id/boards", h.ListProjectBoards)
				workspaces.POST("/projects/:project_id/boards", h.CreateProjectBoard)

				// Board Columns
				workspaces.GET("/boards/:board_id/columns", h.ListBoardColumns)
				workspaces.POST("/boards/:board_id/columns", h.CreateBoardColumn)
				workspaces.PATCH("/boards/:board_id/columns/:column_id", h.UpdateBoardColumn)
				workspaces.DELETE("/boards/:board_id/columns/:column_id", h.DeleteBoardColumn)
				workspaces.DELETE("/columns/:column_id", h.DeleteColumnDirect)

				// Move Task
				workspaces.POST("/tasks/:task_id/move", h.MoveTaskToColumn)

				// Project Tasks
				workspaces.GET("/projects/:project_id/tasks", h.ListProjectTasks)
			}
		}
	}
}
