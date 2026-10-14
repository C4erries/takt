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

	return nil
}

// HandleTaskCompleted executes rules triggered by task completion (e.g., Done & Archive).
func (l *EventListener) HandleTaskCompleted(ctx context.Context, event events.Event) error {
	if event == nil || event.IsFromAutomation() || events.IsAutomationTriggered(ctx) {
		return nil
	}

	taskEvt, ok := event.(events.TaskCompletedEvent)
	if !ok {
		return nil
	}

	autoCtx := events.WithAutomationTrigger(ctx)
	rules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskCompleted)
	if err != nil || len(rules) == 0 {
		return err
	}

	taskCtx := domainAutomation.TaskContext{
		IsCompleted: true,
		ProjectID:   taskEvt.ProjectID,
		ColumnID:    taskEvt.ColumnID,
	}

	for _, rule := range rules {
		match, err := l.evaluator.Evaluate(rule.ConditionsJSON, taskCtx)
		if err != nil || !match {
			continue
		}

		l.executeAction(autoCtx, rule.ActionsJSON, taskEvt.TaskID, event.WorkspaceID(), taskEvt.ProjectID)
	}

	return nil
}

// HandleTaskUpdated executes rules triggered by task modifications (Auto-Focus, Inbox Zero).
func (l *EventListener) HandleTaskUpdated(ctx context.Context, event events.Event) error {
	if event == nil || event.IsFromAutomation() || events.IsAutomationTriggered(ctx) {
		return nil
	}

	taskEvt, ok := event.(events.TaskUpdatedEvent)
	if !ok {
		return nil
	}

	autoCtx := events.WithAutomationTrigger(ctx)
	rules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskUpdated)
	if err != nil || len(rules) == 0 {
		return err
	}

	taskCtx := domainAutomation.TaskContext{
		IsCompleted:       taskEvt.IsCompleted,
		HasDueDate:        taskEvt.DueDate != nil,
		DueDate:           taskEvt.DueDate,
		ProjectID:         taskEvt.ProjectID,
		ColumnID:          taskEvt.ColumnID,
		PreviousProjectID: taskEvt.PreviousProjectID,
		PreviousColumnID:  taskEvt.PreviousColumnID,
	}

	for _, rule := range rules {
		match, err := l.evaluator.Evaluate(rule.ConditionsJSON, taskCtx)
		if err != nil || !match {
			continue
		}

		l.executeAction(autoCtx, rule.ActionsJSON, taskEvt.TaskID, event.WorkspaceID(), taskEvt.ProjectID)
	}

	return nil
}

// HandleTaskCreated checks rules that apply to newly created tasks (e.g. Auto-Focus if due date set).
func (l *EventListener) HandleTaskCreated(ctx context.Context, event events.Event) error {
	if event == nil || event.IsFromAutomation() || events.IsAutomationTriggered(ctx) {
		return nil
	}

	taskEvt, ok := event.(events.TaskCreatedEvent)
	if !ok {
		return nil
	}

	autoCtx := events.WithAutomationTrigger(ctx)
	rules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskUpdated)
	if err != nil || len(rules) == 0 {
		return err
	}

	taskCtx := domainAutomation.TaskContext{
		HasDueDate: taskEvt.DueDate != nil,
		DueDate:    taskEvt.DueDate,
		ProjectID:  taskEvt.ProjectID,
		ColumnID:   taskEvt.ColumnID,
	}

	for _, rule := range rules {
		match, err := l.evaluator.Evaluate(rule.ConditionsJSON, taskCtx)
		if err != nil || !match {
			continue
		}

		l.executeAction(autoCtx, rule.ActionsJSON, taskEvt.TaskID, event.WorkspaceID(), taskEvt.ProjectID)
	}

	return nil
}

type actionPayload struct {
	Action string `json:"action"`
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
	}
}
