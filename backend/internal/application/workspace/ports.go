package workspace

import (
	"context"

	"github.com/google/uuid"
	domainWorkspace "takt/internal/domain/workspace"
)

// Transactor executes database operations within an atomic transaction.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

// PasswordHasher handles password hashing and verification.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

// TokenGenerator generates authentication tokens.
type TokenGenerator interface {
	GenerateToken(userID, workspaceID uuid.UUID) (string, error)
}

// InitialTasksCreator abstracts seeding onboarding tasks for new workspaces.
type InitialTasksCreator interface {
	CreateInitialInboxTasks(ctx context.Context, workspaceID uuid.UUID) error
}

// UserRepository alias to domain repository.
type UserRepository = domainWorkspace.UserRepository

// WorkspaceRepository alias to domain repository.
type WorkspaceRepository = domainWorkspace.WorkspaceRepository
