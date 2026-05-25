package domain

import "time"

// Goal describes the outcome an orchestration run should achieve.
type Goal struct {
	ID          string
	Description string
	Constraints []string
	Context     GoalContext
	Priority    Priority
	Deadline    *time.Time
}

type GoalContext struct {
	ProjectPath   string
	SourceID      string
	Repository    string
	Branch        string
	RelatedIssues []string
	RelatedPRs    []string
	Values        map[string]string
}

type Priority string

const (
	PriorityCritical Priority = "CRITICAL"
	PriorityHigh     Priority = "HIGH"
	PriorityMedium   Priority = "MEDIUM"
	PriorityLow      Priority = "LOW"
)

type AgentState string

const (
	StateIdle       AgentState = "IDLE"
	StateAnalyzing  AgentState = "ANALYZING"
	StatePlanning   AgentState = "PLANNING"
	StateExecuting  AgentState = "EXECUTING"
	StateReviewing  AgentState = "REVIEWING"
	StateReflecting AgentState = "REFLECTING"
	StateComplete   AgentState = "COMPLETE"
	StateFailed     AgentState = "FAILED"
	StateError      AgentState = "ERROR"
)

type ProjectContext struct {
	ProjectPath  string
	ProjectName  string
	ProjectType  string
	Languages    []string
	Frameworks   []string
	Architecture string
	Complexity   string
	GitStatus    string
	Branch       string
	Dependencies map[string]string
	Quality      *QualityMetrics
	Values       map[string]string
}

type QualityMetrics struct {
	TestCoverage    float64
	CodeComplexity  float64
	Maintainability float64
	SecurityIssues  int
	CodeSmells      int
}

type Plan struct {
	ID        string
	Goal      Goal
	Tasks     []Task
	Reasoning string
	CreatedAt time.Time
	Metadata  map[string]string
}

type Task struct {
	ID               string
	Name             string
	Description      string
	Prompt           string
	Type             string
	Priority         Priority
	Dependencies     []string
	Tool             string
	Context          map[string]string
	State            TaskState
	RequiresApproval bool
	Retry            int
	Timeout          time.Duration
	Result           *TaskResult
}

type TaskState string

const (
	TaskStatePending    TaskState = "PENDING"
	TaskStateInProgress TaskState = "IN_PROGRESS"
	TaskStateReviewing  TaskState = "REVIEWING"
	TaskStateComplete   TaskState = "COMPLETE"
	TaskStateFailed     TaskState = "FAILED"
	TaskStateSkipped    TaskState = "SKIPPED"
)

type TaskResult struct {
	TaskID      string
	TaskName    string
	TaskType    string
	Success     bool
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Tool        string
	Changes     []Change
	Artifacts   []Artifact
	SessionID   string
	ActivityID  string
	Output      string
	Error       error
	Metrics     map[string]any
	Review      *ReviewResult
	Diagnostics []Diagnostic
}

type Result struct {
	Goal      Goal
	Success   bool
	State     AgentState
	Plan      *Plan
	Tasks     []TaskResult
	Artifacts []Artifact
	Duration  time.Duration
	Error     error
	Summary   string
	Learnings []string
}

type Change struct {
	FilePath    string
	Type        ChangeType
	Additions   int
	Deletions   int
	Patch       string
	Description string
}

type ChangeType string

const (
	ChangeTypeAdd    ChangeType = "ADD"
	ChangeTypeModify ChangeType = "MODIFY"
	ChangeTypeDelete ChangeType = "DELETE"
	ChangeTypeRename ChangeType = "RENAME"
)

type Artifact struct {
	Type        ArtifactType
	Path        string
	Content     string
	Description string
	Metadata    map[string]string
}

type ArtifactType string

const (
	ArtifactTypeCode          ArtifactType = "CODE"
	ArtifactTypeTest          ArtifactType = "TEST"
	ArtifactTypeDocumentation ArtifactType = "DOCUMENTATION"
	ArtifactTypeConfiguration ArtifactType = "CONFIGURATION"
	ArtifactTypeReport        ArtifactType = "REPORT"
)

