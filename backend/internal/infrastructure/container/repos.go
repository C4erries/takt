package container

import (
	"context"
	"log"

	"takt/internal/infrastructure/config"
	"takt/internal/infrastructure/db"
)

// Repositories container holding all database repositories.
type Repositories struct {
	DB *db.DB
}

// initRepositories initializes database connection pool and repositories.
func initRepositories(ctx context.Context, cfg *config.Config) (*Repositories, error) {
	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("[WARN] Failed to connect to database at startup: %v (service will run in degraded mode)", err)
	}

	return &Repositories{
		DB: database,
	}, nil
}
