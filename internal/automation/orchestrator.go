package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SamyRai/juleson/internal/jules"
)

// Default configuration constants
const (
	DefaultCheckInterval   = 10 * time.Second
	DefaultMaxSessionAge   = 4 * time.Hour
	DefaultRetryAttempts   = 3
	DefaultProgressChanBuf = 100
	DefaultActivityChanBuf = 100
	DefaultErrorChanBuf    = 10
	DefaultTaskWaitTime    = 2 * time.Second
)

// SessionOrchestrator manages multi-phase workflows within a single Jules session
// It ensures efficient session usage by:
// 1. Running multiple tasks in one session
// 2. Monitoring progress in real-time
// 3. Dynamically steering based on results
// 4. Handling plan approval gates
// 5. Managing state persistence
type SessionOrchestrator struct {
	client        *jules.Client
	sessionID     string
	currentPhase  int
	totalPhases   int
	state         OrchestratorState
	startTime     time.Time
	mu            sync.RWMutex
	progressChan  chan ProgressUpdate
	activityChan  chan ActivityUpdate
	errorChan     chan error
	stopChan      chan struct{}
	monitor       *jules.SessionMonitor
	workflow      *WorkflowDefinition
	executionLog  []ExecutionRecord
	autoApprove   bool
	checkInterval time.Duration
	maxSessionAge time.Duration
}

// OrchestratorState represents the current state of the orchestrator
type OrchestratorState string

const (
	StateInitializing OrchestratorState = "INITIALIZING"
	StateRunning      OrchestratorState = "RUNNING"
	StateWaitingPlan  OrchestratorState = "WAITING_PLAN"
	StatePaused       OrchestratorState = "PAUSED"
	StateCompleted    OrchestratorState = "COMPLETED"
	StateFailed       OrchestratorState = "FAILED"
	StateCancelled    OrchestratorState = "CANCELLED"
)

// WorkflowDefinition defines a complete multi-phase workflow
type WorkflowDefinition struct {
	Name               string
	Description        string
	Phases             []Phase
	MaxDuration        time.Duration
	OnPhaseComplete    func(phaseIndex int, result PhaseResult) error
	OnWorkflowComplete func(result WorkflowResult) error
}

// Phase represents a task phase in the orchestration
type Phase struct {
	Name            string
	Description     string
	Tasks           []Task
	Parallel        bool // Execute tasks in parallel
	ContinueOnError bool
	Timeout         time.Duration
	Prerequisites   []string // Phase names that must complete first
}

// Task represents a single task within a phase
type Task struct {
	Name        string
	Description string
	Prompt      string
	WaitForPlan bool
	AutoApprove bool
	Template    string // Optional: template name to execute
	Retry       int    // Number of retry attempts
	Timeout     time.Duration
	Validation  func(result TaskResult) error
}

// ProgressUpdate represents a progress update from the orchestrator
type ProgressUpdate struct {
	SessionID   string
	Phase       int
	TotalPhases int
	Task        string
	Progress    float64
	Message     string
	Timestamp   time.Time
}

// ActivityUpdate represents an activity update from Jules
type ActivityUpdate struct {
	SessionID    string
	ActivityID   string
	ActivityType string
	Timestamp    time.Time
	Artifacts    int
	Message      string
}

// ExecutionRecord tracks the execution of a task
type ExecutionRecord struct {
	PhaseIndex    int
	TaskIndex     int
	TaskName      string
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	Success       bool
	Error         string
	ActivityID    string
	ArtifactCount int
}

// PhaseResult represents the result of a phase execution
type PhaseResult struct {
	PhaseIndex int
	PhaseName  string
	Success    bool
	Tasks      []TaskResult
	Duration   time.Duration
	Error      error
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskName   string
	Success    bool
	ActivityID string
	Artifacts  []jules.Artifact
	Duration   time.Duration
	Error      error
}

// WorkflowResult represents the final result of workflow execution
type WorkflowResult struct {
	WorkflowName  string
	Success       bool
	TotalPhases   int
	PhaseResults  []PhaseResult
	TotalDuration time.Duration
	SessionID     string
	StartTime     time.Time
	EndTime       time.Time
}

// OrchestratorConfig configures the session orchestrator
type OrchestratorConfig struct {
	AutoApprove     bool
	CheckInterval   time.Duration
	MaxSessionAge   time.Duration
	RetryAttempts   int
	ContinueOnError bool
	SaveState       bool
	StateFile       string
}

