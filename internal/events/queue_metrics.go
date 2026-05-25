package events

import "sync"

// QueueMetrics tracks queue metrics.
type QueueMetrics struct {
	MessagesEnqueued  int64
	MessagesDequeued  int64
	MessagesProcessed int64
	MessagesFailed    int64
	DLQSize           int64
	mu                sync.RWMutex
}

// GetMetrics returns current queue metrics.
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

func (m *QueueMetrics) recordEnqueued() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesEnqueued++
}

func (m *QueueMetrics) recordDequeued() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesDequeued++
}

func (m *QueueMetrics) recordProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesProcessed++
}

func (m *QueueMetrics) recordFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesFailed++
	m.DLQSize++
}
