package automation

import (
	"sync"
	"time"

	"github.com/SamyRai/go-jules"
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

// SessionOrchestrator is a legacy compatibility surface. New CLI and MCP
// workflows should use internal/orchestration.Runtime and SessionWorkflowRunner.
//
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
