package core

import (
	"context"
	"testing"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
	"github.com/SamyRai/juleson/internal/agent/tools"
	"github.com/SamyRai/juleson/internal/analyzer"
)

func TestExecuteSingleTask(t *testing.T) {
	// Create a mock tool registry
	registry := tools.NewToolRegistry()

	// Create agent config
	config := &Config{
		MaxIterations:   10,
		DryRun:          false,
		Logger:          nil, // Will use default
		RetryConfig:     DefaultRetryStrategy(),
		CheckpointDir:   "./test_checkpoints",
		AutoSave:        false,
		SaveInterval:    5 * time.Minute,
		EnableTelemetry: false,
	}

	coreAgent := newTestCoreAgent(t, registry, config)

	// Create a test task
	task := &agent.Task{
		ID:          "test-task-1",
		Name:        "Test Task",
		Description: "A test task for unit testing",
		Prompt:      "Execute this test task",
		Priority:    agent.PriorityHigh,
		Tool:        "jules",
		State:       agent.TaskStatePending,
	}

	// Set up agent state
	coreAgent.currentGoal = &agent.Goal{
		Description: "Test goal",
		Priority:    agent.PriorityHigh,
	}

	ctx := context.Background()

	// Execute the task
	result, err := coreAgent.executeSingleTask(ctx, task)

	// In a real test, we'd mock the Jules client and verify the result
	// For now, we just verify no panic occurs
	if err != nil {
		t.Logf("Task execution failed (expected in test environment): %v", err)
	}

	if result != nil {
		t.Logf("Task result: %+v", result)
	}
}

func TestPrepareToolParameters(t *testing.T) {
	// Create agent config
	config := &Config{
		MaxIterations:   10,
		DryRun:          false,
		Logger:          nil,
		RetryConfig:     DefaultRetryStrategy(),
		CheckpointDir:   "./test_checkpoints",
		AutoSave:        false,
		SaveInterval:    5 * time.Minute,
		EnableTelemetry: false,
	}

	// Create agent
	registry := tools.NewToolRegistry()
	coreAgent := newTestCoreAgent(t, registry, config)

	// Create a test task
	task := &agent.Task{
		ID:          "test-task-1",
		Name:        "Test Task",
		Description: "A test task for unit testing",
		Prompt:      "Execute this test task",
		Priority:    agent.PriorityHigh,
		Tool:        "jules",
		State:       agent.TaskStatePending,
		Context: map[string]interface{}{
			"source": "test.go",
			"action": "create",
		},
	}

	// Set up agent state with goal and project context
	coreAgent.currentGoal = &agent.Goal{
		Description: "Test goal",
		Priority:    agent.PriorityHigh,
		Context: agent.GoalContext{
			SourceID: "test-source-123",
		},
	}
	coreAgent.projectContext = &analyzer.ProjectContext{
		ProjectPath: "/test/project",
	}

	// Prepare tool parameters
	params := coreAgent.prepareToolParameters(task)

	// Verify parameters
	if params == nil {
		t.Fatal("Expected non-nil parameters")
	}

	// Check expected parameters
	if action, ok := params["action"]; !ok || action != "create_session" {
		t.Errorf("Expected action parameter 'create_session', got %v", action)
	}

	if prompt, ok := params["prompt"]; !ok || prompt != task.Prompt {
		t.Errorf("Expected prompt parameter '%s', got %v", task.Prompt, prompt)
	}

	if sourceID, ok := params["source_id"]; !ok || sourceID != "test-source-123" {
		t.Errorf("Expected source_id parameter 'test-source-123', got %v", sourceID)
	}

	if projectPath, ok := params["project_path"]; !ok || projectPath != "/test/project" {
		t.Errorf("Expected project_path parameter '/test/project', got %v", projectPath)
	}
}

