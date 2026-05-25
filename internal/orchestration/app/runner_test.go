package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

func TestAgentRunnerStateTransitions(t *testing.T) {
	goal := domain.Goal{
		ID:          "goal-1",
		Description: "ship change",
		Context:     domain.GoalContext{ProjectPath: "/repo"},
	}
	progress := &recordingProgressSink{}
	runner := NewAgentRunner(AgentRunnerDeps{
		ProjectAnalyzer: fakeAnalyzer{},
		Planner: fakePlanner{plan: &domain.Plan{
			ID:   "plan-1",
			Goal: goal,
			Tasks: []domain.Task{
				{ID: "one", Name: "One"},
				{ID: "two", Name: "Two", Dependencies: []string{"one"}},
			},
		}},
		TaskExecutor:  fakeTaskExecutor{},
		Reviewer:      approvingReviewer{},
		ProgressSink:  progress,
		Clock:         fixedClock{now: time.Unix(100, 0)},
		MaxIterations: 4,
	})

	result, err := runner.Run(context.Background(), goal)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !result.Success || result.State != domain.StateComplete {
		t.Fatalf("result success/state = %v/%s", result.Success, result.State)
	}
	if len(result.Tasks) != 2 {
		t.Fatalf("tasks executed = %d, want 2", len(result.Tasks))
	}
	if progress.last.State != domain.StateComplete || progress.last.Progress != 100 {
		t.Fatalf("last progress = %+v", progress.last)
	}
}

func TestAgentRunnerMaxIterations(t *testing.T) {
	goal := domain.Goal{ID: "goal-1", Description: "ship change"}
	runner := NewAgentRunner(AgentRunnerDeps{
		Planner: fakePlanner{plan: &domain.Plan{
			ID:   "plan-1",
			Goal: goal,
			Tasks: []domain.Task{
				{ID: "one", Name: "One"},
				{ID: "two", Name: "Two"},
			},
		}},
		TaskExecutor:  fakeTaskExecutor{},
		Clock:         fixedClock{now: time.Unix(100, 0)},
		MaxIterations: 1,
	})

	result, err := runner.Run(context.Background(), goal)
	if err == nil {
		t.Fatal("Run() error = nil, want max iteration error")
	}
	if result.State != domain.StateFailed {
		t.Fatalf("state = %s, want failed", result.State)
	}
}

func TestTaskSchedulerOrdersDependenciesAndDetectsCycles(t *testing.T) {
	t.Run("orders dependencies", func(t *testing.T) {
		ordered, err := (taskScheduler{}).Order([]domain.Task{
			{ID: "deploy", Dependencies: []string{"test"}},
			{ID: "build"},
			{ID: "test", Dependencies: []string{"build"}},
		})
		if err != nil {
			t.Fatalf("Order() error = %v", err)
		}
		got := []string{ordered[0].ID, ordered[1].ID, ordered[2].ID}
		want := []string{"build", "test", "deploy"}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("order = %v, want %v", got, want)
			}
		}
	})

	t.Run("detects cycles", func(t *testing.T) {
		_, err := (taskScheduler{}).Order([]domain.Task{
			{ID: "a", Dependencies: []string{"b"}},
			{ID: "b", Dependencies: []string{"a"}},
		})
		if err == nil {
			t.Fatal("Order() error = nil, want cycle error")
		}
	})
}

