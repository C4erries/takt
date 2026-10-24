package events

import (
	"time"

	"github.com/google/uuid"
)

// Event is the base interface that all domain events must satisfy.
type Event interface {
	EventName() string
	OccurredAt() time.Time
	WorkspaceID() uuid.UUID
	IsFromAutomation() bool
}

// Event name constants
const (
	EventTaskCreated          = "task_created"
	EventTaskUpdated          = "task_updated"
	EventTaskCompleted        = "task_completed"
	EventTaskMoved            = "task_moved"
	EventBatchSweepCompleted  = "batch_sweep_completed"
)

// TaskCreatedEvent is emitted when a new task is created.
type TaskCreatedEvent struct {
	TaskID         uuid.UUID
	WorkspaceIDVal uuid.UUID
	ParentID       *uuid.UUID
	ProjectID      *uuid.UUID
	ColumnID       *uuid.UUID
	Title          string
	DueDate        *time.Time
	OccurredAtVal  time.Time
	FromAutomation bool
}

func (e TaskCreatedEvent) EventName() string       { return EventTaskCreated }
func (e TaskCreatedEvent) OccurredAt() time.Time   { return e.OccurredAtVal }
func (e TaskCreatedEvent) WorkspaceID() uuid.UUID  { return e.WorkspaceIDVal }
func (e TaskCreatedEvent) IsFromAutomation() bool  { return e.FromAutomation }

// TaskUpdatedEvent is emitted when task fields are modified.
type TaskUpdatedEvent struct {
	TaskID            uuid.UUID
	WorkspaceIDVal    uuid.UUID
	ParentID          *uuid.UUID
	ProjectID         *uuid.UUID
	ColumnID          *uuid.UUID
	Title             string
	DueDate           *time.Time
	IsCompleted       bool
	PreviousProjectID *uuid.UUID
	PreviousColumnID  *uuid.UUID
	OccurredAtVal     time.Time
	FromAutomation    bool
}

func (e TaskUpdatedEvent) EventName() string       { return EventTaskUpdated }
func (e TaskUpdatedEvent) OccurredAt() time.Time   { return e.OccurredAtVal }
func (e TaskUpdatedEvent) WorkspaceID() uuid.UUID  { return e.WorkspaceIDVal }
func (e TaskUpdatedEvent) IsFromAutomation() bool  { return e.FromAutomation }

// TaskCompletedEvent is emitted when a task is toggled to completed.
type TaskCompletedEvent struct {
	TaskID         uuid.UUID
	WorkspaceIDVal uuid.UUID
	ParentID       *uuid.UUID
	ProjectID      *uuid.UUID
	ColumnID       *uuid.UUID
	CompletedAt    time.Time
	OccurredAtVal  time.Time
	FromAutomation bool
}

func (e TaskCompletedEvent) EventName() string       { return EventTaskCompleted }
func (e TaskCompletedEvent) OccurredAt() time.Time   { return e.OccurredAtVal }
func (e TaskCompletedEvent) WorkspaceID() uuid.UUID  { return e.WorkspaceIDVal }
func (e TaskCompletedEvent) IsFromAutomation() bool  { return e.FromAutomation }

// TaskMovedEvent is emitted when a task is moved to a column or board position.
type TaskMovedEvent struct {
	TaskID           uuid.UUID
	WorkspaceIDVal   uuid.UUID
	ProjectID        *uuid.UUID
	ColumnID         *uuid.UUID
	PreviousColumnID *uuid.UUID
	Position         float64
	OccurredAtVal    time.Time
	FromAutomation   bool
}

func (e TaskMovedEvent) EventName() string       { return EventTaskMoved }
func (e TaskMovedEvent) OccurredAt() time.Time   { return e.OccurredAtVal }
func (e TaskMovedEvent) WorkspaceID() uuid.UUID  { return e.WorkspaceIDVal }
func (e TaskMovedEvent) IsFromAutomation() bool  { return e.FromAutomation }

// BatchSweepCompletedEvent is emitted when a batch sweep operation or column flush affects multiple tasks.
type BatchSweepCompletedEvent struct {
	WorkspaceIDVal uuid.UUID
	ColumnIDVal    uuid.UUID
	ActionVal      string
	AffectedCount  int64
	OccurredAtVal  time.Time
}

func (e BatchSweepCompletedEvent) EventName() string       { return EventBatchSweepCompleted }
func (e BatchSweepCompletedEvent) OccurredAt() time.Time   { return e.OccurredAtVal }
func (e BatchSweepCompletedEvent) WorkspaceID() uuid.UUID  { return e.WorkspaceIDVal }
func (e BatchSweepCompletedEvent) IsFromAutomation() bool  { return true }
func (e BatchSweepCompletedEvent) ColumnID() uuid.UUID     { return e.ColumnIDVal }
func (e BatchSweepCompletedEvent) Action() string          { return e.ActionVal }

