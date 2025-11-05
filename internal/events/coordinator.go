package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// EventCoordinator coordinates between EventBus, MessageQueue, and EventStore.
// It provides a unified interface for event-driven communication in the application.
type EventCoordinator struct {
	bus      *EventBus
	queue    *MessageQueue
	store    *EventStore
	breakers *CircuitBreakerPool
	logger   *slog.Logger
	mu       sync.RWMutex
	started  bool
}

// CoordinatorConfig configures the event coordinator
type CoordinatorConfig struct {
	EventStoreConfig *EventStoreConfig
	QueueConfig      *QueueConfig
	EnableStore      bool
	EnableQueue      bool
	Logger           *slog.Logger
}

// DefaultCoordinatorConfig returns default configuration
func DefaultCoordinatorConfig() *CoordinatorConfig {
	return &CoordinatorConfig{
		EventStoreConfig: DefaultEventStoreConfig(),
		QueueConfig:      DefaultQueueConfig(),
		EnableStore:      true,
		EnableQueue:      true,
		Logger:           slog.Default(),
	}
}

// NewEventCoordinator creates a new event coordinator
func NewEventCoordinator(config *CoordinatorConfig) (*EventCoordinator, error) {
	if config == nil {
		config = DefaultCoordinatorConfig()
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	ec := &EventCoordinator{
		bus:      NewEventBus(config.Logger),
		breakers: NewCircuitBreakerPool(config.Logger),
		logger:   config.Logger,
	}

	// Initialize event store if enabled
	if config.EnableStore {
		store, err := NewEventStore(config.EventStoreConfig, config.Logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create event store: %w", err)
		}
		ec.store = store
	}

	// Initialize message queue if enabled
	if config.EnableQueue {
		ec.queue = NewMessageQueue(config.QueueConfig, config.Logger)
	}

	// Set up standard middleware
	ec.setupMiddleware()

	return ec, nil
}

// setupMiddleware configures standard middleware for the event bus
func (ec *EventCoordinator) setupMiddleware() {
	// Add recovery middleware first (outermost)
	ec.bus.Use(RecoveryMiddleware(ec.logger))

	// Add logging middleware
	ec.bus.Use(LoggingMiddleware(ec.logger))

	// Add deduplication middleware (5 minute window)
	ec.bus.Use(DeduplicationMiddleware(5 * time.Minute))

	// Subscribe to all events for storage
	if ec.store != nil {
		ec.bus.Subscribe(TopicAll, Subscriber{
			ID:       "event-store-subscriber",
			Priority: -1, // Low priority, run after others
			Async:    true,
			Handler: func(ctx context.Context, event Event) error {
				return ec.store.Store(event)
			},
		})
	}

	ec.logger.Info("event coordinator middleware configured")
}

// Start starts the event coordinator
func (ec *EventCoordinator) Start(ctx context.Context) error {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if ec.started {
		return fmt.Errorf("coordinator already started")
	}

	ec.started = true
	ec.logger.Info("event coordinator started")

	return nil
}

// PublishEvent publishes an event to the event bus
func (ec *EventCoordinator) PublishEvent(ctx context.Context, event Event) error {
	return ec.bus.Publish(ctx, event.Topic, event)
}

// Subscribe subscribes to events on a topic
func (ec *EventCoordinator) Subscribe(topic string, subscriber Subscriber) error {
	return ec.bus.Subscribe(topic, subscriber)
}

// Unsubscribe removes a subscriber from a topic
func (ec *EventCoordinator) Unsubscribe(topic, subscriberID string) error {
	return ec.bus.Unsubscribe(topic, subscriberID)
}

// EnqueueMessage enqueues a message for asynchronous processing
func (ec *EventCoordinator) EnqueueMessage(msg Message) error {
	if ec.queue == nil {
		return fmt.Errorf("message queue not enabled")
	}
	return ec.queue.Enqueue(msg)
}

// CreateQueue creates a new message queue
func (ec *EventCoordinator) CreateQueue(name string, maxSize int) error {
	if ec.queue == nil {
		return fmt.Errorf("message queue not enabled")
	}
	return ec.queue.CreateQueue(name, maxSize)
}

// RegisterWorker registers a message worker
func (ec *EventCoordinator) RegisterWorker(queueName string, handler MessageHandler) (string, error) {
	if ec.queue == nil {
		return "", fmt.Errorf("message queue not enabled")
	}
	return ec.queue.RegisterWorker(queueName, handler)
}

// GetCircuitBreaker gets or creates a circuit breaker
func (ec *EventCoordinator) GetCircuitBreaker(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	return ec.breakers.GetOrCreate(name, config)
}

// GetEventStore returns the event store
func (ec *EventCoordinator) GetEventStore() *EventStore {
	return ec.store
}

// GetMetrics returns comprehensive metrics from all components
func (ec *EventCoordinator) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Bus metrics
	metrics["bus"] = ec.bus.GetMetrics()

	// Queue metrics
	if ec.queue != nil {
		metrics["queue"] = ec.queue.GetMetrics()
	}

	// Store metrics
	if ec.store != nil {
		metrics["store"] = map[string]interface{}{
			"event_count": ec.store.Count(),
		}
	}

	// Circuit breaker metrics
	breakerMetrics := make(map[string]interface{})
	for name, cb := range ec.breakers.GetAll() {
		breakerMetrics[name] = cb.GetMetrics()
	}
	metrics["circuit_breakers"] = breakerMetrics

	return metrics
}

