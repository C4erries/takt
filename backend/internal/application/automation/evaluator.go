package automation

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/domain/events"
)

// Evaluator evaluates conditions and executes actions for Low-Code rules.
type Evaluator struct {
	taskModifier  TaskModifier
	projectReader ProjectReader
}

// NewEvaluator creates a new Evaluator.
func NewEvaluator(taskModifier TaskModifier, projectReader ProjectReader) *Evaluator {
	return &Evaluator{
		taskModifier:  taskModifier,
		projectReader: projectReader,
	}
}

// EvaluateConditions evaluates all conditions with AND logic.
// Returns true if all conditions pass or if conditions slice is empty.
func (e *Evaluator) EvaluateConditions(conditions []domainAutomation.Condition, task *TaskInfo) bool {
	for _, c := range conditions {
		if !e.evaluateCondition(c, task) {
			return false
		}
	}
	return true
}

func (e *Evaluator) evaluateCondition(c domainAutomation.Condition, task *TaskInfo) bool {
	switch c.Field {
	case domainAutomation.ConditionFieldPriority:
		return evaluateStringCondition(task.Priority, c.Operator, c.Value)
	case domainAutomation.ConditionFieldIsCompleted:
		return evaluateBoolCondition(task.IsCompleted, c.Operator, c.Value)
	case domainAutomation.ConditionFieldDueDate:
		return evaluateDateCondition(task.DueDate, c.Operator, c.Value)
	case domainAutomation.ConditionFieldTimeInColumnHours:
		return evaluateNumericCondition(task.TimeInColumnHours, c.Operator, c.Value)
	case domainAutomation.ConditionFieldTags:
		return evaluateTagsCondition(task.Tags, c.Operator, c.Value)
	default:
		return false
	}
}

func evaluateStringCondition(actual string, op string, expectedVal interface{}) bool {
	expected := fmt.Sprint(expectedVal)
	switch op {
	case domainAutomation.OperatorEquals:
		return strings.EqualFold(actual, expected)
	case domainAutomation.OperatorNotEquals:
		return !strings.EqualFold(actual, expected)
	case domainAutomation.OperatorContains:
		return strings.Contains(strings.ToLower(actual), strings.ToLower(expected))
	case domainAutomation.OperatorNotContains:
		return !strings.Contains(strings.ToLower(actual), strings.ToLower(expected))
	case domainAutomation.OperatorIsEmpty:
		return actual == ""
	case domainAutomation.OperatorIsNotEmpty:
		return actual != ""
	default:
		return false
	}
}

func evaluateBoolCondition(actual bool, op string, expectedVal interface{}) bool {
	var expected bool
	switch v := expectedVal.(type) {
	case bool:
		expected = v
	case string:
		expected = strings.EqualFold(v, "true") || v == "1"
	default:
		expected = false
	}

	switch op {
	case domainAutomation.OperatorEquals:
		return actual == expected
	case domainAutomation.OperatorNotEquals:
		return actual != expected
	case domainAutomation.OperatorIsEmpty:
		return !actual
	case domainAutomation.OperatorIsNotEmpty:
		return actual
	default:
		return false
	}
}

func evaluateDateCondition(actual *time.Time, op string, expectedVal interface{}) bool {
	switch op {
	case domainAutomation.OperatorIsEmpty:
		return actual == nil
	case domainAutomation.OperatorIsNotEmpty:
		return actual != nil
	}

	if actual == nil {
		return false
	}

	expStr := fmt.Sprint(expectedVal)
	parsedExp, err := time.Parse("2006-01-02", expStr)
	if err != nil {
		parsedExp, err = time.Parse(time.RFC3339, expStr)
	}

	actDate := actual.UTC().Truncate(24 * time.Hour)
	expDate := parsedExp.UTC().Truncate(24 * time.Hour)

	switch op {
	case domainAutomation.OperatorEquals:
		return err == nil && actDate.Equal(expDate)
	case domainAutomation.OperatorNotEquals:
		return err == nil && !actDate.Equal(expDate)
	case domainAutomation.OperatorGreaterThan:
		return err == nil && actDate.After(expDate)
	case domainAutomation.OperatorLessThan:
		return err == nil && actDate.Before(expDate)
	default:
		return false
	}
}

func evaluateNumericCondition(actual float64, op string, expectedVal interface{}) bool {
	expected := toFloat(expectedVal)
	switch op {
	case domainAutomation.OperatorEquals:
		return actual == expected
	case domainAutomation.OperatorNotEquals:
		return actual != expected
	case domainAutomation.OperatorGreaterThan:
		return actual > expected
	case domainAutomation.OperatorLessThan:
		return actual < expected
	case domainAutomation.OperatorIsEmpty:
		return actual == 0
	case domainAutomation.OperatorIsNotEmpty:
		return actual > 0
	default:
		return false
	}
}

