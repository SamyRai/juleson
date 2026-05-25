package app

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/internal/orchestration/ports"
)

const defaultAIMaxIterations = 20

type AIWorkflowRunnerDeps struct {
	ProjectAnalyzer ports.ProjectAnalyzer
	Planner         ports.Planner
	DecisionMaker   ports.DecisionMaker
	TaskExecutor    ports.TaskExecutor
	Reviewer        ports.Reviewer
	MemoryStore     ports.MemoryStore
	ProgressSink    ports.ProgressSink
	Clock           ports.Clock
	MaxIterations   int
}

type AIWorkflowRunner struct {
	deps  AIWorkflowRunnerDeps
	clock ports.Clock
}

func NewAIWorkflowRunner(deps AIWorkflowRunnerDeps) *AIWorkflowRunner {
	return &AIWorkflowRunner{deps: deps, clock: clockOrDefault(deps.Clock)}
}

func (r *AIWorkflowRunner) Run(ctx context.Context, goal domain.Goal) (*domain.Result, error) {
	if goal.Description == "" {
		return nil, fmt.Errorf("goal description cannot be empty")
	}
	if r.deps.Planner == nil {
		return nil, fmt.Errorf("planner is required")
	}
	if r.deps.DecisionMaker == nil {
		return nil, fmt.Errorf("decision maker is required")
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

	var project *domain.ProjectContext
	var err error
	if r.deps.ProjectAnalyzer != nil && goal.Context.ProjectPath != "" {
		project, err = r.deps.ProjectAnalyzer.AnalyzeProject(ctx, goal.Context.ProjectPath)
		if err != nil {
			return r.finish(result, start, err), err
		}
	}

	result.State = domain.StatePlanning
	plan, err := r.deps.Planner.Plan(ctx, goal, project)
	if err != nil {
		return r.finish(result, start, err), err
	}
	result.Plan = plan

	execution := domain.ExecutionContext{
		Goal:      goal,
		Project:   project,
		Plan:      plan,
		StartedAt: start,
	}

	maxIterations := r.deps.MaxIterations
	if maxIterations == 0 {
		maxIterations = defaultAIMaxIterations
	}
	for iteration := 0; iteration < maxIterations; iteration++ {
		if err := ctx.Err(); err != nil {
			return r.finish(result, start, err), err
		}
		execution.Iteration = iteration + 1
		execution.Completed = append([]domain.TaskResult(nil), result.Tasks...)
		decision, err := r.deps.DecisionMaker.Decide(ctx, execution)
		if err != nil {
			return r.finish(result, start, err), err
		}
		if decision == nil {
			return r.finish(result, start, fmt.Errorf("decision maker returned nil decision")), fmt.Errorf("decision maker returned nil decision")
		}
		execution.Decisions = append(execution.Decisions, *decision)
		if r.deps.MemoryStore != nil {
			if err := r.deps.MemoryStore.RecordDecision(ctx, *decision); err != nil {
				return r.finish(result, start, err), err
			}
		}

		switch decision.Type {
		case domain.DecisionTypeNextTask:
			task, err := selectTask(plan.Tasks, result.Tasks, decision.TaskID)
			if err != nil {
				return r.finish(result, start, err), err
			}
			result.State = domain.StateExecuting
			taskResult, err := r.deps.TaskExecutor.ExecuteTask(ctx, task, execution)
			if taskResult != nil {
				result.Tasks = append(result.Tasks, *taskResult)
				result.Artifacts = append(result.Artifacts, taskResult.Artifacts...)
			}
			if err != nil {
				return r.finish(result, start, err), err
			}
		case domain.DecisionTypeReviewNeeded:
			result.State = domain.StateReviewing
			if r.deps.Reviewer != nil {
				if _, err := r.deps.Reviewer.Review(ctx, execution); err != nil {
					return r.finish(result, start, err), err
				}
			}
		case domain.DecisionTypeAdaptPlan:
			result.State = domain.StatePlanning
			adapted, err := r.deps.Planner.AdaptPlan(ctx, execution, decision.Reasoning)
			if err != nil {
				return r.finish(result, start, err), err
			}
			result.Plan = adapted
			plan = adapted
			execution.Plan = adapted
		case domain.DecisionTypeComplete:
			result.State = domain.StateComplete
			result.Success = true
			result.Duration = r.clock.Now().Sub(start)
			return result, nil
		case domain.DecisionTypeAbort:
			err := fmt.Errorf("AI workflow aborted: %s", decision.Reasoning)
			return r.finish(result, start, err), err
		default:
			err := fmt.Errorf("unsupported decision type %q", decision.Type)
			return r.finish(result, start, err), err
		}
	}

	err = fmt.Errorf("max iterations (%d) reached", maxIterations)
	return r.finish(result, start, err), err
}

func (r *AIWorkflowRunner) finish(result *domain.Result, start time.Time, err error) *domain.Result {
	result.State = domain.StateFailed
	result.Success = false
	result.Error = err
	result.Duration = r.clock.Now().Sub(start)
	return result
}

func selectTask(tasks []domain.Task, completed []domain.TaskResult, taskID string) (domain.Task, error) {
	completedByID := map[string]bool{}
	for _, task := range completed {
		if task.TaskID != "" {
			completedByID[task.TaskID] = true
		}
		if task.TaskName != "" {
			completedByID[task.TaskName] = true
		}
	}
	if taskID != "" {
		for _, task := range tasks {
			if task.ID == taskID || task.Name == taskID {
				return task, nil
			}
		}
		return domain.Task{}, fmt.Errorf("task %q not found", taskID)
	}
	for _, task := range tasks {
		key := task.ID
		if key == "" {
			key = task.Name
		}
		if !completedByID[key] {
			return task, nil
		}
	}
	return domain.Task{}, fmt.Errorf("no pending tasks")
}
