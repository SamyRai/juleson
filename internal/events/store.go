package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EventStore provides event persistence and replay capabilities.
// It stores events on disk for audit trails, debugging, and system recovery.
type EventStore struct {
	events        []StoredEvent
	mu            sync.RWMutex
	logger        *slog.Logger
	storageDir    string
	maxEvents     int
	autoFlush     bool
	flushInterval time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// StoredEvent represents an event stored in the event store
type StoredEvent struct {
	Event
	StoredAt time.Time `json:"stored_at"`
	Sequence int64     `json:"sequence"`
}

// EventStoreConfig configures the event store
type EventStoreConfig struct {
	StorageDir    string
	MaxEvents     int
	AutoFlush     bool
	FlushInterval time.Duration
}

// DefaultEventStoreConfig returns default configuration
func DefaultEventStoreConfig() *EventStoreConfig {
	return &EventStoreConfig{
		StorageDir:    "./data/events",
		MaxEvents:     100000,
		AutoFlush:     true,
		FlushInterval: 10 * time.Second,
	}
}

// NewEventStore creates a new event store
func NewEventStore(config *EventStoreConfig, logger *slog.Logger) (*EventStore, error) {
	if config == nil {
		config = DefaultEventStoreConfig()
	}
	if logger == nil {
		logger = slog.Default()
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(config.StorageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	es := &EventStore{
		events:        make([]StoredEvent, 0),
		logger:        logger,
		storageDir:    config.StorageDir,
		maxEvents:     config.MaxEvents,
		autoFlush:     config.AutoFlush,
		flushInterval: config.FlushInterval,
		stopChan:      make(chan struct{}),
	}

	// Load existing events
	if err := es.load(); err != nil {
		logger.Warn("failed to load existing events", "error", err)
	}

	// Start auto-flush if enabled
	if es.autoFlush {
		es.wg.Add(1)
		go es.autoFlushLoop()
	}

	return es, nil
}

// Store stores an event in the event store
func (es *EventStore) Store(event Event) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	storedEvent := StoredEvent{
		Event:    event,
		StoredAt: time.Now(),
		Sequence: int64(len(es.events) + 1),
	}

	es.events = append(es.events, storedEvent)

	// Trim if exceeds max
	if es.maxEvents > 0 && len(es.events) > es.maxEvents {
		// Keep most recent events
		es.events = es.events[len(es.events)-es.maxEvents:]
		// Renumber sequences
		for i := range es.events {
			es.events[i].Sequence = int64(i + 1)
		}
	}

	es.logger.Debug("event stored",
		"event_id", event.ID,
		"event_type", event.Type,
		"sequence", storedEvent.Sequence)

	return nil
}

// Get retrieves an event by ID
func (es *EventStore) Get(eventID string) (*StoredEvent, error) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for i := range es.events {
		if es.events[i].ID == eventID {
			return &es.events[i], nil
		}
	}

	return nil, fmt.Errorf("event not found: %s", eventID)
}

// GetByType retrieves all events of a specific type
func (es *EventStore) GetByType(eventType EventType) []StoredEvent {
	es.mu.RLock()
	defer es.mu.RUnlock()

	result := make([]StoredEvent, 0)
	for _, event := range es.events {
		if event.Type == eventType {
			result = append(result, event)
		}
	}

	return result
}

// GetByTimeRange retrieves events within a time range
func (es *EventStore) GetByTimeRange(start, end time.Time) []StoredEvent {
	es.mu.RLock()
	defer es.mu.RUnlock()

	result := make([]StoredEvent, 0)
	for _, event := range es.events {
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			result = append(result, event)
		}
	}

	return result
}

// GetRecent retrieves the most recent N events
func (es *EventStore) GetRecent(count int) []StoredEvent {
	es.mu.RLock()
	defer es.mu.RUnlock()

	if count <= 0 || count > len(es.events) {
		count = len(es.events)
	}

	result := make([]StoredEvent, count)
	start := len(es.events) - count
	copy(result, es.events[start:])

	return result
}

// Replay replays events by calling a handler for each event
func (es *EventStore) Replay(ctx context.Context, handler func(Event) error) error {
	es.mu.RLock()
	eventsCopy := make([]StoredEvent, len(es.events))
	copy(eventsCopy, es.events)
	es.mu.RUnlock()

	es.logger.Info("replaying events", "count", len(eventsCopy))

	for i, storedEvent := range eventsCopy {
		select {
		case <-ctx.Done():
			return fmt.Errorf("replay cancelled: %w", ctx.Err())
		default:
		}

		if err := handler(storedEvent.Event); err != nil {
			return fmt.Errorf("replay failed at event %d: %w", i, err)
		}
	}

	es.logger.Info("replay complete", "events_replayed", len(eventsCopy))
	return nil
}

// Flush writes events to disk
func (es *EventStore) Flush() error {
	es.mu.RLock()
	eventsCopy := make([]StoredEvent, len(es.events))
	copy(eventsCopy, es.events)
	es.mu.RUnlock()

	if len(eventsCopy) == 0 {
		return nil
	}

	filename := filepath.Join(es.storageDir, fmt.Sprintf("events_%s.json", time.Now().Format("20060102_150405")))

	data, err := json.MarshalIndent(eventsCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write events to file: %w", err)
	}

	es.logger.Info("events flushed to disk",
		"filename", filename,
		"event_count", len(eventsCopy))

	return nil
}

// load loads events from the most recent file
func (es *EventStore) load() error {
	files, err := filepath.Glob(filepath.Join(es.storageDir, "events_*.json"))
	if err != nil {
		return fmt.Errorf("failed to list event files: %w", err)
	}

	if len(files) == 0 {
		es.logger.Info("no existing events found")
		return nil
	}

	// Load most recent file
	mostRecent := files[len(files)-1]

	data, err := os.ReadFile(mostRecent)
	if err != nil {
		return fmt.Errorf("failed to read event file: %w", err)
	}

	var events []StoredEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return fmt.Errorf("failed to unmarshal events: %w", err)
	}

	es.mu.Lock()
	es.events = events
	es.mu.Unlock()

	es.logger.Info("events loaded from disk",
		"filename", mostRecent,
		"event_count", len(events))

	return nil
}

// autoFlushLoop periodically flushes events to disk
func (es *EventStore) autoFlushLoop() {
	defer es.wg.Done()

	ticker := time.NewTicker(es.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := es.Flush(); err != nil {
				es.logger.Error("auto-flush failed", "error", err)
			}
		case <-es.stopChan:
			// Final flush before shutdown
			if err := es.Flush(); err != nil {
				es.logger.Error("final flush failed", "error", err)
			}
			return
		}
	}
}

// Clear clears all events from the store
func (es *EventStore) Clear() {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.events = make([]StoredEvent, 0)
	es.logger.Info("event store cleared")
}

// Count returns the total number of events in the store
func (es *EventStore) Count() int {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return len(es.events)
}

// Shutdown gracefully shuts down the event store
func (es *EventStore) Shutdown(ctx context.Context) error {
	close(es.stopChan)

	// Wait for auto-flush to complete
	done := make(chan struct{})
	go func() {
		es.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		es.logger.Info("event store shutdown complete")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}
