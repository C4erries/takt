package workspace

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type WorkspaceRepository interface {
	Create(ctx context.Context, ws *Workspace) error
	GetByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
	CreateMember(ctx context.Context, member *WorkspaceMember) error
	GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*WorkspaceMember, error)
	GetWorkspacesByUserID(ctx context.Context, userID uuid.UUID) ([]*UserWorkspace, error)
	GetPersonalWorkspaceByUserID(ctx context.Context, userID uuid.UUID) (*UserWorkspace, error)
}
