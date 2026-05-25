package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/internal/orchestration/ports"
)

type SessionWorkflowRunnerDeps struct {
	TaskExecutor ports.TaskExecutor
	ProgressSink ports.ProgressSink
	Clock        ports.Clock
}

type SessionWorkflowRunner struct {
	deps  SessionWorkflowRunnerDeps
	clock ports.Clock
}

func NewSessionWorkflowRunner(deps SessionWorkflowRunnerDeps) *SessionWorkflowRunner {
	return &SessionWorkflowRunner{deps: deps, clock: clockOrDefault(deps.Clock)}
}

func (r *SessionWorkflowRunner) Run(ctx context.Context, workflow domain.Workflow, execution domain.ExecutionContext) (*domain.WorkflowResult, error) {
	if workflow.Name == "" {
		return nil, fmt.Errorf("workflow name cannot be empty")
	}
	if r.deps.TaskExecutor == nil {
		return nil, fmt.Errorf("task executor is required")
	}

	start := r.clock.Now()
	result := &domain.WorkflowResult{
		WorkflowName: workflow.Name,
		TotalPhases:  len(workflow.Phases),
		StartTime:    start,
		PhaseResults: make([]domain.PhaseResult, 0, len(workflow.Phases)),
		SessionID:    execution.SessionID,
	}

	execution.Workflow = &workflow
	for i, phase := range workflow.Phases {
		phaseResult, err := r.runPhase(ctx, i, phase, execution)
		result.PhaseResults = append(result.PhaseResults, phaseResult)
		execution.Completed = append(execution.Completed, phaseResult.Tasks...)
		if err != nil {
			result.Success = false
			result.Error = err
			result.EndTime = r.clock.Now()
			result.TotalDuration = result.EndTime.Sub(start)
			return result, err
		}
	}

	result.Success = true
	result.EndTime = r.clock.Now()
	result.TotalDuration = result.EndTime.Sub(start)
	return result, nil
}

func (r *SessionWorkflowRunner) runPhase(ctx context.Context, index int, phase domain.Phase, execution domain.ExecutionContext) (domain.PhaseResult, error) {
	if phase.Parallel {
		return r.runPhaseParallel(ctx, index, phase, execution)
	}

	start := r.clock.Now()
	result := domain.PhaseResult{
		PhaseIndex: index,
		PhaseName:  phase.Name,
		Success:    true,
		Tasks:      make([]domain.TaskResult, 0, len(phase.Tasks)),
	}
	for taskIndex, task := range phase.Tasks {
		taskResult, err := r.runTask(ctx, taskIndex, len(phase.Tasks), task, execution)
		if taskResult != nil {
			result.Tasks = append(result.Tasks, *taskResult)
		}
		if err != nil {
			result.Success = false
			result.Error = err
			result.Duration = r.clock.Now().Sub(start)
			if !phase.ContinueOnError {
				return result, err
			}
		}
	}
	result.Duration = r.clock.Now().Sub(start)
	return result, result.Error
}

func (r *SessionWorkflowRunner) runPhaseParallel(ctx context.Context, index int, phase domain.Phase, execution domain.ExecutionContext) (domain.PhaseResult, error) {
	start := r.clock.Now()
	result := domain.PhaseResult{
		PhaseIndex: index,
		PhaseName:  phase.Name,
		Success:    true,
		Tasks:      make([]domain.TaskResult, len(phase.Tasks)),
	}

	var wg sync.WaitGroup
	results := make(chan struct {
		index int
		task  *domain.TaskResult
		err   error
	}, len(phase.Tasks))

	for i, task := range phase.Tasks {
		wg.Add(1)
		go func(taskIndex int, task domain.Task) {
			defer wg.Done()
			taskResult, err := r.runTask(ctx, taskIndex, len(phase.Tasks), task, execution)
			results <- struct {
				index int
				task  *domain.TaskResult
				err   error
			}{index: taskIndex, task: taskResult, err: err}
		}(i, task)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var firstErr error
	for item := range results {
		if item.task != nil {
			result.Tasks[item.index] = *item.task
		}
		if item.err != nil && firstErr == nil {
			firstErr = item.err
			result.Success = false
			result.Error = item.err
		}
	}
	result.Duration = r.clock.Now().Sub(start)
	if firstErr != nil && !phase.ContinueOnError {
		return result, firstErr
	}
	return result, nil
}

func (r *SessionWorkflowRunner) runTask(ctx context.Context, taskIndex, total int, task domain.Task, execution domain.ExecutionContext) (*domain.TaskResult, error) {
	if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:          domain.StateExecuting,
		CurrentTask:    task.Name,
		CompletedTasks: taskIndex,
		TotalTasks:     total,
		Progress:       percent(taskIndex, total),
		Message:        "Executing workflow task",
		Timestamp:      r.clock.Now(),
	}); err != nil {
		return nil, err
	}
	return r.deps.TaskExecutor.ExecuteTask(ctx, task, execution)
}
