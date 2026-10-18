package automation

import (
	"encoding/json"
)

// DSL limits defined in ADR-0004
const (
	MaxConditions = 5
	MaxActions    = 5
)

// Valid DSL Trigger types (ADR-0004)
var validDSLTriggers = map[string]bool{
	TriggerTaskEnteredColumn:     true,
	TriggerTaskCompletedInColumn: true,
	TriggerTaskFieldChanged:      true,
	TriggerColumnSLATimer:        true,
	TriggerScheduledSweep:        true,
}

// IsValidDSLTrigger validates if the trigger type is recognized in Low-Code DSL.
func IsValidDSLTrigger(trigger string) bool {
	return validDSLTriggers[trigger]
}

// Valid Condition fields
const (
	ConditionFieldPriority          = "priority"
	ConditionFieldIsCompleted       = "is_completed"
	ConditionFieldDueDate           = "due_date"
	ConditionFieldDueDateStatus     = "due_date_status"
	ConditionFieldTimeInColumnHours = "time_in_column_hours"
	ConditionFieldTags              = "tags"
)

var validConditionFields = map[string]bool{
	ConditionFieldPriority:          true,
	ConditionFieldIsCompleted:       true,
	ConditionFieldDueDate:           true,
	ConditionFieldDueDateStatus:     true,
	ConditionFieldTimeInColumnHours: true,
	ConditionFieldTags:              true,
}

// IsValidConditionField validates if the condition field is recognized.
func IsValidConditionField(field string) bool {
	return validConditionFields[field]
}

// Valid Condition operators
const (
	OperatorEquals      = "equals"
	OperatorNotEquals   = "not_equals"
	OperatorContains    = "contains"
	OperatorNotContains = "not_contains"
	OperatorIsEmpty     = "is_empty"
	OperatorIsNotEmpty  = "is_not_empty"
	OperatorGreaterThan = "greater_than"
	OperatorLessThan    = "less_than"
	OperatorIn          = "in"
	OperatorNotIn       = "not_in"
)

var validConditionOperators = map[string]bool{
	OperatorEquals:      true,
	OperatorNotEquals:   true,
	OperatorContains:    true,
	OperatorNotContains: true,
	OperatorIsEmpty:     true,
	OperatorIsNotEmpty:  true,
	OperatorGreaterThan: true,
	OperatorLessThan:    true,
	OperatorIn:          true,
	OperatorNotIn:       true,
}

// IsValidConditionOperator validates if the condition operator is recognized.
func IsValidConditionOperator(op string) bool {
	return validConditionOperators[op]
}

// Valid Action types
const (
	ActionMoveToColumn    = "move_to_column"
	ActionSetPriority     = "set_priority"
	ActionAddTag          = "add_tag"
	ActionRemoveTag       = "remove_tag"
	ActionSetDueDate      = "set_due_date"
	ActionArchiveTask     = "archive_task"
	ActionMarkCompleted   = "mark_completed"
	ActionMarkUncompleted = "mark_uncompleted"
	ActionClearDueDate    = "clear_due_date"
)

var validActions = map[string]bool{
	ActionMoveToColumn:    true,
	ActionSetPriority:     true,
	ActionAddTag:          true,
	ActionRemoveTag:       true,
	ActionSetDueDate:      true,
	ActionArchiveTask:     true,
	ActionMarkCompleted:   true,
	ActionMarkUncompleted: true,
	ActionClearDueDate:    true,
}

// IsValidActionType validates if the action type is recognized.
func IsValidActionType(action string) bool {
	return validActions[action]
}

// Trigger represents the trigger configuration of an automation rule.
type Trigger struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// Condition represents a single predicate in an automation rule.
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value,omitempty"`
}

// Logical operators for conditions evaluation
const (
	LogicAnd = "and"
	LogicOr  = "or"
)

// Action represents an operation to execute when an automation rule triggers and matches.
type Action struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// RuleDefinition represents the full Trigger-Condition-Action (TCA) specification of an automation.
type RuleDefinition struct {
	Trigger         Trigger     `json:"trigger"`
	LogicalOperator string      `json:"logical_operator,omitempty"`
	Conditions      []Condition `json:"conditions"`
	Actions         []Action    `json:"actions"`
}

// UnmarshalJSON supports both snake_case (logical_operator) and camelCase (logicalOperator).
func (r *RuleDefinition) UnmarshalJSON(data []byte) error {
	type Alias RuleDefinition
	aux := &struct {
		*Alias
		LogicalOpAlt string `json:"logicalOperator"`
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if r.LogicalOperator == "" && aux.LogicalOpAlt != "" {
		r.LogicalOperator = aux.LogicalOpAlt
	}
	return nil
}

// Validate verifies that the rule definition satisfies all domain invariants according to ADR-0004:
// 1. Trigger type is a valid DSL trigger.
// 2. Conditions count does not exceed MaxConditions (5).
// 3. Every condition uses a recognized field and operator.
// 4. Logical operator is either empty, "and", or "or".
// 5. Actions slice contains between 1 and MaxActions (5) actions.
// 6. Every action uses a recognized action type.
func (r *RuleDefinition) Validate() error {
	if r == nil {
		return ErrRuleDefinitionNil
	}

	if !IsValidDSLTrigger(r.Trigger.Type) {
		return ErrInvalidTriggerType
	}

	if r.LogicalOperator != "" && r.LogicalOperator != LogicAnd && r.LogicalOperator != LogicOr {
		return ErrInvalidLogicalOperator
	}

	if len(r.Conditions) > MaxConditions {
		return ErrTooManyConditions
	}

	for _, c := range r.Conditions {
		if !IsValidConditionField(c.Field) {
			return ErrInvalidConditionField
		}
		if !IsValidConditionOperator(c.Operator) {
			return ErrInvalidConditionOperator
		}
	}

	if len(r.Actions) == 0 {
		return ErrEmptyActions
	}

	if len(r.Actions) > MaxActions {
		return ErrTooManyActions
	}

	for _, a := range r.Actions {
		if !IsValidActionType(a.Type) {
			return ErrInvalidActionType
		}
	}

	return nil
}

// ValidateRuleDefinition validates a RuleDefinition instance.
func ValidateRuleDefinition(r *RuleDefinition) error {
	if r == nil {
		return ErrRuleDefinitionNil
	}
	return r.Validate()
}

// ToJSON marshals the rule definition to a json.RawMessage.
func (r *RuleDefinition) ToJSON() (json.RawMessage, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(r)
}

// RuleDefinitionFromJSON parses and validates a RuleDefinition from JSON bytes.
func RuleDefinitionFromJSON(data []byte) (*RuleDefinition, error) {
	if len(data) == 0 || string(data) == "{}" || string(data) == "null" {
		return nil, nil
	}
	var def RuleDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, err
	}
	if err := def.Validate(); err != nil {
		return nil, err
	}
	return &def, nil
}
