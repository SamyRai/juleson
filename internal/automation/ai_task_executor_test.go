package automation

import (
	"context"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
)

func TestAITaskExecutorSendsTaskAndBuildsCompletedTask(t *testing.T) {
	client := &fakeAITaskSessionClient{
		activities: []jules.Activity{
			{
				ID: "activity-1",
				Artifacts: []jules.Artifact{
					{ChangeSet: &jules.ChangeSet{GitPatch: &jules.GitPatch{UnidiffPatch: "diff"}}},
				},
			},
		},
	}
	executor := newAITaskExecutorWithServices(client, client, func() string { return "session-1" })
	executor.currentTime = func() time.Time { return time.Unix(100, 0) }
	executor.sleep = func(time.Duration) {}

	completed, err := executor.execute(context.Background(), PendingTask{
		Name:        "Refactor",
		Description: "Improve ownership",
		Prompt:      "do refactor",
	})

	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if client.sessionID != "session-1" {
		t.Fatalf("sessionID = %q, want session-1", client.sessionID)
	}
	if client.prompt != "do refactor" {
		t.Fatalf("prompt = %q, want do refactor", client.prompt)
	}
	if completed.Name != "Refactor" {
		t.Fatalf("Name = %q, want Refactor", completed.Name)
	}
	if completed.Result.ActivityID != "activity-1" {
		t.Fatalf("ActivityID = %q, want activity-1", completed.Result.ActivityID)
	}
	if len(completed.Artifacts) != 1 {
		t.Fatalf("Artifacts = %d, want 1", len(completed.Artifacts))
	}
}

type fakeAITaskSessionClient struct {
	sessionID  string
	prompt     string
	activities []jules.Activity
}

func (f *fakeAITaskSessionClient) SendMessage(_ context.Context, sessionID string, req *jules.SendMessageRequest) error {
	f.sessionID = sessionID
	f.prompt = req.Prompt
	return nil
}

func (f *fakeAITaskSessionClient) List(context.Context, string, *jules.ListActivitiesOptions) (*jules.ActivitiesResponse, error) {
	return &jules.ActivitiesResponse{Activities: f.activities}, nil
}
