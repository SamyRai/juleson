# Event System for Juleson

A comprehensive event-driven architecture providing pub/sub messaging, asynchronous task queues, event persistence, and fault tolerance for the Juleson automation platform.

## Architecture Overview

The event system consists of five main components:

1. **Event Bus** - Central pub/sub system for real-time event distribution
2. **Message Queue** - Asynchronous task processing with priority queues
3. **Event Store** - Event persistence for audit trails and replay
4. **Circuit Breaker** - Fault tolerance and cascading failure prevention
5. **Event Coordinator** - Unified interface coordinating all components

## Components

### 1. Event Bus

The Event Bus provides publish-subscribe messaging with:

- Topic-based routing
- Priority subscribers
- Synchronous and asynchronous delivery
- Middleware support
- Automatic metrics collection

**Features:**

- Subscribe to specific topics or all events (`TopicAll`)
- Filter events before processing
- Chain middleware for cross-cutting concerns
- Automatic subscriber ordering by priority

### 2. Message Queue

Asynchronous message processing with:

- Priority queues
- Multiple workers per queue
- Automatic retries with exponential backoff
- Dead letter queue (DLQ) for failed messages
- Backpressure control

**Features:**

- Create multiple named queues
- Register workers to process messages
- Automatic retry on failure
- Failed messages moved to DLQ after max retries

### 3. Event Store

Event persistence providing:

- Disk-based storage
- Event replay capabilities
- Time-range queries
- Automatic flushing

**Features:**

- Store events to disk for audit trails
- Replay events for debugging or recovery
- Query events by type, time range, or ID
- Automatic rotation and cleanup

### 4. Circuit Breaker

Fault tolerance with:

- Automatic failure detection
- Three states: Closed, Open, Half-Open
- Configurable thresholds
- Pool management for multiple breakers

**Features:**

- Prevent cascading failures
- Automatic recovery detection
- State change callbacks
- Per-component circuit breakers

### 5. Event Coordinator

Unified interface providing:

- Single entry point for all event operations
- Automatic integration between components
- Helper methods for common event types
- Centralized metrics

## Event Types

The system defines events for all major operations:

### Agent Events

- `agent.started` - Agent execution started
- `agent.stopped` - Agent stopped
- `agent.state_changed` - Agent state transition
- `agent.decision` - Agent made a decision
- `agent.progress` - Progress update

### Session Events

- `session.created` - Jules session created
- `session.updated` - Session updated
- `session.completed` - Session completed
- `session.failed` - Session failed

### Task Events

- `task.created` - Task created
- `task.started` - Task started
- `task.completed` - Task completed
- `task.failed` - Task failed

### Activity Events

- `activity.received` - Activity received from Jules
- `activity.processed` - Activity processed
- `plan.generated` - Plan generated
- `plan.approved` - Plan approved

### Tool Events

- `tool.invoked` - Tool invoked
- `tool.completed` - Tool completed
- `tool.failed` - Tool failed

### And more

## Usage Examples

### Basic Event Publishing

```go
import (
    "context"
    "github.com/SamyRai/juleson/internal/events"
)

// Create coordinator
coordinator, err := events.NewEventCoordinator(nil) // uses defaults
if err != nil {
    panic(err)
}

// Start coordinator
ctx := context.Background()
coordinator.Start(ctx)
defer coordinator.Shutdown(ctx)

// Publish an agent event
data := events.AgentStateChangedData{
    OldState: "IDLE",
    NewState: "ANALYZING",
    GoalID:   "goal-123",
}

err = coordinator.EmitAgentEvent(ctx, events.EventAgentStateChanged, data)
```

### Subscribing to Events

```go
// Subscribe to agent events
subscriber := events.Subscriber{
    ID:       "my-subscriber",
    Priority: 10,
    Async:    true,
    Handler: func(ctx context.Context, event events.Event) error {
        // Process event
        fmt.Printf("Received: %s\n", event.Type)
        return nil
    },
}

coordinator.Subscribe(events.TopicAgent, subscriber)
```

### Filtering Events

```go
// Subscribe only to state change events
subscriber := events.Subscriber{
    ID:       "state-watcher",
    Priority: 5,
    Filter: func(event events.Event) bool {
        return event.Type == events.EventAgentStateChanged
    },
    Handler: func(ctx context.Context, event events.Event) error {
        data := event.Data.(events.AgentStateChangedData)
        fmt.Printf("State: %s -> %s\n", data.OldState, data.NewState)
        return nil
    },
}

coordinator.Subscribe(events.TopicAgent, subscriber)
```

### Using Message Queue

```go
// Create a queue for background tasks
coordinator.CreateQueue("background-tasks", 1000)

// Register a worker
workerID, err := coordinator.RegisterWorker("background-tasks",
    func(ctx context.Context, msg events.Message) error {
        fmt.Printf("Processing: %s\n", msg.Type)
        // Do work...
        return nil
    },
)

// Enqueue a message
msg := events.Message{
    Type:    "process-session",
    Queue:   "background-tasks",
    Payload: sessionData,
    Metadata: map[string]interface{}{
        "priority": 5,
    },
}

coordinator.EnqueueMessage(msg)
```

### Using Circuit Breaker

```go
// Get circuit breaker for Jules API calls
cb := coordinator.GetCircuitBreaker("jules-api", nil)

// Execute with circuit breaker protection
err := cb.Execute(ctx, func(ctx context.Context) error {
    // Make Jules API call
    return julesClient.CreateSession(ctx, req)
})

if err != nil {
    // Circuit might be open
    state := cb.GetState()
    fmt.Printf("Circuit state: %s\n", state)
}
```