func TestSessionWorkflowRunnerPlanApprovalFlow(t *testing.T) {
	executor := fakeTaskExecutor{}
	runner := NewSessionWorkflowRunner(SessionWorkflowRunnerDeps{
		TaskExecutor: executor,
		Clock:        fixedClock{now: time.Unix(100, 0)},
	})
	workflow := domain.Workflow{
		Name: "release",
		Phases: []domain.Phase{{
			Name: "approval",
			Tasks: []domain.Task{{
				ID:               "plan",
				Name:             "Approve plan",
				RequiresApproval: true,
			}},
		}},
	}

	result, err := runner.Run(context.Background(), workflow, domain.ExecutionContext{SessionID: "session-1"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !result.Success || result.SessionID != "session-1" {
		t.Fatalf("result = %+v", result)
	}
}

func TestAIWorkflowRunnerDecisionRouting(t *testing.T) {
	goal := domain.Goal{ID: "goal-1", Description: "ship change"}
	decisionMaker := &scriptedDecisionMaker{decisions: []domain.Decision{
		{Type: domain.DecisionTypeNextTask, TaskID: "one"},
		{Type: domain.DecisionTypeComplete},
	}}
	runner := NewAIWorkflowRunner(AIWorkflowRunnerDeps{
		Planner: fakePlanner{plan: &domain.Plan{
			ID:    "plan-1",
			Goal:  goal,
			Tasks: []domain.Task{{ID: "one", Name: "One"}},
		}},
		DecisionMaker: decisionMaker,
		TaskExecutor:  fakeTaskExecutor{},
		Clock:         fixedClock{now: time.Unix(100, 0)},
		MaxIterations: 4,
	})

	result, err := runner.Run(context.Background(), goal)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !result.Success || len(result.Tasks) != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestTemplateRunnerRendersSchedulesExecutesAndWritesOutput(t *testing.T) {
	runner := NewTemplateRunner(TemplateRunnerDeps{
		TemplateStore: fakeTemplateStore{template: &domain.Template{
			Name:        "tmpl",
			Description: "Template",
			Tasks: []domain.Task{
				{ID: "write", Name: "Write", Prompt: "do {{thing}}", Dependencies: []string{"plan"}},
				{ID: "plan", Name: "Plan", Prompt: "plan {{thing}}"},
			},
			OutputFiles: []domain.OutputFile{{Path: "out.md", Template: "done"}},
		}},
		PromptRenderer: fakePromptRenderer{},
		TaskExecutor:   fakeTaskExecutor{},
		OutputWriter:   fakeOutputWriter{outputs: []string{"out.md"}},
		Clock:          fixedClock{now: time.Unix(100, 0)},
	})

	result, outputs, err := runner.Run(context.Background(), "tmpl", "/repo", map[string]string{"thing": "work"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !result.Success || len(result.Tasks) != 2 {
		t.Fatalf("result = %+v", result)
	}
	if outputs[0] != "out.md" {
		t.Fatalf("outputs = %v", outputs)
	}
	if result.Plan.Tasks[0].ID != "plan" {
		t.Fatalf("first task = %s, want dependency first", result.Plan.Tasks[0].ID)
	}
	if result.Plan.Tasks[0].Prompt != "plan work" {
		t.Fatalf("rendered prompt = %q", result.Plan.Tasks[0].Prompt)
	}
}

type fakeAnalyzer struct{}

func (fakeAnalyzer) AnalyzeProject(ctx context.Context, projectPath string) (*domain.ProjectContext, error) {
	return &domain.ProjectContext{ProjectPath: projectPath, ProjectName: "repo"}, nil
}

type fakePlanner struct {
	plan *domain.Plan
	err  error
}

func (p fakePlanner) Plan(ctx context.Context, goal domain.Goal, project *domain.ProjectContext) (*domain.Plan, error) {
	return p.plan, p.err
}

func (p fakePlanner) AdaptPlan(ctx context.Context, execution domain.ExecutionContext, reason string) (*domain.Plan, error) {
	return p.plan, p.err
}

type fakeTaskExecutor struct{}

func (fakeTaskExecutor) ExecuteTask(ctx context.Context, task domain.Task, execution domain.ExecutionContext) (*domain.TaskResult, error) {
	return &domain.TaskResult{
		TaskID:   task.ID,
		TaskName: task.Name,
		Success:  true,
	}, nil
}

type approvingReviewer struct{}

func (approvingReviewer) Review(ctx context.Context, execution domain.ExecutionContext) (*domain.ReviewResult, error) {
	return &domain.ReviewResult{Approved: true, Score: 100}, nil
}

type recordingProgressSink struct {
	last domain.Progress
}

func (s *recordingProgressSink) ReportProgress(ctx context.Context, progress domain.Progress) error {
	s.last = progress
	return nil
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func (c fixedClock) Sleep(ctx context.Context, duration time.Duration) error {
	return nil
}

type scriptedDecisionMaker struct {
	decisions []domain.Decision
	index     int
}

func (d *scriptedDecisionMaker) Decide(ctx context.Context, execution domain.ExecutionContext) (*domain.Decision, error) {
	if d.index >= len(d.decisions) {
		return nil, errors.New("no decisions left")
	}
	decision := d.decisions[d.index]
	d.index++
	return &decision, nil
}

type fakeTemplateStore struct {
	template *domain.Template
}

func (s fakeTemplateStore) LoadTemplate(ctx context.Context, name string) (*domain.Template, error) {
	return s.template, nil
}

type fakePromptRenderer struct{}

func (fakePromptRenderer) RenderPrompt(ctx context.Context, template string, values map[string]string) (string, error) {
	return strings.ReplaceAll(template, "{{thing}}", values["thing"]), nil
}

type fakeOutputWriter struct {
	outputs []string
}

func (w fakeOutputWriter) WriteOutputs(ctx context.Context, template domain.Template, result domain.Result) ([]string, error) {
	return w.outputs, nil
}
