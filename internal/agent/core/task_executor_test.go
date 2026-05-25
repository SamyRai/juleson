package core

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
	"github.com/SamyRai/juleson/internal/agent/tools"
	"github.com/SamyRai/juleson/internal/analyzer"
)

func TestTaskExecutorPrepareParameters(t *testing.T) {
	executor := newTaskExecutor(
		fakeToolFinder{},
		nil,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		false,
	)

	task := &agent.Task{Prompt: "implement task"}
	params := executor.prepareParameters(task, taskExecutionContext{
		goal: &agent.Goal{
			Context: agent.GoalContext{SourceID: "source-123"},
		},
		projectContext: &analyzer.ProjectContext{ProjectPath: "/repo"},
	})

	if params["action"] != "create_session" {
		t.Fatalf("action = %v, want create_session", params["action"])
	}
	if params["prompt"] != "implement task" {
		t.Fatalf("prompt = %v, want implement task", params["prompt"])
	}
	if params["source_id"] != "source-123" {
		t.Fatalf("source_id = %v, want source-123", params["source_id"])
	}
	if params["project_path"] != "/repo" {
		t.Fatalf("project_path = %v, want /repo", params["project_path"])
	}
}

func TestTaskExecutorExecutePassesOwnedParametersToTool(t *testing.T) {
	tool := &fakeExecutionTool{
		result: &tools.ToolResult{
			Success: true,
			Changes: []agent.Change{
				{FilePath: "main.go", Type: agent.ChangeTypeModify},
			},
		},
	}
	executor := newTaskExecutor(
		fakeToolFinder{tools: []tools.Tool{tool}},
		nil,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		false,
	)

	result, err := executor.execute(context.Background(), &agent.Task{
		ID:     "task-1",
		Name:   "Task 1",
		Prompt: "make change",
	}, taskExecutionContext{
		goal:           &agent.Goal{Context: agent.GoalContext{SourceID: "source-123"}},
		projectContext: &analyzer.ProjectContext{ProjectPath: "/repo"},
	})

	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if !result.Success {
		t.Fatalf("Success = false, want true")
	}
	if result.Tool != "fake-tool" {
		t.Fatalf("Tool = %q, want fake-tool", result.Tool)
	}
	if tool.receivedParams["prompt"] != "make change" {
		t.Fatalf("tool prompt param = %v, want make change", tool.receivedParams["prompt"])
	}
	if tool.receivedParams["source_id"] != "source-123" {
		t.Fatalf("tool source_id param = %v, want source-123", tool.receivedParams["source_id"])
	}
}

func TestTaskExecutorDryRunSkipsToolExecution(t *testing.T) {
	tool := &fakeExecutionTool{
		result: &tools.ToolResult{Success: true},
	}
	executor := newTaskExecutor(
		fakeToolFinder{tools: []tools.Tool{tool}},
		nil,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		true,
	)

	result, err := executor.execute(context.Background(), &agent.Task{
		ID:   "task-1",
		Name: "Task 1",
	}, taskExecutionContext{})

	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if !result.Success {
		t.Fatalf("Success = false, want true")
	}
	if result.Duration != time.Second {
		t.Fatalf("Duration = %v, want %v", result.Duration, time.Second)
	}
	if tool.executed {
		t.Fatal("tool executed during dry run")
	}
}

type fakeToolFinder struct {
	tools []tools.Tool
}

func (f fakeToolFinder) FindForTask(agent.Task) []tools.Tool {
	return f.tools
}

type fakeExecutionTool struct {
	result         *tools.ToolResult
	receivedParams map[string]interface{}
	executed       bool
}

func (t *fakeExecutionTool) Name() string {
	return "fake-tool"
}

func (t *fakeExecutionTool) Description() string {
	return "fake tool"
}

func (t *fakeExecutionTool) Parameters() []tools.Parameter {
	return nil
}

func (t *fakeExecutionTool) Execute(_ context.Context, params map[string]interface{}) (*tools.ToolResult, error) {
	t.executed = true
	t.receivedParams = params
	return t.result, nil
}

func (t *fakeExecutionTool) RequiresApproval() bool {
	return false
}

func (t *fakeExecutionTool) CanHandle(agent.Task) bool {
	return true
}