### Event Store Queries

```go
// Get event store
store := coordinator.GetEventStore()

// Get recent events
recentEvents := store.GetRecent(100)

// Get events by type
agentEvents := store.GetByType(events.EventAgentStateChanged)

// Get events in time range
start := time.Now().Add(-1 * time.Hour)
end := time.Now()
hourEvents := store.GetByTimeRange(start, end)

// Replay events
store.Replay(ctx, func(event events.Event) error {
    fmt.Printf("Replaying: %s\n", event.Type)
    return nil
})
```

### Middleware

```go
// Add custom middleware to event bus
coordinator.bus.Use(func(next events.EventHandler) events.EventHandler {
    return func(ctx context.Context, event events.Event) error {
        // Pre-processing
        fmt.Printf("Before: %s\n", event.Type)

        err := next(ctx, event)

        // Post-processing
        fmt.Printf("After: %s\n", event.Type)

        return err
    }
})
```

## Integration with Juleson Components

### Agent Integration

```go
// In agent code
func (a *CoreAgent) setState(newState agent.AgentState) {
    oldState := a.state
    a.state = newState

    // Emit state change event
    a.eventCoordinator.EmitAgentEvent(ctx, events.EventAgentStateChanged,
        events.AgentStateChangedData{
            OldState: string(oldState),
            NewState: string(newState),
            GoalID:   a.currentGoal.ID,
        },
    )
}
```

### Session Orchestrator Integration

```go
// In orchestrator
func (o *SessionOrchestrator) Start(ctx context.Context, sourceID string) error {
    // Emit workflow started event
    o.eventCoordinator.EmitWorkflowEvent(ctx, events.EventWorkflowStarted,
        events.WorkflowEventData{
            WorkflowName: o.workflow.Name,
            TotalPhases:  len(o.workflow.Phases),
        },
    )

    // ... rest of start logic
}
```

### Jules Client Integration

```go
// In Jules client
func (c *Client) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
    // Emit session creation event
    c.eventCoordinator.EmitSessionEvent(ctx, events.EventSessionCreated,
        events.SessionEventData{
            SessionID: session.ID,
            State:     session.State,
            Title:     session.Title,
        },
    )

    // ... create session
}
```

## Configuration

```go
config := &events.CoordinatorConfig{
    EventStoreConfig: &events.EventStoreConfig{
        StorageDir:    "./data/events",
        MaxEvents:     100000,
        AutoFlush:     true,
        FlushInterval: 10 * time.Second,
    },
    QueueConfig: &events.QueueConfig{
        MaxQueueSize:    10000,
        MaxRetries:      3,
        RetryDelay:      5 * time.Second,
        WorkerCount:     5,
        DLQMaxSize:      1000,
    },
    EnableStore: true,
    EnableQueue: true,
}

coordinator, err := events.NewEventCoordinator(config)
```

## Metrics

Get comprehensive metrics from all components:

```go
metrics := coordinator.GetMetrics()

// Bus metrics
busMetrics := metrics["bus"].(events.BusMetrics)
fmt.Printf("Events published: %d\n", busMetrics.EventsPublished)

// Queue metrics
queueMetrics := metrics["queue"].(events.QueueMetrics)
fmt.Printf("Messages processed: %d\n", queueMetrics.MessagesProcessed)

// Circuit breaker metrics
breakers := metrics["circuit_breakers"].(map[string]interface{})
```

## Best Practices

1. **Use Topics Wisely**: Subscribe to specific topics instead of `TopicAll` when possible
2. **Async for Heavy Work**: Use async subscribers for time-consuming operations
3. **Priority Matters**: Set priority to control event processing order
4. **Filter Early**: Use filters to reduce unnecessary processing
5. **Circuit Breakers**: Wrap external API calls in circuit breakers
6. **Monitor Metrics**: Regularly check metrics for performance issues
7. **DLQ Monitoring**: Monitor dead letter queue for failed messages
8. **Event Store Size**: Configure appropriate max events to manage disk usage

## Performance Considerations

- **Async Subscribers**: Don't block event processing
- **Filter Efficiency**: Filters should be fast
- **Queue Size**: Monitor queue sizes to prevent memory issues
- **Store Rotation**: Configure auto-flush to manage disk space
- **Circuit Breaker Tuning**: Adjust thresholds based on service behavior

## Error Handling

The system provides multiple layers of error handling:

1. **Event Bus**: Recovery middleware catches panics
2. **Message Queue**: Automatic retries with DLQ
3. **Circuit Breaker**: Prevents cascading failures
4. **Event Store**: Persists events even on errors

## Testing

```go
// Create test coordinator with in-memory storage
config := &events.CoordinatorConfig{
    EnableStore: false,  // Disable disk storage
    EnableQueue: true,
}

coordinator, _ := events.NewEventCoordinator(config)

// Test event publishing
err := coordinator.PublishEvent(ctx, testEvent)
assert.NoError(t, err)
```

## Migration Guide

To integrate the event system into existing code:

1. Add `EventCoordinator` to your service container
2. Emit events at key state transitions
3. Subscribe to events for cross-cutting concerns (logging, metrics, etc.)
4. Replace polling with event-driven updates
5. Use message queue for background tasks

## Roadmap

- [ ] WebSocket support for real-time client updates
- [ ] Event schema validation
- [ ] Event versioning
- [ ] Distributed event bus (Redis/NATS backend)
- [ ] GraphQL subscription support
- [ ] Event replay UI
- [ ] Advanced analytics on event streams

## License

This is part of the Juleson project. See main LICENSE file.
