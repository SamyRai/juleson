package automation

import (
	"context"
	"sync"
	"time"
)

// executePhase executes a single phase.
func (o *SessionOrchestrator) executePhase(ctx context.Context, phaseIndex int, phase Phase) (PhaseResult, error) {
	result := PhaseResult{
		PhaseIndex: phaseIndex,
		PhaseName:  phase.Name,
		Success:    true,
		Tasks:      make([]TaskResult, 0),
	}

	startTime := time.Now()

	if phase.Parallel {
		return o.executeTasksParallel(ctx, phaseIndex, phase)
	}

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

// executeTask executes a single task.
func (o *SessionOrchestrator) executeTask(ctx context.Context, phaseIndex, taskIndex int, task Task) (TaskResult, error) {
	return o.newOrchestrationTaskRunner().execute(ctx, phaseIndex, taskIndex, task)
}

// executeTasksParallel executes tasks in parallel.
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