// DefaultOrchestratorConfig returns default configuration
func DefaultOrchestratorConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		AutoApprove:     false,
		CheckInterval:   DefaultCheckInterval,
		MaxSessionAge:   DefaultMaxSessionAge,
		RetryAttempts:   DefaultRetryAttempts,
		ContinueOnError: false,
		SaveState:       true,
	}
}

// NewSessionOrchestrator creates a new session orchestrator
func NewSessionOrchestrator(client *jules.Client, workflow *WorkflowDefinition, config *OrchestratorConfig) *SessionOrchestrator {
	if client == nil {
		panic("client cannot be nil")
	}
	if workflow == nil {
		panic("workflow cannot be nil")
	}

	if config == nil {
		config = DefaultOrchestratorConfig()
	}

	return &SessionOrchestrator{
		client:        client,
		workflow:      workflow,
		state:         StateInitializing,
		progressChan:  make(chan ProgressUpdate, DefaultProgressChanBuf),
		activityChan:  make(chan ActivityUpdate, DefaultActivityChanBuf),
		errorChan:     make(chan error, DefaultErrorChanBuf),
		stopChan:      make(chan struct{}),
		executionLog:  make([]ExecutionRecord, 0),
		autoApprove:   config.AutoApprove,
		checkInterval: config.CheckInterval,
		maxSessionAge: config.MaxSessionAge,
	}
}

// Start initiates the workflow orchestration
func (o *SessionOrchestrator) Start(ctx context.Context, sourceID string) error {
	o.mu.Lock()
	o.state = StateRunning
	o.startTime = time.Now()
	o.mu.Unlock()

	// Create initial session with comprehensive prompt
	initialPrompt := o.buildInitialPrompt()

	session, err := o.client.CreateSession(ctx, &jules.CreateSessionRequest{
		Prompt: initialPrompt,
		SourceContext: &jules.SourceContext{
			Source: fmt.Sprintf("sources/%s", sourceID),
		},
		RequirePlanApproval: !o.autoApprove,
	})
	if err != nil {
		o.setState(StateFailed)
		return fmt.Errorf("failed to create session: %w", err)
	}

	o.mu.Lock()
	o.sessionID = session.ID
	o.mu.Unlock()

	// Set up session monitor
	o.monitor = jules.NewSessionMonitor(o.client, session.ID).
		WithInterval(o.checkInterval).
		WithMaxWait(o.maxSessionAge).
		OnProgress(o.handleProgress)

	// Send progress update
	o.sendProgress(0, 0, "Session created", 0)

	// Execute workflow phases
	return o.executeWorkflow(ctx)
}

// executeWorkflow executes all workflow phases
func (o *SessionOrchestrator) executeWorkflow(ctx context.Context) error {
	o.totalPhases = len(o.workflow.Phases)

	for i, phase := range o.workflow.Phases {
		// Check context cancellation before each phase
		select {
		case <-ctx.Done():
			o.setState(StateCancelled)
			return fmt.Errorf("workflow cancelled: %w", ctx.Err())
		default:
		}

		o.mu.Lock()
		o.currentPhase = i
		o.mu.Unlock()

		o.sendProgress(i, 0, fmt.Sprintf("Starting phase: %s", phase.Name), 0)

		// Check prerequisites
		if err := o.checkPrerequisites(phase.Prerequisites); err != nil {
			return fmt.Errorf("phase %d prerequisites not met: %w", i, err)
		}

		// Execute phase
		result, err := o.executePhase(ctx, i, phase)
		if err != nil {
			if !phase.ContinueOnError {
				o.setState(StateFailed)
				return fmt.Errorf("phase %d failed: %w", i, err)
			}
		}

		// Call phase completion callback
		if o.workflow.OnPhaseComplete != nil {
			if err := o.workflow.OnPhaseComplete(i, result); err != nil {
				return fmt.Errorf("phase completion callback failed: %w", err)
			}
		}

		o.sendProgress(i, len(phase.Tasks), fmt.Sprintf("Phase completed: %s", phase.Name), 100)
	}

	o.setState(StateCompleted)

	// Build final result
	workflowResult := o.buildWorkflowResult()

	// Call workflow completion callback
	if o.workflow.OnWorkflowComplete != nil {
		if err := o.workflow.OnWorkflowComplete(workflowResult); err != nil {
			return fmt.Errorf("workflow completion callback failed: %w", err)
		}
	}

	return nil
}

