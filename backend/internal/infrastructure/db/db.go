package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps pgxpool connection pool.
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new database connection pool.
func New(ctx context.Context, dsn string) (*DB, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 2
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// Short timeout for initial ping to avoid blocking app boot if DB is temporarily down
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err := pool.Ping(pingCtx); err != nil {
		// Log or return pool anyway to allow service to start in degraded mode if desired
		return &DB{Pool: pool}, fmt.Errorf("ping db: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Ping checks database availability.
func (d *DB) Ping(ctx context.Context) error {
	if d == nil || d.Pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	return d.Pool.Ping(ctx)
}

// Close closes database connection pool.
func (d *DB) Close() {
	if d != nil && d.Pool != nil {
		d.Pool.Close()
	}
}

