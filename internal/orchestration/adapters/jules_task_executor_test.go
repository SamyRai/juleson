package adapters

import (
	"context"
	"strings"
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

func TestJulesTaskExecutorAddsContextAndGuidelinesToPrompt(t *testing.T) {
	gateway := &recordingSessionGateway{
		created: &domain.Session{ID: "session-1", URL: "https://jules.example/session-1"},
	}
	executor := NewJulesTaskExecutor(gateway, nil)

	_, err := executor.ExecuteTask(context.Background(), domain.Task{
		ID:          "task-1",
		Name:        "Improve coverage",
		Description: "Add tests",
		Prompt:      "Cover the session apply workflow.",
		Type:        "testing",
		Priority:    domain.PriorityHigh,
		Context:     map[string]string{"focus": "patch application"},
	}, domain.ExecutionContext{
		Goal: domain.Goal{
			Description: "Improve test confidence",
			Constraints: []string{
				"Keep public CLI behavior unchanged",
			},
			Context: domain.GoalContext{
				SourceID:   "source-1",
				Repository: "acme/widgets",
				Branch:     "main",
			},
		},
		Project: &domain.ProjectContext{
			ProjectPath:  "/repo",
			ProjectName:  "widgets",
			ProjectType:  "go-cli",
			Languages:    []string{"Go"},
			Architecture: "CLI and MCP server",
			GitStatus:    "clean",
		},
		Completed: []domain.TaskResult{{
			TaskID:   "analysis",
			TaskName: "Analyze tests",
			Success:  true,
			Output:   "found gaps",
		}},
	})
	if err != nil {
		t.Fatalf("ExecuteTask() error = %v", err)
	}

	for _, want := range []string{
		"Cover the session apply workflow.",
		"## Juleson Context",
		"- Goal: Improve test confidence",
		"- Constraints: Keep public CLI behavior unchanged",
		"- Repository: acme/widgets",
		"- Project name: widgets",
		"- Architecture: CLI and MCP server",
		"- Task context: focus=patch application",
		"- Analyze tests: succeeded - found gaps",
		"## Engineering Guidelines",
		"Make the smallest correct change that satisfies the goal.",
		"Run the relevant format, lint, or test commands when possible",
	} {
		if !strings.Contains(gateway.request.Prompt, want) {
			t.Fatalf("created prompt missing %q:\n%s", want, gateway.request.Prompt)
		}
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
