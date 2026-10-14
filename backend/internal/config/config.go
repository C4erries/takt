package config

import (
	"os"
	"strconv"
)

// Config holds all service configuration loaded from environment.
type Config struct {
	Port        int
	Environment string
	LogLevel    string
	DatabaseURL string
}

// Load loads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "debug"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/takt_db?sslmode=disable"
	}

	return &Config{
		Port:        port,
		Environment: env,
		LogLevel:    logLevel,
		DatabaseURL: databaseURL,
	}, nil
}