type ReviewResult struct {
	Approved         bool
	ChangesRequested bool
	Score            float64
	Summary          string
	Diagnostics      []Diagnostic
}

type Diagnostic struct {
	Severity   Severity
	Category   string
	Message    string
	Suggestion string
	Location   *Location
}

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

type Location struct {
	FilePath  string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}

type Decision struct {
	ID           string
	Timestamp    time.Time
	State        AgentState
	Type         DecisionType
	Reasoning    string
	Action       string
	Confidence   float64
	Alternatives []string
	NextSteps    []string
	TaskID       string
	Outcome      *DecisionOutcome
}

type DecisionType string

const (
	DecisionTypeNextTask      DecisionType = "NEXT_TASK"
	DecisionTypeReviewNeeded  DecisionType = "REVIEW_NEEDED"
	DecisionTypeAdaptPlan     DecisionType = "ADAPT_PLAN"
	DecisionTypeComplete      DecisionType = "COMPLETE"
	DecisionTypeRetry         DecisionType = "RETRY"
	DecisionTypeAbort         DecisionType = "ABORT"
	DecisionTypeApprove       DecisionType = "APPROVE"
	DecisionTypeRequestChange DecisionType = "REQUEST_CHANGE"
)

type DecisionOutcome struct {
	Success   bool
	Duration  time.Duration
	Result    string
	Error     error
	Learnings []string
}

type Progress struct {
	State          AgentState
	CurrentTask    string
	CompletedTasks int
	TotalTasks     int
	Progress       float64
	Message        string
	NextSteps      []string
	Timestamp      time.Time
}

type Workflow struct {
	ID          string
	Name        string
	Description string
	Goal        Goal
	Phases      []Phase
	MaxDuration time.Duration
	Metadata    map[string]string
}

type Phase struct {
	ID              string
	Name            string
	Description     string
	Tasks           []Task
	Parallel        bool
	ContinueOnError bool
	Timeout         time.Duration
	Prerequisites   []string
}

type WorkflowResult struct {
	WorkflowName  string
	Success       bool
	TotalPhases   int
	PhaseResults  []PhaseResult
	TotalDuration time.Duration
	SessionID     string
	StartTime     time.Time
	EndTime       time.Time
	Error         error
}

type PhaseResult struct {
	PhaseIndex int
	PhaseName  string
	Success    bool
	Tasks      []TaskResult
	Duration   time.Duration
	Error      error
}

type ExecutionContext struct {
	Goal             Goal
	Project          *ProjectContext
	Plan             *Plan
	Workflow         *Workflow
	Completed        []TaskResult
	Decisions        []Decision
	SessionID        string
	Iteration        int
	StartedAt        time.Time
	Values           map[string]string
	ApprovalPolicy   ApprovalPolicy
	DryRun           bool
	ReviewStrictness string
}

type ApprovalPolicy struct {
	RequirePlanApproval bool
	AutoApprove         bool
}

type Checkpoint struct {
	ID        string
	GoalID    string
	State     AgentState
	Context   ExecutionContext
	CreatedAt time.Time
	Metadata  map[string]string
}

type Template struct {
	Name        string
	Description string
	Tasks       []Task
	OutputFiles []OutputFile
	Metadata    map[string]string
}

type OutputFile struct {
	Path     string
	Template string
}

type Source struct {
	ID         string
	Name       string
	Repository string
	URL        string
	Metadata   map[string]string
}

type SessionRequest struct {
	Prompt              string
	Title               string
	Source              Source
	Branch              string
	RequirePlanApproval bool
	AutomationMode      string
	Metadata            map[string]string
}

type Session struct {
	ID       string
	Name     string
	Title    string
	URL      string
	State    string
	Source   Source
	Metadata map[string]string
}
