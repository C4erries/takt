package automation

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	domainAutomation "takt/internal/domain/automation"
	"takt/internal/domain/events"
)

// Worker runs background checks for SLA timers and scheduled sweeps.
type Worker struct {
	repo           AutomationRepository
	taskReader     TaskReader
	evaluator      *Evaluator
	projectReader  ProjectReader
	batchExecutor  BatchTaskExecutor
	eventPublisher EventPublisher
	interval       time.Duration

	stopCh  chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	running bool
}

// NewWorker constructs an automation Worker.
func NewWorker(
	repo AutomationRepository,
	taskReader TaskReader,
	evaluator *Evaluator,
	interval time.Duration,
	projectReader ...ProjectReader,
) *Worker {
	if interval <= 0 {
		interval = 1 * time.Minute
	}
	var pr ProjectReader
	if len(projectReader) > 0 {
		pr = projectReader[0]
	} else if evaluator != nil {
		pr = evaluator.projectReader
	}
	return &Worker{
		repo:          repo,
		taskReader:    taskReader,
		evaluator:     evaluator,
		projectReader: pr,
		interval:      interval,
		stopCh:        make(chan struct{}),
	}
}

// SetProjectReader sets the project reader for inspecting board structure.
func (w *Worker) SetProjectReader(reader ProjectReader) {
	w.projectReader = reader
}

// SetBatchExecutor injects a BatchTaskExecutor for fast-path batch executions (ADR-0008).
func (w *Worker) SetBatchExecutor(executor BatchTaskExecutor) {
	w.batchExecutor = executor
}

// SetEventPublisher sets event bus for publishing batch sweep completion events.
func (w *Worker) SetEventPublisher(pub EventPublisher) {
	w.eventPublisher = pub
}

// CanUseBatchPath returns true if the rule can be executed via a single batch SQL query (ADR-0008).
func CanUseBatchPath(rule *domainAutomation.RuleDefinition) bool {
	if rule == nil || len(rule.Actions) != 1 {
		return false
	}
	if rule.Actions[0].Type != domainAutomation.ActionArchiveTask {
		return false
	}
	for _, c := range rule.Conditions {
		if c.Field != domainAutomation.ConditionFieldIsCompleted {
			return false
		}
		if strVal, ok := c.Value.(string); ok && strVal != "true" {
			return false
		}
		if boolVal, ok := c.Value.(bool); ok && !boolVal {
			return false
		}
	}
	return true
}

// Start begins periodic timer and schedule checks.
func (w *Worker) Start(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.stopCh = make(chan struct{})
	w.mu.Unlock()

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-w.stopCh:
				return
			case <-ticker.C:
				_ = w.Tick(ctx)
			}
		}
	}()
}

// Stop terminates the background ticker gracefully.
func (w *Worker) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	close(w.stopCh)
	w.mu.Unlock()

	w.wg.Wait()
}

// Tick executes a single evaluation pass for all enabled timer automations.
func (w *Worker) Tick(ctx context.Context) error {
	if w.repo == nil || w.taskReader == nil || w.evaluator == nil {
		return nil
	}

	rules, err := w.repo.ListAllEnabledTimerAutomations(ctx)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		if rule.RuleDefinition == nil {
			continue
		}

		switch rule.RuleDefinition.Trigger.Type {
		case domainAutomation.TriggerColumnSLATimer:
			w.processSLATimerRule(ctx, rule)
		case domainAutomation.TriggerScheduledSweep:
			w.processScheduledSweepRule(ctx, rule)
		case domainAutomation.TriggerTaskTimer:
			w.processTaskTimerRule(ctx, rule)
		}
	}

	return nil
}

func (w *Worker) processTaskTimerRule(ctx context.Context, rule *domainAutomation.Automation) {
	if rule.TaskID == nil || rule.RuleDefinition == nil {
		return
	}
	if rule.RuleDefinition.ID == nil {
		rule.RuleDefinition.ID = &rule.ID
	}
	if rule.RuleDefinition.Name == "" {
		rule.RuleDefinition.Name = rule.Name
	}

	task, err := w.taskReader.GetTaskByID(ctx, *rule.TaskID, rule.WorkspaceID)
	if err != nil || task == nil || task.IsCompleted {
		return
	}

	delayMinutes := toFloat(rule.RuleDefinition.Trigger.Params["delay_minutes"])
	if delayMinutes <= 0 {
		delayMinutes = toFloat(rule.RuleDefinition.Trigger.Params["duration_minutes"])
	}
	var requiredDuration time.Duration
	if delayMinutes > 0 {
		requiredDuration = time.Duration(delayMinutes) * time.Minute
	} else {
		durationHours := toFloat(rule.RuleDefinition.Trigger.Params["duration_hours"])
		if durationHours <= 0 {
			durationHours = toFloat(rule.RuleDefinition.Trigger.Params["delay_hours"])
		}
		if durationHours <= 0 {
			durationHours = 24
		}
		requiredDuration = time.Duration(durationHours * float64(time.Hour))
	}

	refTime := rule.CreatedAt
	if refTime.IsZero() {
		refTime = task.UpdatedAt
		if refTime.IsZero() {
			refTime = task.CreatedAt
		}
	}

	now := time.Now().UTC()
	if now.Sub(refTime) >= requiredDuration {
		_ = w.evaluator.ExecuteRule(ctx, rule.RuleDefinition, task)
	}
}