func evaluateTagsCondition(tags []string, op string, expectedVal interface{}) bool {
	switch op {
	case domainAutomation.OperatorIsEmpty:
		return len(tags) == 0
	case domainAutomation.OperatorIsNotEmpty:
		return len(tags) > 0
	}

	exp := strings.ToLower(strings.TrimSpace(fmt.Sprint(expectedVal)))
	found := false
	for _, t := range tags {
		if strings.EqualFold(strings.TrimSpace(t), exp) {
			found = true
			break
		}
	}

	switch op {
	case domainAutomation.OperatorEquals, domainAutomation.OperatorContains:
		return found
	case domainAutomation.OperatorNotEquals, domainAutomation.OperatorNotContains:
		return !found
	default:
		return false
	}
}

func toFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}

// ExecuteRule checks Cycle Guard, evaluates conditions, and runs actions sequentially.
func (e *Evaluator) ExecuteRule(ctx context.Context, rule *domainAutomation.RuleDefinition, task *TaskInfo) error {
	if rule == nil || task == nil {
		return nil
	}

	if !events.CanExecuteAutomation(ctx) {
		return nil
	}

	if !e.EvaluateConditions(rule.Conditions, task) {
		return nil
	}

	autoCtx := events.WithAutomationTrigger(ctx)
	for _, action := range rule.Actions {
		_ = e.executeAction(autoCtx, action, task)
	}
	return nil
}

func (e *Evaluator) executeAction(ctx context.Context, action domainAutomation.Action, task *TaskInfo) error {
	if e.taskModifier == nil {
		return nil
	}

	switch action.Type {
	case domainAutomation.ActionMoveToColumn:
		targetColIDStr := extractStringParam(action.Params, "target_column_id", "targetColumnId")
		if targetColIDStr != "" {
			targetColUUID, err := uuid.Parse(targetColIDStr)
			if err == nil && targetColUUID != uuid.Nil {
				return e.taskModifier.MoveTaskToColumn(ctx, task.ID, task.WorkspaceID, &targetColUUID, task.ProjectID, 1000.0)
			}
		}

	case domainAutomation.ActionSetPriority:
		priority := extractStringParam(action.Params, "priority")
		if priority != "" {
			return e.taskModifier.SetTaskPriority(ctx, task.ID, task.WorkspaceID, priority)
		}

	case domainAutomation.ActionAddTag:
		tagIDStr := extractStringParam(action.Params, "tag_id", "tagId")
		tagName := extractStringParam(action.Params, "name", "tag_name")
		var tagID *uuid.UUID
		if tagIDStr != "" {
			if parsed, err := uuid.Parse(tagIDStr); err == nil && parsed != uuid.Nil {
				tagID = &parsed
			}
		}
		return e.taskModifier.AddTagToTask(ctx, task.ID, task.WorkspaceID, tagID, tagName)

	case domainAutomation.ActionRemoveTag:
		tagIDStr := extractStringParam(action.Params, "tag_id", "tagId")
		tagName := extractStringParam(action.Params, "name", "tag_name")
		var tagID *uuid.UUID
		if tagIDStr != "" {
			if parsed, err := uuid.Parse(tagIDStr); err == nil && parsed != uuid.Nil {
				tagID = &parsed
			}
		}
		return e.taskModifier.RemoveTagFromTask(ctx, task.ID, task.WorkspaceID, tagID, tagName)

	case domainAutomation.ActionSetDueDate:
		preset := extractStringParam(action.Params, "preset")
		offsetDays := int(toFloat(action.Params["offset_days"]))
		var dueDate *time.Time
		now := time.Now().UTC()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		switch preset {
		case "today":
			dueDate = &today
		case "tomorrow":
			tmrw := today.AddDate(0, 0, 1)
			dueDate = &tmrw
		case "clear":
			dueDate = nil
		default:
			if offsetDays > 0 {
				dt := today.AddDate(0, 0, offsetDays)
				dueDate = &dt
			}
		}
		return e.taskModifier.SetTaskDueDate(ctx, task.ID, task.WorkspaceID, dueDate)

	case domainAutomation.ActionArchiveTask:
		return e.taskModifier.ArchiveTask(ctx, task.ID, task.WorkspaceID)
	}

	return nil
}

func extractStringParam(params map[string]interface{}, keys ...string) string {
	if params == nil {
		return ""
	}
	for _, k := range keys {
		if val, ok := params[k]; ok && val != nil {
			s := strings.TrimSpace(fmt.Sprint(val))
			if s != "" {
				return s
			}
		}
	}
	return ""
}
