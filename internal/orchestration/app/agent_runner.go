package app

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/internal/orchestration/ports"
)

const defaultAgentMaxIterations = 20

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
	if goal.Description == "" {
		return nil, fmt.Errorf("goal description cannot be empty")
	}
	if r.deps.Planner == nil {
		return nil, fmt.Errorf("planner is required")
	}
	if r.deps.TaskExecutor == nil {
		return nil, fmt.Errorf("task executor is required")
	}

	start := r.clock.Now()
	result := &domain.Result{
		Goal:  goal,
		State: domain.StateAnalyzing,
		Tasks: make([]domain.TaskResult, 0),
	}

	project, err := r.analyze(ctx, goal)
	if err != nil {
		return r.finish(result, start, domain.StateFailed, err), err
	}

	result.State = domain.StatePlanning
	if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:     domain.StatePlanning,
		Message:   "Planning",
		Progress:  20,
		Timestamp: r.clock.Now(),
	}); err != nil {
		return r.finish(result, start, domain.StateFailed, err), err
	}

	plan, err := r.deps.Planner.Plan(ctx, goal, project)
	if err != nil {
		return r.finish(result, start, domain.StateFailed, fmt.Errorf("plan goal: %w", err)), err
	}
	result.Plan = plan

	ordered, err := r.scheduler.Order(plan.Tasks)
	if err != nil {
		return r.finish(result, start, domain.StateFailed, err), err
	}

	execution := domain.ExecutionContext{
		Goal:      goal,
		Project:   project,
		Plan:      plan,
		StartedAt: start,
	}

	maxIterations := r.deps.MaxIterations
	if maxIterations == 0 {
		maxIterations = defaultAgentMaxIterations
	}
	if len(ordered) > maxIterations {
		err := fmt.Errorf("max iterations (%d) reached", maxIterations)
		return r.finish(result, start, domain.StateFailed, err), err
	}

	for i, task := range ordered {
		if err := ctx.Err(); err != nil {
			return r.finish(result, start, domain.StateFailed, err), err
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
			return r.finish(result, start, domain.StateFailed, err), err
		}

		taskResult, err := r.deps.TaskExecutor.ExecuteTask(ctx, task, execution)
		if taskResult != nil {
			result.Tasks = append(result.Tasks, *taskResult)
			result.Artifacts = append(result.Artifacts, taskResult.Artifacts...)
		}
		if err != nil {
			return r.finish(result, start, domain.StateFailed, err), err
		}
	}

	result.State = domain.StateReviewing
	if r.deps.Reviewer != nil {
		execution.Completed = append([]domain.TaskResult(nil), result.Tasks...)
		review, err := r.deps.Reviewer.Review(ctx, execution)
		if err != nil {
			return r.finish(result, start, domain.StateFailed, err), err
		}
		if review != nil && review.ChangesRequested {
			err := fmt.Errorf("review requested changes: %s", review.Summary)
			return r.finish(result, start, domain.StateFailed, err), err
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

func percent(done, total int) float64 {
	if total == 0 {
		return 100
	}
	return float64(done) / float64(total) * 100
}
