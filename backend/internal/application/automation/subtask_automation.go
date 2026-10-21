package automation

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

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

// handleParentAutoComplete checks if the parent task should be automatically marked
// completed when all of its subtasks are finished.
func (l *EventListener) handleParentAutoComplete(ctx context.Context, workspaceID, parentID uuid.UUID) {
	if l.taskReader == nil || l.taskModifier == nil {
		return
	}

	rules, err := l.repo.ListWorkspaceAutomations(ctx, workspaceID)
	if err != nil {
		return
	}

	shouldAutoComplete := false
	for _, rule := range rules {
		if !rule.IsEnabled {
			continue
		}
		if rule.Name == "auto_complete_parent_on_subtasks_done" {
			shouldAutoComplete = true
			break
		}
		var payload actionPayload
		if err := json.Unmarshal(rule.ActionsJSON, &payload); err == nil && payload.Action == "mark_parent_completed" {
			shouldAutoComplete = true
			break
		}
		if rule.RuleDefinition != nil {
			for _, act := range rule.RuleDefinition.Actions {
				if act.Type == "mark_parent_completed" {
					shouldAutoComplete = true
					break
				}
			}
			if shouldAutoComplete {
				break
			}
		}
	}

	if !shouldAutoComplete {
		return
	}

	subtasks, err := l.taskReader.ListSubtasks(ctx, parentID)
	if err != nil || len(subtasks) == 0 {
		return
	}

	for _, st := range subtasks {
		if !st.IsCompleted {
			return
		}
	}

	_ = l.taskModifier.SetTaskCompleted(ctx, parentID, workspaceID, true)
}
