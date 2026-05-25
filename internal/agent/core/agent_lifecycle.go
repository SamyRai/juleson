package core

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
)

// Execute implements the main agent loop: Perceive -> Plan -> Act -> Review -> Reflect.
func (a *CoreAgent) Execute(ctx context.Context, goal agent.Goal) (*agent.Result, error) {
	a.currentGoal = &goal
	startTime := time.Now()

	a.logger.Info("agent.execute.start",
		"goal_id", goal.ID,
		"description", goal.Description,
		"priority", goal.Priority)

	if a.checkpointMgr != nil {
		a.checkpointMgr.StartAutoSave(ctx, a)
	}

	result := &agent.Result{
		Goal:      goal,
		Success:   false,
		State:     agent.StateIdle,
		Tasks:     make([]agent.TaskResult, 0),
		Artifacts: make([]agent.Artifact, 0),
		PRs:       make([]string, 0),
		Issues:    make([]string, 0),
		Learnings: make([]agent.Learning, 0),
	}

	if a.telemetry != nil {
		a.telemetry.RecordExecution(false, 0)
	}

	iteration := 0
	for {
		iteration++

		a.logger.Info("agent.iteration.start", "iteration", iteration, "state", a.state)

		if a.state == agent.StateComplete || a.state == agent.StateFailed {
			a.logger.Info("agent.terminal_state_reached", "state", a.state, "iteration", iteration)
			break
		}

		if iteration > a.maxIterations {
			a.logger.Warn("agent.max_iterations_reached", "iterations", iteration-1)
			return a.finalizeResult(result, startTime, fmt.Errorf("max iterations (%d) reached", a.maxIterations))
		}

		select {
		case <-ctx.Done():
			return a.finalizeResult(result, startTime, fmt.Errorf("execution cancelled: %w", ctx.Err()))
		default:
		}

		var err error
		if a.retryStrategy != nil {
			err = a.retryStrategy.Execute(ctx, func(ctx context.Context, attempt int) error {
				return a.executeState(ctx, goal, result)
			}, fmt.Sprintf("state-%s", a.state))
		} else {
			err = a.executeState(ctx, goal, result)
		}

		if err != nil {
			a.logger.Error("agent.state.error", "state", a.state, "error", err)
			a.setState(agent.StateFailed)
			return a.finalizeResult(result, startTime, err)
		}
	}

	if a.state == agent.StateComplete || a.state == agent.StateFailed {
		return a.finalizeResult(result, startTime, nil)
	}

	a.logger.Warn("agent.max_iterations_reached", "iterations", iteration)
	return a.finalizeResult(result, startTime, fmt.Errorf("max iterations (%d) reached", a.maxIterations))
}

// executeState executes the current agent state.
func (a *CoreAgent) executeState(ctx context.Context, goal agent.Goal, result *agent.Result) error {
	switch a.state {
	case agent.StateIdle:
		return a.perceive(ctx, goal)
	case agent.StateAnalyzing:
		return a.plan(ctx, goal)
	case agent.StatePlanning:
		return a.act(ctx)
	case agent.StateExecuting:
		return a.review(ctx, result)
	case agent.StateReviewing:
		return a.reflect(ctx, result)
	case agent.StateReflecting:
		if a.needsMoreWork(result) {
			a.setState(agent.StatePlanning)
		} else {
			a.setState(agent.StateComplete)
		}
		return nil
	case agent.StateComplete:
		result.Success = true
		result.Duration = time.Since(time.Now())
		a.logger.Info("agent.execute.complete")
		return nil
	case agent.StateFailed:
		return fmt.Errorf("agent execution failed")
	default:
		return fmt.Errorf("unknown state: %s", a.state)
	}
}
