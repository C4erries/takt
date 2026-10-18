package events

import "context"

type contextKey string

const (
	automationTriggerKey contextKey = "automation_triggered"
	automationDepthKey   contextKey = "automation_depth"
	MaxAutomationDepth              = 3
)

// WithAutomationTrigger returns a context marked as originating from an automation and increments execution depth.
func WithAutomationTrigger(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	depth := GetAutomationDepth(ctx) + 1
	ctx = context.WithValue(ctx, automationDepthKey, depth)
	return context.WithValue(ctx, automationTriggerKey, true)
}

// WithAutomationDepth returns a context with explicit depth for testing.
func WithAutomationDepth(ctx context.Context, depth int) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, automationDepthKey, depth)
	return context.WithValue(ctx, automationTriggerKey, depth > 0)
}

// GetAutomationDepth returns current automation execution chain depth.
func GetAutomationDepth(ctx context.Context) int {
	if ctx == nil {
		return 0
	}
	val, ok := ctx.Value(automationDepthKey).(int)
	if !ok {
		return 0
	}
	return val
}

// IsAutomationTriggered checks whether the context was triggered by an automation action.
func IsAutomationTriggered(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	val, ok := ctx.Value(automationTriggerKey).(bool)
	return ok && val
}

// CanExecuteAutomation checks whether the chain depth is within MaxAutomationDepth.
func CanExecuteAutomation(ctx context.Context) bool {
	return GetAutomationDepth(ctx) < MaxAutomationDepth
}
