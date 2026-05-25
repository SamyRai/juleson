package events

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestPriorityQueueOrdersHighestPriorityFirst(t *testing.T) {
	queue := NewPriorityQueue(10)
	for _, item := range []QueueItem{
		{Message: Message{ID: "low"}, Priority: 1},
		{Message: Message{ID: "high"}, Priority: 10},
		{Message: Message{ID: "medium"}, Priority: 5},
	} {
		if err := queue.Push(item); err != nil {
			t.Fatalf("push: %v", err)
		}
	}

	for _, want := range []string{"high", "medium", "low"} {
		item, ok := queue.Pop()
		if !ok {
			t.Fatalf("pop %s: queue empty", want)
		}
		if item.Message.ID != want {
			t.Fatalf("message ID = %s, want %s", item.Message.ID, want)
		}
	}
}

func TestMessageQueueMovesFailedMessageToDLQAndTracksMetrics(t *testing.T) {
	queue := NewMessageQueue(&QueueConfig{
		MaxRetries:  1,
		RetryDelay:  1 * time.Millisecond,
		DLQMaxSize:  10,
		WorkerCount: 1,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := queue.CreateQueue("work", 10); err != nil {
		t.Fatalf("create queue: %v", err)
	}
	if _, err := queue.RegisterWorker("work", func(ctx context.Context, msg Message) error {
		return fmt.Errorf("failed")
	}); err != nil {
		t.Fatalf("register worker: %v", err)
	}

	if err := queue.Enqueue(Message{ID: "msg-1", Queue: "work"}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	eventually(t, func() bool {
		return len(queue.GetDLQMessages()) == 1
	})
	metrics := queue.GetMetrics()
	if metrics.MessagesEnqueued != 1 || metrics.MessagesDequeued != 1 || metrics.MessagesFailed != 1 || metrics.DLQSize != 1 {
		t.Fatalf("unexpected metrics: %#v", metrics)
	}

	if err := queue.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestMessageQueueShutdownStopsWorker(t *testing.T) {
	queue := NewMessageQueue(DefaultQueueConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := queue.CreateQueue("work", 10); err != nil {
		t.Fatalf("create queue: %v", err)
	}
	if _, err := queue.RegisterWorker("work", func(ctx context.Context, msg Message) error {
		return nil
	}); err != nil {
		t.Fatalf("register worker: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := queue.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func eventually(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition was not satisfied before timeout")
}
