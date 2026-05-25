package automation

import (
	"strings"
	"testing"
	"time"

	"github.com/SamyRai/juleson/pkg/jules"
)

func TestSessionOrchestratorBuildInitialPrompt(t *testing.T) {
	orchestrator := NewSessionOrchestrator(&jules.Client{}, &WorkflowDefinition{
		Name:        "Quality Sprint",
		Description: "Improve ownership",
		Phases: []Phase{
			{
				Name:        "Agent",
				Description: "Split agent core",
				Tasks: []Task{
					{Name: "Lifecycle", Description: "Move lifecycle methods"},
					{Name: "Execution", Description: "Move task execution"},
				},
			},
		},
	}, nil)

	prompt := orchestrator.buildInitialPrompt()

	for _, want := range []string{
		"Execute workflow: Quality Sprint",
		"Improve ownership",
		"Starting with Phase 1: Agent",
		"- Lifecycle: Move lifecycle methods",
		"- Execution: Move task execution",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestSessionOrchestratorStateAccessorsReturnCopies(t *testing.T) {
	orchestrator := NewSessionOrchestrator(&jules.Client{}, &WorkflowDefinition{Name: "Workflow"}, nil)
	orchestrator.sessionID = "session-1"
	orchestrator.totalPhases = 3
	orchestrator.currentPhase = 1
	orchestrator.state = StateRunning
	orchestrator.addExecutionRecord(ExecutionRecord{TaskName: "task-1"})

	phase, total, state := orchestrator.GetProgress()
	if phase != 1 || total != 3 || state != StateRunning {
		t.Fatalf("progress = %d/%d %s, want 1/3 %s", phase, total, state, StateRunning)
	}
	if orchestrator.GetSessionID() != "session-1" {
		t.Fatalf("session ID = %q, want session-1", orchestrator.GetSessionID())
	}

	logCopy := orchestrator.GetExecutionLog()
	logCopy[0].TaskName = "mutated"
	if orchestrator.GetExecutionLog()[0].TaskName != "task-1" {
		t.Fatalf("GetExecutionLog should return a copy")
	}
}

func TestAIOrchestratorStateHelpers(t *testing.T) {
	orchestrator := NewAIOrchestrator(&jules.Client{}, nil, &AIOrchestrationConfig{
		MaxIterations: 1,
		CheckInterval: time.Millisecond,
	})
	orchestrator.goal = "Improve maintainability"
	orchestrator.context = &ProjectContext{
		Languages:    []string{"Go"},
		Architecture: "CLI",
		CurrentState: "Working",
	}
	orchestrator.pendingTasks = []PendingTask{
		{Name: "One", Description: "First"},
		{Name: "Two", Description: "Second"},
	}

	prompt := orchestrator.buildAIPrompt()
	if !strings.Contains(prompt, "Improve maintainability") || !strings.Contains(prompt, "First") {
		t.Fatalf("unexpected AI prompt:\n%s", prompt)
	}

	decision := &AIDecision{DecisionType: "next_task"}
	orchestrator.recordDecision(decision)
	history := orchestrator.GetDecisionHistory()
	history[0].DecisionType = "mutated"
	if orchestrator.GetDecisionHistory()[0].DecisionType != "next_task" {
		t.Fatalf("GetDecisionHistory should return a copy")
	}
}
