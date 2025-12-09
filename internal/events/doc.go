// Package events provides a comprehensive event-driven architecture for Juleson.
//
// The events package implements a sophisticated event system with the following components:
//
// 1. Event Bus - Central pub/sub system for real-time event distribution
// 2. Message Queue - Asynchronous task processing with priority and retries
// 3. Event Store - Event persistence for audit trails and replay capabilities
// 4. Circuit Breaker - Fault tolerance and cascading failure prevention
// 5. Event Coordinator - Unified interface coordinating all components
//
// # Architecture
//
// The event system replaces polling-based communication with event-driven patterns,
// enabling:
// - Real-time updates without polling
// - Decoupled components
// - Asynchronous task processing
// - Event sourcing and replay
// - Fault-tolerant operations
//
// # Quick Start
//
// Basic usage:
//
//	// Create coordinator
//	coordinator, err := events.NewEventCoordinator(nil)
//	if err != nil {
//	    panic(err)
//	}
//
//	// Start coordinator
//	ctx := context.Background()
//	coordinator.Start(ctx)
//	defer coordinator.Shutdown(ctx)
//
//	// Publish event
//	coordinator.EmitAgentEvent(ctx, events.EventAgentStateChanged, data)
//
//	// Subscribe to events
//	coordinator.Subscribe(events.TopicAgent, events.Subscriber{
//	    ID: "my-subscriber",
//	    Handler: func(ctx context.Context, event events.Event) error {
//	        fmt.Printf("Event: %s\n", event.Type)
//	        return nil
//	    },
//	})
//
// # Event Types
//
// The package defines events for all major operations:
// - Agent events (state changes, decisions, progress)
// - Session events (created, updated, completed, failed)
// - Task events (created, started, completed, failed)
// - Activity events (received, processed, plan generated)
// - Tool events (invoked, completed, failed)
// - Workflow events (started, phase changes, completed)
// - GitHub events (PR created, merged, closed)
//
// # Topics
//
// Events are organized into topics:
// - TopicAgent - Agent-related events
// - TopicSession - Jules session events
// - TopicTask - Task execution events
// - TopicActivity - Jules activity events
// - TopicTool - Tool invocation events
// - TopicReview - Code review events
// - TopicOrchestration - Workflow orchestration events
// - TopicGitHub - GitHub integration events
// - TopicAll - Subscribe to all events
//
// # Message Queue
//
// For asynchronous task processing:
//
//	// Create queue
//	coordinator.CreateQueue("background-tasks", 1000)
//
//	// Register worker
//	coordinator.RegisterWorker("background-tasks",
//	    func(ctx context.Context, msg events.Message) error {
//	        // Process message
//	        return nil
//	    },
//	)
//
//	// Enqueue message
//	coordinator.EnqueueMessage(events.Message{
//	    Type:    "process-session",
//	    Queue:   "background-tasks",
//	    Payload: data,
//	})
//
// # Circuit Breaker
//
// For fault-tolerant external API calls:
//
//	cb := coordinator.GetCircuitBreaker("jules-api", nil)
//
//	err := cb.Execute(ctx, func(ctx context.Context) error {
//	    return julesClient.CreateSession(ctx, req)
//	})
//
// # Event Store
//
// For event persistence and replay:
//
//	store := coordinator.GetEventStore()
//
//	// Get recent events
//	events := store.GetRecent(100)
//
//	// Replay events
//	store.Replay(ctx, func(event events.Event) error {
//	    // Process replayed event
//	    return nil
//	})
//
// # Middleware
//
// The event bus supports middleware for cross-cutting concerns:
// - LoggingMiddleware - Logs all event processing
// - RetryMiddleware - Retries failed handlers
// - TimeoutMiddleware - Adds timeout to processing
// - RecoveryMiddleware - Recovers from panics
// - FilterMiddleware - Filters events
// - DeduplicationMiddleware - Prevents duplicate processing
//
// # Integration
//
// The event system integrates with Juleson components:
//
//	// In agent
//	a.eventCoordinator.EmitAgentEvent(ctx, events.EventAgentStateChanged, data)
//
//	// In session orchestrator
//	o.eventCoordinator.EmitSessionEvent(ctx, events.EventSessionCreated, data)
//
//	// In Jules client
//	c.eventCoordinator.EmitActivityEvent(ctx, events.EventActivityReceived, data)
//
// # Metrics
//
// Get comprehensive metrics from all components:
//
//	metrics := coordinator.GetMetrics()
//	fmt.Printf("Events published: %d\n", metrics["bus"].EventsPublished)
//
// See README.md for detailed documentation and examples.
package events
