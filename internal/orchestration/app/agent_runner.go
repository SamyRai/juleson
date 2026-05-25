package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/internal/orchestration/ports"
)

const defaultAgentMaxIterations = 20
const defaultReviewStrictness = "medium"

const (
	agentAnalyzeNodeName         = "agent.analyze"
	agentPlanNodeName            = "agent.plan"
	agentOrderTasksNodeName      = "agent.orderTasks"
	agentDryRunOrExecuteNodeName = "agent.maybeDryRunOrExecute"
	agentExecuteTasksNodeName    = "agent.executeTasks"
	agentReviewNodeName          = "agent.review"
	agentCompleteNodeName        = "agent.complete"
)

type AgentRunOptions struct {
	DryRun           bool
	MaxIterations    int
	ReviewStrictness string
}

type AgentRunnerDeps struct {
	ProjectAnalyzer ports.ProjectAnalyzer
	Planner         ports.Planner
	TaskExecutor    ports.TaskExecutor
	Reviewer        ports.Reviewer
	MemoryStore     ports.MemoryStore
	CheckpointStore ports.CheckpointStore
	ProgressSink    ports.ProgressSink
	Clock           ports.Clock
	MaxIterations   int
}

type AgentRunner struct {
	deps      AgentRunnerDeps
	clock     ports.Clock
	scheduler taskScheduler
}

func NewAgentRunner(deps AgentRunnerDeps) *AgentRunner {
	return &AgentRunner{
		deps:  deps,
		clock: clockOrDefault(deps.Clock),
	}
}

func (r *AgentRunner) Run(ctx context.Context, goal domain.Goal) (*domain.Result, error) {
	return r.RunWithOptions(ctx, goal, AgentRunOptions{})
}

func (r *AgentRunner) RunWithOptions(ctx context.Context, goal domain.Goal, opts AgentRunOptions) (*domain.Result, error) {
	if goal.Description == "" {
		return nil, fmt.Errorf("goal description cannot be empty")
	}
	if r.deps.Planner == nil {
		return nil, fmt.Errorf("planner is required")
	}
	if r.deps.TaskExecutor == nil {
		return nil, fmt.Errorf("task executor is required")
	}
	strictness, err := normalizeReviewStrictness(opts.ReviewStrictness)
	if err != nil {
		return nil, err
	}

	start := r.clock.Now()
	state := &appRunState{
		goal:         goal,
		agentOptions: opts,
		startedAt:    start,
		result: &domain.Result{
			Goal:  goal,
			State: domain.StateAnalyzing,
			Tasks: make([]domain.TaskResult, 0),
		},
		execution: domain.ExecutionContext{
			Goal:             goal,
			StartedAt:        start,
			DryRun:           opts.DryRun,
			ReviewStrictness: strictness,
		},
	}

	graph, err := newAgentGraph(agentAnalyzeNodeName, map[string]graphNode{
		agentAnalyzeNodeName:         r.agentAnalyzeNode,
		agentPlanNodeName:            r.agentPlanNode,
		agentOrderTasksNodeName:      r.agentOrderNode,
		agentDryRunOrExecuteNodeName: r.agentDryRunOrExecuteNode,
		agentExecuteTasksNodeName:    r.agentExecuteNode,
		agentReviewNodeName:          r.agentReviewNode,
		agentCompleteNodeName:        r.agentCompleteNode,
	})
	if err != nil {
		return r.finishWithCheckpoint(ctx, state.result, start, domain.StateFailed, err, state.execution, "failed"), err
	}

	if err := graph.run(ctx, state); err != nil {
		return r.finishWithCheckpoint(ctx, state.result, start, domain.StateFailed, err, state.execution, "failed"), err
	}
	return state.result, nil
}

func (r *AgentRunner) agentAnalyzeNode(ctx context.Context, state *appRunState) (string, error) {
	project, err := r.analyze(ctx, state.goal)
	if err != nil {
		state.execution.Project = project
		return "", err
	}
	state.project = project
	state.execution.Project = project
	return agentPlanNodeName, nil
}

