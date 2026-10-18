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

// EvaluateConditions evaluates conditions using the specified logical operator ("and" or "or").
// If logicalOp is LogicOr ("or"), it returns true if any condition matches.
// If logicalOp is LogicAnd ("and") or empty, it returns true only if all conditions match.
// Empty conditions always returns true regardless of logicalOp.
func (e *Evaluator) EvaluateConditions(logicalOp string, conditions []domainAutomation.Condition, task *TaskInfo) bool {
	if len(conditions) == 0 {
		return true
	}
	if logicalOp == domainAutomation.LogicOr {
		for _, c := range conditions {
			if e.evaluateCondition(c, task) {
				return true
			}
		}
		return false
	}
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
	case domainAutomation.ConditionFieldDueDateStatus:
		return evaluateDueDateStatus(task.DueDate, c.Operator, c.Value)
	case domainAutomation.ConditionFieldTimeInColumnHours:
		return evaluateNumericCondition(task.TimeInColumnHours, c.Operator, c.Value)
	case domainAutomation.ConditionFieldTags:
		return evaluateTagsCondition(task.Tags, c.Operator, c.Value)
	default:
		return false
	}
}

func evaluateDueDateStatus(dueDate *time.Time, op string, expectedVal interface{}) bool {
	if dueDate == nil {
		return evaluateStringCondition("no_due_date", op, expectedVal)
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	due := dueDate.UTC().Truncate(24 * time.Hour)
	status := "on_track"
	if due.Before(today) {
		status = "overdue"
	} else if !due.After(today.AddDate(0, 0, 2)) {
		status = "due_soon"
	}
	return evaluateStringCondition(status, op, expectedVal)
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
	case domainAutomation.OperatorIn, domainAutomation.OperatorNotIn:
		in := false
		for _, item := range toStringSlice(expectedVal) {
			if strings.EqualFold(actual, item) {
				in = true
				break
			}
		}
		if op == domainAutomation.OperatorIn {
			return in
		}
		return !in
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

	expStr := strings.ToLower(strings.TrimSpace(fmt.Sprint(expectedVal)))
	today := time.Now().UTC().Truncate(24 * time.Hour)
	var expDate time.Time
	var err error

	if expStr == "today" {
		expDate = today
	} else if expStr == "tomorrow" {
		expDate = today.AddDate(0, 0, 1)
	} else {
		parsedExp, parseErr := time.Parse("2006-01-02", expStr)
		if parseErr != nil {
			parsedExp, parseErr = time.Parse(time.RFC3339, expStr)
		}
		err = parseErr
		expDate = parsedExp.UTC().Truncate(24 * time.Hour)
	}

	actDate := actual.UTC().Truncate(24 * time.Hour)

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

	expItems := toStringSlice(expectedVal)
	hasIntersection := false
	for _, t := range tags {
		for _, exp := range expItems {
			if strings.EqualFold(strings.TrimSpace(t), exp) {
				hasIntersection = true
				break
			}
		}
		if hasIntersection {
			break
		}
	}

	switch op {
	case domainAutomation.OperatorEquals, domainAutomation.OperatorContains, domainAutomation.OperatorIn:
		return hasIntersection
	case domainAutomation.OperatorNotEquals, domainAutomation.OperatorNotContains, domainAutomation.OperatorNotIn:
		return !hasIntersection
	default:
		return false
	}
}

func toFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	return 0
}

// ExecuteRule checks Cycle Guard, evaluates conditions, and runs actions sequentially.
func (e *Evaluator) ExecuteRule(ctx context.Context, rule *domainAutomation.RuleDefinition, task *TaskInfo) error {
	if rule == nil || task == nil {
		return nil
	}

	if !events.CanExecuteAutomation(ctx) {
		return nil
	}

	if !e.EvaluateConditions(rule.LogicalOperator, rule.Conditions, task) {
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
			if targetColUUID, err := uuid.Parse(targetColIDStr); err == nil && targetColUUID != uuid.Nil {
				return e.taskModifier.MoveTaskToColumn(ctx, task.ID, task.WorkspaceID, &targetColUUID, task.ProjectID, 1000.0)
			}
		}

	case domainAutomation.ActionSetPriority:
		if priority := extractStringParam(action.Params, "priority"); priority != "" {
			return e.taskModifier.SetTaskPriority(ctx, task.ID, task.WorkspaceID, priority)
		}

	case domainAutomation.ActionAddTag:
		tid, name := extractTagParams(action.Params)
		return e.taskModifier.AddTagToTask(ctx, task.ID, task.WorkspaceID, tid, name)

	case domainAutomation.ActionRemoveTag:
		tid, name := extractTagParams(action.Params)
		return e.taskModifier.RemoveTagFromTask(ctx, task.ID, task.WorkspaceID, tid, name)

	case domainAutomation.ActionSetDueDate:
		preset := extractStringParam(action.Params, "preset")
		offsetDays := int(toFloat(action.Params["offset_days"]))
		var dueDate *time.Time
		today := time.Now().UTC().Truncate(24 * time.Hour)
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

	case domainAutomation.ActionMarkCompleted:
		return e.taskModifier.SetTaskCompleted(ctx, task.ID, task.WorkspaceID, true)

	case domainAutomation.ActionMarkUncompleted:
		return e.taskModifier.SetTaskCompleted(ctx, task.ID, task.WorkspaceID, false)

	case domainAutomation.ActionClearDueDate:
		return e.taskModifier.SetTaskDueDate(ctx, task.ID, task.WorkspaceID, nil)
	}

	return nil
}

func extractTagParams(params map[string]interface{}) (*uuid.UUID, string) {
	tagIDStr := extractStringParam(params, "tag_id", "tagId")
	tagName := extractStringParam(params, "name", "tag_name")
	if tagIDStr != "" {
		if parsed, err := uuid.Parse(tagIDStr); err == nil && parsed != uuid.Nil {
			return &parsed, tagName
		}
	}
	return nil, tagName
}

func toStringSlice(val interface{}) []string {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case []string:
		return v
	case []interface{}:
		res := make([]string, 0, len(v))
		for _, item := range v {
			res = append(res, strings.TrimSpace(fmt.Sprint(item)))
		}
		return res
	case string:
		parts := strings.Split(v, ",")
		res := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				res = append(res, s)
			}
		}
		return res
	default:
		return []string{strings.TrimSpace(fmt.Sprint(val))}
	}
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
