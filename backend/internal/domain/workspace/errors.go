package workspace

import "errors"

var (
	ErrUserAlreadyExists   = errors.New("user with this email already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrWorkspaceNotFound   = errors.New("workspace not found")
	ErrInvalidDisplayName  = errors.New("display name cannot be empty")
	ErrInvalidEmail        = errors.New("invalid email address")
	ErrInvalidPassword     = errors.New("password must be at least 6 characters")
	ErrAccessDenied        = errors.New("access denied to workspace")
)
