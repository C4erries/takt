package automation

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	domainAutomation "takt/internal/domain/automation"
)

// Worker runs background checks for SLA timers and scheduled sweeps.
type Worker struct {
	repo       AutomationRepository
	taskReader TaskReader
	evaluator  *Evaluator
	interval   time.Duration

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
) *Worker {
	if interval <= 0 {
		interval = 1 * time.Minute
	}
	return &Worker{
		repo:       repo,
		taskReader: taskReader,
		evaluator:  evaluator,
		interval:   interval,
		stopCh:     make(chan struct{}),
	}
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
		}
	}

	return nil
}

func (w *Worker) processSLATimerRule(ctx context.Context, rule *domainAutomation.Automation) {
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
	if colIDStr == "" {
		return
	}
	colUUID, err := uuid.Parse(colIDStr)
	if err != nil {
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
