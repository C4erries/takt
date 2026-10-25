package workspace

import "errors"

var (
	ErrUserAlreadyExists    = errors.New("user with this email already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrWorkspaceNotFound    = errors.New("workspace not found")
	ErrInvalidDisplayName   = errors.New("display name cannot be empty")
	ErrInvalidEmail         = errors.New("invalid email address")
	ErrInvalidPassword      = errors.New("password must be at least 6 characters")
	ErrAccessDenied         = errors.New("access denied to workspace")
	ErrInvalidWorkspaceName = errors.New("workspace name cannot be empty or exceed 255 characters")
	ErrInviteNotFound       = errors.New("invite not found")
	ErrInviteExpired        = errors.New("invite has expired")
	ErrAlreadyMember        = errors.New("user is already a member of this workspace")
	ErrCannotRemoveOwner    = errors.New("cannot remove workspace owner")
	ErrInvalidInviteToken   = errors.New("invalid invite token")
	ErrInvalidAvatarFormat  = errors.New("invalid avatar format: only jpeg, png and webp are allowed")
	ErrAvatarTooLarge       = errors.New("avatar file exceeds maximum size of 5MB")
)
