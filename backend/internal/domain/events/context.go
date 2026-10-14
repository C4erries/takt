package events

import "context"

type contextKey string

const (
	automationTriggerKey contextKey = "automation_triggered"
)

// WithAutomationTrigger returns a context marked as originating from an automation.
func WithAutomationTrigger(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, automationTriggerKey, true)
}

// IsAutomationTriggered checks whether the context was triggered by an automation action.
func IsAutomationTriggered(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	val, ok := ctx.Value(automationTriggerKey).(bool)
	return ok && val
}

