package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Worker processes messages from a queue.
type Worker struct {
	ID       string
	Queue    string
	Handler  MessageHandler
	stopChan chan struct{}
	logger   *slog.Logger
}

// MessageHandler processes a message.
type MessageHandler func(ctx context.Context, msg Message) error

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
			item, ok := queue.Pop()
			if !ok {
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

func (mq *MessageQueue) processMessage(worker *Worker, item QueueItem) {
	msg := item.Message
	item.Attempts++
	mq.metrics.recordDequeued()

	mq.logger.Debug("processing message",
		"worker_id", worker.ID,
		"message_id", msg.ID,
		"attempt", item.Attempts)

	ctx := context.Background()
	if msg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, msg.Timeout)
		defer cancel()
	}

	if err := worker.Handler(ctx, msg); err != nil {
		mq.handleMessageFailure(worker, item, err)
		return
	}

	mq.metrics.recordProcessed()
	mq.logger.Debug("message processed successfully",
		"worker_id", worker.ID,
		"message_id", msg.ID)
}

func (mq *MessageQueue) handleMessageFailure(worker *Worker, item QueueItem, err error) {
	msg := item.Message
	mq.logger.Error("message processing failed",
		"worker_id", worker.ID,
		"message_id", msg.ID,
		"error", err,
		"attempt", item.Attempts)

	if item.Attempts < msg.MaxRetries {
		mq.logger.Info("retrying message",
			"message_id", msg.ID,
			"attempt", item.Attempts+1,
			"max_retries", msg.MaxRetries)
		time.Sleep(mq.retryDelay)
		mq.mu.RLock()
		queue, exists := mq.queues[msg.Queue]
		mq.mu.RUnlock()
		if exists {
			_ = queue.Push(item)
		}
		return
	}

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
	mq.metrics.recordFailed()
}
