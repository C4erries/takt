package automation

import (
	"context"

	domainAutomation "takt/internal/domain/automation"
	"takt/internal/domain/events"
)

// EventListener listens to domain events and executes enabled automation rules.
type EventListener struct {
	repo          AutomationRepository
	taskModifier  TaskModifier
	projectReader ProjectReader
	taskReader    TaskReader
	evaluator     *domainAutomation.RuleEvaluator
	dslEvaluator  *Evaluator
}

// NewEventListener constructs an automation EventListener.
func NewEventListener(
	repo AutomationRepository,
	taskModifier TaskModifier,
	projectReader ProjectReader,
	evaluator *domainAutomation.RuleEvaluator,
	taskReader ...TaskReader,
) *EventListener {
	if evaluator == nil {
		evaluator = domainAutomation.NewRuleEvaluator()
	}
	var tr TaskReader
	if len(taskReader) > 0 {
		tr = taskReader[0]
	} else if r, ok := taskModifier.(TaskReader); ok {
		tr = r
	}
	return &EventListener{
		repo:          repo,
		taskModifier:  taskModifier,
		projectReader: projectReader,
		taskReader:    tr,
		evaluator:     evaluator,
		dslEvaluator:  NewEvaluator(taskModifier, projectReader, tr),
	}
}

// SetTaskReader assigns the task reader used for subtasks and task inspection.
func (l *EventListener) SetTaskReader(tr TaskReader) {
	l.taskReader = tr
	if l.dslEvaluator != nil {
		l.dslEvaluator.SetTaskReader(tr)
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

	// 3. Subtask auto-completion of parent task and subtask triggers
	parentID := taskEvt.ParentID
	if parentID == nil && l.taskReader != nil {
		if ti, err := l.taskReader.GetTaskByID(autoCtx, taskEvt.TaskID, event.WorkspaceID()); err == nil && ti != nil {
			parentID = ti.ParentID
		}
	}
	if parentID != nil {
		l.handleParentAutoComplete(autoCtx, event.WorkspaceID(), *parentID)

		if l.taskReader != nil {
			anyRules, errAny := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerAnySubtaskCompleted)
			allRules, errAll := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerAllSubtasksCompleted)

			if (len(anyRules) > 0 || len(allRules) > 0) && errAny == nil && errAll == nil {
				parentTask, err := l.taskReader.GetTaskByID(autoCtx, *parentID, event.WorkspaceID())
				if err == nil && parentTask != nil {
					parentCtx := domainAutomation.TaskContext{
						IsCompleted: parentTask.IsCompleted,
						HasDueDate:  parentTask.DueDate != nil,
						DueDate:     parentTask.DueDate,
						ProjectID:   parentTask.ProjectID,
						ColumnID:    parentTask.ColumnID,
					}

					for _, rule := range anyRules {
						l.evaluateAndExecute(autoCtx, rule, parentCtx, parentTask)
					}

					if len(allRules) > 0 {
						subtasks, err := l.taskReader.ListSubtasks(autoCtx, *parentID)
						if err == nil && len(subtasks) > 0 {
							allCompleted := true
							for _, st := range subtasks {
								if st.ID == taskEvt.TaskID {
									continue
								}
								if !st.IsCompleted {
									allCompleted = false
									break
								}
							}
							if allCompleted {
								for _, rule := range allRules {
									l.evaluateAndExecute(autoCtx, rule, parentCtx, parentTask)
								}
							}
						}
					}
				}
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

	// Check if task completion status changed to uncompleted
	if !taskEvt.IsCompleted {
		uncompletedRules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerTaskUncompleted)
		if err == nil {
			for _, rule := range uncompletedRules {
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

	// Trigger rules for parent task when a subtask is created
	parentID := taskEvt.ParentID
	if parentID == nil && l.taskReader != nil {
		if ti, err := l.taskReader.GetTaskByID(autoCtx, taskEvt.TaskID, event.WorkspaceID()); err == nil && ti != nil {
			parentID = ti.ParentID
		}
	}
	if parentID != nil && l.taskReader != nil {
		subCreatedRules, err := l.repo.ListEnabledAutomationsByTrigger(autoCtx, event.WorkspaceID(), domainAutomation.TriggerSubtaskCreated)
		if err == nil && len(subCreatedRules) > 0 {
			parentTask, err := l.taskReader.GetTaskByID(autoCtx, *parentID, event.WorkspaceID())
			if err == nil && parentTask != nil {
				parentCtx := domainAutomation.TaskContext{
					IsCompleted: parentTask.IsCompleted,
					HasDueDate:  parentTask.DueDate != nil,
					DueDate:     parentTask.DueDate,
					ProjectID:   parentTask.ProjectID,
					ColumnID:    parentTask.ColumnID,
				}
				for _, rule := range subCreatedRules {
					l.evaluateAndExecute(autoCtx, rule, parentCtx, parentTask)
				}
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
	if rule.BoardID != nil && (taskInfo.ProjectID == nil || *rule.BoardID != *taskInfo.ProjectID) {
		return
	}

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

