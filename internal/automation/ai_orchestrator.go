package automation

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/SamyRai/juleson/internal/gemini"
	"github.com/SamyRai/juleson/pkg/jules"
)

// AIOrchestrator uses Gemini AI to intelligently orchestrate complex workflows
// The AI decides:
// - What tasks to execute and in what order
// - When to review and when to auto-approve
// - How to adapt based on progress and results
// - When the workflow is complete
type AIOrchestrator struct {
	julesClient  *jules.Client
	geminiClient *gemini.Client
	sessionID    string

	// Workflow state
	goal        string
	projectPath string
	constraints []string
	context     *ProjectContext

	// Execution state
	currentPhase   string
	completedTasks []CompletedTask
	pendingTasks   []PendingTask

	// AI decision making
	lastAIDecision  *AIDecision
	decisionHistory []AIDecision

	// Monitoring
	state        AIState
	startTime    time.Time
	mu           sync.RWMutex
	progressChan chan AIProgress
	decisionChan chan AIDecision
	stopChan     chan struct{}

	// Configuration
	maxIterations int
	checkInterval time.Duration
	allowedTools  []string
	taskExecutor  *aiTaskExecutor
}

// AIState represents the current state of AI orchestration
type AIState string

const (
	AIStateAnalyzing AIState = "ANALYZING"
	AIStatePlanning  AIState = "PLANNING"
	AIStateExecuting AIState = "EXECUTING"
	AIStateReviewing AIState = "REVIEWING"
	AIStateAdapting  AIState = "ADAPTING"
	AIStateCompleted AIState = "COMPLETED"
	AIStateFailed    AIState = "FAILED"
)

// ProjectContext contains AI's understanding of the project
type ProjectContext struct {
	Languages     []string
	Frameworks    []string
	Architecture  string
	Complexity    string
	CurrentState  string
	KeyFiles      []string
	Dependencies  map[string]string
	TestCoverage  float64
	Issues        []string
	Opportunities []string
}

// CompletedTask represents a task that has been completed
type CompletedTask struct {
	Name           string
	Description    string
	Result         AITaskResult
	Timestamp      time.Time
	Artifacts      []jules.Artifact
	LessonsLearned []string
}

// PendingTask represents a task the AI wants to execute
type PendingTask struct {
	Name         string
	Description  string
	Prompt       string
	Priority     int
	Dependencies []string
	Rationale    string
}

// AITaskResult represents the result of executing an AI-orchestrated task
type AITaskResult struct {
	Success      bool
	ActivityID   string
	FilesChanged []string
	TestsPassed  bool
	Errors       []string
	Warnings     []string
}

// AIDecision represents a decision made by the AI
type AIDecision struct {
	Timestamp    time.Time
	DecisionType string // "next_task", "review_needed", "adapt_plan", "complete"
	Reasoning    string
	Action       string
	Confidence   float64
	Alternatives []string
}

// AIProgress represents progress update from AI orchestration
type AIProgress struct {
	Phase       string
	CurrentTask string
	Progress    float64
	Message     string
	Timestamp   time.Time
	NextSteps   []string
}

// AIOrchestrationConfig configures the AI orchestrator
type AIOrchestrationConfig struct {
	MaxIterations  int
	CheckInterval  time.Duration
	AllowedTools   []string
	AutoApprove    bool
	MaxSessionTime time.Duration
}

// DefaultAIOrchestrationConfig returns default configuration
func DefaultAIOrchestrationConfig() *AIOrchestrationConfig {
	return &AIOrchestrationConfig{
		MaxIterations: 20,
		CheckInterval: 15 * time.Second,
		AllowedTools: []string{
			"execute_template",
			"run_tests",
			"apply_patches",
			"create_issue",
			"create_milestone",
		},
		AutoApprove:    false,
		MaxSessionTime: 4 * time.Hour,
	}
}

// NewAIOrchestrator creates a new AI-powered orchestrator
func NewAIOrchestrator(
	julesClient *jules.Client,
	geminiClient *gemini.Client,
	config *AIOrchestrationConfig,
) *AIOrchestrator {
	if config == nil {
		config = DefaultAIOrchestrationConfig()
	}

	orchestrator := &AIOrchestrator{
		julesClient:     julesClient,
		geminiClient:    geminiClient,
		state:           AIStateAnalyzing,
		progressChan:    make(chan AIProgress, 100),
		decisionChan:    make(chan AIDecision, 100),
		stopChan:        make(chan struct{}),
		maxIterations:   config.MaxIterations,
		checkInterval:   config.CheckInterval,
		allowedTools:    config.AllowedTools,
		completedTasks:  make([]CompletedTask, 0),
		pendingTasks:    make([]PendingTask, 0),
		decisionHistory: make([]AIDecision, 0),
	}
	orchestrator.taskExecutor = newAITaskExecutor(julesClient, orchestrator.GetSessionID)
	return orchestrator
}

