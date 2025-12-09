package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// EventBus provides centralized event distribution for the application.
// It supports pub/sub pattern with topic-based routing and priority queues.
type EventBus struct {
	subscribers map[string][]Subscriber
	mu          sync.RWMutex
	logger      *slog.Logger
	metrics     *BusMetrics
	middleware  []Middleware
	stopping    bool
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// Subscriber represents an event subscriber
type Subscriber struct {
	ID       string
	Handler  EventHandler
	Filter   EventFilter
	Priority int // Higher priority subscribers receive events first
	Async    bool
}

// EventHandler handles an event
type EventHandler func(ctx context.Context, event Event) error

// EventFilter filters events based on criteria
type EventFilter func(event Event) bool

// Middleware provides event processing middleware
type Middleware func(next EventHandler) EventHandler

// BusMetrics tracks event bus metrics
type BusMetrics struct {
	EventsPublished int64
	EventsDelivered int64
	EventsFailed    int64
	SubscriberCount int
	AverageLatency  time.Duration
	mu              sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus(logger *slog.Logger) *EventBus {
	if logger == nil {
		logger = slog.Default()
	}

	return &EventBus{
		subscribers: make(map[string][]Subscriber),
		logger:      logger,
		metrics:     &BusMetrics{},
		middleware:  make([]Middleware, 0),
		stopChan:    make(chan struct{}),
	}
}

// Subscribe subscribes to events on a topic
func (eb *EventBus) Subscribe(topic string, subscriber Subscriber) error {
	if topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}
	if subscriber.ID == "" {
		return fmt.Errorf("subscriber ID cannot be empty")
	}
	if subscriber.Handler == nil {
		return fmt.Errorf("subscriber handler cannot be nil")
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.stopping {
		return fmt.Errorf("event bus is stopping")
	}

	// Check for duplicate subscriber ID
	for _, sub := range eb.subscribers[topic] {
		if sub.ID == subscriber.ID {
			return fmt.Errorf("subscriber with ID %s already exists for topic %s", subscriber.ID, topic)
		}
	}

	eb.subscribers[topic] = append(eb.subscribers[topic], subscriber)

	// Sort by priority (higher first)
	subs := eb.subscribers[topic]
	for i := 0; i < len(subs); i++ {
		for j := i + 1; j < len(subs); j++ {
			if subs[j].Priority > subs[i].Priority {
				subs[i], subs[j] = subs[j], subs[i]
			}
		}
	}

	eb.metrics.mu.Lock()
	eb.metrics.SubscriberCount++
	eb.metrics.mu.Unlock()

	eb.logger.Info("subscriber added",
		"topic", topic,
		"subscriber_id", subscriber.ID,
		"priority", subscriber.Priority,
		"async", subscriber.Async)

	return nil
}

// Unsubscribe removes a subscriber from a topic
func (eb *EventBus) Unsubscribe(topic, subscriberID string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs, exists := eb.subscribers[topic]
	if !exists {
		return fmt.Errorf("topic %s not found", topic)
	}

	for i, sub := range subs {
		if sub.ID == subscriberID {
			eb.subscribers[topic] = append(subs[:i], subs[i+1:]...)

			eb.metrics.mu.Lock()
			eb.metrics.SubscriberCount--
			eb.metrics.mu.Unlock()

			eb.logger.Info("subscriber removed",
				"topic", topic,
				"subscriber_id", subscriberID)
			return nil
		}
	}

	return fmt.Errorf("subscriber %s not found for topic %s", subscriberID, topic)
}

// Publish publishes an event to all subscribers of a topic
func (eb *EventBus) Publish(ctx context.Context, topic string, event Event) error {
	if topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}

	eb.mu.RLock()
	if eb.stopping {
		eb.mu.RUnlock()
		return fmt.Errorf("event bus is stopping")
	}

	subs, exists := eb.subscribers[topic]
	if !exists || len(subs) == 0 {
		eb.mu.RUnlock()
		eb.logger.Debug("no subscribers for topic", "topic", topic)
		return nil
	}

	// Make a copy to avoid holding lock during delivery
	subsCopy := make([]Subscriber, len(subs))
	copy(subsCopy, subs)
	eb.mu.RUnlock()

	// Set topic if not already set
	if event.Topic == "" {
		event.Topic = topic
	}

	// Set timestamp if not already set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	eb.metrics.mu.Lock()
	eb.metrics.EventsPublished++
	eb.metrics.mu.Unlock()

	startTime := time.Now()

	// Deliver to all matching subscribers
	var wg sync.WaitGroup
	errors := make([]error, 0)
	errorsMu := sync.Mutex{}

	for _, sub := range subsCopy {
		// Apply filter if present
		if sub.Filter != nil && !sub.Filter(event) {
			continue
		}

		// Build handler chain with middleware
		handler := sub.Handler
		for i := len(eb.middleware) - 1; i >= 0; i-- {
			handler = eb.middleware[i](handler)
		}

		if sub.Async {
			wg.Add(1)
			eb.wg.Add(1)
			go func(s Subscriber, h EventHandler) {
				defer wg.Done()
				defer eb.wg.Done()

				if err := h(ctx, event); err != nil {
					errorsMu.Lock()
					errors = append(errors, fmt.Errorf("subscriber %s failed: %w", s.ID, err))
					errorsMu.Unlock()

					eb.metrics.mu.Lock()
					eb.metrics.EventsFailed++
					eb.metrics.mu.Unlock()

					eb.logger.Error("async subscriber error",
						"subscriber_id", s.ID,
						"topic", topic,
						"error", err)
				} else {
					eb.metrics.mu.Lock()
					eb.metrics.EventsDelivered++
					eb.metrics.mu.Unlock()
				}
			}(sub, handler)
		} else {
			if err := handler(ctx, event); err != nil {
				errors = append(errors, fmt.Errorf("subscriber %s failed: %w", sub.ID, err))

				eb.metrics.mu.Lock()
				eb.metrics.EventsFailed++
				eb.metrics.mu.Unlock()

				eb.logger.Error("subscriber error",
					"subscriber_id", sub.ID,
					"topic", topic,
					"error", err)
			} else {
				eb.metrics.mu.Lock()
				eb.metrics.EventsDelivered++
				eb.metrics.mu.Unlock()
			}
		}
	}

	wg.Wait()

	// Update average latency
	latency := time.Since(startTime)
	eb.metrics.mu.Lock()
	if eb.metrics.AverageLatency == 0 {
		eb.metrics.AverageLatency = latency
	} else {
		eb.metrics.AverageLatency = (eb.metrics.AverageLatency + latency) / 2
	}
	eb.metrics.mu.Unlock()

	if len(errors) > 0 {
		return fmt.Errorf("event delivery had %d errors: %v", len(errors), errors[0])
	}

	eb.logger.Debug("event published",
		"topic", topic,
		"event_type", event.Type,
		"subscribers", len(subsCopy),
		"latency", latency)

	return nil
}

// Use adds middleware to the event bus
func (eb *EventBus) Use(middleware Middleware) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.middleware = append(eb.middleware, middleware)
}

