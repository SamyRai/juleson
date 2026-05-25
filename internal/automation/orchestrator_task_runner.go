package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/pkg/jules"
)

type orchestrationSessionClient interface {
	SendMessage(ctx context.Context, sessionID string, req *jules.SendMessageRequest) error
	ApprovePlan(ctx context.Context, sessionID string) error
	ListActivities(ctx context.Context, sessionID string, pageSize int) ([]jules.Activity, error)
}

type planWaiter interface {
	WaitForPlan(ctx context.Context) (*jules.SessionStatus, error)
}

type orchestrationTaskRunner struct {
	client           orchestrationSessionClient
	monitor          planWaiter
	sessionID        string
	autoApprove      bool
	waitAfterSend    time.Duration
	setState         func(OrchestratorState)
	sendProgress     func(phase, task int, message string, progress float64)
	addRecord        func(ExecutionRecord)
	currentTime      func() time.Time
	sleepAfterSend   func(time.Duration)
	activityPageSize int
}

func (o *SessionOrchestrator) newOrchestrationTaskRunner() *orchestrationTaskRunner {
	return &orchestrationTaskRunner{
		client:           o.client,
		monitor:          o.monitor,
		sessionID:        o.sessionID,
		autoApprove:      o.autoApprove,
		waitAfterSend:    DefaultTaskWaitTime,
		setState:         o.setState,
		sendProgress:     o.sendProgress,
		addRecord:        o.addExecutionRecord,
		currentTime:      time.Now,
		sleepAfterSend:   time.Sleep,
		activityPageSize: 10,
	}
}

func (r *orchestrationTaskRunner) execute(ctx context.Context, phaseIndex, taskIndex int, task Task) (TaskResult, error) {
	result := TaskResult{
		TaskName: task.Name,
		Success:  false,
	}

	startTime := r.currentTime()
	record := ExecutionRecord{
		PhaseIndex: phaseIndex,
		TaskIndex:  taskIndex,
		TaskName:   task.Name,
		StartTime:  startTime,
	}

	r.sendProgress(phaseIndex, taskIndex, fmt.Sprintf("Executing: %s", task.Name), 0)

	if err := r.sendTaskMessage(ctx, task); err != nil {
		return r.failTask(result, record, startTime, err, "failed to send task message")
	}

	if task.WaitForPlan {
		if err := r.waitForPlan(ctx); err != nil {
			return r.failTask(result, record, startTime, err, "failed to wait for plan")
		}
		if err := r.approvePlanIfNeeded(ctx, task); err != nil {
			return r.failTask(result, record, startTime, err, "failed to approve plan")
		}
		r.completePlanGate(phaseIndex, taskIndex, task)
	}

	r.sleepAfterSend(r.waitAfterSend)
	result, record = r.attachLatestActivity(ctx, result, record)

	if err := r.validateTask(task, result); err != nil {
		return r.failTaskWithRecordError(
			result,
			record,
			startTime,
			err,
			"task validation failed",
			fmt.Sprintf("validation failed: %v", err),
		)
	}

	record.Success = true
	record.EndTime = r.currentTime()
	record.Duration = record.EndTime.Sub(startTime)
	r.addRecord(record)

	result.Success = true
	result.Duration = record.Duration

	r.sendProgress(phaseIndex, taskIndex, fmt.Sprintf("Completed: %s", task.Name), 100)

	return result, nil
}

func (r *orchestrationTaskRunner) sendTaskMessage(ctx context.Context, task Task) error {
	if err := r.client.SendMessage(ctx, r.sessionID, &jules.SendMessageRequest{Prompt: task.Prompt}); err != nil {
		return err
	}
	return nil
}

func (r *orchestrationTaskRunner) waitForPlan(ctx context.Context) error {
	r.setState(StateWaitingPlan)
	if _, err := r.monitor.WaitForPlan(ctx); err != nil {
		return err
	}
	return nil
}

func (r *orchestrationTaskRunner) approvePlanIfNeeded(ctx context.Context, task Task) error {
	if task.AutoApprove || r.autoApprove {
		if err := r.client.ApprovePlan(ctx, r.sessionID); err != nil {
			return err
		}
	}
	return nil
}

func (r *orchestrationTaskRunner) completePlanGate(phaseIndex, taskIndex int, task Task) {
	r.setState(StateRunning)
	r.sendProgress(phaseIndex, taskIndex, fmt.Sprintf("Plan approved: %s", task.Name), 50)
}

func (r *orchestrationTaskRunner) attachLatestActivity(ctx context.Context, result TaskResult, record ExecutionRecord) (TaskResult, ExecutionRecord) {
	activities, err := r.client.ListActivities(ctx, r.sessionID, r.activityPageSize)
	if err != nil || len(activities) == 0 {
		return result, record
	}

	latestActivity := activities[0]
	record.ActivityID = latestActivity.ID
	record.ArtifactCount = len(latestActivity.Artifacts)
	result.ActivityID = latestActivity.ID
	result.Artifacts = latestActivity.Artifacts

	return result, record
}

func (r *orchestrationTaskRunner) validateTask(task Task, result TaskResult) error {
	if task.Validation == nil {
		return nil
	}
	return task.Validation(result)
}

func (r *orchestrationTaskRunner) failTask(result TaskResult, record ExecutionRecord, startTime time.Time, err error, message string) (TaskResult, error) {
	return r.failTaskWithRecordError(result, record, startTime, err, message, err.Error())
}

func (r *orchestrationTaskRunner) failTaskWithRecordError(
	result TaskResult,
	record ExecutionRecord,
	startTime time.Time,
	err error,
	message string,
	recordError string,
) (TaskResult, error) {
	record.Success = false
	record.Error = recordError
	record.EndTime = r.currentTime()
	record.Duration = record.EndTime.Sub(startTime)
	r.addRecord(record)

	result.Error = err
	result.Duration = record.Duration
	return result, fmt.Errorf("%s: %w", message, err)
}