// Execute runs the AI orchestration for the given goal
func (ao *AIOrchestrator) Execute(ctx context.Context, sourceID, goal, projectPath string, constraints []string) error {
	ao.mu.Lock()
	ao.goal = goal
	ao.projectPath = projectPath
	ao.constraints = constraints
	ao.startTime = time.Now()
	ao.mu.Unlock()

	// Phase 1: AI analyzes the project
	ao.setState(AIStateAnalyzing)
	ao.sendProgress("Analyzing project", "AI is understanding your codebase", 0, nil)

	if err := ao.analyzeProject(ctx); err != nil {
		return fmt.Errorf("project analysis failed: %w", err)
	}

	// Phase 2: AI creates initial plan
	ao.setState(AIStatePlanning)
	ao.sendProgress("Planning workflow", "AI is creating execution plan", 10, nil)

	if err := ao.createInitialPlan(ctx); err != nil {
		return fmt.Errorf("planning failed: %w", err)
	}

	// Phase 3: Create Jules session with AI-generated comprehensive prompt
	ao.setState(AIStateExecuting)
	initialPrompt := ao.buildAIPrompt()

	session, err := ao.julesClient.CreateSession(ctx, &jules.CreateSessionRequest{
		Prompt: initialPrompt,
		SourceContext: &jules.SourceContext{
			Source: fmt.Sprintf("sources/%s", sourceID),
		},
		RequirePlanApproval: true, // Always review AI decisions
	})
	if err != nil {
		ao.setState(AIStateFailed)
		return fmt.Errorf("failed to create session: %w", err)
	}

	ao.mu.Lock()
	ao.sessionID = session.ID
	ao.mu.Unlock()

	ao.sendProgress("Session started", fmt.Sprintf("Session %s created", session.ID), 20, nil)

	// Phase 4: AI-driven execution loop
	return ao.executionLoop(ctx)
}

