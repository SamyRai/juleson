package events

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventStore(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "events.db")

	config := &EventStoreConfig{
		StorageDir:    storePath,
		MaxEvents:     100,
		FlushInterval: 10 * time.Millisecond,
	}

	store, err := NewEventStore(config, nil)
	require.NoError(t, err)

	ctx := context.Background()
	event1 := NewEvent(EventSystemStarted, "source-a", map[string]string{"k": "v1"})
	event2 := NewEvent(EventSystemStarted, "source-b", map[string]string{"k": "v2"})

	err = store.Store(event1)
	require.NoError(t, err)
	err = store.Store(event2)
	require.NoError(t, err)

	assert.Equal(t, 2, store.Count())

	retrieved, err := store.Get(event1.ID)
	require.NoError(t, err)
	assert.Equal(t, event1.ID, retrieved.ID)

	byType := store.GetByType(EventSystemStarted)
	assert.Len(t, byType, 2)

	// Test flush and persistence
	err = store.Flush()
	require.NoError(t, err)
	err = store.Shutdown(ctx)
	require.NoError(t, err)

	// Re-load
	store2, err := NewEventStore(config, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, store2.Count())

	store2.Clear()
	assert.Equal(t, 0, store2.Count())
	_ = store2.Shutdown(ctx)
}

func TestEventStore_Replay(t *testing.T) {
	tempDir := t.TempDir()
	storePath := tempDir

	store, _ := NewEventStore(&EventStoreConfig{
		StorageDir: storePath,
		MaxEvents:  10,
	}, nil)

	ctx := context.Background()
	e1 := NewEvent(EventSystemStarted, "s1", nil)
	e2 := NewEvent(EventSystemStarted, "s2", nil)
	_ = store.Store(e1)
	_ = store.Store(e2)

	replayed := []Event{}
	err := store.Replay(ctx, func(e Event) error {
		replayed = append(replayed, e)
		return nil
	})
	require.NoError(t, err)
	assert.Len(t, replayed, 2)
}
