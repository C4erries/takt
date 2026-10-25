package workspace

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	domainWorkspace "takt/internal/domain/workspace"
)

// Transactor executes database operations within an atomic transaction.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

// AvatarStorage defines binary storage operations for user avatars.
type AvatarStorage interface {
	PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, key string) (io.ReadCloser, string, int64, error)
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

// AutomationSeeder abstracts seeding preset automations for new workspaces.
type AutomationSeeder interface {
	SeedDefaultAutomations(ctx context.Context, workspaceID uuid.UUID) error
}

// WorkspaceMemberInfo represents member details combined with user profile.
type WorkspaceMemberInfo struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Email       string
	DisplayName string
	Role        domainWorkspace.MemberRole
	CreatedAt   time.Time
}

// InvitePreviewInfo represents public preview information for an invite.
type InvitePreviewInfo struct {
	Token         string
	WorkspaceName string
	InvitedByName string
	ExpiresAt     *time.Time
}

// SwitchWorkspaceOutput contains token and active workspace info.
type SwitchWorkspaceOutput struct {
	Token     string
	Workspace *domainWorkspace.UserWorkspace
}

// UserRepository alias to domain repository.
type UserRepository = domainWorkspace.UserRepository

// WorkspaceRepository defines storage methods for workspaces, members, and invites.
type WorkspaceRepository interface {
	Create(ctx context.Context, ws *domainWorkspace.Workspace) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainWorkspace.Workspace, error)
	UpdateName(ctx context.Context, id, ownerID uuid.UUID, name string) (*domainWorkspace.Workspace, error)
	CreateMember(ctx context.Context, member *domainWorkspace.WorkspaceMember) error
	GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*domainWorkspace.WorkspaceMember, error)
	ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]*WorkspaceMemberInfo, error)
	DeleteMember(ctx context.Context, workspaceID, userID uuid.UUID) error
	GetWorkspacesByUserID(ctx context.Context, userID uuid.UUID) ([]*domainWorkspace.UserWorkspace, error)
	GetPersonalWorkspaceByUserID(ctx context.Context, userID uuid.UUID) (*domainWorkspace.UserWorkspace, error)
	CreateInvite(ctx context.Context, invite *domainWorkspace.WorkspaceInvite) error
	GetInviteByToken(ctx context.Context, token string) (*domainWorkspace.WorkspaceInvite, *InvitePreviewInfo, error)
	GetActiveInvite(ctx context.Context, workspaceID uuid.UUID) (*domainWorkspace.WorkspaceInvite, error)
	DeleteInvite(ctx context.Context, id uuid.UUID) error
	DeleteInviteByToken(ctx context.Context, token string) error
}
