package core

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
	"github.com/SamyRai/juleson/internal/agent/memory"
	"github.com/SamyRai/juleson/internal/agent/review"
	"github.com/SamyRai/juleson/internal/agent/tools"
	"github.com/SamyRai/juleson/internal/analyzer"
	"github.com/SamyRai/juleson/internal/gemini"
)

// Default agent configuration constants
const (
	DefaultMaxIterations      = 20
	DefaultCheckpointInterval = 5 * time.Minute
	DefaultMinTestCoverage    = 0.8
	// DefaultConfidenceThreshold is the default confidence threshold for initial observations
	DefaultConfidenceThreshold = 0.5
	// PercentageScale is used to convert scores to percentage (0-1 range)
	PercentageScale = 100.0
)

// CoreAgent implements the main agent loop
type CoreAgent struct {
	state        agent.AgentState
	toolRegistry tools.ToolRegistry
	reviewer     review.Reviewer
	memory       memory.Memory
	logger       *slog.Logger
	analyzer     *analyzer.ProjectAnalyzer

	// New production-ready components
	planner       *Planner
	retryStrategy *RetryStrategy
	checkpointMgr *CheckpointManager
	telemetry     *Metrics
	validator     *ConstraintValidator
	geminiClient  *gemini.Client

	// Current execution context
	currentGoal    *agent.Goal
	currentPlan    []agent.Task
	decisions      []agent.Decision
	projectContext *analyzer.ProjectContext

	// Configuration
	maxIterations int
	dryRun        bool
}

// Config holds agent configuration
type Config struct {
	MaxIterations   int
	DryRun          bool
	ReviewConfig    *review.Config
	Logger          *slog.Logger
	GeminiClient    *gemini.Client
	CheckpointDir   string
	AutoSave        bool
	SaveInterval    time.Duration
	RetryConfig     *RetryStrategy
	EnableTelemetry bool
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxIterations:   DefaultMaxIterations,
		DryRun:          false,
		ReviewConfig:    review.DefaultConfig(),
		Logger:          slog.Default(),
		CheckpointDir:   "./checkpoints",
		AutoSave:        true,
		SaveInterval:    DefaultCheckpointInterval,
		RetryConfig:     DefaultRetryStrategy(),
		EnableTelemetry: true,
	}
}

// NewAgent creates a new core agent
func NewAgent(toolRegistry tools.ToolRegistry, config *Config) agent.Agent {
	if config == nil {
		config = DefaultConfig()
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if config.RetryConfig == nil {
		config.RetryConfig = DefaultRetryStrategy()
	}

	agent := &CoreAgent{
		state:         agent.StateIdle,
		toolRegistry:  toolRegistry,
		reviewer:      review.NewReviewer(config.ReviewConfig),
		memory:        memory.NewMemory(),
		logger:        config.Logger,
		analyzer:      analyzer.NewProjectAnalyzer(),
		maxIterations: config.MaxIterations,
		dryRun:        config.DryRun,
		decisions:     make([]agent.Decision, 0),
	}

	// Initialize new components
	if config.GeminiClient != nil {
		agent.geminiClient = config.GeminiClient
		agent.planner = NewPlanner(config.GeminiClient, config.Logger)
	}

	agent.retryStrategy = config.RetryConfig
	agent.checkpointMgr = NewCheckpointManager(config.CheckpointDir, config.AutoSave, config.SaveInterval, config.Logger)

	if config.EnableTelemetry {
		agent.telemetry = NewMetrics()
	}

	// Initialize validator with empty constraints (can be set later)
	agent.validator = NewConstraintValidator([]string{})

	return agent
}

// Execute implements the main agent loop: Perceive → Plan → Act → Review → Reflect
func (a *CoreAgent) Execute(ctx context.Context, goal agent.Goal) (*agent.Result, error) {
	a.currentGoal = &goal
	startTime := time.Now()

	a.logger.Info("agent.execute.start",
		"goal_id", goal.ID,
		"description", goal.Description,
		"priority", goal.Priority)

	// Start auto-save if configured
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

	// Record execution start in telemetry
	if a.telemetry != nil {
		a.telemetry.RecordExecution(false, 0) // Will update on completion
	}

	// Main agent loop
	iteration := 0
	for {
		iteration++

		a.logger.Info("agent.iteration.start", "iteration", iteration, "state", a.state)

		// Check if we've reached a terminal state BEFORE executing
		if a.state == agent.StateComplete || a.state == agent.StateFailed {
			a.logger.Info("agent.terminal_state_reached", "state", a.state, "iteration", iteration)
			break
		}

		// Check if we've reached max iterations (only if not in terminal state)
		if iteration > a.maxIterations {
			a.logger.Warn("agent.max_iterations_reached", "iterations", iteration-1)
			return a.finalizeResult(result, startTime, fmt.Errorf("max iterations (%d) reached", a.maxIterations))
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return a.finalizeResult(result, startTime, fmt.Errorf("execution cancelled: %w", ctx.Err()))
		default:
		}

		// Execute current state with retry logic
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

	// Loop exited - check if it was due to terminal state or max iterations
	if a.state == agent.StateComplete || a.state == agent.StateFailed {
		// Successful completion or controlled failure
		return a.finalizeResult(result, startTime, nil)
	} else {
		// Max iterations reached (should not happen due to check inside loop, but safety net)
		a.logger.Warn("agent.max_iterations_reached", "iterations", iteration)
		return a.finalizeResult(result, startTime, fmt.Errorf("max iterations (%d) reached", a.maxIterations))
	}
}

// executeState executes the current agent state
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
		// Check if we need more iterations
		if a.needsMoreWork(result) {
			a.setState(agent.StatePlanning) // Adapt plan
		} else {
			a.setState(agent.StateComplete)
		}
		return nil
	case agent.StateComplete:
		result.Success = true
		result.Duration = time.Since(time.Now()) // This will be overridden in finalizeResult
		a.logger.Info("agent.execute.complete")
		return nil
	case agent.StateFailed:
		return fmt.Errorf("agent execution failed")
	default:
		return fmt.Errorf("unknown state: %s", a.state)
	}
}

