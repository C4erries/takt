package workspace

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type WorkspaceType string

const (
	TypePersonal WorkspaceType = "personal"
	TypeShared   WorkspaceType = "shared"
)

type MemberRole string

const (
	RoleOwner  MemberRole = "owner"
	RoleMember MemberRole = "member"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	DisplayName  string
	CreatedAt    time.Time
}

func NewUser(email, passwordHash, displayName string) (*User, error) {
	trimmedEmail := strings.TrimSpace(strings.ToLower(email))
	if trimmedEmail == "" || !strings.Contains(trimmedEmail, "@") {
		return nil, ErrInvalidEmail
	}
	trimmedName := strings.TrimSpace(displayName)
	if trimmedName == "" {
		return nil, ErrInvalidDisplayName
	}

	return &User{
		ID:           uuid.New(),
		Email:        trimmedEmail,
		PasswordHash: passwordHash,
		DisplayName:  trimmedName,
		CreatedAt:    time.Now().UTC(),
	}, nil
}

type Workspace struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	Type      WorkspaceType
	CreatedAt time.Time
}

func NewPersonalWorkspace(ownerID uuid.UUID, name string) (*Workspace, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		trimmedName = "Личное"
	}

	return &Workspace{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		Name:      trimmedName,
		Type:      TypePersonal,
		CreatedAt: time.Now().UTC(),
	}, nil
}

type WorkspaceMember struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Role        MemberRole
	CreatedAt   time.Time
}

func NewWorkspaceMember(workspaceID, userID uuid.UUID, role MemberRole) (*WorkspaceMember, error) {
	return &WorkspaceMember{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        role,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

type UserWorkspace struct {
	ID        uuid.UUID
	Name      string
	Type      WorkspaceType
	Role      MemberRole
	CreatedAt time.Time
}
