package automation

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TaskContext provides contextual task properties evaluated by rules.
type TaskContext struct {
	IsCompleted       bool
	HasDueDate        bool
	DueDate           *time.Time
	ProjectID         *uuid.UUID
	ColumnID          *uuid.UUID
	PreviousProjectID *uuid.UUID
	PreviousColumnID  *uuid.UUID
}

// Conditions represents supported rule condition criteria.
type Conditions struct {
	HasDueDate      *bool `json:"has_due_date,omitempty"`
	DueDateLeToday  *bool `json:"due_date_le_today,omitempty"`
	ProjectAssigned *bool `json:"project_assigned,omitempty"`
	IsCompleted     *bool `json:"is_completed,omitempty"`
	InboxOnly       *bool `json:"inbox_only,omitempty"`
}

// RuleEvaluator evaluates rule conditions against task context.
type RuleEvaluator struct{}

// NewRuleEvaluator creates a new pure RuleEvaluator.
func NewRuleEvaluator() *RuleEvaluator {
	return &RuleEvaluator{}
}

// Evaluate evaluates conditionsJSON against task context. Returns true if all conditions pass.
func (e *RuleEvaluator) Evaluate(conditionsJSON json.RawMessage, ctx TaskContext) (bool, error) {
	trimmed := bytes.TrimSpace(conditionsJSON)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("{}")) || bytes.Equal(trimmed, []byte("null")) {
		return true, nil
	}

	var cond Conditions
	if err := json.Unmarshal(trimmed, &cond); err != nil {
		return false, err
	}

	if cond.HasDueDate != nil {
		if ctx.HasDueDate != *cond.HasDueDate {
			return false, nil
		}
	}

	if cond.DueDateLeToday != nil {
		if ctx.DueDate == nil {
			if *cond.DueDateLeToday {
				return false, nil
			}
		} else {
			todayStr := time.Now().UTC().Format("2006-01-02")
			dueStr := ctx.DueDate.Format("2006-01-02")
			isLeToday := dueStr <= todayStr
			if isLeToday != *cond.DueDateLeToday {
				return false, nil
			}
		}
	}

	if cond.ProjectAssigned != nil {
		hasProject := ctx.ProjectID != nil
		if hasProject != *cond.ProjectAssigned {
			return false, nil
		}
	}

	if cond.IsCompleted != nil {
		if ctx.IsCompleted != *cond.IsCompleted {
			return false, nil
		}
	}

	if cond.InboxOnly != nil {
		isInbox := ctx.ColumnID == nil
		if isInbox != *cond.InboxOnly {
			return false, nil
		}
	}

	return true, nil
}