// perceive gathers context about the goal and environment
func (a *CoreAgent) perceive(ctx context.Context, goal agent.Goal) error {
	a.logger.Info("agent.perceive.start")

	// Analyze project if project path is provided
	if goal.Context.ProjectPath != "" {
		a.logger.Info("agent.perceive.analyzing_project", "path", goal.Context.ProjectPath)

		projectContext, err := a.analyzer.Analyze(goal.Context.ProjectPath)
		if err != nil {
			a.logger.Warn("agent.perceive.project_analysis_failed", "error", err)
			// Continue without project context
		} else {
			a.projectContext = projectContext
			a.logger.Info("agent.perceive.project_analyzed",
				"name", projectContext.ProjectName,
				"type", projectContext.ProjectType,
				"languages", len(projectContext.Languages),
				"frameworks", len(projectContext.Frameworks))
		}
	}

	// Recall relevant learnings from memory
	learnings, err := a.memory.Recall(ctx, goal.Description)
	if err != nil {
		a.logger.Warn("agent.perceive.memory_recall_failed", "error", err)
	} else {
		a.logger.Info("agent.perceive.recalled_learnings", "count", len(learnings))
	}

	// Record perception decision
	decision := agent.Decision{
		Timestamp:  time.Now(),
		State:      agent.StateAnalyzing,
		Type:       agent.DecisionTypeSelectTool,
		Reasoning:  fmt.Sprintf("Perceived goal: %s. Found %d relevant learnings.", goal.Description, len(learnings)),
		Confidence: 0.8,
	}
	a.recordDecision(decision)

	// Transition to planning
	a.setState(agent.StateAnalyzing)
	return nil
}

