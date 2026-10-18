package automation

import "errors"

var (
	// ErrAutomationNotFound indicates that the requested automation rule was not found.
	ErrAutomationNotFound = errors.New("automation not found")

	// ErrInvalidTriggerType indicates that the trigger type is unsupported or unknown.
	ErrInvalidTriggerType = errors.New("invalid automation trigger type")

	// ErrInvalidAutomationName indicates that the rule name is empty or too long.
	ErrInvalidAutomationName = errors.New("invalid automation name")

	// ErrInvalidColumnID indicates that the provided column ID is invalid.
	ErrInvalidColumnID = errors.New("invalid column ID")
)

