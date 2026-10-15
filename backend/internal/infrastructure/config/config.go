package config

import (
	"os"
	"strconv"
)

// Config holds all service configuration loaded from environment.
type Config struct {
	Port                   int
	Environment            string
	LogLevel               string
	DatabaseURL            string
	JWTSecret              string
	S3Endpoint             string
	S3Region               string
	S3Bucket               string
	S3AccessKey            string
	S3SecretKey            string
	S3UseSSL               bool
	MaxAttachmentSizeBytes int64
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "takt-dev-secret-key-change-in-production-123456"
	}

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint == "" {
		s3Endpoint = "http://localhost:9000"
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		s3Region = "us-east-1"
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		s3Bucket = "takt-attachments"
	}

	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	if s3AccessKey == "" {
		s3AccessKey = "minioadmin"
	}

	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	if s3SecretKey == "" {
		s3SecretKey = "minioadmin"
	}

	s3UseSSL := false
	if sslStr := os.Getenv("S3_USE_SSL"); sslStr != "" {
		if ssl, err := strconv.ParseBool(sslStr); err == nil {
			s3UseSSL = ssl
		}
	}

	maxAttachmentSize := int64(52428800) // 50 MB
	if sizeStr := os.Getenv("MAX_ATTACHMENT_SIZE_BYTES"); sizeStr != "" {
		if s, err := strconv.ParseInt(sizeStr, 10, 64); err == nil && s > 0 {
			maxAttachmentSize = s
		}
	}

	return &Config{
		Port:                   port,
		Environment:            env,
		LogLevel:               logLevel,
		DatabaseURL:            databaseURL,
		JWTSecret:              jwtSecret,
		S3Endpoint:             s3Endpoint,
		S3Region:               s3Region,
		S3Bucket:               s3Bucket,
		S3AccessKey:            s3AccessKey,
		S3SecretKey:            s3SecretKey,
		S3UseSSL:               s3UseSSL,
		MaxAttachmentSizeBytes: maxAttachmentSize,
	}, nil
}
