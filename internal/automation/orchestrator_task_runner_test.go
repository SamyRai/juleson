package automation

import (
	"context"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
)

func TestOrchestrationTaskRunnerExecutesPlanGateAndRecordsActivity(t *testing.T) {
	client := &fakeOrchestrationSessionClient{
		activities: []jules.Activity{
			{
				ID:        "activity-1",
				Artifacts: []jules.Artifact{{BashOutput: &jules.BashOutput{Command: "go test ./..."}}},
			},
		},
	}
	monitor := &fakePlanWaiter{}
	states := make([]OrchestratorState, 0)
	progress := make([]ProgressUpdate, 0)
	records := make([]ExecutionRecord, 0)
	now := time.Unix(100, 0)

	runner := &orchestrationTaskRunner{
		sessions:      client,
		activities:    client,
		monitor:       monitor,
		sessionID:     "session-1",
		autoApprove:   false,
		waitAfterSend: DefaultTaskWaitTime,
		setState:      func(state OrchestratorState) { states = append(states, state) },
		sendProgress: func(phase, task int, message string, value float64) {
			progress = append(progress, ProgressUpdate{Phase: phase, Task: message, Progress: value})
		},
		addRecord:        func(record ExecutionRecord) { records = append(records, record) },
		currentTime:      func() time.Time { current := now; now = now.Add(time.Second); return current },
		sleepAfterSend:   func(time.Duration) {},
		activityPageSize: 10,
	}

	result, err := runner.execute(context.Background(), 2, 3, Task{
		Name:        "Implement",
		Prompt:      "do work",
		WaitForPlan: true,
		AutoApprove: true,
		Validation: func(result TaskResult) error {
			if result.ActivityID != "activity-1" {
				t.Fatalf("validation ActivityID = %q, want activity-1", result.ActivityID)
			}
			return nil
		},
	})

	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if !result.Success {
		t.Fatal("result.Success = false, want true")
	}
	if client.sentPrompt != "do work" {
		t.Fatalf("sentPrompt = %q, want do work", client.sentPrompt)
	}
	if !client.approved {
		t.Fatal("expected plan approval")
	}
	if !monitor.waited {
		t.Fatal("expected WaitForPlan")
	}
	if len(states) != 2 || states[0] != StateWaitingPlan || states[1] != StateRunning {
		t.Fatalf("states = %v, want [%s %s]", states, StateWaitingPlan, StateRunning)
	}
	if len(records) != 1 || !records[0].Success || records[0].ActivityID != "activity-1" {
		t.Fatalf("records = %+v, want one successful activity-1 record", records)
	}
	if len(progress) != 3 {
		t.Fatalf("progress updates = %d, want 3", len(progress))
	}
}

type fakeOrchestrationSessionClient struct {
	sentPrompt string
	approved   bool
	activities []jules.Activity
}

func (f *fakeOrchestrationSessionClient) SendMessage(_ context.Context, _ string, req *jules.SendMessageRequest) error {
	f.sentPrompt = req.Prompt
	return nil
}

func (f *fakeOrchestrationSessionClient) ApprovePlan(context.Context, string) error {
	f.approved = true
	return nil
}

func (f *fakeOrchestrationSessionClient) List(context.Context, string, *jules.ListActivitiesOptions) (*jules.ActivitiesResponse, error) {
	return &jules.ActivitiesResponse{Activities: f.activities}, nil
}

type fakePlanWaiter struct {
	waited bool
}

func (f *fakePlanWaiter) WaitForPlan(context.Context) (*jules.SessionStatus, error) {
	f.waited = true
	return &jules.SessionStatus{}, nil
}
