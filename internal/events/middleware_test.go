package events

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryMiddleware(t *testing.T) {
	logger := slog.Default()
	middleware := RetryMiddleware(3, 10*time.Millisecond, logger)

	var attempts int
	handler := middleware(func(ctx context.Context, e Event) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	err := handler(context.Background(), NewEvent(EventSystemStarted, "test", nil))
	require.NoError(t, err)
	assert.Equal(t, 3, attempts)

	// Test max retries exceeded
	attempts = 0
	handlerFail := middleware(func(ctx context.Context, e Event) error {
		attempts++
		return errors.New("permanent error")
	})

	err = handlerFail(context.Background(), NewEvent(EventSystemStarted, "test", nil))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max retries (3) exceeded")
	assert.Equal(t, 3, attempts)
}

func TestTimeoutMiddleware(t *testing.T) {
	middleware := TimeoutMiddleware(50 * time.Millisecond)

	handler := middleware(func(ctx context.Context, e Event) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	err := handler(context.Background(), NewEvent(EventSystemStarted, "test", nil))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event processing timeout")
}

func TestFilterMiddleware(t *testing.T) {
	middleware := FilterMiddleware(func(e Event) bool {
		return e.Type == EventSystemStarted
	})

	var processed int
	handler := middleware(func(ctx context.Context, e Event) error {
		processed++
		return nil
	})

	_ = handler(context.Background(), NewEvent(EventSystemStarted, "test", nil))
	assert.Equal(t, 1, processed)

	_ = handler(context.Background(), NewEvent(EventSystemError, "test", nil))
	assert.Equal(t, 1, processed) // Should be skipped
}

func TestMetricsMiddleware(t *testing.T) {
	metrics := NewEventMetrics()
	middleware := MetricsMiddleware(metrics)

	handler := middleware(func(ctx context.Context, e Event) error {
		if e.Source == "fail" {
			return errors.New("fail")
		}
		return nil
	})

	_ = handler(context.Background(), NewEvent(EventSystemStarted, "success", nil))
	_ = handler(context.Background(), NewEvent(EventSystemStarted, "fail", nil))

	stats := metrics.GetStats(string(EventSystemStarted))
	assert.Equal(t, int64(2), stats["processed"])
	assert.Equal(t, int64(1), stats["succeeded"])
	assert.Equal(t, int64(1), stats["failed"])
}

func TestDeduplicationMiddleware(t *testing.T) {
	middleware := DeduplicationMiddleware(1 * time.Second)

	var processed int
	handler := middleware(func(ctx context.Context, e Event) error {
		processed++
		return nil
	})

	event := NewEvent(EventSystemStarted, "test", nil)

	_ = handler(context.Background(), event)
	assert.Equal(t, 1, processed)

	_ = handler(context.Background(), event)
	assert.Equal(t, 1, processed) // Should skip duplicate
}
