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
	result := &domain.Result{
		Goal:  goal,
		State: domain.StateAnalyzing,
		Tasks: make([]domain.TaskResult, 0),
	}

	project, err := r.analyze(ctx, goal)
	if err != nil {
		execution := domain.ExecutionContext{
			Goal:             goal,
			Project:          project,
			StartedAt:        start,
			DryRun:           opts.DryRun,
			ReviewStrictness: strictness,
		}
		return r.finishWithCheckpoint(ctx, result, start, domain.StateFailed, err, execution, "failed"), err
	}

	result.State = domain.StatePlanning
	if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:     domain.StatePlanning,
		Message:   "Planning",
		Progress:  20,
		Timestamp: r.clock.Now(),
	}); err != nil {
		execution := domain.ExecutionContext{
			Goal:             goal,
			Project:          project,
			StartedAt:        start,
			DryRun:           opts.DryRun,
			ReviewStrictness: strictness,
		}
		return r.finishWithCheckpoint(ctx, result, start, domain.StateFailed, err, execution, "failed"), err
	}

	plan, err := r.deps.Planner.Plan(ctx, goal, project)
	if err != nil {
		wrapped := fmt.Errorf("plan goal: %w", err)
		execution := domain.ExecutionContext{
			Goal:             goal,
			Project:          project,
			StartedAt:        start,
			DryRun:           opts.DryRun,
			ReviewStrictness: strictness,
		}
		return r.finishWithCheckpoint(ctx, result, start, domain.StateFailed, wrapped, execution, "failed"), err
	}
	result.Plan = plan

	ordered, err := r.scheduler.Order(plan.Tasks)
	if err != nil {
		execution := domain.ExecutionContext{
			Goal:             goal,
			Project:          project,
			Plan:             plan,
			StartedAt:        start,
			DryRun:           opts.DryRun,
			ReviewStrictness: strictness,
		}
		return r.finishWithCheckpoint(ctx, result, start, domain.StateFailed, err, execution, "failed"), err
	}
	result.Plan.Tasks = ordered

	execution := domain.ExecutionContext{
		Goal:             goal,
		Project:          project,
		Plan:             result.Plan,
		StartedAt:        start,
		DryRun:           opts.DryRun,
		ReviewStrictness: strictness,
		ApprovalPolicy: domain.ApprovalPolicy{
			RequirePlanApproval: true,
		},
	}
	if err := r.saveCheckpoint(ctx, result, execution, "planned"); err != nil {
		return r.finish(result, start, domain.StateFailed, err), err
	}

	maxIterations := opts.MaxIterations
	if maxIterations == 0 {
		maxIterations = r.deps.MaxIterations
	}
	if maxIterations == 0 {
		maxIterations = defaultAgentMaxIterations
	}
	if len(ordered) > maxIterations {
		err := fmt.Errorf("max iterations (%d) reached", maxIterations)
		return r.finishWithCheckpoint(ctx, result, start, domain.StateFailed, err, execution, "failed"), err
	}

	if opts.DryRun {
		result.Summary = fmt.Sprintf("Dry run planned %d task(s); no orchestration side effects were executed.", len(ordered))
		result = r.finish(result, start, domain.StateComplete, nil)
		if err := r.saveCheckpoint(ctx, result, execution, "complete"); err != nil {
			return result, err
		}
		_ = reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
			State:          domain.StateComplete,
			CompletedTasks: 0,
			TotalTasks:     len(ordered),
			Progress:       100,
			Message:        "Dry run complete",
			Timestamp:      r.clock.Now(),
		})
		return result, nil
	}

	for i, task := range ordered {
		if err := ctx.Err(); err != nil {
			return r.finishWithCheckpoint(ctx, result, start, domain.StateFailed, err, execution, "failed"), err
		}
		result.State = domain.StateExecuting
		execution.Iteration = i + 1
		execution.Completed = append([]domain.TaskResult(nil), result.Tasks...)
		if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
			State:          domain.StateExecuting,
			CurrentTask:    task.Name,
			CompletedTasks: i,
			TotalTasks:     len(ordered),
			Progress:       percent(i, len(ordered)),
			Message:        "Executing task",
			Timestamp:      r.clock.Now(),
		}); err != nil {
			return r.finishWithCheckpoint(ctx, result, start, domain.StateFailed, err, execution, "failed"), err
		}

		taskResult, err := r.deps.TaskExecutor.ExecuteTask(ctx, task, execution)
		if taskResult != nil {
			result.Tasks = append(result.Tasks, *taskResult)
			result.Artifacts = append(result.Artifacts, taskResult.Artifacts...)
		}
		execution.Completed = append([]domain.TaskResult(nil), result.Tasks...)
		if err != nil {
			return r.finishWithCheckpoint(ctx, result, start, domain.StateFailed, err, execution, "failed"), err
		}
		if err := r.saveCheckpoint(ctx, result, execution, "task"); err != nil {
			return r.finish(result, start, domain.StateFailed, err), err
		}
	}

	result.State = domain.StateReviewing
	if r.deps.Reviewer != nil {
		execution.Completed = append([]domain.TaskResult(nil), result.Tasks...)
		review, err := r.deps.Reviewer.Review(ctx, execution)
		if err != nil {
			return r.finishWithCheckpoint(ctx, result, start, domain.StateFailed, err, execution, "failed"), err
		}
		if review != nil && review.ChangesRequested {
			err := fmt.Errorf("review requested changes: %s", review.Summary)
			return r.finishWithCheckpoint(ctx, result, start, domain.StateFailed, err, execution, "failed"), err
		}
	}

	result = r.finish(result, start, domain.StateComplete, nil)
	if r.deps.MemoryStore != nil {
		if err := r.deps.MemoryStore.RecordResult(ctx, *result); err != nil {
			return result, err
		}
	}
	_ = reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:          domain.StateComplete,
		CompletedTasks: len(result.Tasks),
		TotalTasks:     len(result.Tasks),
		Progress:       100,
		Message:        "Complete",
		Timestamp:      r.clock.Now(),
	})
	if err := r.saveCheckpoint(ctx, result, execution, "complete"); err != nil {
		return result, err
	}
	return result, nil
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
