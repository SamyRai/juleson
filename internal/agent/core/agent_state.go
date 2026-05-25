package core

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
)

func (a *CoreAgent) setState(newState agent.AgentState) {
	oldState := a.state
	a.state = newState
	a.logger.Info("agent.state.transition", "from", oldState, "to", newState)
}

func (a *CoreAgent) recordDecision(decision agent.Decision) {
	if decision.ID == "" {
		decision.ID = fmt.Sprintf("decision-%d", len(a.decisions)+1)
	}
	a.decisions = append(a.decisions, decision)
	if err := a.memory.RecordDecision(context.Background(), decision); err != nil {
		a.logger.Error("failed to record decision", "error", err)
	}
}

func (a *CoreAgent) needsMoreWork(result *agent.Result) bool {
	for _, task := range a.currentPlan {
		if task.State == agent.TaskStateFailed || task.State == agent.TaskStatePending {
			return true
		}
	}
	return false
}

func (a *CoreAgent) finalizeResult(result *agent.Result, startTime time.Time, err error) (*agent.Result, error) {
	result.Duration = time.Since(startTime)
	result.State = a.state
	result.Error = err

	if err != nil {
		result.Success = false
		result.Summary = fmt.Sprintf("Failed: %v", err)
	} else {
		result.Success = true
		result.Summary = "Completed successfully"
	}

	if a.telemetry != nil {
		a.telemetry.RecordExecution(result.Success, result.Duration)
	}

	return result, err
}

// GetState returns the current agent state.
func (a *CoreAgent) GetState() agent.AgentState {
	return a.state
}

// GetHistory returns the recorded decision history.
func (a *CoreAgent) GetHistory() []agent.Decision {
	return a.decisions
}

// Pause pauses the agent.
func (a *CoreAgent) Pause() error {
	return fmt.Errorf("pause not implemented")
}

// Resume resumes the agent.
func (a *CoreAgent) Resume() error {
	return fmt.Errorf("resume not implemented")
}

// Stop stops the agent by moving it to failed state.
func (a *CoreAgent) Stop() error {
	a.setState(agent.StateFailed)
	return nil
}

// ProvideFeedback accepts external feedback for future processing.
func (a *CoreAgent) ProvideFeedback(feedback agent.Feedback) error {
	a.logger.Info("agent.feedback.received", "type", feedback.Type, "message", feedback.Message)
	return nil
}

// GetProgress returns current task progress.
func (a *CoreAgent) GetProgress() *agent.Progress {
	completedTasks := 0
	for _, task := range a.currentPlan {
		if task.State == agent.TaskStateComplete {
			completedTasks++
		}
	}

	progress := float64(0)
	if len(a.currentPlan) > 0 {
		progress = float64(completedTasks) / float64(len(a.currentPlan)) * 100
	}

	currentTaskName := ""
	for _, task := range a.currentPlan {
		if task.State == agent.TaskStateInProgress {
			currentTaskName = task.Name
			break
		}
	}

	return &agent.Progress{
		State:          a.state,
		CurrentTask:    currentTaskName,
		CompletedTasks: completedTasks,
		TotalTasks:     len(a.currentPlan),
		Progress:       progress,
		Message:        fmt.Sprintf("State: %s", a.state),
		Timestamp:      time.Now(),
	}
}

// SetConstraints updates the agent's constraint validator.
func (a *CoreAgent) SetConstraints(constraints []string) {
	if a.validator != nil {
		a.validator = NewConstraintValidator(constraints)
		if a.executor != nil {
			a.executor.setValidator(a.validator)
		}
	}
}

// GetTelemetrySummary returns telemetry metrics summary.
func (a *CoreAgent) GetTelemetrySummary() map[string]interface{} {
	if a.telemetry != nil {
		return a.telemetry.Summary()
	}
	return nil
}

// GetCheckpoints returns available checkpoints.
func (a *CoreAgent) GetCheckpoints() ([]Checkpoint, error) {
	if a.checkpointMgr != nil {
		return a.checkpointMgr.List()
	}
	return nil, fmt.Errorf("checkpoint manager not initialized")
}

// RestoreFromCheckpoint restores agent state from a checkpoint.
func (a *CoreAgent) RestoreFromCheckpoint(ctx context.Context, checkpointID string) error {
	if a.checkpointMgr != nil {
		return a.checkpointMgr.Restore(ctx, checkpointID, a)
	}
	return fmt.Errorf("checkpoint manager not initialized")
}
