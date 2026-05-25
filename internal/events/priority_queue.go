package events

import (
	"fmt"
	"sync"
	"time"
)

// PriorityQueue implements a priority queue for messages.
type PriorityQueue struct {
	items    []QueueItem
	mu       sync.RWMutex
	notEmpty chan struct{}
	maxSize  int
}

// QueueItem represents an item in the queue.
type QueueItem struct {
	Message   Message
	Priority  int
	EnqueueAt time.Time
	Attempts  int
}

// NewPriorityQueue creates a new priority queue.
func NewPriorityQueue(maxSize int) *PriorityQueue {
	return &PriorityQueue{
		items:    make([]QueueItem, 0),
		notEmpty: make(chan struct{}, 1),
		maxSize:  maxSize,
	}
}

// Push adds an item to the queue.
func (pq *PriorityQueue) Push(item QueueItem) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.maxSize > 0 && len(pq.items) >= pq.maxSize {
		return fmt.Errorf("queue is full (max size: %d)", pq.maxSize)
	}

	pq.items = append(pq.items, item)
	for i := len(pq.items) - 1; i > 0; i-- {
		if pq.items[i].Priority > pq.items[i-1].Priority {
			pq.items[i], pq.items[i-1] = pq.items[i-1], pq.items[i]
		} else {
			break
		}
	}

	select {
	case pq.notEmpty <- struct{}{}:
	default:
	}
	return nil
}

// Pop removes and returns the highest priority item.
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

// Size returns the current size of the queue.
func (pq *PriorityQueue) Size() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items)
}
