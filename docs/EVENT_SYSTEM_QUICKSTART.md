# Event System Quick Start

The event system is initialized through `internal/events`.

```go
coordinator, err := events.NewEventCoordinator(nil)
if err != nil {
    return err
}

ctx := context.Background()
if err := coordinator.Start(ctx); err != nil {
    return err
}
defer coordinator.Shutdown(ctx)
```

## Publish An Event

```go
event := events.NewEvent(events.EventSessionCreated, "session", map[string]any{
    "session_id": "session-123",
}).WithTopic(events.TopicSession)

if err := coordinator.PublishEvent(ctx, event); err != nil {
    return err
}
```

## Subscribe

```go
subscriber := events.Subscriber{
    ID: "logger",
    Handler: func(ctx context.Context, event events.Event) error {
        log.Printf("event=%s id=%s", event.Type, event.ID)
        return nil
    },
}

if err := coordinator.Subscribe(events.TopicSession, subscriber); err != nil {
    return err
}
```

Typed helpers are available for common domains:

```go
err := coordinator.EmitSessionEvent(ctx, events.EventSessionCreated, events.SessionEventData{
    SessionID: "session-123",
    State:     "created",
})
```

## Queue Work

Use queues for background work that should be retried or prioritized. Keep
handlers idempotent where possible.

```go
if err := coordinator.CreateQueue("work", 100); err != nil {
    return err
}

_, err := coordinator.RegisterWorker("work", func(ctx context.Context, msg events.Message) error {
    return nil
})
if err != nil {
    return err
}

return coordinator.EnqueueMessage(events.Message{
    Queue: "work",
    Type:  "example",
})
```

## Store Events

The event store writes JSON files under `./data/events/`. Use `t.TempDir()` in
tests and avoid relying on file names as a public API.

See [Event System Architecture](EVENT_SYSTEM_ARCHITECTURE.md) for component details.
