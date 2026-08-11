package events

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeadLetterQueue_AddAndGetAll(t *testing.T) {
	dlq := NewDeadLetterQueue(5)

	msg1 := DeadLetterMessage{
		Message:  Message{ID: "msg-1"},
		Error:    "timeout",
		FailedAt: time.Now(),
	}

	dlq.Add(msg1)

	all := dlq.GetAll()
	assert.Len(t, all, 1)
	assert.Equal(t, "msg-1", all[0].Message.ID)
}

func TestDeadLetterQueue_MaxSizeEviction(t *testing.T) {
	dlq := NewDeadLetterQueue(3)

	for i := 0; i < 5; i++ {
		dlq.Add(DeadLetterMessage{
			Message: Message{ID: string(rune('A' + i))}, // A, B, C, D, E
		})
	}

	all := dlq.GetAll()
	assert.Len(t, all, 3)
	// Should have evicted A and B, keeping C, D, E
	assert.Equal(t, "C", all[0].Message.ID)
	assert.Equal(t, "D", all[1].Message.ID)
	assert.Equal(t, "E", all[2].Message.ID)
}

func TestDeadLetterQueue_Clear(t *testing.T) {
	dlq := NewDeadLetterQueue(5)

	dlq.Add(DeadLetterMessage{Message: Message{ID: "msg-1"}})
	dlq.Add(DeadLetterMessage{Message: Message{ID: "msg-2"}})

	assert.Len(t, dlq.GetAll(), 2)

	dlq.Clear()

	assert.Len(t, dlq.GetAll(), 0)
}

func TestDeadLetterQueue_Concurrency(t *testing.T) {
	dlq := NewDeadLetterQueue(100)

	var wg sync.WaitGroup
	workers := 50

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				dlq.Add(DeadLetterMessage{
					Message: Message{ID: "concurrent"},
				})
			}
		}(i)
	}

	wg.Wait()

	// Max size is 100, so we expect exactly 100 messages,
	// even though 250 were added concurrently.
	assert.Len(t, dlq.GetAll(), 100)
}
