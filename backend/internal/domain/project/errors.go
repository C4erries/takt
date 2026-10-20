package project

import "errors"

var (
	ErrProjectNotFound       = errors.New("project not found")
	ErrInvalidProjectName    = errors.New("project name must be between 1 and 255 characters")
	ErrInvalidProjectColor   = errors.New("invalid project color")
	ErrBoardNotFound         = errors.New("board not found")
	ErrInvalidBoardName      = errors.New("board name must be between 1 and 255 characters")
	ErrInvalidBoardID        = errors.New("invalid board id")
	ErrColumnNotFound        = errors.New("column not found")
	ErrInvalidColumnName     = errors.New("column name must be between 1 and 255 characters")
	ErrInvalidStatusCategory = errors.New("invalid status category")
	ErrInvalidPosition       = errors.New("position must be greater than zero")
	ErrColumnNotEmpty        = errors.New("cannot delete column containing tasks")
	ErrLastBoard             = errors.New("cannot delete the only board in a project")
	ErrAccessDenied          = errors.New("access denied to project")
)

