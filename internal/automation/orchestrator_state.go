package automation

import (
	"fmt"
	"time"

	"github.com/SamyRai/go-jules"
)

func (o *SessionOrchestrator) setState(state OrchestratorState) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.state = state
}

func (o *SessionOrchestrator) sendProgress(phase, task int, message string, progress float64) {
	update := ProgressUpdate{
		SessionID:   o.sessionID,
		Phase:       phase,
		TotalPhases: o.totalPhases,
		Task:        message,
		Progress:    progress,
		Message:     message,
		Timestamp:   time.Now(),
	}

	select {
	case o.progressChan <- update:
	default:
	}
}

func (o *SessionOrchestrator) addExecutionRecord(record ExecutionRecord) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.executionLog = append(o.executionLog, record)
}

func (o *SessionOrchestrator) handleProgress(status *jules.SessionStatus) {
	update := ActivityUpdate{
		SessionID: o.sessionID,
		Message:   fmt.Sprintf("Session state: %s", status.State),
		Timestamp: time.Now(),
	}

	select {
	case o.activityChan <- update:
	default:
	}
}

func (o *SessionOrchestrator) checkPrerequisites(prerequisites []string) error {
	return nil
}

func (o *SessionOrchestrator) buildWorkflowResult() WorkflowResult {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return WorkflowResult{
		WorkflowName:  o.workflow.Name,
		Success:       o.state == StateCompleted,
		TotalPhases:   o.totalPhases,
		TotalDuration: time.Since(o.startTime),
		SessionID:     o.sessionID,
		StartTime:     o.startTime,
		EndTime:       time.Now(),
	}
}

// GetProgress returns the current progress.
func (o *SessionOrchestrator) GetProgress() (int, int, OrchestratorState) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.currentPhase, o.totalPhases, o.state
}

// GetSessionID returns the current session ID.
func (o *SessionOrchestrator) GetSessionID() string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.sessionID
}

// GetExecutionLog returns the execution log.
func (o *SessionOrchestrator) GetExecutionLog() []ExecutionRecord {
	o.mu.RLock()
	defer o.mu.RUnlock()
	logCopy := make([]ExecutionRecord, len(o.executionLog))
	copy(logCopy, o.executionLog)
	return logCopy
}

// ProgressChannel returns the progress update channel.
func (o *SessionOrchestrator) ProgressChannel() <-chan ProgressUpdate {
	return o.progressChan
}

// ActivityChannel returns the activity update channel.
func (o *SessionOrchestrator) ActivityChannel() <-chan ActivityUpdate {
	return o.activityChan
}

// ErrorChannel returns the error channel.
func (o *SessionOrchestrator) ErrorChannel() <-chan error {
	return o.errorChan
}

// Stop gracefully stops the orchestrator.
func (o *SessionOrchestrator) Stop() {
	o.setState(StateCancelled)
	close(o.stopChan)
}

// Pause pauses the orchestrator.
func (o *SessionOrchestrator) Pause() {
	o.setState(StatePaused)
}

// Resume resumes the orchestrator.
func (o *SessionOrchestrator) Resume() {
	o.setState(StateRunning)
}
