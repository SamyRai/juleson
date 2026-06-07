package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoordinatorLifecycle(t *testing.T) {
	config := DefaultCoordinatorConfig()
	// Disable store and queue to simplify test
	config.EnableStore = false
	config.EnableQueue = false

	coordinator, err := NewEventCoordinator(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = coordinator.Start(ctx)
	require.NoError(t, err)

	err = coordinator.Shutdown(ctx)
	require.NoError(t, err)
}

func TestCoordinatorPublishSubscribe(t *testing.T) {
	config := DefaultCoordinatorConfig()
	config.EnableStore = false
	config.EnableQueue = false

	coordinator, err := NewEventCoordinator(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = coordinator.Start(ctx)
	require.NoError(t, err)

	var received bool
	err = coordinator.Subscribe(TopicSession, Subscriber{
		ID: "test-sub",
		Handler: func(ctx context.Context, e Event) error {
			received = true
			return nil
		},
	})
	require.NoError(t, err)

	err = coordinator.EmitSessionEvent(ctx, EventSessionCreated, SessionEventData{
		SessionID: "s-123",
	})
	require.NoError(t, err)

	// Wait for delivery (synchronous by default)
	time.Sleep(50 * time.Millisecond)

	assert.True(t, received)

	metrics := coordinator.GetMetrics()
	assert.NotNil(t, metrics["bus"])

	err = coordinator.Shutdown(ctx)
	require.NoError(t, err)
}
