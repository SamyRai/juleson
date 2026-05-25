package automation

import (
	"fmt"
	"strings"
	"time"

	"github.com/SamyRai/juleson/pkg/jules"
)

func (ao *AIOrchestrator) buildAIPrompt() string {
	prompt := fmt.Sprintf(`I need you to help achieve this goal: %s

`, ao.goal)

	if ao.context != nil {
		prompt += fmt.Sprintf(`Project Context:
- Languages: %s
- Architecture: %s
- Current State: %s

`, strings.Join(ao.context.Languages, ", "), ao.context.Architecture, ao.context.CurrentState)
	}

	if len(ao.pendingTasks) > 0 {
		prompt += "Initial tasks to execute:\n"
		for i, task := range ao.pendingTasks[:min(3, len(ao.pendingTasks))] {
			prompt += fmt.Sprintf("%d. %s\n", i+1, task.Description)
		}
	}

	prompt += "\nThis is a multi-step process. I'll send follow-up messages as we progress."

	return prompt
}

func (ao *AIOrchestrator) setState(state AIState) {
	ao.mu.Lock()
	defer ao.mu.Unlock()
	ao.state = state
}

func (ao *AIOrchestrator) sendProgress(phase, message string, progress float64, nextSteps []string) {
	update := AIProgress{
		Phase:       phase,
		CurrentTask: message,
		Progress:    progress,
		Message:     message,
		Timestamp:   time.Now(),
		NextSteps:   nextSteps,
	}

	select {
	case ao.progressChan <- update:
	default:
	}
}

func (ao *AIOrchestrator) sendDecision(decision AIDecision) {
	select {
	case ao.decisionChan <- decision:
	default:
	}
}

func (ao *AIOrchestrator) recordDecision(decision *AIDecision) {
	ao.mu.Lock()
	defer ao.mu.Unlock()
	ao.lastAIDecision = decision
	ao.decisionHistory = append(ao.decisionHistory, *decision)
}

func (ao *AIOrchestrator) formatCompletedTasks() string {
	ao.mu.RLock()
	defer ao.mu.RUnlock()

	if len(ao.completedTasks) == 0 {
		return "None yet"
	}

	var sb strings.Builder
	for i, task := range ao.completedTasks {
		sb.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, task.Name, task.Description))
	}
	return sb.String()
}

func (ao *AIOrchestrator) formatCompletedTasksDetailed() string {
	ao.mu.RLock()
	defer ao.mu.RUnlock()

	var sb strings.Builder
	for i, task := range ao.completedTasks {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, task.Name))
		sb.WriteString(fmt.Sprintf("   Description: %s\n", task.Description))
		sb.WriteString(fmt.Sprintf("   Success: %v\n", task.Result.Success))
		if len(task.Result.FilesChanged) > 0 {
			sb.WriteString(fmt.Sprintf("   Files Changed: %d\n", len(task.Result.FilesChanged)))
		}
	}
	return sb.String()
}

func (ao *AIOrchestrator) formatPendingTasks() string {
	ao.mu.RLock()
	defer ao.mu.RUnlock()

	if len(ao.pendingTasks) == 0 {
		return "None"
	}

	var sb strings.Builder
	for i, task := range ao.pendingTasks {
		sb.WriteString(fmt.Sprintf("%d. %s (Priority: %d)\n", i+1, task.Name, task.Priority))
	}
	return sb.String()
}

func (ao *AIOrchestrator) applyAdaptations(adaptations map[string]interface{}) {
	// Apply AI's recommended adaptations to the pending task list.
}

// ProgressChannel returns progress updates from AI orchestration.
func (ao *AIOrchestrator) ProgressChannel() <-chan AIProgress {
	return ao.progressChan
}

// DecisionChannel returns AI decisions.
func (ao *AIOrchestrator) DecisionChannel() <-chan AIDecision {
	return ao.decisionChan
}

// Stop stops the AI orchestrator.
func (ao *AIOrchestrator) Stop() {
	close(ao.stopChan)
}

// GetState returns the current AI orchestration state.
func (ao *AIOrchestrator) GetState() AIState {
	ao.mu.RLock()
	defer ao.mu.RUnlock()
	return ao.state
}

// GetSessionID returns the current Jules session ID.
func (ao *AIOrchestrator) GetSessionID() string {
	ao.mu.RLock()
	defer ao.mu.RUnlock()
	return ao.sessionID
}

// GetDecisionHistory returns a copy of AI decision history.
func (ao *AIOrchestrator) GetDecisionHistory() []AIDecision {
	ao.mu.RLock()
	defer ao.mu.RUnlock()
	history := make([]AIDecision, len(ao.decisionHistory))
	copy(history, ao.decisionHistory)
	return history
}

func extractFilesChanged(activity jules.Activity) []string {
	files := make([]string, 0)
	for _, artifact := range activity.Artifacts {
		if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
			files = append(files, "modified_file.go")
		}
	}
	return files
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
