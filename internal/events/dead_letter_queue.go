package events

import (
	"sync"
	"time"
)

// DeadLetterQueue stores messages that failed processing.
type DeadLetterQueue struct {
	messages []DeadLetterMessage
	mu       sync.RWMutex
	maxSize  int
}

// DeadLetterMessage represents a failed message.
type DeadLetterMessage struct {
	Message   Message
	Error     string
	FailedAt  time.Time
	Attempts  int
	LastError string
}

// NewDeadLetterQueue creates a new dead letter queue.
func NewDeadLetterQueue(maxSize int) *DeadLetterQueue {
	return &DeadLetterQueue{
		messages: make([]DeadLetterMessage, 0),
		maxSize:  maxSize,
	}
}

// Add adds a message to the DLQ.
func (dlq *DeadLetterQueue) Add(msg DeadLetterMessage) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	if dlq.maxSize > 0 && len(dlq.messages) >= dlq.maxSize {
		dlq.messages = dlq.messages[1:]
	}
	dlq.messages = append(dlq.messages, msg)
}

// GetAll returns all messages in the DLQ.
func (dlq *DeadLetterQueue) GetAll() []DeadLetterMessage {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	result := make([]DeadLetterMessage, len(dlq.messages))
	copy(result, dlq.messages)
	return result
}

// Clear clears all messages from the DLQ.
func (dlq *DeadLetterQueue) Clear() {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	dlq.messages = make([]DeadLetterMessage, 0)
}
