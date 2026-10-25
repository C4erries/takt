package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/domain/events"
)

// Evaluator evaluates conditions and executes actions for Low-Code rules.
type Evaluator struct {
	taskModifier  TaskModifier
	projectReader ProjectReader
	taskReader    TaskReader
}

// NewEvaluator creates a new Evaluator.
func NewEvaluator(taskModifier TaskModifier, projectReader ProjectReader, taskReader ...TaskReader) *Evaluator {
	var tr TaskReader
	if len(taskReader) > 0 {
		tr = taskReader[0]
	} else if r, ok := taskModifier.(TaskReader); ok {
		tr = r
	}
	return &Evaluator{
		taskModifier:  taskModifier,
		projectReader: projectReader,
		taskReader:    tr,
	}
}

// SetTaskReader assigns the task reader used for subtasks and activity history.
func (e *Evaluator) SetTaskReader(tr TaskReader) {
	e.taskReader = tr
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
	case domainAutomation.ConditionFieldSubtasksAllCompleted:
		allDone := task.SubtasksCount > 0 && task.SubtasksCount == task.SubtasksDoneCount
		if c.Operator == "" {
			return allDone
		}
		return evaluateBoolCondition(allDone, c.Operator, c.Value)
	case domainAutomation.ConditionFieldSubtasksCount:
		return evaluateNumericCondition(float64(task.SubtasksCount), c.Operator, c.Value)
	default:
		return false
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

	case domainAutomation.ActionReturnToPreviousColumn:
		return e.executeReturnToPreviousColumn(ctx, task)

	case domainAutomation.ActionCompleteAllSubtasks:
		if e.taskReader == nil {
			return nil
		}
		subtasks, err := e.taskReader.ListSubtasks(ctx, task.ID)
		if err != nil {
			return err
		}
		for _, st := range subtasks {
			if !st.IsCompleted {
				_ = e.taskModifier.SetTaskCompleted(ctx, st.ID, task.WorkspaceID, true)
			}
		}
		return nil

	case domainAutomation.ActionReopenAllSubtasks:
		if e.taskReader == nil {
			return nil
		}
		subtasks, err := e.taskReader.ListSubtasks(ctx, task.ID)
		if err != nil {
			return err
		}
		for _, st := range subtasks {
			if st.IsCompleted {
				_ = e.taskModifier.SetTaskCompleted(ctx, st.ID, task.WorkspaceID, false)
			}
		}
		return nil

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

func (e *Evaluator) executeReturnToPreviousColumn(ctx context.Context, task *TaskInfo) error {
	var targetColID *uuid.UUID

	if e.taskReader != nil {
		activities, err := e.taskReader.ListTaskActivities(ctx, task.ID)
		if err == nil && len(activities) > 0 {
			var latestMoved *TaskActivityInfo
			for _, act := range activities {
				if act.ActivityType == "column_moved" {
					if latestMoved == nil || act.CreatedAt.After(latestMoved.CreatedAt) {
						latestMoved = act
					}
				}
			}

			if latestMoved != nil && latestMoved.Payload != nil {
				if fromColRaw, ok := latestMoved.Payload["from_column_id"]; ok && fromColRaw != nil {
					fromColStr := fmt.Sprint(fromColRaw)
					if parsed, err := uuid.Parse(fromColStr); err == nil && parsed != uuid.Nil {
						if e.projectReader != nil {
							exists, err := e.projectReader.ColumnExists(ctx, parsed)
							if err == nil && exists {
								targetColID = &parsed
							}
						}
					}
				}
			}
		}
	}

	if targetColID != nil {
		return e.taskModifier.MoveTaskToColumn(ctx, task.ID, task.WorkspaceID, targetColID, task.ProjectID, 1000.0)
	}

	// Fallback 1: Inbox column of project
	if task.ProjectID != nil && e.projectReader != nil {
		inboxCol, err := e.projectReader.GetInboxColumn(ctx, *task.ProjectID, task.WorkspaceID)
		if err == nil && inboxCol != nil {
			return e.taskModifier.MoveTaskToColumn(ctx, task.ID, task.WorkspaceID, inboxCol, task.ProjectID, 1000.0)
		}
	}

	// Fallback 2: Inbox of workspace (columnID = nil)
	return e.taskModifier.MoveTaskToColumn(ctx, task.ID, task.WorkspaceID, nil, task.ProjectID, 1000.0)
}
