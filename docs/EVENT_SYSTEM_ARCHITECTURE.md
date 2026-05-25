# Event System Architecture

The event system provides in-process pub/sub, message queues, event storage,
circuit breakers, and a coordinator used by Juleson services.

## Components

- `internal/events/bus.go`: topic-based event bus with middleware and delivery
  modes.
- `internal/events/queue.go`: priority queues with worker processing, retry, and
  dead-letter handling.
- `internal/events/store.go`: JSON event persistence and replay queries.
- `internal/events/circuit_breaker.go`: closed, open, and half-open circuit breaker states.
- `internal/events/coordinator.go`: setup and shared access to event components.
- `internal/events/types.go`: event names and payload structures.

## Integration

The service container creates the event coordinator:

```go
container := services.NewContainer(cfg)
coordinator := container.EventCoordinator()
defer container.Close()
```

Jules client operations emit session, activity, and plan events where supported by the client path.

## Queues

The default setup creates high, normal, and low priority queues. Use the
coordinator when code needs to enqueue background work.

## Storage

Events are stored as JSON files under `./data/events/` by default. Treat this as
local operational data, not as a stable public database format.

## Failure Handling

- Event handlers should return errors rather than panic.
- Queue workers retry according to queue settings.
- Circuit breakers should wrap external services that can fail repeatedly.
- Shutdown should call the container or coordinator close path to drain work.

## Testing

Prefer unit tests around event publication, subscriber filtering, retry behavior,
and circuit breaker transitions. Use temporary directories for event store tests.
