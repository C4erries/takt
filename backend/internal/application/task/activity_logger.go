package task

import (
	"context"

	"takt/internal/domain/events"
	domainTask "takt/internal/domain/task"

	"github.com/google/uuid"
)

// ActivityLogger listens to domain events and creates immutable task activity records.
type ActivityLogger struct {
	taskRepo TaskRepository
}

// NewActivityLogger creates a new ActivityLogger.
func NewActivityLogger(taskRepo TaskRepository) *ActivityLogger {
	return &ActivityLogger{taskRepo: taskRepo}
}

// Register subscribes the listener to domain events on the event bus.
func (l *ActivityLogger) Register(bus events.EventBus) error {
	if bus == nil {
		return nil
	}
	if err := bus.Subscribe(events.EventTaskMoved, l.HandleTaskMoved); err != nil {
		return err
	}
	if err := bus.Subscribe(events.EventTaskCompleted, l.HandleTaskCompleted); err != nil {
		return err
	}
	if err := bus.Subscribe(events.EventTaskUpdated, l.HandleTaskUpdated); err != nil {
		return err
	}
	return nil
}

// HandleTaskMoved logs column movements.
func (l *ActivityLogger) HandleTaskMoved(ctx context.Context, event events.Event) error {
	e, ok := event.(events.TaskMovedEvent)
	if !ok {
		return nil
	}

	payload := make(map[string]interface{})
	if e.PreviousColumnID != nil {
		payload["from_column_id"] = e.PreviousColumnID.String()
	}
	if e.ColumnID != nil {
		payload["to_column_id"] = e.ColumnID.String()
	}
	if e.FromColumnName != "" {
		payload["from_column"] = e.FromColumnName
	}
	if e.ToColumnName != "" {
		payload["to_column"] = e.ToColumnName
	}
	if e.FromAutomation {
		if e.AutomationID != nil {
			payload["automation_id"] = e.AutomationID.String()
		}
		if e.AutomationName != "" {
			payload["automation_name"] = e.AutomationName
		}
	}

	actorID := e.ActorID
	if actorID == nil {
		actorID = getActorIDFromContext(ctx)
	}
	activity, err := domainTask.NewTaskActivity(e.TaskID, actorID, domainTask.ActivityColumnMoved, payload)
	if err != nil {
		return err
	}

	return l.taskRepo.CreateActivity(ctx, activity)
}

// HandleTaskCompleted logs task completion.
func (l *ActivityLogger) HandleTaskCompleted(ctx context.Context, event events.Event) error {
	e, ok := event.(events.TaskCompletedEvent)
	if !ok {
		return nil
	}

	actorID := e.ActorID
	if actorID == nil {
		actorID = getActorIDFromContext(ctx)
	}
	activity, err := domainTask.NewTaskActivity(e.TaskID, actorID, domainTask.ActivityStatusChanged, map[string]interface{}{
		"is_completed": true,
	})
	if err != nil {
		return err
	}

	return l.taskRepo.CreateActivity(ctx, activity)
}

// HandleTaskUpdated logs column moves and status changes emitted via TaskUpdatedEvent.
func (l *ActivityLogger) HandleTaskUpdated(ctx context.Context, event events.Event) error {
	e, ok := event.(events.TaskUpdatedEvent)
	if !ok {
		return nil
	}

	actorID := e.ActorID
	if actorID == nil {
		actorID = getActorIDFromContext(ctx)
	}

	if !e.IsCompleted {
		act, err := domainTask.NewTaskActivity(e.TaskID, actorID, domainTask.ActivityStatusChanged, map[string]interface{}{
			"is_completed": false,
		})
		if err == nil {
			_ = l.taskRepo.CreateActivity(ctx, act)
		}
	}

	if e.PreviousColumnID != nil && e.ColumnID != nil && *e.PreviousColumnID != *e.ColumnID {
		payload := map[string]interface{}{
			"from_column_id": e.PreviousColumnID.String(),
			"to_column_id":   e.ColumnID.String(),
		}
		act, err := domainTask.NewTaskActivity(e.TaskID, actorID, domainTask.ActivityColumnMoved, payload)
		if err == nil {
			_ = l.taskRepo.CreateActivity(ctx, act)
		}
	}
	return nil
}

// ListTaskActivities returns chronological activities for a given task, enforcing workspace access.
func (s *Service) ListTaskActivities(ctx context.Context, workspaceID, taskID uuid.UUID) ([]*domainTask.TaskActivity, error) {
	if err := s.checkWorkspaceAccess(ctx, workspaceID); err != nil {
		return nil, err
	}

	if _, err := s.taskRepo.GetByID(ctx, taskID, workspaceID); err != nil {
		return nil, err
	}

	return s.taskRepo.ListActivitiesByTaskID(ctx, taskID)
}

func getActorIDFromContext(ctx context.Context) *uuid.UUID {
	if ctx == nil {
		return nil
	}
	if val := ctx.Value("userID"); val != nil {
		if uid, ok := val.(uuid.UUID); ok && uid != uuid.Nil {
			return &uid
		}
	}
	return nil
}
