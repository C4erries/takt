package task

import "errors"

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrInvalidTaskTitle   = errors.New("task title cannot be empty")
	ErrInvalidPriority    = errors.New("invalid task priority")
	ErrAccessDenied       = errors.New("access denied to task")
	ErrInvalidTagName     = errors.New("tag name must be between 1 and 64 characters and contain only allowed characters")
	ErrInvalidTagColor    = errors.New("invalid tag color")
	ErrTagNotFound        = errors.New("tag not found")
	ErrTagAlreadyAttached = errors.New("tag already attached to task")
	ErrInvalidFocusOrder  = errors.New("focus order must be non-negative")
)
