package automation

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/domain/events"
)

// EventListener listens to domain events and executes enabled automation rules.
type EventListener struct {
	repo          AutomationRepository
	taskModifier  TaskModifier
	projectReader ProjectReader
	evaluator     *domainAutomation.RuleEvaluator
	dslEvaluator  *Evaluator
}

// NewEventListener constructs an automation EventListener.
func NewEventListener(
	repo AutomationRepository,
	taskModifier TaskModifier,
	projectReader ProjectReader,
	evaluator *domainAutomation.RuleEvaluator,
) *EventListener {
	if evaluator == nil {
		evaluator = domainAutomation.NewRuleEvaluator()
	}
	return &EventListener{
		repo:          repo,
		taskModifier:  taskModifier,
		projectReader: projectReader,
		evaluator:     evaluator,
		dslEvaluator:  NewEvaluator(taskModifier, projectReader),
	}
}

// Register subscribes the listener to domain events on the provided EventBus.
func (l *EventListener) Register(bus events.EventBus) error {
	if bus == nil {
		return nil
	}

	if err := bus.Subscribe(events.EventTaskCompleted, l.HandleTaskCompleted); err != nil {
		return err
	}
	if err := bus.Subscribe(events.EventTaskUpdated, l.HandleTaskUpdated); err != nil {
		return err
	}
	if err := bus.Subscribe(events.EventTaskCreated, l.HandleTaskCreated); err != nil {
		return err
	}
	if err := bus.Subscribe(events.EventTaskMoved, l.HandleTaskMoved); err != nil {
		return err
	}

	return nil
}

// HandleTaskCompleted executes rules triggered by task completion.
func (l *EventListener) HandleTaskCompleted(ctx context.Context, event events.Event) error {
	if event == nil || !events.CanExecuteAutomation(ctx) {
		return nil
	}

	taskEvt, ok := event.(events.TaskCompletedEvent)
	if !ok {
		return nil
	}

	autoCtx := events.WithAutomationTrigger(ctx)
	taskCtx := domainAutomation.TaskContext{
		IsCompleted: true,
		ProjectID:   taskEvt.ProjectID,
		ColumnID:    taskEvt.ColumnID,
	}
	taskInfo := &TaskInfo{
		ID:          taskEvt.TaskID,
		WorkspaceID: event.WorkspaceID(),
		ProjectID:   taskEvt.ProjectID,
		ColumnID:    taskEvt.ColumnID,
		IsCompleted: true,
	}

	// 1. Workspace-level completion rules
	rules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskCompleted)
	if err == nil {
		for _, rule := range rules {
			l.evaluateAndExecute(autoCtx, rule, taskCtx, taskInfo)
		}
	}

	// 2. Column-level completion rules
	if taskEvt.ColumnID != nil {
		colRules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskCompletedInColumn)
		if err == nil {
			for _, rule := range colRules {
				if rule.ColumnID != nil && *rule.ColumnID != *taskEvt.ColumnID {
					continue
				}
				l.evaluateAndExecute(autoCtx, rule, taskCtx, taskInfo)
			}
		}
	}

	return nil
}

// HandleTaskUpdated executes rules triggered by task modifications.
func (l *EventListener) HandleTaskUpdated(ctx context.Context, event events.Event) error {
	if event == nil || !events.CanExecuteAutomation(ctx) {
		return nil
	}

	taskEvt, ok := event.(events.TaskUpdatedEvent)
	if !ok {
		return nil
	}

	autoCtx := events.WithAutomationTrigger(ctx)
	taskCtx := domainAutomation.TaskContext{
		IsCompleted:       taskEvt.IsCompleted,
		HasDueDate:        taskEvt.DueDate != nil,
		DueDate:           taskEvt.DueDate,
		ProjectID:         taskEvt.ProjectID,
		ColumnID:          taskEvt.ColumnID,
		PreviousProjectID: taskEvt.PreviousProjectID,
		PreviousColumnID:  taskEvt.PreviousColumnID,
	}
	taskInfo := &TaskInfo{
		ID:                taskEvt.TaskID,
		WorkspaceID:       event.WorkspaceID(),
		ProjectID:         taskEvt.ProjectID,
		ColumnID:          taskEvt.ColumnID,
		PreviousColumnID:  taskEvt.PreviousColumnID,
		PreviousProjectID: taskEvt.PreviousProjectID,
		IsCompleted:       taskEvt.IsCompleted,
		DueDate:           taskEvt.DueDate,
	}

	rules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskUpdated)
	if err == nil {
		for _, rule := range rules {
			l.evaluateAndExecute(autoCtx, rule, taskCtx, taskInfo)
		}
	}

	// Check if task entered a new column
	if taskEvt.ColumnID != nil && (taskEvt.PreviousColumnID == nil || *taskEvt.PreviousColumnID != *taskEvt.ColumnID) {
		colRules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskEnteredColumn)
		if err == nil {
			for _, rule := range colRules {
				if rule.ColumnID != nil && *rule.ColumnID != *taskEvt.ColumnID {
					continue
				}
				l.evaluateAndExecute(autoCtx, rule, taskCtx, taskInfo)
			}
		}
	}

	// Check if task was completed in a column
	if taskEvt.IsCompleted && taskEvt.ColumnID != nil {
		colRules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskCompletedInColumn)
		if err == nil {
			for _, rule := range colRules {
				if rule.ColumnID != nil && *rule.ColumnID != *taskEvt.ColumnID {
					continue
				}
				l.evaluateAndExecute(autoCtx, rule, taskCtx, taskInfo)
			}
		}
	}

	return nil
}

