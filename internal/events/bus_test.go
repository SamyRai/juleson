package events

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventBusSubscribeAndPublish(t *testing.T) {
	bus := NewEventBus(nil)
	ctx := context.Background()

	var received Event
	var mu sync.Mutex

	err := bus.Subscribe("test.topic", Subscriber{
		ID: "sub-1",
		Handler: func(ctx context.Context, e Event) error {
			mu.Lock()
			received = e
			mu.Unlock()
			return nil
		},
	})
	require.NoError(t, err)

	evt := Event{
		Type: "TEST",
		Data: "hello",
	}

	err = bus.Publish(ctx, "test.topic", evt)
	require.NoError(t, err)

	mu.Lock()
	assert.Equal(t, EventType("TEST"), received.Type)
	assert.Equal(t, "hello", received.Data)
	assert.Equal(t, "test.topic", received.Topic)
	mu.Unlock()

	// Verify metrics
	metrics := bus.GetMetrics()
	assert.Equal(t, int64(1), metrics.EventsPublished)
	assert.Equal(t, int64(1), metrics.EventsDelivered)
}

func TestEventBusUnsubscribe(t *testing.T) {
	bus := NewEventBus(nil)

	err := bus.Subscribe("test.topic", Subscriber{
		ID: "sub-1",
		Handler: func(ctx context.Context, e Event) error {
			return nil
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"sub-1"}, bus.GetSubscribers("test.topic"))

	err = bus.Unsubscribe("test.topic", "sub-1")
	require.NoError(t, err)

	assert.Empty(t, bus.GetSubscribers("test.topic"))
}

func TestEventBusFilter(t *testing.T) {
	bus := NewEventBus(nil)
	ctx := context.Background()

	var count int
	var mu sync.Mutex

	err := bus.Subscribe("test.topic", Subscriber{
		ID: "sub-1",
		Filter: func(e Event) bool {
			return e.Type == "ALLOW"
		},
		Handler: func(ctx context.Context, e Event) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		},
	})
	require.NoError(t, err)

	// Publish blocked event
	bus.Publish(ctx, "test.topic", Event{Type: "BLOCK"})

	// Publish allowed event
	bus.Publish(ctx, "test.topic", Event{Type: "ALLOW"})

	mu.Lock()
	assert.Equal(t, 1, count)
	mu.Unlock()
}