func (r *AgentRunner) agentPlanNode(ctx context.Context, state *appRunState) (string, error) {
	state.result.State = domain.StatePlanning
	if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:     domain.StatePlanning,
		Message:   "Planning",
		Progress:  20,
		Timestamp: r.clock.Now(),
	}); err != nil {
		return "", err
	}

	plan, err := r.deps.Planner.Plan(ctx, state.goal, state.project)
	if err != nil {
		return "", fmt.Errorf("plan goal: %w", err)
	}
	state.plan = plan
	state.result.Plan = plan
	state.execution.Plan = plan
	return agentOrderTasksNodeName, nil
}

func (r *AgentRunner) agentOrderNode(ctx context.Context, state *appRunState) (string, error) {
	ordered, err := r.scheduler.Order(state.plan.Tasks)
	if err != nil {
		return "", err
	}
	state.ordered = ordered
	state.result.Plan.Tasks = ordered
	state.execution.Plan = state.result.Plan
	state.execution.ApprovalPolicy = domain.ApprovalPolicy{
		RequirePlanApproval: true,
	}
	if err := r.saveCheckpoint(ctx, state.result, state.execution, "planned"); err != nil {
		return "", err
	}

	maxIterations := state.agentOptions.MaxIterations
	if maxIterations == 0 {
		maxIterations = r.deps.MaxIterations
	}
	if maxIterations == 0 {
		maxIterations = defaultAgentMaxIterations
	}
	if len(ordered) > maxIterations {
		return "", fmt.Errorf("max iterations (%d) reached", maxIterations)
	}
	return agentDryRunOrExecuteNodeName, nil
}

func (r *AgentRunner) agentDryRunOrExecuteNode(ctx context.Context, state *appRunState) (string, error) {
	if !state.agentOptions.DryRun {
		return agentExecuteTasksNodeName, nil
	}
	state.result.Summary = fmt.Sprintf("Dry run planned %d task(s); no orchestration side effects were executed.", len(state.ordered))
	state.result = r.finish(state.result, state.startedAt, domain.StateComplete, nil)
	if err := r.saveCheckpoint(ctx, state.result, state.execution, "complete"); err != nil {
		return "", err
	}
	_ = reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:          domain.StateComplete,
		CompletedTasks: 0,
		TotalTasks:     len(state.ordered),
		Progress:       100,
		Message:        "Dry run complete",
		Timestamp:      r.clock.Now(),
	})
	return graphEndNode, nil
}

func (r *AgentRunner) agentExecuteNode(ctx context.Context, state *appRunState) (string, error) {
	for i, task := range state.ordered {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		state.result.State = domain.StateExecuting
		state.execution.Iteration = i + 1
		state.execution.Completed = append([]domain.TaskResult(nil), state.result.Tasks...)
		if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
			State:          domain.StateExecuting,
			CurrentTask:    task.Name,
			CompletedTasks: i,
			TotalTasks:     len(state.ordered),
			Progress:       percent(i, len(state.ordered)),
			Message:        "Executing task",
			Timestamp:      r.clock.Now(),
		}); err != nil {
			return "", err
		}

		taskResult, err := r.deps.TaskExecutor.ExecuteTask(ctx, task, state.execution)
		if taskResult != nil {
			state.result.Tasks = append(state.result.Tasks, *taskResult)
			state.result.Artifacts = append(state.result.Artifacts, taskResult.Artifacts...)
		}
		state.execution.Completed = append([]domain.TaskResult(nil), state.result.Tasks...)
		if err != nil {
			return "", err
		}
		if err := r.saveCheckpoint(ctx, state.result, state.execution, "task"); err != nil {
			return "", err
		}
	}
	return agentReviewNodeName, nil
}

