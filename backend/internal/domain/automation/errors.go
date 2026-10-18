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

	// ErrRuleDefinitionNil indicates that the rule definition cannot be nil.
	ErrRuleDefinitionNil = errors.New("rule definition cannot be nil")

	// ErrTooManyConditions indicates that the number of conditions exceeds the maximum limit (5).
	ErrTooManyConditions = errors.New("rule conditions exceed maximum allowed limit of 5")

	// ErrTooManyActions indicates that the number of actions exceeds the maximum limit (5).
	ErrTooManyActions = errors.New("rule actions exceed maximum allowed limit of 5")

	// ErrEmptyActions indicates that the rule definition contains no actions.
	ErrEmptyActions = errors.New("rule must contain at least one action")

	// ErrInvalidActionType indicates that an action type is unsupported or unknown.
	ErrInvalidActionType = errors.New("invalid automation action type")

	// ErrInvalidConditionField indicates that a condition field is unsupported or unknown.
	ErrInvalidConditionField = errors.New("invalid condition field")

	// ErrInvalidConditionOperator indicates that a condition operator is unsupported or unknown.
	ErrInvalidConditionOperator = errors.New("invalid condition operator")

	// ErrInvalidLogicalOperator indicates that the logical operator is unsupported (must be "and" or "or").
	ErrInvalidLogicalOperator = errors.New("invalid logical operator")
)
