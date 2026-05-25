# Event System Quick Start

The event system is initialized through the service container.

```go
container := services.NewContainer(cfg)
defer container.Close()

coordinator := container.EventCoordinator()
```

## Publish An Event

```go
event := events.NewEvent(events.EventSessionCreated, "session", map[string]any{
    "session_id": "session-123",
})

if err := coordinator.Publish(ctx, event); err != nil {
    return err
}
```

## Subscribe

```go
subscriber := events.NewSubscriber("logger", func(ctx context.Context, event events.Event) error {
    log.Printf("event=%s id=%s", event.Type, event.ID)
    return nil
})

coordinator.Subscribe(events.EventSessionCreated, subscriber)
```

## Queue Work

Use queues for background work that should be retried or prioritized. Keep
handlers idempotent where possible.

## Store Events

The event store writes JSON files under `./data/events/`. Use `t.TempDir()` in
tests and avoid relying on file names as a public API.

See [Event System Architecture](EVENT_SYSTEM_ARCHITECTURE.md) for component details.