// plan creates a multi-step execution plan
func (a *CoreAgent) plan(ctx context.Context, goal agent.Goal) error {
	a.logger.Info("agent.plan.start")

	// Use AI-powered planner if available
	if a.planner != nil {
		codebaseContext := "Go project with agent, tools, and automation components" // TODO: Extract real codebase context

		tasks, reasoning, err := a.planner.GeneratePlan(ctx, goal, codebaseContext, a.projectContext)
		if err != nil {
			a.logger.Warn("agent.plan.ai_failed", "error", err)
			// Fall back to simple planning
		} else {
			a.currentPlan = tasks

			// Record planning decision with reasoning
			decision := agent.Decision{
				Timestamp:  time.Now(),
				State:      agent.StatePlanning,
				Type:       agent.DecisionTypeSelectTool,
				Reasoning:  reasoning.ChainOfThought[0],
				Action:     fmt.Sprintf("Generated AI plan with %d task(s)", len(a.currentPlan)),
				Confidence: reasoning.Confidence,
			}
			a.recordDecision(decision)

			// Record telemetry
			if a.telemetry != nil {
				a.telemetry.RecordDecision(agent.DecisionTypeAdaptPlan, time.Since(time.Now()))
			}

			a.logger.Info("agent.plan.ai_complete", "tasks", len(a.currentPlan))
			a.setState(agent.StatePlanning)
			return nil
		}
	}

	// Fallback to simple planning
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

// act executes tasks using appropriate tools
func (a *CoreAgent) act(ctx context.Context) error {
	a.logger.Info("agent.act.start", "tasks", len(a.currentPlan))

	if len(a.currentPlan) == 0 {
		a.setState(agent.StateReviewing)
		return nil
	}

	// Find next pending task
	var currentTask *agent.Task
	for i := range a.currentPlan {
		if a.currentPlan[i].State == agent.TaskStatePending {
			currentTask = &a.currentPlan[i]
			currentTask.State = agent.TaskStateInProgress
			break
		}
	}

	if currentTask == nil {
		// All tasks complete, move to review
		a.setState(agent.StateExecuting)
		return nil
	}

	a.logger.Info("agent.act.task", "task_id", currentTask.ID, "name", currentTask.Name)

	// Delegate to tool execution
	taskResult, err := a.executeSingleTask(ctx, currentTask)
	if err != nil {
		currentTask.State = agent.TaskStateFailed
		return err
	}

	currentTask.State = agent.TaskStateComplete
	currentTask.Result = taskResult

	// Move to review state
	a.setState(agent.StateExecuting)
	return nil
}

// executeSingleTask executes a single task using the appropriate tool
// This method encapsulates all tool interaction logic
func (a *CoreAgent) executeSingleTask(ctx context.Context, task *agent.Task) (*agent.TaskResult, error) {
	// Find appropriate tool
	matchingTools := a.toolRegistry.FindForTask(*task)
	if len(matchingTools) == 0 {
		return nil, fmt.Errorf("no tool found for task: %s", task.Name)
	}

	tool := matchingTools[0] // Use first matching tool

	// Prepare task result
	taskResult := &agent.TaskResult{
		TaskID:  task.ID,
		Name:    task.Name,
		Success: false,
		Tool:    tool.Name(),
		Changes: make([]agent.Change, 0),
	}

	// Handle dry-run mode
	if a.dryRun {
		a.logger.Info("agent.act.dry_run", "task_id", task.ID)
		taskResult.Success = true
		taskResult.Duration = time.Second

		// Record telemetry
		if a.telemetry != nil {
			a.telemetry.RecordToolInvocation(tool.Name(), true, time.Second)
		}

		return taskResult, nil
	}

	// Prepare parameters for tool execution
	params := a.prepareToolParameters(task)

	// Execute tool with telemetry tracking
	startTime := time.Now()
	toolResult, err := tool.Execute(ctx, params)
	taskResult.Duration = time.Since(startTime)

	if err != nil {
		taskResult.Error = err
		a.logger.Error("agent.act.tool_failed", "error", err)

		// Record telemetry
		if a.telemetry != nil {
			a.telemetry.RecordToolInvocation(tool.Name(), false, taskResult.Duration)
		}

		return taskResult, err
	}

	// Process tool results
	taskResult.Success = toolResult.Success
	taskResult.Changes = toolResult.Changes

	// Validate constraints on changes
	if a.validator != nil && len(taskResult.Changes) > 0 {
		if err := a.validator.ValidateChanges(taskResult.Changes); err != nil {
			a.logger.Warn("agent.act.constraint_violation", "error", err)
			// Continue but log the violation
		}
	}

	// Record telemetry
	if a.telemetry != nil {
		a.telemetry.RecordToolInvocation(tool.Name(), taskResult.Success, taskResult.Duration)
		a.telemetry.RecordTask(true, taskResult.Success, !taskResult.Success, taskResult.Duration)
	}

	a.logger.Info("agent.act.task_complete",
		"task_id", task.ID,
		"success", taskResult.Success,
		"duration", taskResult.Duration)

	return taskResult, nil
}

// prepareToolParameters prepares parameters for tool execution
func (a *CoreAgent) prepareToolParameters(task *agent.Task) map[string]interface{} {
	params := map[string]interface{}{
		"action": "create_session",
		"prompt": task.Prompt,
	}

	if a.currentGoal != nil && a.currentGoal.Context.SourceID != "" {
		params["source_id"] = a.currentGoal.Context.SourceID
	}

	if a.projectContext != nil {
		params["project_path"] = a.projectContext.ProjectPath
	}

	return params
}

// review performs code review on changes
func (a *CoreAgent) review(ctx context.Context, result *agent.Result) error {
	a.logger.Info("agent.review.start")

	// Collect all changes from completed tasks
	var allChanges []agent.Change
	for _, task := range a.currentPlan {
		if task.Result != nil {
			allChanges = append(allChanges, task.Result.Changes...)
		}
	}

	if len(allChanges) == 0 {
		a.logger.Info("agent.review.no_changes")
		a.setState(agent.StateReflecting)
		return nil
	}

	// Perform review
	reviewResult, err := a.reviewer.Review(ctx, allChanges)
	if err != nil {
		a.logger.Error("agent.review.failed", "error", err)
		return err
	}

	a.logger.Info("agent.review.complete",
		"decision", reviewResult.Decision,
		"score", reviewResult.Score,
		"comments", len(reviewResult.Comments))

	// Store review result
	for i := range a.currentPlan {
		if a.currentPlan[i].Result != nil {
			a.currentPlan[i].Result.ReviewResult = reviewResult
		}
	}

	// Make decision based on review
	decision := agent.Decision{
		Timestamp:  time.Now(),
		State:      agent.StateReviewing,
		Type:       agent.DecisionTypeApprove,
		Reasoning:  reviewResult.Summary,
		Confidence: reviewResult.Score / PercentageScale,
	}

	if reviewResult.Decision == agent.ReviewDecisionReject {
		decision.Type = agent.DecisionTypeReject
		decision.Action = "Reject changes and retry"
		a.recordDecision(decision)

		// Adapt plan: reset tasks and try again
		for i := range a.currentPlan {
			if a.currentPlan[i].State != agent.TaskStateFailed {
				a.currentPlan[i].State = agent.TaskStatePending
			}
		}

		a.setState(agent.StatePlanning)
		return nil
	}

	if reviewResult.Decision == agent.ReviewDecisionRequestChanges {
		decision.Type = agent.DecisionTypeRequestChange
		decision.Action = "Request improvements"
		a.recordDecision(decision)

		// TODO: Provide feedback to Jules and request improvements
		// For now, accept with warnings
	}

	a.recordDecision(decision)
	a.setState(agent.StateReflecting)
	return nil
}

// reflect learns from the execution and decides next steps
func (a *CoreAgent) reflect(ctx context.Context, result *agent.Result) error {
	a.logger.Info("agent.reflect.start")

	// Collect task results
	for _, task := range a.currentPlan {
		if task.Result != nil {
			result.Tasks = append(result.Tasks, *task.Result)
		}
	}

	// Extract learnings
	learning := agent.Learning{
		Timestamp:  time.Now(),
		Context:    a.currentGoal.Description,
		Pattern:    "Jules tool execution",
		Lesson:     fmt.Sprintf("Completed %d tasks", len(result.Tasks)),
		Confidence: 0.7,
	}

	// Adjust confidence based on success
	successCount := 0
	for _, taskResult := range result.Tasks {
		if taskResult.Success {
			successCount++
		}
	}

	if successCount == len(result.Tasks) {
		learning.Confidence = 0.9
		learning.Lesson += " - all tasks succeeded"
	}

	// Store learning
	if err := a.memory.Store(ctx, learning); err != nil {
		a.logger.Warn("agent.reflect.store_learning_failed", "error", err)
	}

	result.Learnings = append(result.Learnings, learning)

	a.logger.Info("agent.reflect.complete", "learnings", len(result.Learnings))
	a.setState(agent.StateReflecting)
	return nil
}

// Helper methods

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
	// Check if any tasks failed
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

	// Record final telemetry
	if a.telemetry != nil {
		a.telemetry.RecordExecution(result.Success, result.Duration)
	}

	return result, err
}

