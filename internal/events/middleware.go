package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Common middleware for event processing

// LoggingMiddleware logs event processing
func LoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			start := time.Now()

			logger.Debug("event processing started",
				"event_id", event.ID,
				"event_type", event.Type,
				"topic", event.Topic)

			err := next(ctx, event)

			duration := time.Since(start)

			if err != nil {
				logger.Error("event processing failed",
					"event_id", event.ID,
					"event_type", event.Type,
					"duration", duration,
					"error", err)
			} else {
				logger.Debug("event processing completed",
					"event_id", event.ID,
					"event_type", event.Type,
					"duration", duration)
			}

			return err
		}
	}
}

// RetryMiddleware retries failed event processing
func RetryMiddleware(maxRetries int, delay time.Duration, logger *slog.Logger) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			var lastErr error

			for attempt := 0; attempt < maxRetries; attempt++ {
				err := next(ctx, event)
				if err == nil {
					return nil
				}

				lastErr = err

				if attempt < maxRetries-1 {
					logger.Warn("event processing failed, retrying",
						"event_id", event.ID,
						"attempt", attempt+1,
						"max_retries", maxRetries,
						"error", err)

					select {
					case <-time.After(delay):
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}

			return fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, lastErr)
		}
	}
}

// TimeoutMiddleware adds timeout to event processing
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			errChan := make(chan error, 1)

			go func() {
				errChan <- next(ctx, event)
			}()

			select {
			case err := <-errChan:
				return err
			case <-ctx.Done():
				return fmt.Errorf("event processing timeout after %v: %w", timeout, ctx.Err())
			}
		}
	}
}

// RecoveryMiddleware recovers from panics in event handlers
func RecoveryMiddleware(logger *slog.Logger) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("event handler panicked",
						"event_id", event.ID,
						"event_type", event.Type,
						"panic", r)

					err = fmt.Errorf("handler panicked: %v", r)
				}
			}()

			return next(ctx, event)
		}
	}
}

// FilterMiddleware filters events based on a predicate
func FilterMiddleware(predicate func(Event) bool) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			if !predicate(event) {
				return nil // Skip processing
			}
			return next(ctx, event)
		}
	}
}

// MetricsMiddleware tracks event processing metrics
func MetricsMiddleware(metrics *EventMetrics) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			start := time.Now()
			metrics.RecordEvent(string(event.Type))

			err := next(ctx, event)
			duration := time.Since(start)

			if err != nil {
				metrics.RecordError(string(event.Type))
			} else {
				metrics.RecordSuccess(string(event.Type))
			}

			metrics.RecordDuration(string(event.Type), duration)

			return err
		}
	}
}

// EventMetrics tracks event processing metrics
type EventMetrics struct {
	eventsProcessed map[string]int64
	eventsSucceeded map[string]int64
	eventsFailed    map[string]int64
	totalDuration   map[string]time.Duration
	eventCount      map[string]int64
}

// NewEventMetrics creates new event metrics
func NewEventMetrics() *EventMetrics {
	return &EventMetrics{
		eventsProcessed: make(map[string]int64),
		eventsSucceeded: make(map[string]int64),
		eventsFailed:    make(map[string]int64),
		totalDuration:   make(map[string]time.Duration),
		eventCount:      make(map[string]int64),
	}
}

// RecordEvent records an event
func (em *EventMetrics) RecordEvent(eventType string) {
	em.eventsProcessed[eventType]++
}

// RecordSuccess records a successful event
func (em *EventMetrics) RecordSuccess(eventType string) {
	em.eventsSucceeded[eventType]++
}

// RecordError records a failed event
func (em *EventMetrics) RecordError(eventType string) {
	em.eventsFailed[eventType]++
}

// RecordDuration records event processing duration
func (em *EventMetrics) RecordDuration(eventType string, duration time.Duration) {
	em.totalDuration[eventType] += duration
	em.eventCount[eventType]++
}

// GetAverageDuration returns average duration for an event type
func (em *EventMetrics) GetAverageDuration(eventType string) time.Duration {
	count := em.eventCount[eventType]
	if count == 0 {
		return 0
	}
	return em.totalDuration[eventType] / time.Duration(count)
}

// GetStats returns metrics for an event type
func (em *EventMetrics) GetStats(eventType string) map[string]interface{} {
	return map[string]interface{}{
		"processed":        em.eventsProcessed[eventType],
		"succeeded":        em.eventsSucceeded[eventType],
		"failed":           em.eventsFailed[eventType],
		"average_duration": em.GetAverageDuration(eventType),
	}
}

// DeduplicationMiddleware prevents duplicate event processing
func DeduplicationMiddleware(window time.Duration) Middleware {
	seen := newSeenCache(window)

	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			if seen.Has(event.ID) {
				return nil // Skip duplicate
			}

			seen.Add(event.ID)
			return next(ctx, event)
		}
	}
}

// seenCache tracks recently seen event IDs
type seenCache struct {
	items  map[string]time.Time
	window time.Duration
}

func newSeenCache(window time.Duration) *seenCache {
	return &seenCache{
		items:  make(map[string]time.Time),
		window: window,
	}
}

func (sc *seenCache) Has(id string) bool {
	expiry, exists := sc.items[id]
	if !exists {
		return false
	}

	if time.Since(expiry) > sc.window {
		delete(sc.items, id)
		return false
	}

	return true
}

func (sc *seenCache) Add(id string) {
	sc.items[id] = time.Now()

	// Clean up old entries periodically
	if len(sc.items) > 1000 {
		now := time.Now()
		for id, expiry := range sc.items {
			if now.Sub(expiry) > sc.window {
				delete(sc.items, id)
			}
		}
	}
}
