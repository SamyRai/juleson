package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SamyRai/juleson/pkg/jules"
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
	result := TaskResult{
		TaskName: task.Name,
		Success:  false,
	}

	startTime := time.Now()
	record := ExecutionRecord{
		PhaseIndex: phaseIndex,
		TaskIndex:  taskIndex,
		TaskName:   task.Name,
		StartTime:  startTime,
	}

	o.sendProgress(phaseIndex, taskIndex, fmt.Sprintf("Executing: %s", task.Name), 0)

	err := o.client.SendMessage(ctx, o.sessionID, &jules.SendMessageRequest{
		Prompt: task.Prompt,
	})
	if err != nil {
		record.Success = false
		record.Error = err.Error()
		record.EndTime = time.Now()
		record.Duration = time.Since(startTime)
		o.addExecutionRecord(record)

		result.Error = err
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("failed to send task message: %w", err)
	}

	if task.WaitForPlan {
		o.setState(StateWaitingPlan)
		_, err := o.monitor.WaitForPlan(ctx)
		if err != nil {
			record.Success = false
			record.Error = err.Error()
			record.EndTime = time.Now()
			record.Duration = time.Since(startTime)
			o.addExecutionRecord(record)

			result.Error = err
			result.Duration = time.Since(startTime)
			return result, fmt.Errorf("failed to wait for plan: %w", err)
		}

		if task.AutoApprove || o.autoApprove {
			if err := o.client.ApprovePlan(ctx, o.sessionID); err != nil {
				record.Success = false
				record.Error = err.Error()
				record.EndTime = time.Now()
				record.Duration = time.Since(startTime)
				o.addExecutionRecord(record)

				result.Error = err
				result.Duration = time.Since(startTime)
				return result, fmt.Errorf("failed to approve plan: %w", err)
			}
		}

		o.setState(StateRunning)
		o.sendProgress(phaseIndex, taskIndex, fmt.Sprintf("Plan approved: %s", task.Name), 50)
	}

	time.Sleep(DefaultTaskWaitTime)

	activities, err := o.client.ListActivities(ctx, o.sessionID, 10)
	if err == nil && len(activities) > 0 {
		latestActivity := activities[0]
		record.ActivityID = latestActivity.ID
		record.ArtifactCount = len(latestActivity.Artifacts)

		result.ActivityID = latestActivity.ID
		result.Artifacts = latestActivity.Artifacts
	}

	if task.Validation != nil {
		if err := task.Validation(result); err != nil {
			record.Success = false
			record.Error = fmt.Sprintf("validation failed: %v", err)
			record.EndTime = time.Now()
			record.Duration = time.Since(startTime)
			o.addExecutionRecord(record)

			result.Error = err
			result.Duration = time.Since(startTime)
			return result, fmt.Errorf("task validation failed: %w", err)
		}
	}

	record.Success = true
	record.EndTime = time.Now()
	record.Duration = time.Since(startTime)
	o.addExecutionRecord(record)

	result.Success = true
	result.Duration = time.Since(startTime)

	o.sendProgress(phaseIndex, taskIndex, fmt.Sprintf("Completed: %s", task.Name), 100)

	return result, nil
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