// GetMetrics returns current bus metrics
func (eb *EventBus) GetMetrics() BusMetrics {
	eb.metrics.mu.RLock()
	defer eb.metrics.mu.RUnlock()

	return BusMetrics{
		EventsPublished: eb.metrics.EventsPublished,
		EventsDelivered: eb.metrics.EventsDelivered,
		EventsFailed:    eb.metrics.EventsFailed,
		SubscriberCount: eb.metrics.SubscriberCount,
		AverageLatency:  eb.metrics.AverageLatency,
	}
}

// Shutdown gracefully shuts down the event bus
func (eb *EventBus) Shutdown(ctx context.Context) error {
	eb.mu.Lock()
	eb.stopping = true
	eb.mu.Unlock()

	// Wait for all async handlers to complete
	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		eb.logger.Info("event bus shutdown complete")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}

// Clear removes all subscribers from all topics
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers = make(map[string][]Subscriber)
	eb.metrics.mu.Lock()
	eb.metrics.SubscriberCount = 0
	eb.metrics.mu.Unlock()

	eb.logger.Info("event bus cleared")
}

// GetTopics returns all topics with subscribers
func (eb *EventBus) GetTopics() []string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	topics := make([]string, 0, len(eb.subscribers))
	for topic := range eb.subscribers {
		topics = append(topics, topic)
	}
	return topics
}

// GetSubscribers returns all subscribers for a topic
func (eb *EventBus) GetSubscribers(topic string) []string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	subs, exists := eb.subscribers[topic]
	if !exists {
		return []string{}
	}

	ids := make([]string, len(subs))
	for i, sub := range subs {
		ids[i] = sub.ID
	}
	return ids
}