// executePhase executes a single phase
func (o *SessionOrchestrator) executePhase(ctx context.Context, phaseIndex int, phase Phase) (PhaseResult, error) {
	result := PhaseResult{
		PhaseIndex: phaseIndex,
		PhaseName:  phase.Name,
		Success:    true,
		Tasks:      make([]TaskResult, 0),
	}

	startTime := time.Now()

	if phase.Parallel {
		// Execute tasks in parallel
		return o.executeTasksParallel(ctx, phaseIndex, phase)
	}

	// Execute tasks sequentially
	for j, task := range phase.Tasks {
		taskResult, err := o.executeTask(ctx, phaseIndex, j, task)
		result.Tasks = append(result.Tasks, taskResult)

		if err != nil {
			result.Success = false
			result.Error = err
			if !phase.ContinueOnError {
				result.Duration = time.Since(startTime)
				return result, err
			}
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// executeTask executes a single task
func (o *SessionOrchestrator) executeTask(ctx context.Context, phaseIndex, taskIndex int, task Task) (TaskResult, error) {
	result := TaskResult{
		TaskName: task.Name,
		Success:  false,
	}

	startTime := time.Now()
	record := ExecutionRecord{
		PhaseIndex: phaseIndex,
		TaskIndex:  taskIndex,
		TaskName:   task.Name,
		StartTime:  startTime,
	}

	o.sendProgress(phaseIndex, taskIndex, fmt.Sprintf("Executing: %s", task.Name), 0)

	// Send message to session to execute this task
	err := o.client.SendMessage(ctx, o.sessionID, &jules.SendMessageRequest{
		Prompt: task.Prompt,
	})
	if err != nil {
		record.Success = false
		record.Error = err.Error()
		record.EndTime = time.Now()
		record.Duration = time.Since(startTime)
		o.addExecutionRecord(record)

		result.Error = err
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("failed to send task message: %w", err)
	}

	// Wait for plan if required
	if task.WaitForPlan {
		o.setState(StateWaitingPlan)
		_, err := o.monitor.WaitForPlan(ctx)
		if err != nil {
			record.Success = false
			record.Error = err.Error()
			record.EndTime = time.Now()
			record.Duration = time.Since(startTime)
			o.addExecutionRecord(record)

			result.Error = err
			result.Duration = time.Since(startTime)
			return result, fmt.Errorf("failed to wait for plan: %w", err)
		}

		// Approve plan if auto-approve or task specifies
		if task.AutoApprove || o.autoApprove {
			if err := o.client.ApprovePlan(ctx, o.sessionID); err != nil {
				record.Success = false
				record.Error = err.Error()
				record.EndTime = time.Now()
				record.Duration = time.Since(startTime)
				o.addExecutionRecord(record)

				result.Error = err
				result.Duration = time.Since(startTime)
				return result, fmt.Errorf("failed to approve plan: %w", err)
			}
		}

		o.setState(StateRunning)
		o.sendProgress(phaseIndex, taskIndex, fmt.Sprintf("Plan approved: %s", task.Name), 50)
	}

	// Monitor task completion by checking for new activities
	// This is a simplified approach - production version would track specific activities
	time.Sleep(DefaultTaskWaitTime) // Allow time for task to start

	// Get latest activities to find task result
	activities, err := o.client.ListActivities(ctx, o.sessionID, 10)
	if err == nil && len(activities) > 0 {
		// Use the most recent activity
		latestActivity := activities[0]
		record.ActivityID = latestActivity.ID
		record.ArtifactCount = len(latestActivity.Artifacts)

		result.ActivityID = latestActivity.ID
		result.Artifacts = latestActivity.Artifacts
	}

	// Run validation if provided
	if task.Validation != nil {
		if err := task.Validation(result); err != nil {
			record.Success = false
			record.Error = fmt.Sprintf("validation failed: %v", err)
			record.EndTime = time.Now()
			record.Duration = time.Since(startTime)
			o.addExecutionRecord(record)

			result.Error = err
			result.Duration = time.Since(startTime)
			return result, fmt.Errorf("task validation failed: %w", err)
		}
	}

	record.Success = true
	record.EndTime = time.Now()
	record.Duration = time.Since(startTime)
	o.addExecutionRecord(record)

	result.Success = true
	result.Duration = time.Since(startTime)

	o.sendProgress(phaseIndex, taskIndex, fmt.Sprintf("Completed: %s", task.Name), 100)

	return result, nil
}

// executeTasksParallel executes tasks in parallel
func (o *SessionOrchestrator) executeTasksParallel(ctx context.Context, phaseIndex int, phase Phase) (PhaseResult, error) {
	result := PhaseResult{
		PhaseIndex: phaseIndex,
		PhaseName:  phase.Name,
		Success:    true,
		Tasks:      make([]TaskResult, len(phase.Tasks)),
	}

	startTime := time.Now()

	var wg sync.WaitGroup
	resultChan := make(chan struct {
		index  int
		result TaskResult
		err    error
	}, len(phase.Tasks))

	for j, task := range phase.Tasks {
		wg.Add(1)
		go func(index int, t Task) {
			defer wg.Done()
			taskResult, err := o.executeTask(ctx, phaseIndex, index, t)
			resultChan <- struct {
				index  int
				result TaskResult
				err    error
			}{index, taskResult, err}
		}(j, task)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var firstError error
	for res := range resultChan {
		result.Tasks[res.index] = res.result
		if res.err != nil && firstError == nil {
			firstError = res.err
			result.Success = false
		}
	}

	result.Duration = time.Since(startTime)
	result.Error = firstError

	return result, firstError
}

// buildInitialPrompt creates a comprehensive initial prompt from workflow
func (o *SessionOrchestrator) buildInitialPrompt() string {
	prompt := fmt.Sprintf("Execute workflow: %s\n\n%s\n\n", o.workflow.Name, o.workflow.Description)

	prompt += "This is a multi-phase workflow that will be executed progressively. "
	prompt += "Please be ready to receive follow-up messages for each phase.\n\n"

	// Add first phase details
	if len(o.workflow.Phases) > 0 {
		firstPhase := o.workflow.Phases[0]
		prompt += fmt.Sprintf("Starting with Phase 1: %s\n%s\n\n", firstPhase.Name, firstPhase.Description)

		if len(firstPhase.Tasks) > 0 {
			prompt += "Initial tasks:\n"
			for _, task := range firstPhase.Tasks {
				prompt += fmt.Sprintf("- %s: %s\n", task.Name, task.Description)
			}
		}
	}

	return prompt
}

// Helper methods

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
		// Channel full, skip update
	}
}

func (o *SessionOrchestrator) addExecutionRecord(record ExecutionRecord) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.executionLog = append(o.executionLog, record)
}

func (o *SessionOrchestrator) handleProgress(status *jules.SessionStatus) {
	// Handle session progress updates
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
	// Check if prerequisite phases have completed
	// This would need to track completed phase names
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

// Public methods for monitoring and control

// GetProgress returns the current progress
func (o *SessionOrchestrator) GetProgress() (int, int, OrchestratorState) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.currentPhase, o.totalPhases, o.state
}

// GetSessionID returns the current session ID
func (o *SessionOrchestrator) GetSessionID() string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.sessionID
}

// GetExecutionLog returns the execution log
func (o *SessionOrchestrator) GetExecutionLog() []ExecutionRecord {
	o.mu.RLock()
	defer o.mu.RUnlock()
	logCopy := make([]ExecutionRecord, len(o.executionLog))
	copy(logCopy, o.executionLog)
	return logCopy
}

// ProgressChannel returns the progress update channel
func (o *SessionOrchestrator) ProgressChannel() <-chan ProgressUpdate {
	return o.progressChan
}

// ActivityChannel returns the activity update channel
func (o *SessionOrchestrator) ActivityChannel() <-chan ActivityUpdate {
	return o.activityChan
}

// ErrorChannel returns the error channel
func (o *SessionOrchestrator) ErrorChannel() <-chan error {
	return o.errorChan
}

// Stop gracefully stops the orchestrator
func (o *SessionOrchestrator) Stop() {
	o.setState(StateCancelled)
	close(o.stopChan)
}

// Pause pauses the orchestrator
func (o *SessionOrchestrator) Pause() {
	o.setState(StatePaused)
}

// Resume resumes the orchestrator
func (o *SessionOrchestrator) Resume() {
	o.setState(StateRunning)
}
