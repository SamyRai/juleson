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

// PriorityQueue implements a priority queue for messages
type PriorityQueue struct {
	items    []QueueItem
	mu       sync.RWMutex
	notEmpty chan struct{}
	maxSize  int
}

// QueueItem represents an item in the queue
type QueueItem struct {
	Message   Message
	Priority  int
	EnqueueAt time.Time
	Attempts  int
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

// Worker processes messages from a queue
type Worker struct {
	ID       string
	Queue    string
	Handler  MessageHandler
	stopChan chan struct{}
	logger   *slog.Logger
}

// MessageHandler processes a message
type MessageHandler func(ctx context.Context, msg Message) error

// DeadLetterQueue stores messages that failed processing
type DeadLetterQueue struct {
	messages []DeadLetterMessage
	mu       sync.RWMutex
	maxSize  int
}

// DeadLetterMessage represents a failed message
type DeadLetterMessage struct {
	Message   Message
	Error     string
	FailedAt  time.Time
	Attempts  int
	LastError string
}

// QueueMetrics tracks queue metrics
type QueueMetrics struct {
	MessagesEnqueued  int64
	MessagesDequeued  int64
	MessagesProcessed int64
	MessagesFailed    int64
	DLQSize           int64
	mu                sync.RWMutex
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

	mq.metrics.mu.Lock()
	mq.metrics.MessagesEnqueued++
	mq.metrics.mu.Unlock()

	mq.logger.Debug("message enqueued",
		"message_id", msg.ID,
		"queue", msg.Queue,
		"type", msg.Type)

	return nil
}

// runWorker runs a worker that processes messages
func (mq *MessageQueue) runWorker(worker *Worker, queue *PriorityQueue) {
	defer mq.wg.Done()

	mq.logger.Info("worker started", "worker_id", worker.ID, "queue", worker.Queue)

	for {
		select {
		case <-worker.stopChan:
			mq.logger.Info("worker stopped", "worker_id", worker.ID)
			return
		case <-mq.stopChan:
			mq.logger.Info("worker stopped (queue shutdown)", "worker_id", worker.ID)
			return
		default:
			// Try to get message from queue
			item, ok := queue.Pop()
			if !ok {
				// Queue is empty, wait for notification
				select {
				case <-queue.notEmpty:
					continue
				case <-worker.stopChan:
					return
				case <-mq.stopChan:
					return
				case <-time.After(1 * time.Second):
					continue
				}
			}

			mq.processMessage(worker, item)
		}
	}
}

// processMessage processes a single message
func (mq *MessageQueue) processMessage(worker *Worker, item QueueItem) {
	msg := item.Message
	item.Attempts++

	mq.metrics.mu.Lock()
	mq.metrics.MessagesDequeued++
	mq.metrics.mu.Unlock()

	mq.logger.Debug("processing message",
		"worker_id", worker.ID,
		"message_id", msg.ID,
		"attempt", item.Attempts)

	// Create context with timeout
	ctx := context.Background()
	if msg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, msg.Timeout)
		defer cancel()
	}

	// Process message
	err := worker.Handler(ctx, msg)

	if err != nil {
		mq.logger.Error("message processing failed",
			"worker_id", worker.ID,
			"message_id", msg.ID,
			"error", err,
			"attempt", item.Attempts)

		// Check if should retry
		if item.Attempts < msg.MaxRetries {
			mq.logger.Info("retrying message",
				"message_id", msg.ID,
				"attempt", item.Attempts+1,
				"max_retries", msg.MaxRetries)

			// Re-enqueue with delay
			time.Sleep(mq.retryDelay)

			mq.mu.RLock()
			queue, exists := mq.queues[msg.Queue]
			mq.mu.RUnlock()

			if exists {
				queue.Push(item)
			}
		} else {
			// Max retries reached, send to DLQ
			mq.logger.Warn("message moved to DLQ",
				"message_id", msg.ID,
				"attempts", item.Attempts)

			mq.dlq.Add(DeadLetterMessage{
				Message:   msg,
				Error:     fmt.Sprintf("max retries (%d) exceeded", msg.MaxRetries),
				FailedAt:  time.Now(),
				Attempts:  item.Attempts,
				LastError: err.Error(),
			})

			mq.metrics.mu.Lock()
			mq.metrics.MessagesFailed++
			mq.metrics.DLQSize++
			mq.metrics.mu.Unlock()
		}
	} else {
		mq.metrics.mu.Lock()
		mq.metrics.MessagesProcessed++
		mq.metrics.mu.Unlock()

		mq.logger.Debug("message processed successfully",
			"worker_id", worker.ID,
			"message_id", msg.ID)
	}
}

// GetMetrics returns current queue metrics
func (mq *MessageQueue) GetMetrics() QueueMetrics {
	mq.metrics.mu.RLock()
	defer mq.metrics.mu.RUnlock()

	return QueueMetrics{
		MessagesEnqueued:  mq.metrics.MessagesEnqueued,
		MessagesDequeued:  mq.metrics.MessagesDequeued,
		MessagesProcessed: mq.metrics.MessagesProcessed,
		MessagesFailed:    mq.metrics.MessagesFailed,
		DLQSize:           mq.metrics.DLQSize,
	}
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

// PriorityQueue implementation

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue(maxSize int) *PriorityQueue {
	return &PriorityQueue{
		items:    make([]QueueItem, 0),
		notEmpty: make(chan struct{}, 1),
		maxSize:  maxSize,
	}
}

// Push adds an item to the queue
func (pq *PriorityQueue) Push(item QueueItem) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.maxSize > 0 && len(pq.items) >= pq.maxSize {
		return fmt.Errorf("queue is full (max size: %d)", pq.maxSize)
	}

	pq.items = append(pq.items, item)

	// Sort by priority (higher first)
	for i := len(pq.items) - 1; i > 0; i-- {
		if pq.items[i].Priority > pq.items[i-1].Priority {
			pq.items[i], pq.items[i-1] = pq.items[i-1], pq.items[i]
		} else {
			break
		}
	}

	// Notify waiting consumers
	select {
	case pq.notEmpty <- struct{}{}:
	default:
	}

	return nil
}

// Pop removes and returns the highest priority item
func (pq *PriorityQueue) Pop() (QueueItem, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.items) == 0 {
		return QueueItem{}, false
	}

	item := pq.items[0]
	pq.items = pq.items[1:]

	return item, true
}

// Size returns the current size of the queue
func (pq *PriorityQueue) Size() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items)
}

// DeadLetterQueue implementation

// NewDeadLetterQueue creates a new dead letter queue
func NewDeadLetterQueue(maxSize int) *DeadLetterQueue {
	return &DeadLetterQueue{
		messages: make([]DeadLetterMessage, 0),
		maxSize:  maxSize,
	}
}

// Add adds a message to the DLQ
func (dlq *DeadLetterQueue) Add(msg DeadLetterMessage) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	if dlq.maxSize > 0 && len(dlq.messages) >= dlq.maxSize {
		// Remove oldest message
		dlq.messages = dlq.messages[1:]
	}

	dlq.messages = append(dlq.messages, msg)
}

// GetAll returns all messages in the DLQ
func (dlq *DeadLetterQueue) GetAll() []DeadLetterMessage {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	result := make([]DeadLetterMessage, len(dlq.messages))
	copy(result, dlq.messages)
	return result
}

// Clear clears all messages from the DLQ
func (dlq *DeadLetterQueue) Clear() {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	dlq.messages = make([]DeadLetterMessage, 0)
}
