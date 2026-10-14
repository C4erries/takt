package task

import "errors"

var (
	ErrTaskNotFound     = errors.New("task not found")
	ErrInvalidTaskTitle = errors.New("task title cannot be empty")
	ErrInvalidPriority  = errors.New("invalid task priority")
	ErrAccessDenied     = errors.New("access denied to task")
)
