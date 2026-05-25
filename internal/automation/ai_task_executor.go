package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/pkg/jules"
)

type aiTaskSessionClient interface {
	SendMessage(ctx context.Context, sessionID string, req *jules.SendMessageRequest) error
	ListActivities(ctx context.Context, sessionID string, pageSize int) ([]jules.Activity, error)
}

type aiTaskExecutor struct {
	client         aiTaskSessionClient
	sessionID      func() string
	currentTime    func() time.Time
	sleep          func(time.Duration)
	executionDelay time.Duration
}

func newAITaskExecutor(client aiTaskSessionClient, sessionID func() string) *aiTaskExecutor {
	return &aiTaskExecutor{
		client:         client,
		sessionID:      sessionID,
		currentTime:    time.Now,
		sleep:          time.Sleep,
		executionDelay: 5 * time.Second,
	}
}

func (e *aiTaskExecutor) execute(ctx context.Context, task PendingTask) (CompletedTask, error) {
	if err := e.client.SendMessage(ctx, e.sessionID(), &jules.SendMessageRequest{Prompt: task.Prompt}); err != nil {
		return CompletedTask{}, fmt.Errorf("failed to send task to session: %w", err)
	}

	e.sleep(e.executionDelay)

	activities, err := e.client.ListActivities(ctx, e.sessionID(), 5)
	if err != nil {
		return CompletedTask{}, fmt.Errorf("failed to list activities: %w", err)
	}

	result := AITaskResult{Success: true}
	completed := CompletedTask{
		Name:        task.Name,
		Description: task.Description,
		Result:      result,
		Timestamp:   e.currentTime(),
	}

	if len(activities) > 0 {
		latestActivity := activities[0]
		completed.Result.ActivityID = latestActivity.ID
		completed.Result.FilesChanged = extractFilesChanged(latestActivity)
		completed.Artifacts = latestActivity.Artifacts
	}

	return completed, nil
}