// analyzeProject uses AI to deeply understand the project
func (ao *AIOrchestrator) analyzeProject(ctx context.Context) error {
	// Build analysis prompt for Gemini
	analysisPrompt := fmt.Sprintf(`Analyze this software project to understand its context:

Project Path: %s
User Goal: %s
Constraints: %s

Please analyze:
1. Programming languages and frameworks used
2. Current architecture and design patterns
3. Code quality and technical debt
4. Test coverage and quality
5. Key areas that need attention
6. Potential risks and challenges
7. Recommended approach to achieve the goal

Provide a structured JSON response with your analysis.`,
		ao.projectPath,
		ao.goal,
		strings.Join(ao.constraints, ", "),
	)

	resp, err := ao.geminiClient.GenerateContent("", analysisPrompt)
	if err != nil {
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	// Parse AI's analysis
	analysis := extractAnalysisFromResponse(resp)

	ao.mu.Lock()
	ao.context = analysis
	ao.mu.Unlock()

	return nil
}

// createInitialPlan asks AI to create an execution plan
func (ao *AIOrchestrator) createInitialPlan(ctx context.Context) error {
	planningPrompt := fmt.Sprintf(`Based on the project analysis, create a detailed execution plan to achieve this goal:

Goal: %s

Project Context:
- Languages: %s
- Architecture: %s
- Complexity: %s
- Current Issues: %s

Constraints: %s

Create a plan with:
1. High-priority tasks that should be executed first
2. Dependencies between tasks
3. Estimated effort for each task
4. Risk assessment
5. Success criteria

Be adaptive - the plan may change based on execution results.
Provide a structured JSON response with tasks prioritized and sequenced.`,
		ao.goal,
		strings.Join(ao.context.Languages, ", "),
		ao.context.Architecture,
		ao.context.Complexity,
		strings.Join(ao.context.Issues, "; "),
		strings.Join(ao.constraints, ", "),
	)

	resp, err := ao.geminiClient.GenerateContent("", planningPrompt)
	if err != nil {
		return fmt.Errorf("AI planning failed: %w", err)
	}

	// Parse AI's plan into pending tasks
	tasks := extractTasksFromResponse(resp)

	ao.mu.Lock()
	ao.pendingTasks = tasks
	ao.mu.Unlock()

	return nil
}

// executionLoop is the main AI-driven execution loop
func (ao *AIOrchestrator) executionLoop(ctx context.Context) error {
	iteration := 0

	for iteration < ao.maxIterations {
		iteration++

		// Check if we should stop
		select {
		case <-ao.stopChan:
			ao.setState(AIStateFailed)
			return fmt.Errorf("orchestration stopped")
		case <-ctx.Done():
			ao.setState(AIStateFailed)
			return ctx.Err()
		default:
		}

		// AI decides what to do next
		decision, err := ao.makeNextDecision(ctx)
		if err != nil {
			return fmt.Errorf("AI decision failed: %w", err)
		}

		ao.recordDecision(decision)

		// Execute based on AI decision
		switch decision.DecisionType {
		case "next_task":
			if err := ao.executeNextTask(ctx, decision); err != nil {
				return fmt.Errorf("task execution failed: %w", err)
			}

		case "review_needed":
			ao.setState(AIStateReviewing)
			if err := ao.reviewAndAdapt(ctx, decision); err != nil {
				return fmt.Errorf("review failed: %w", err)
			}

		case "adapt_plan":
			ao.setState(AIStateAdapting)
			if err := ao.adaptPlan(ctx, decision); err != nil {
				return fmt.Errorf("adaptation failed: %w", err)
			}

		case "complete":
			ao.setState(AIStateCompleted)
			ao.sendProgress("Workflow complete", decision.Reasoning, 100, nil)
			return nil

		default:
			return fmt.Errorf("unknown decision type: %s", decision.DecisionType)
		}

		// Brief pause before next iteration
		time.Sleep(ao.checkInterval)
	}

	return fmt.Errorf("max iterations (%d) reached", ao.maxIterations)
}

// makeNextDecision asks AI what to do next
func (ao *AIOrchestrator) makeNextDecision(ctx context.Context) (*AIDecision, error) {
	ao.mu.RLock()
	completedCount := len(ao.completedTasks)
	pendingCount := len(ao.pendingTasks)
	ao.mu.RUnlock()

	// Get latest session status
	session, err := ao.julesClient.GetSession(ctx, ao.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Build decision prompt for AI
	decisionPrompt := fmt.Sprintf(`You are orchestrating a software development workflow. Make the next decision:

Original Goal: %s
Session State: %s
Completed Tasks: %d
Pending Tasks: %d

Recent Completed Tasks:
%s

Current Pending Tasks:
%s

What should happen next? Choose one:
1. "next_task" - Execute the next task from pending list
2. "review_needed" - Review progress and get user feedback
3. "adapt_plan" - Adjust the plan based on results
4. "complete" - Goal is achieved, workflow complete

Provide JSON response with:
- decision_type: (next_task|review_needed|adapt_plan|complete)
- reasoning: Why this decision
- action: Specific action to take
- confidence: 0.0-1.0
- next_steps: What happens after this`,
		ao.goal,
		session.State,
		completedCount,
		pendingCount,
		ao.formatCompletedTasks(),
		ao.formatPendingTasks(),
	)

	resp, err := ao.geminiClient.GenerateContent("", decisionPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI decision generation failed: %w", err)
	}

	// Parse AI's decision
	decision := extractDecisionFromResponse(resp)
	decision.Timestamp = time.Now()

	ao.sendDecision(*decision)

	return decision, nil
}

// executeNextTask executes the task AI decided on
func (ao *AIOrchestrator) executeNextTask(ctx context.Context, decision *AIDecision) error {
	ao.setState(AIStateExecuting)

	// Get next task from pending list
	ao.mu.Lock()
	if len(ao.pendingTasks) == 0 {
		ao.mu.Unlock()
		return fmt.Errorf("no pending tasks")
	}
	task := ao.pendingTasks[0]
	ao.pendingTasks = ao.pendingTasks[1:]
	ao.mu.Unlock()

	ao.sendProgress("Executing task", task.Name,
		float64(len(ao.completedTasks))/float64(len(ao.completedTasks)+len(ao.pendingTasks)+1)*100,
		[]string{task.Description},
	)

	completed, err := ao.taskExecutor.execute(ctx, task)
	if err != nil {
		return err
	}

	ao.mu.Lock()
	ao.completedTasks = append(ao.completedTasks, completed)
	ao.mu.Unlock()

	return nil
}

// reviewAndAdapt asks AI to review progress and adapt
func (ao *AIOrchestrator) reviewAndAdapt(ctx context.Context, decision *AIDecision) error {
	// AI reviews completed tasks and decides on adaptations
	reviewPrompt := fmt.Sprintf(`Review the progress and decide if we should adapt the plan:

Original Goal: %s
Completed Tasks: %d
Tasks Summary:
%s

Should we:
1. Continue with current plan?
2. Adjust remaining tasks based on learnings?
3. Add new tasks we discovered are needed?
4. Skip tasks that are no longer relevant?

Provide detailed reasoning and updated task list if needed.`,
		ao.goal,
		len(ao.completedTasks),
		ao.formatCompletedTasksDetailed(),
	)

	resp, err := ao.geminiClient.GenerateContent("", reviewPrompt)
	if err != nil {
		return fmt.Errorf("review failed: %w", err)
	}

	// Parse AI's review and apply changes
	adaptations := extractAdaptationsFromResponse(resp)
	ao.applyAdaptations(adaptations)

	return nil
}

// adaptPlan adapts the plan based on AI's decisions
func (ao *AIOrchestrator) adaptPlan(ctx context.Context, decision *AIDecision) error {
	// Similar to reviewAndAdapt but focused on plan changes
	return ao.reviewAndAdapt(ctx, decision)
}