func (w *Worker) processSLATimerRule(ctx context.Context, rule *domainAutomation.Automation) {
	if rule.RuleDefinition == nil {
		return
	}
	if rule.RuleDefinition.ID == nil {
		rule.RuleDefinition.ID = &rule.ID
	}
	if rule.RuleDefinition.Name == "" {
		rule.RuleDefinition.Name = rule.Name
	}
	colIDStr := extractStringParam(rule.RuleDefinition.Trigger.Params, "column_id", "columnId")
	if colIDStr == "" && rule.ColumnID != nil {
		colIDStr = rule.ColumnID.String()
	}
	if colIDStr == "" {
		return
	}
	colUUID, err := uuid.Parse(colIDStr)
	if err != nil {
		return
	}

	durationHours := toFloat(rule.RuleDefinition.Trigger.Params["duration_hours"])
	if durationHours <= 0 {
		durationHours = 24
	}

	// ADR-0008 Fast Path: execute single batch query if rule is simple archive sweep
	if w.batchExecutor != nil && CanUseBatchPath(rule.RuleDefinition) {
		affected, err := w.batchExecutor.ArchiveCompletedTasksBatchSQL(ctx, rule.WorkspaceID, colUUID, durationHours)
		if err == nil && affected > 0 && w.eventPublisher != nil {
			_ = w.eventPublisher.Publish(ctx, events.BatchSweepCompletedEvent{
				WorkspaceIDVal: rule.WorkspaceID,
				ColumnIDVal:    colUUID,
				ActionVal:      domainAutomation.ActionArchiveTask,
				AffectedCount:  affected,
				OccurredAtVal:  time.Now().UTC(),
			})
		}
		return
	}

	tasks, err := w.taskReader.ListTasksByColumn(ctx, rule.WorkspaceID, colUUID)
	if err != nil {
		return
	}

	now := time.Now().UTC()
	for _, task := range tasks {
		if task.IsCompleted {
			continue
		}
		refTime := task.UpdatedAt
		if refTime.IsZero() {
			refTime = task.CreatedAt
		}
		hoursInCol := now.Sub(refTime).Hours()
		task.TimeInColumnHours = hoursInCol

		if hoursInCol >= durationHours {
			_ = w.evaluator.ExecuteRule(ctx, rule.RuleDefinition, task)
		}
	}
}

func (w *Worker) processScheduledSweepRule(ctx context.Context, rule *domainAutomation.Automation) {
	colIDStr := extractStringParam(rule.RuleDefinition.Trigger.Params, "column_id", "columnId")
	if colIDStr == "" && rule.ColumnID != nil {
		colIDStr = rule.ColumnID.String()
	}
	if colIDStr != "" {
		colUUID, err := uuid.Parse(colIDStr)
		if err != nil {
			return
		}
		w.sweepColumn(ctx, rule, colUUID)
		return
	}

	if rule.BoardID != nil && w.projectReader != nil {
		columns, err := w.projectReader.ListColumnIDsByBoard(ctx, *rule.BoardID)
		if err != nil {
			return
		}
		for _, colUUID := range columns {
			w.sweepColumn(ctx, rule, colUUID)
		}
	}
}

func (w *Worker) sweepColumn(ctx context.Context, rule *domainAutomation.Automation, colUUID uuid.UUID) {
	if rule.RuleDefinition == nil {
		return
	}
	if rule.RuleDefinition.ID == nil {
		rule.RuleDefinition.ID = &rule.ID
	}
	if rule.RuleDefinition.Name == "" {
		rule.RuleDefinition.Name = rule.Name
	}

	// ADR-0008 Fast Path: execute single batch query for scheduled sweep
	if w.batchExecutor != nil && CanUseBatchPath(rule.RuleDefinition) {
		affected, err := w.batchExecutor.ArchiveCompletedTasksBatchSQL(ctx, rule.WorkspaceID, colUUID, 0)
		if err == nil && affected > 0 && w.eventPublisher != nil {
			_ = w.eventPublisher.Publish(ctx, events.BatchSweepCompletedEvent{
				WorkspaceIDVal: rule.WorkspaceID,
				ColumnIDVal:    colUUID,
				ActionVal:      domainAutomation.ActionArchiveTask,
				AffectedCount:  affected,
				OccurredAtVal:  time.Now().UTC(),
			})
		}
		return
	}

	tasks, err := w.taskReader.ListTasksByColumn(ctx, rule.WorkspaceID, colUUID)
	if err != nil {
		return
	}

	for _, task := range tasks {
		_ = w.evaluator.ExecuteRule(ctx, rule.RuleDefinition, task)
	}
}
