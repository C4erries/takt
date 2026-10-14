package eventbus_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"takt/internal/domain/events"
	"takt/internal/infrastructure/eventbus"
)

type dummyEvent struct {
	name string
	wsID uuid.UUID
}

func (d dummyEvent) EventName() string      { return d.name }
func (d dummyEvent) OccurredAt() time.Time  { return time.Now().UTC() }
func (d dummyEvent) WorkspaceID() uuid.UUID { return d.wsID }
func (d dummyEvent) IsFromAutomation() bool { return false }

func TestInMemoryEventBus_PublishSubscribe(t *testing.T) {
	bus := eventbus.NewInMemoryEventBus(
		eventbus.WithQueueSize(16),
		eventbus.WithWorkerCount(2),
	)
	defer func() {
		_ = bus.Close(context.Background())
	}()

	wsID := uuid.New()
	received := make(chan events.Event, 2)

	err := bus.Subscribe("test_event", func(ctx context.Context, e events.Event) error {
		received <- e
		return nil
	})
	require.NoError(t, err)

	evt := dummyEvent{name: "test_event", wsID: wsID}
	err = bus.Publish(context.Background(), evt)
	require.NoError(t, err)

	select {
	case got := <-received:
		assert.Equal(t, "test_event", got.EventName())
		assert.Equal(t, wsID, got.WorkspaceID())
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestInMemoryEventBus_MultipleSubscribers(t *testing.T) {
	bus := eventbus.NewInMemoryEventBus(eventbus.WithWorkerCount(4))
	defer func() {
		_ = bus.Close(context.Background())
	}()

	var count int32
	var wg sync.WaitGroup
	wg.Add(3)

	handler := func(ctx context.Context, e events.Event) error {
		atomic.AddInt32(&count, 1)
		wg.Done()
		return nil
	}

	_ = bus.Subscribe("multi_event", handler)
	_ = bus.Subscribe("multi_event", handler)
	_ = bus.Subscribe("multi_event", handler)

	err := bus.Publish(context.Background(), dummyEvent{name: "multi_event", wsID: uuid.New()})
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.Equal(t, int32(3), atomic.LoadInt32(&count))
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for all subscribers")
	}
}

func TestInMemoryEventBus_Close(t *testing.T) {
	bus := eventbus.NewInMemoryEventBus(eventbus.WithQueueSize(8), eventbus.WithWorkerCount(2))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := bus.Close(ctx)
	require.NoError(t, err)

	// Publishing to closed bus returns ErrBusClosed
	err = bus.Publish(context.Background(), dummyEvent{name: "post_close"})
	assert.ErrorIs(t, err, eventbus.ErrBusClosed)
}