// HandleTaskCreated checks rules for newly created tasks.
func (l *EventListener) HandleTaskCreated(ctx context.Context, event events.Event) error {
	if event == nil || !events.CanExecuteAutomation(ctx) {
		return nil
	}

	taskEvt, ok := event.(events.TaskCreatedEvent)
	if !ok {
		return nil
	}

	autoCtx := events.WithAutomationTrigger(ctx)
	taskCtx := domainAutomation.TaskContext{
		HasDueDate: taskEvt.DueDate != nil,
		DueDate:    taskEvt.DueDate,
		ProjectID:  taskEvt.ProjectID,
		ColumnID:   taskEvt.ColumnID,
	}
	taskInfo := &TaskInfo{
		ID:          taskEvt.TaskID,
		WorkspaceID: event.WorkspaceID(),
		ProjectID:   taskEvt.ProjectID,
		ColumnID:    taskEvt.ColumnID,
		DueDate:     taskEvt.DueDate,
	}

	rules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskUpdated)
	if err == nil {
		for _, rule := range rules {
			l.evaluateAndExecute(autoCtx, rule, taskCtx, taskInfo)
		}
	}

	// Task created directly in a column
	if taskEvt.ColumnID != nil {
		colRules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskEnteredColumn)
		if err == nil {
			for _, rule := range colRules {
				if rule.ColumnID != nil && *rule.ColumnID != *taskEvt.ColumnID {
					continue
				}
				l.evaluateAndExecute(autoCtx, rule, taskCtx, taskInfo)
			}
		}
	}

	return nil
}

// HandleTaskMoved executes rules when a task is moved to another column.
func (l *EventListener) HandleTaskMoved(ctx context.Context, event events.Event) error {
	if event == nil || !events.CanExecuteAutomation(ctx) {
		return nil
	}

	taskEvt, ok := event.(events.TaskMovedEvent)
	if !ok || taskEvt.ColumnID == nil {
		return nil
	}

	autoCtx := events.WithAutomationTrigger(ctx)
	rules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskEnteredColumn)
	if err != nil || len(rules) == 0 {
		return err
	}

	taskCtx := domainAutomation.TaskContext{
		ProjectID:        taskEvt.ProjectID,
		ColumnID:         taskEvt.ColumnID,
		PreviousColumnID: taskEvt.PreviousColumnID,
	}
	taskInfo := &TaskInfo{
		ID:               taskEvt.TaskID,
		WorkspaceID:      event.WorkspaceID(),
		ProjectID:        taskEvt.ProjectID,
		ColumnID:         taskEvt.ColumnID,
		PreviousColumnID: taskEvt.PreviousColumnID,
	}

	for _, rule := range rules {
		if rule.ColumnID != nil && *rule.ColumnID != *taskEvt.ColumnID {
			continue
		}
		l.evaluateAndExecute(autoCtx, rule, taskCtx, taskInfo)
	}

	return nil
}

func (l *EventListener) evaluateAndExecute(
	ctx context.Context,
	rule *domainAutomation.Automation,
	taskCtx domainAutomation.TaskContext,
	taskInfo *TaskInfo,
) {
	if rule.RuleDefinition != nil {
		_ = l.dslEvaluator.ExecuteRule(ctx, rule.RuleDefinition, taskInfo)
		return
	}

	match, err := l.evaluator.Evaluate(rule.ConditionsJSON, taskCtx)
	if err != nil || !match {
		return
	}
	l.executeAction(ctx, rule.ActionsJSON, taskInfo.ID, taskInfo.WorkspaceID, taskInfo.ProjectID)
}

type actionPayload struct {
	Action         string  `json:"action"`
	Priority       string  `json:"priority"`
	TargetColumnID *string `json:"target_column_id"`
	TargetColumnId *string `json:"targetColumnId"`
}

func (l *EventListener) executeAction(ctx context.Context, actionsJSON json.RawMessage, taskID, workspaceID uuid.UUID, projectID *uuid.UUID) {
	if l.taskModifier == nil {
		return
	}

	var payload actionPayload
	if err := json.Unmarshal(actionsJSON, &payload); err != nil {
		return
	}

	switch payload.Action {
	case "move_to_done":
		if projectID != nil && l.projectReader != nil {
			doneCol, err := l.projectReader.GetDoneColumn(ctx, *projectID, workspaceID)
			if err == nil && doneCol != nil {
				_ = l.taskModifier.MoveTaskToColumn(ctx, taskID, workspaceID, doneCol, projectID, 1000.0)
			}
		}
	case "set_focus":
		focusOrder := 1
		_ = l.taskModifier.SetTaskFocus(ctx, taskID, workspaceID, &focusOrder)
	case "move_to_first_column":
		if projectID != nil && l.projectReader != nil {
			todoCol, err := l.projectReader.GetFirstTodoColumn(ctx, *projectID, workspaceID)
			if err == nil && todoCol != nil {
				_ = l.taskModifier.MoveTaskToColumn(ctx, taskID, workspaceID, todoCol, projectID, 1000.0)
			}
		}
	case "set_priority":
		if payload.Priority != "" {
			_ = l.taskModifier.SetTaskPriority(ctx, taskID, workspaceID, payload.Priority)
		}
	case "move_to_column":
		targetStr := ""
		if payload.TargetColumnID != nil && *payload.TargetColumnID != "" {
			targetStr = *payload.TargetColumnID
		} else if payload.TargetColumnId != nil && *payload.TargetColumnId != "" {
			targetStr = *payload.TargetColumnId
		}
		if targetStr != "" {
			targetColUUID, err := uuid.Parse(targetStr)
			if err == nil && targetColUUID != uuid.Nil {
				_ = l.taskModifier.MoveTaskToColumn(ctx, taskID, workspaceID, &targetColUUID, projectID, 1000.0)
			}
		}
	}
}
