package attachment

import "errors"

var (
	ErrInvalidFileName    = errors.New("file name cannot be empty")
	ErrFileTooLarge       = errors.New("file size exceeds maximum allowed size")
	ErrInvalidFileSize    = errors.New("file size must be greater than zero")
	ErrInvalidWorkspaceID = errors.New("workspace ID cannot be nil")
	ErrInvalidTaskID      = errors.New("task ID cannot be nil")
	ErrInvalidCreatedByID = errors.New("created by ID cannot be nil")
	ErrAttachmentNotFound = errors.New("attachment not found")
)

