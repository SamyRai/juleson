package core

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
)

// perceive gathers context about the goal and environment.
func (a *CoreAgent) perceive(ctx context.Context, goal agent.Goal) error {
	a.logger.Info("agent.perceive.start")

	if goal.Context.ProjectPath != "" {
		a.logger.Info("agent.perceive.analyzing_project", "path", goal.Context.ProjectPath)

		projectContext, err := a.analyzer.Analyze(goal.Context.ProjectPath)
		if err != nil {
			a.logger.Warn("agent.perceive.project_analysis_failed", "error", err)
		} else {
			a.projectContext = projectContext
			a.logger.Info("agent.perceive.project_analyzed",
				"name", projectContext.ProjectName,
				"type", projectContext.ProjectType,
				"languages", len(projectContext.Languages),
				"frameworks", len(projectContext.Frameworks))
		}
	}

	learnings, err := a.memory.Recall(ctx, goal.Description)
	if err != nil {
		a.logger.Warn("agent.perceive.memory_recall_failed", "error", err)
	} else {
		a.logger.Info("agent.perceive.recalled_learnings", "count", len(learnings))
	}

	decision := agent.Decision{
		Timestamp:  time.Now(),
		State:      agent.StateAnalyzing,
		Type:       agent.DecisionTypeSelectTool,
		Reasoning:  fmt.Sprintf("Perceived goal: %s. Found %d relevant learnings.", goal.Description, len(learnings)),
		Confidence: 0.8,
	}
	a.recordDecision(decision)

	a.setState(agent.StateAnalyzing)
	return nil
}

// plan creates a multi-step execution plan.
func (a *CoreAgent) plan(ctx context.Context, goal agent.Goal) error {
	a.logger.Info("agent.plan.start")

	if a.planner != nil {
		codebaseContext := "Go project with agent, tools, and automation components"

		tasks, reasoning, err := a.planner.GeneratePlan(ctx, goal, codebaseContext, a.projectContext)
		if err != nil {
			a.logger.Warn("agent.plan.ai_failed", "error", err)
		} else {
			a.currentPlan = tasks

			decision := agent.Decision{
				Timestamp:  time.Now(),
				State:      agent.StatePlanning,
				Type:       agent.DecisionTypeSelectTool,
				Reasoning:  reasoning.ChainOfThought[0],
				Action:     fmt.Sprintf("Generated AI plan with %d task(s)", len(a.currentPlan)),
				Confidence: reasoning.Confidence,
			}
			a.recordDecision(decision)

			if a.telemetry != nil {
				a.telemetry.RecordDecision(agent.DecisionTypeAdaptPlan, time.Since(time.Now()))
			}

			a.logger.Info("agent.plan.ai_complete", "tasks", len(a.currentPlan))
			a.setState(agent.StatePlanning)
			return nil
		}
	}

	a.logger.Info("agent.plan.fallback")
	task := agent.Task{
		ID:          "task-1",
		Name:        "Execute goal with Jules",
		Description: goal.Description,
		Prompt:      goal.Description,
		Priority:    goal.Priority,
		Tool:        "jules",
		State:       agent.TaskStatePending,
	}

	a.currentPlan = []agent.Task{task}

	decision := agent.Decision{
		Timestamp:  time.Now(),
		State:      agent.StatePlanning,
		Type:       agent.DecisionTypeSelectTool,
		Reasoning:  "Using fallback simple planning",
		Action:     "Use Jules tool for code generation",
		Confidence: DefaultConfidenceThreshold,
	}
	a.recordDecision(decision)

	a.logger.Info("agent.plan.complete", "tasks", len(a.currentPlan))
	a.setState(agent.StatePlanning)
	return nil
}
