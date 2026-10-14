package http

import (
	"context"
)

// DBChecker defines an interface for pinging the database.
type DBChecker interface {
	Ping(ctx context.Context) error
}

// Handler holds HTTP dependencies and handlers.
type Handler struct {
	dbChecker DBChecker
	version   string
}

// NewHandler constructs a new HTTP Handler.
func NewHandler(dbChecker DBChecker, version string) *Handler {
	return &Handler{
		dbChecker: dbChecker,
		version:   version,
	}
}

