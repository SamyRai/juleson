package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventBuilder(t *testing.T) {
	e := NewEvent(EventSystemStarted, "test-source", "test-data")

	e = e.WithTopic("test-topic").
		WithPriority(10).
		WithTTL(1*time.Minute).
		WithMetadata("key1", "value1").
		WithMetadata("key2", "value2")

	assert.Equal(t, "test-topic", e.Topic)
	assert.Equal(t, 10, e.Priority)
	assert.Equal(t, 1*time.Minute, e.TTL)
	assert.Equal(t, "value1", e.Metadata["key1"])
	assert.Equal(t, "value2", e.Metadata["key2"])
}

func TestEventIsExpired(t *testing.T) {
	e := NewEvent(EventSystemStarted, "test-source", nil)
	assert.False(t, e.IsExpired())

	e = e.WithTTL(1 * time.Millisecond)
	time.Sleep(2 * time.Millisecond)
	assert.True(t, e.IsExpired())
}

func TestEventClone(t *testing.T) {
	e1 := NewEvent(EventSystemStarted, "test-source", nil).
		WithMetadata("k", "v")

	e2 := e1.Clone()

	assert.Equal(t, e1.ID, e2.ID)
	assert.Equal(t, e1.Metadata["k"], e2.Metadata["k"])

	// Modify clone, should not affect original
	e2.Metadata["k"] = "new-v"
	assert.Equal(t, "v", e1.Metadata["k"])
	assert.Equal(t, "new-v", e2.Metadata["k"])
}

func TestEventJSONMarshaling(t *testing.T) {
	e := NewEvent(EventSystemStarted, "test-source", "data")

	bytes, err := json.Marshal(e)
	require.NoError(t, err)

	var e2 Event
	err = json.Unmarshal(bytes, &e2)
	require.NoError(t, err)

	assert.Equal(t, e.ID, e2.ID)
	assert.Equal(t, e.Type, e2.Type)
	assert.Equal(t, e.Source, e2.Source)
	assert.Equal(t, e.Data, e2.Data)
	assert.Equal(t, e.Timestamp.UnixNano(), e2.Timestamp.UnixNano())
}