// Agent interface implementation

func (a *CoreAgent) GetState() agent.AgentState {
	return a.state
}

func (a *CoreAgent) GetHistory() []agent.Decision {
	return a.decisions
}

func (a *CoreAgent) Pause() error {
	// TODO: Implement pause functionality
	return fmt.Errorf("pause not implemented")
}

func (a *CoreAgent) Resume() error {
	// TODO: Implement resume functionality
	return fmt.Errorf("resume not implemented")
}

func (a *CoreAgent) Stop() error {
	a.setState(agent.StateFailed)
	return nil
}

func (a *CoreAgent) ProvideFeedback(feedback agent.Feedback) error {
	// TODO: Implement feedback processing
	a.logger.Info("agent.feedback.received", "type", feedback.Type, "message", feedback.Message)
	return nil
}

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

// SetConstraints updates the agent's constraint validator
func (a *CoreAgent) SetConstraints(constraints []string) {
	if a.validator != nil {
		a.validator = NewConstraintValidator(constraints)
	}
}

// GetTelemetrySummary returns telemetry metrics summary
func (a *CoreAgent) GetTelemetrySummary() map[string]interface{} {
	if a.telemetry != nil {
		return a.telemetry.Summary()
	}
	return nil
}

// GetCheckpoints returns available checkpoints
func (a *CoreAgent) GetCheckpoints() ([]Checkpoint, error) {
	if a.checkpointMgr != nil {
		return a.checkpointMgr.List()
	}
	return nil, fmt.Errorf("checkpoint manager not initialized")
}

// RestoreFromCheckpoint restores agent state from a checkpoint
func (a *CoreAgent) RestoreFromCheckpoint(ctx context.Context, checkpointID string) error {
	if a.checkpointMgr != nil {
		return a.checkpointMgr.Restore(ctx, checkpointID, a)
	}
	return fmt.Errorf("checkpoint manager not initialized")
}
