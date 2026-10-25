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
	AvatarURL    *string
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

// UpdateProfile updates the display name of the user enforcing domain invariants.
func (u *User) UpdateProfile(displayName *string) error {
	if displayName != nil {
		trimmed := strings.TrimSpace(*displayName)
		if trimmed == "" {
			return ErrInvalidDisplayName
		}
		if len([]rune(trimmed)) > 100 {
			return ErrInvalidDisplayName
		}
		u.DisplayName = trimmed
	}
	return nil
}

// SetAvatar updates the avatar URL of the user.
func (u *User) SetAvatar(avatarURL string) {
	trimmed := strings.TrimSpace(avatarURL)
	if trimmed == "" {
		u.AvatarURL = nil
	} else {
		u.AvatarURL = &trimmed
	}
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

// NewWorkspace constructs a unified workspace.
func NewWorkspace(ownerID uuid.UUID, name string) (*Workspace, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > 255 {
		return nil, ErrInvalidWorkspaceName
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
	OwnerID   uuid.UUID
	Name      string
	Type      WorkspaceType
	Role      MemberRole
	CreatedAt time.Time
}

type WorkspaceInvite struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	InvitedByID uuid.UUID
	Token       string
	ExpiresAt   *time.Time
	CreatedAt   time.Time
}

// NewWorkspaceInvite creates a new workspace invite entity.
func NewWorkspaceInvite(workspaceID, invitedByID uuid.UUID, token string, expiresAt *time.Time) (*WorkspaceInvite, error) {
	trimmedToken := strings.TrimSpace(token)
	if trimmedToken == "" {
		return nil, ErrInvalidInviteToken
	}

	return &WorkspaceInvite{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		InvitedByID: invitedByID,
		Token:       trimmedToken,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// IsExpired checks whether the invite has expired.
func (i *WorkspaceInvite) IsExpired() bool {
	if i.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*i.ExpiresAt)
}
