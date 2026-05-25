package adapters

import (
	"context"
	"testing"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

func TestJulesTaskExecutorDryRunDoesNotTouchGateway(t *testing.T) {
	gateway := &recordingSessionGateway{}
	executor := NewJulesTaskExecutor(gateway, nil)

	result, err := executor.ExecuteTask(context.Background(), domain.Task{
		ID:          "task-1",
		Name:        "Task",
		Description: "Do work",
	}, domain.ExecutionContext{DryRun: true})
	if err != nil {
		t.Fatalf("ExecuteTask() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("result success = false: %+v", result)
	}
	if result.Metrics["dry_run"] != true {
		t.Fatalf("dry_run metric = %v, want true", result.Metrics["dry_run"])
	}
	if result.Metrics["session_action"] != "dry_run" {
		t.Fatalf("session action = %v, want dry_run", result.Metrics["session_action"])
	}
	if gateway.calls != 0 {
		t.Fatalf("gateway calls = %d, want 0", gateway.calls)
	}
}

func TestJulesTaskExecutorRequiresPlanApprovalForCreatedSessions(t *testing.T) {
	gateway := &recordingSessionGateway{
		created: &domain.Session{ID: "session-1", URL: "https://jules.example/session-1"},
	}
	executor := NewJulesTaskExecutor(gateway, nil)

	result, err := executor.ExecuteTask(context.Background(), domain.Task{
		ID:          "task-1",
		Name:        "Task",
		Description: "Do work",
	}, domain.ExecutionContext{
		Goal: domain.Goal{Context: domain.GoalContext{SourceID: "source-1"}},
	})
	if err != nil {
		t.Fatalf("ExecuteTask() error = %v", err)
	}
	if !result.Success || result.SessionID != "session-1" {
		t.Fatalf("result = %+v", result)
	}
	if !gateway.request.RequirePlanApproval {
		t.Fatal("created session did not require plan approval")
	}
	if result.Metrics["require_plan_approval"] != true {
		t.Fatalf("require approval metric = %v, want true", result.Metrics["require_plan_approval"])
	}
	if result.Metrics["session_action"] != "created" {
		t.Fatalf("session action = %v, want created", result.Metrics["session_action"])
	}
}

func TestJulesTaskExecutorHonorsAutoApprovePolicy(t *testing.T) {
	gateway := &recordingSessionGateway{
		created: &domain.Session{ID: "session-1", URL: "https://jules.example/session-1"},
	}
	executor := NewJulesTaskExecutor(gateway, nil)

	result, err := executor.ExecuteTask(context.Background(), domain.Task{
		ID:          "task-1",
		Name:        "Task",
		Description: "Do work",
	}, domain.ExecutionContext{
		Goal: domain.Goal{Context: domain.GoalContext{SourceID: "source-1"}},
		ApprovalPolicy: domain.ApprovalPolicy{
			AutoApprove: true,
		},
	})
	if err != nil {
		t.Fatalf("ExecuteTask() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("result success = false: %+v", result)
	}
	if gateway.request.RequirePlanApproval {
		t.Fatal("created session required plan approval despite auto-approve policy")
	}
	if result.Metrics["require_plan_approval"] != false {
		t.Fatalf("require approval metric = %v, want false", result.Metrics["require_plan_approval"])
	}
}

type recordingSessionGateway struct {
	calls   int
	request domain.SessionRequest
	created *domain.Session
}

func (g *recordingSessionGateway) ListSources(ctx context.Context, limit int) ([]domain.Source, error) {
	g.calls++
	return []domain.Source{{ID: "source-1", Name: "source-1"}}, nil
}

func (g *recordingSessionGateway) FindReusableSession(ctx context.Context, title string) (*domain.Session, error) {
	g.calls++
	return nil, nil
}

func (g *recordingSessionGateway) CreateSession(ctx context.Context, request domain.SessionRequest) (*domain.Session, error) {
	g.calls++
	g.request = request
	if g.created != nil {
		return g.created, nil
	}
	return &domain.Session{ID: "session-1", URL: "https://jules.example/session-1"}, nil
}

func (g *recordingSessionGateway) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	g.calls++
	return &domain.Session{ID: sessionID}, nil
}
