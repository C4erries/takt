package eventbus

import (
	"context"
	"errors"
	"sync"

	"takt/internal/domain/events"
)

var (
	// ErrBusClosed is returned when publishing to a closed event bus.
	ErrBusClosed = errors.New("event bus is closed")
)

const (
	defaultQueueSize   = 256
	defaultWorkerCount = 4
)

// Option allows configuring InMemoryEventBus.
type Option func(*InMemoryEventBus)

// WithQueueSize configures the buffer size for incoming events.
func WithQueueSize(size int) Option {
	return func(b *InMemoryEventBus) {
		if size > 0 {
			b.queueSize = size
		}
	}
}

// WithWorkerCount configures the number of background worker goroutines.
func WithWorkerCount(count int) Option {
	return func(b *InMemoryEventBus) {
		if count > 0 {
			b.workerCount = count
		}
	}
}

// InMemoryEventBus implements events.EventBus with buffered channels and a worker pool.
type InMemoryEventBus struct {
	queueSize   int
	workerCount int

	subscribers map[string][]events.EventHandler
	subMu       sync.RWMutex

	eventCh chan events.Event
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc

	closedMu sync.Mutex
	isClosed bool
}

// NewInMemoryEventBus creates and starts a new in-memory event bus.
func NewInMemoryEventBus(opts ...Option) *InMemoryEventBus {
	ctx, cancel := context.WithCancel(context.Background())

	b := &InMemoryEventBus{
		queueSize:   defaultQueueSize,
		workerCount: defaultWorkerCount,
		subscribers: make(map[string][]events.EventHandler),
		ctx:         ctx,
		cancel:      cancel,
	}

	for _, opt := range opts {
		opt(b)
	}

	b.eventCh = make(chan events.Event, b.queueSize)

	for i := 0; i < b.workerCount; i++ {
		b.wg.Add(1)
		go b.worker()
	}

	return b
}

func (b *InMemoryEventBus) worker() {
	defer b.wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			// Drain remaining events in channel before exit
			for {
				select {
				case event, ok := <-b.eventCh:
					if !ok {
						return
					}
					b.dispatch(context.Background(), event)
				default:
					return
				}
			}
		case event, ok := <-b.eventCh:
			if !ok {
				return
			}
			b.dispatch(b.ctx, event)
		}
	}
}

func (b *InMemoryEventBus) dispatch(ctx context.Context, event events.Event) {
	if event == nil {
		return
	}

	b.subMu.RLock()
	handlers := make([]events.EventHandler, len(b.subscribers[event.EventName()]))
	copy(handlers, b.subscribers[event.EventName()])
	b.subMu.RUnlock()

	for _, handler := range handlers {
		// Run handler; ignore handler errors so one subscriber doesn't block others
		_ = handler(ctx, event)
	}
}

// Publish queues an event for asynchronous dispatch to registered subscribers.
func (b *InMemoryEventBus) Publish(ctx context.Context, event events.Event) error {
	b.closedMu.Lock()
	if b.isClosed {
		b.closedMu.Unlock()
		return ErrBusClosed
	}
	b.closedMu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.ctx.Done():
		return ErrBusClosed
	case b.eventCh <- event:
		return nil
	}
}

// Subscribe registers an event handler for a specific event name.
func (b *InMemoryEventBus) Subscribe(eventName string, handler events.EventHandler) error {
	if handler == nil {
		return nil
	}

	b.subMu.Lock()
	defer b.subMu.Unlock()

	b.subscribers[eventName] = append(b.subscribers[eventName], handler)
	return nil
}

// Close gracefully stops the event bus, waiting for pending events to finish.
func (b *InMemoryEventBus) Close(ctx context.Context) error {
	b.closedMu.Lock()
	if b.isClosed {
		b.closedMu.Unlock()
		return nil
	}
	b.isClosed = true
	b.closedMu.Unlock()

	close(b.eventCh)
	b.cancel()

	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

