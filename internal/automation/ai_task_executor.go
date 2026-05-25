package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/go-jules"
)

type aiTaskSessionClient interface {
	SendMessage(ctx context.Context, sessionID string, req *jules.SendMessageRequest) error
}

type aiTaskActivityClient interface {
	List(ctx context.Context, sessionID string, options *jules.ListActivitiesOptions) (*jules.ActivitiesResponse, error)
}

type aiTaskExecutor struct {
	sessions       aiTaskSessionClient
	activities     aiTaskActivityClient
	sessionID      func() string
	currentTime    func() time.Time
	sleep          func(time.Duration)
	executionDelay time.Duration
}

func newAITaskExecutor(client *jules.Client, sessionID func() string) *aiTaskExecutor {
	return newAITaskExecutorWithServices(client.Sessions(), client.Activities(), sessionID)
}

func newAITaskExecutorWithServices(sessions aiTaskSessionClient, activities aiTaskActivityClient, sessionID func() string) *aiTaskExecutor {
	return &aiTaskExecutor{
		sessions:       sessions,
		activities:     activities,
		sessionID:      sessionID,
		currentTime:    time.Now,
		sleep:          time.Sleep,
		executionDelay: 5 * time.Second,
	}
}

func (e *aiTaskExecutor) execute(ctx context.Context, task PendingTask) (CompletedTask, error) {
	if err := e.sessions.SendMessage(ctx, e.sessionID(), &jules.SendMessageRequest{Prompt: task.Prompt}); err != nil {
		return CompletedTask{}, fmt.Errorf("failed to send task to session: %w", err)
	}

	e.sleep(e.executionDelay)

	response, err := e.activities.List(ctx, e.sessionID(), &jules.ListActivitiesOptions{PageSize: 5})
	if err != nil {
		return CompletedTask{}, fmt.Errorf("failed to list activities: %w", err)
	}
	if response == nil {
		return CompletedTask{}, fmt.Errorf("failed to list activities: empty response")
	}
	activities := response.Activities

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
