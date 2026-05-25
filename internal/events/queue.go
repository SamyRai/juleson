package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// MessageQueue provides asynchronous task processing with priority queues,
// retry logic, dead letter queues, and backpressure control.
type MessageQueue struct {
	queues     map[string]*PriorityQueue
	workers    map[string][]*Worker
	dlq        *DeadLetterQueue
	mu         sync.RWMutex
	logger     *slog.Logger
	metrics    *QueueMetrics
	maxRetries int
	retryDelay time.Duration
	stopping   bool
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// Message represents a message in the queue
type Message struct {
	ID         string
	Type       string
	Queue      string
	Payload    interface{}
	Timestamp  time.Time
	Metadata   map[string]interface{}
	Retries    int
	MaxRetries int
	Timeout    time.Duration
}

// QueueConfig configures the message queue
type QueueConfig struct {
	MaxQueueSize int
	MaxRetries   int
	RetryDelay   time.Duration
	WorkerCount  int
	DLQMaxSize   int
}

// DefaultQueueConfig returns default queue configuration
func DefaultQueueConfig() *QueueConfig {
	return &QueueConfig{
		MaxQueueSize: 10000,
		MaxRetries:   3,
		RetryDelay:   5 * time.Second,
		WorkerCount:  5,
		DLQMaxSize:   1000,
	}
}

// NewMessageQueue creates a new message queue
func NewMessageQueue(config *QueueConfig, logger *slog.Logger) *MessageQueue {
	if config == nil {
		config = DefaultQueueConfig()
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &MessageQueue{
		queues:     make(map[string]*PriorityQueue),
		workers:    make(map[string][]*Worker),
		dlq:        NewDeadLetterQueue(config.DLQMaxSize),
		logger:     logger,
		metrics:    &QueueMetrics{},
		maxRetries: config.MaxRetries,
		retryDelay: config.RetryDelay,
		stopChan:   make(chan struct{}),
	}
}

// CreateQueue creates a new queue with the specified configuration
func (mq *MessageQueue) CreateQueue(name string, maxSize int) error {
	if name == "" {
		return fmt.Errorf("queue name cannot be empty")
	}

	mq.mu.Lock()
	defer mq.mu.Unlock()

	if _, exists := mq.queues[name]; exists {
		return fmt.Errorf("queue %s already exists", name)
	}

	mq.queues[name] = NewPriorityQueue(maxSize)
	mq.workers[name] = make([]*Worker, 0)

	mq.logger.Info("queue created", "queue", name, "max_size", maxSize)
	return nil
}

// RegisterWorker registers a worker for a queue
func (mq *MessageQueue) RegisterWorker(queueName string, handler MessageHandler) (string, error) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	queue, exists := mq.queues[queueName]
	if !exists {
		return "", fmt.Errorf("queue %s not found", queueName)
	}

	workerID := fmt.Sprintf("worker-%s-%d", queueName, len(mq.workers[queueName])+1)
	worker := &Worker{
		ID:       workerID,
		Queue:    queueName,
		Handler:  handler,
		stopChan: make(chan struct{}),
		logger:   mq.logger,
	}

	mq.workers[queueName] = append(mq.workers[queueName], worker)

	// Start worker
	mq.wg.Add(1)
	go mq.runWorker(worker, queue)

	mq.logger.Info("worker registered", "worker_id", workerID, "queue", queueName)
	return workerID, nil
}

// Enqueue adds a message to a queue
func (mq *MessageQueue) Enqueue(msg Message) error {
	if msg.Queue == "" {
		return fmt.Errorf("message queue name cannot be empty")
	}

	mq.mu.RLock()
	queue, exists := mq.queues[msg.Queue]
	mq.mu.RUnlock()

	if !exists {
		return fmt.Errorf("queue %s not found", msg.Queue)
	}

	// Set defaults
	if msg.ID == "" {
		msg.ID = generateEventID()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	if msg.MaxRetries == 0 {
		msg.MaxRetries = mq.maxRetries
	}

	priority := 0
	if meta, ok := msg.Metadata["priority"].(int); ok {
		priority = meta
	}

	item := QueueItem{
		Message:   msg,
		Priority:  priority,
		EnqueueAt: time.Now(),
		Attempts:  0,
	}

	if err := queue.Push(item); err != nil {
		return fmt.Errorf("failed to enqueue message: %w", err)
	}

	mq.metrics.recordEnqueued()

	mq.logger.Debug("message enqueued",
		"message_id", msg.ID,
		"queue", msg.Queue,
		"type", msg.Type)

	return nil
}

// GetQueueSize returns the size of a queue
func (mq *MessageQueue) GetQueueSize(queueName string) int {
	mq.mu.RLock()
	queue, exists := mq.queues[queueName]
	mq.mu.RUnlock()

	if !exists {
		return 0
	}

	return queue.Size()
}

// GetDLQMessages returns messages from the dead letter queue
func (mq *MessageQueue) GetDLQMessages() []DeadLetterMessage {
	return mq.dlq.GetAll()
}

// Shutdown gracefully shuts down the message queue
func (mq *MessageQueue) Shutdown(ctx context.Context) error {
	mq.mu.Lock()
	mq.stopping = true
	close(mq.stopChan)
	mq.mu.Unlock()

	// Wait for all workers to finish
	done := make(chan struct{})
	go func() {
		mq.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		mq.logger.Info("message queue shutdown complete")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}