func TestGetProgressReportsCurrentTask(t *testing.T) {
	registry := tools.NewToolRegistry()
	coreAgent := newTestCoreAgent(t, registry, &Config{
		MaxIterations:   10,
		Logger:          nil,
		RetryConfig:     DefaultRetryStrategy(),
		CheckpointDir:   "./test_checkpoints",
		AutoSave:        false,
		SaveInterval:    5 * time.Minute,
		EnableTelemetry: false,
	})
	coreAgent.state = agent.StateExecuting
	coreAgent.currentPlan = []agent.Task{
		{ID: "done", Name: "Done", State: agent.TaskStateComplete},
		{ID: "active", Name: "Active", State: agent.TaskStateInProgress},
		{ID: "next", Name: "Next", State: agent.TaskStatePending},
	}

	progress := coreAgent.GetProgress()

	if progress.State != agent.StateExecuting {
		t.Fatalf("State = %s, want %s", progress.State, agent.StateExecuting)
	}
	if progress.CurrentTask != "Active" {
		t.Fatalf("CurrentTask = %q, want Active", progress.CurrentTask)
	}
	if progress.CompletedTasks != 1 || progress.TotalTasks != 3 {
		t.Fatalf("task counts = %d/%d, want 1/3", progress.CompletedTasks, progress.TotalTasks)
	}
	if progress.Progress != float64(1)/float64(3)*100 {
		t.Fatalf("Progress = %v, want one third", progress.Progress)
	}
}

func TestNeedsMoreWork(t *testing.T) {
	registry := tools.NewToolRegistry()
	coreAgent := newTestCoreAgent(t, registry, &Config{
		MaxIterations:   10,
		Logger:          nil,
		RetryConfig:     DefaultRetryStrategy(),
		CheckpointDir:   "./test_checkpoints",
		AutoSave:        false,
		SaveInterval:    5 * time.Minute,
		EnableTelemetry: false,
	})

	tests := []struct {
		name  string
		tasks []agent.Task
		want  bool
	}{
		{
			name:  "all complete",
			tasks: []agent.Task{{State: agent.TaskStateComplete}},
			want:  false,
		},
		{
			name:  "pending task",
			tasks: []agent.Task{{State: agent.TaskStatePending}},
			want:  true,
		},
		{
			name:  "failed task",
			tasks: []agent.Task{{State: agent.TaskStateFailed}},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coreAgent.currentPlan = tt.tasks
			if got := coreAgent.needsMoreWork(&agent.Result{}); got != tt.want {
				t.Fatalf("needsMoreWork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if config.MaxIterations != DefaultMaxIterations {
		t.Errorf("Expected MaxIterations %d, got %d", DefaultMaxIterations, config.MaxIterations)
	}

	if config.SaveInterval != DefaultCheckpointInterval {
		t.Errorf("Expected SaveInterval %v, got %v", DefaultCheckpointInterval, config.SaveInterval)
	}

	if config.RetryConfig == nil {
		t.Error("Expected non-nil RetryConfig")
	}
}

func TestNewAgent(t *testing.T) {
	registry := tools.NewToolRegistry()
	config := DefaultConfig()

	agent := NewAgent(registry, config)

	if agent == nil {
		t.Fatal("Expected non-nil agent")
	}

	coreAgent, ok := agent.(*CoreAgent)
	if !ok {
		t.Fatal("Expected CoreAgent type")
	}

	if coreAgent.toolRegistry == nil {
		t.Error("Expected non-nil tool registry")
	}

	if coreAgent.maxIterations != DefaultMaxIterations {
		t.Errorf("Expected maxIterations %d, got %d", DefaultMaxIterations, coreAgent.maxIterations)
	}
}

func newTestCoreAgent(t *testing.T, registry tools.ToolRegistry, config *Config) *CoreAgent {
	t.Helper()
	coreAgent, ok := NewAgent(registry, config).(*CoreAgent)
	if !ok {
		t.Fatal("Expected CoreAgent type")
	}
	return coreAgent
}
