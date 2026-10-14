package events

import (
	"context"
)

// EventHandler defines the function signature for consuming domain events.
type EventHandler func(ctx context.Context, event Event) error

// EventBus defines the publishing and subscription contract for domain events.
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventName string, handler EventHandler) error
}

