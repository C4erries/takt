package container

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"takt/internal/infrastructure/config"
	httpDelivery "takt/internal/infrastructure/http"
)

const AppVersion = "0.3.1"

// Run boots up the application server and listens for graceful shutdown.
func Run(ctx context.Context, cfg *config.Config) error {
	repos, err := initRepositories(ctx, cfg)
	if err != nil {
		return fmt.Errorf("init repositories: %w", err)
	}
	defer func() {
		if repos.DB != nil {
			repos.DB.Close()
		}
	}()

	services := initServices(repos, cfg)

	var dbChecker httpDelivery.DBChecker
	if repos.DB != nil {
		dbChecker = repos.DB
	}

	isProduction := cfg.Environment == "production"
	handler := httpDelivery.NewHandler(
		dbChecker,
		AppVersion,
		isProduction,
		services.WorkspaceService,
		services.TaskService,
		services.ProjectService,
		services.JWTManager,
		services.AutomationService,
	)

	if services.AttachmentService != nil {
		handler.SetAttachmentService(services.AttachmentService, cfg.MaxAttachmentSizeBytes)
	}

	if services.SearchService != nil {
		handler.SetSearchService(services.SearchService)
	}

	e := echo.New()
	e.HideBanner = true
	setupRouter(e, handler)

	serverAddr := fmt.Sprintf(":%d", cfg.Port)

	// Start server in background
	serverErrCh := make(chan error, 1)
	go func() {
		log.Printf("[INFO] Takt Backend server v%s listening on %s (env: %s)", AppVersion, serverAddr, cfg.Environment)
		if err := e.Start(serverAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
		close(serverErrCh)
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Printf("[INFO] Shutdown signal received, shutting down gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if services.EventBus != nil {
			if err := services.EventBus.Close(shutdownCtx); err != nil {
				log.Printf("[WARN] EventBus shutdown error: %v", err)
			}
		}

		if err := e.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
		log.Printf("[INFO] Takt Backend server stopped successfully")
		return nil

	case err := <-serverErrCh:
		return fmt.Errorf("server startup failure: %w", err)
	}
}