func (r *AgentRunner) agentReviewNode(ctx context.Context, state *appRunState) (string, error) {
	state.result.State = domain.StateReviewing
	if r.deps.Reviewer == nil {
		return agentCompleteNodeName, nil
	}
	state.execution.Completed = append([]domain.TaskResult(nil), state.result.Tasks...)
	review, err := r.deps.Reviewer.Review(ctx, state.execution)
	if err != nil {
		return "", err
	}
	if review != nil && review.ChangesRequested {
		return "", fmt.Errorf("review requested changes: %s", review.Summary)
	}
	return agentCompleteNodeName, nil
}

func (r *AgentRunner) agentCompleteNode(ctx context.Context, state *appRunState) (string, error) {
	state.result = r.finish(state.result, state.startedAt, domain.StateComplete, nil)
	if r.deps.MemoryStore != nil {
		if err := r.deps.MemoryStore.RecordResult(ctx, *state.result); err != nil {
			return "", err
		}
	}
	_ = reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:          domain.StateComplete,
		CompletedTasks: len(state.result.Tasks),
		TotalTasks:     len(state.result.Tasks),
		Progress:       100,
		Message:        "Complete",
		Timestamp:      r.clock.Now(),
	})
	if err := r.saveCheckpoint(ctx, state.result, state.execution, "complete"); err != nil {
		return "", err
	}
	return graphEndNode, nil
}

func (r *AgentRunner) analyze(ctx context.Context, goal domain.Goal) (*domain.ProjectContext, error) {
	if r.deps.ProjectAnalyzer == nil || goal.Context.ProjectPath == "" {
		return nil, nil
	}
	if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:     domain.StateAnalyzing,
		Message:   "Analyzing project",
		Progress:  5,
		Timestamp: r.clock.Now(),
	}); err != nil {
		return nil, err
	}
	return r.deps.ProjectAnalyzer.AnalyzeProject(ctx, goal.Context.ProjectPath)
}

func (r *AgentRunner) finish(result *domain.Result, start time.Time, state domain.AgentState, err error) *domain.Result {
	result.State = state
	result.Success = state == domain.StateComplete && err == nil
	result.Error = err
	result.Duration = r.clock.Now().Sub(start)
	return result
}

func (r *AgentRunner) finishWithCheckpoint(ctx context.Context, result *domain.Result, start time.Time, state domain.AgentState, err error, execution domain.ExecutionContext, phase string) *domain.Result {
	result = r.finish(result, start, state, err)
	_ = r.saveCheckpoint(ctx, result, execution, phase)
	return result
}

func (r *AgentRunner) saveCheckpoint(ctx context.Context, result *domain.Result, execution domain.ExecutionContext, phase string) error {
	if r.deps.CheckpointStore == nil {
		return nil
	}
	if execution.Completed == nil {
		execution.Completed = append([]domain.TaskResult(nil), result.Tasks...)
	}
	if execution.Plan == nil {
		execution.Plan = result.Plan
	}
	checkpoint := domain.Checkpoint{
		ID:        checkpointID(result.Goal.ID, phase, execution.Iteration),
		GoalID:    result.Goal.ID,
		State:     result.State,
		Context:   execution,
		CreatedAt: r.clock.Now(),
		Metadata: map[string]string{
			"phase": phase,
		},
	}
	return r.deps.CheckpointStore.SaveCheckpoint(ctx, checkpoint)
}

func checkpointID(goalID, phase string, iteration int) string {
	if goalID == "" {
		goalID = "goal"
	}
	if phase == "" {
		phase = "checkpoint"
	}
	return fmt.Sprintf("%s-%s-%d", goalID, phase, iteration)
}

func normalizeReviewStrictness(value string) (string, error) {
	strictness := strings.ToLower(strings.TrimSpace(value))
	if strictness == "" {
		return defaultReviewStrictness, nil
	}
	switch strictness {
	case "low", "medium", "high":
		return strictness, nil
	default:
		return "", fmt.Errorf("invalid review strictness %q: expected low, medium, or high", value)
	}
}

func percent(done, total int) float64 {
	if total == 0 {
		return 100
	}
	return float64(done) / float64(total) * 100
}