// Shutdown gracefully shuts down all components
func (ec *EventCoordinator) Shutdown(ctx context.Context) error {
	ec.mu.Lock()
	if !ec.started {
		ec.mu.Unlock()
		return nil
	}
	ec.mu.Unlock()

	ec.logger.Info("shutting down event coordinator")

	// Shutdown components in order
	var shutdownErrors []error

	if ec.queue != nil {
		if err := ec.queue.Shutdown(ctx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("queue shutdown error: %w", err))
		}
	}

	if err := ec.bus.Shutdown(ctx); err != nil {
		shutdownErrors = append(shutdownErrors, fmt.Errorf("bus shutdown error: %w", err))
	}

	if ec.store != nil {
		if err := ec.store.Shutdown(ctx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("store shutdown error: %w", err))
		}
	}

	ec.mu.Lock()
	ec.started = false
	ec.mu.Unlock()

	if len(shutdownErrors) > 0 {
		return fmt.Errorf("shutdown errors: %v", shutdownErrors)
	}

	ec.logger.Info("event coordinator shutdown complete")
	return nil
}

// Helper methods for common operations

// EmitAgentEvent emits an agent-related event
func (ec *EventCoordinator) EmitAgentEvent(ctx context.Context, eventType EventType, data interface{}) error {
	event := NewEvent(eventType, "agent", data).WithTopic(TopicAgent)
	return ec.PublishEvent(ctx, event)
}

// EmitSessionEvent emits a session-related event
func (ec *EventCoordinator) EmitSessionEvent(ctx context.Context, eventType EventType, data SessionEventData) error {
	event := NewEvent(eventType, "session", data).WithTopic(TopicSession)
	return ec.PublishEvent(ctx, event)
}

// EmitTaskEvent emits a task-related event
func (ec *EventCoordinator) EmitTaskEvent(ctx context.Context, eventType EventType, data TaskEventData) error {
	event := NewEvent(eventType, "task", data).WithTopic(TopicTask)
	return ec.PublishEvent(ctx, event)
}

// EmitActivityEvent emits an activity-related event
func (ec *EventCoordinator) EmitActivityEvent(ctx context.Context, eventType EventType, data ActivityEventData) error {
	event := NewEvent(eventType, "activity", data).WithTopic(TopicActivity)
	return ec.PublishEvent(ctx, event)
}

// EmitToolEvent emits a tool-related event
func (ec *EventCoordinator) EmitToolEvent(ctx context.Context, eventType EventType, data ToolEventData) error {
	event := NewEvent(eventType, "tool", data).WithTopic(TopicTool)
	return ec.PublishEvent(ctx, event)
}

// EmitWorkflowEvent emits a workflow-related event
func (ec *EventCoordinator) EmitWorkflowEvent(ctx context.Context, eventType EventType, data WorkflowEventData) error {
	event := NewEvent(eventType, "workflow", data).WithTopic(TopicOrchestration)
	return ec.PublishEvent(ctx, event)
}
