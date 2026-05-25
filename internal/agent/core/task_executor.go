package core

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
	"github.com/SamyRai/juleson/internal/agent/tools"
	"github.com/SamyRai/juleson/internal/analyzer"
)

type taskToolFinder interface {
	FindForTask(task agent.Task) []tools.Tool
}

type taskExecutionContext struct {
	goal           *agent.Goal
	projectContext *analyzer.ProjectContext
}

type taskExecutor struct {
	tools     taskToolFinder
	validator *ConstraintValidator
	telemetry *Metrics
	logger    *slog.Logger
	dryRun    bool
}

func newTaskExecutor(
	toolFinder taskToolFinder,
	validator *ConstraintValidator,
	telemetry *Metrics,
	logger *slog.Logger,
	dryRun bool,
) *taskExecutor {
	return &taskExecutor{
		tools:     toolFinder,
		validator: validator,
		telemetry: telemetry,
		logger:    logger,
		dryRun:    dryRun,
	}
}

func (e *taskExecutor) execute(ctx context.Context, task *agent.Task, executionContext taskExecutionContext) (*agent.TaskResult, error) {
	matchingTools := e.tools.FindForTask(*task)
	if len(matchingTools) == 0 {
		return nil, fmt.Errorf("no tool found for task: %s", task.Name)
	}

	tool := matchingTools[0]
	taskResult := &agent.TaskResult{
		TaskID:  task.ID,
		Name:    task.Name,
		Success: false,
		Tool:    tool.Name(),
		Changes: make([]agent.Change, 0),
	}

	if e.dryRun {
		e.logger.Info("agent.act.dry_run", "task_id", task.ID)
		taskResult.Success = true
		taskResult.Duration = time.Second
		e.recordToolInvocation(tool.Name(), true, taskResult.Duration)
		return taskResult, nil
	}

	params := e.prepareParameters(task, executionContext)

	startTime := time.Now()
	toolResult, err := tool.Execute(ctx, params)
	taskResult.Duration = time.Since(startTime)

	if err != nil {
		taskResult.Error = err
		e.logger.Error("agent.act.tool_failed", "error", err)
		e.recordToolInvocation(tool.Name(), false, taskResult.Duration)
		return taskResult, err
	}

	taskResult.Success = toolResult.Success
	taskResult.Changes = toolResult.Changes

	e.validateChanges(taskResult.Changes)
	e.recordTaskResult(tool.Name(), taskResult.Success, taskResult.Duration)

	e.logger.Info("agent.act.task_complete",
		"task_id", task.ID,
		"success", taskResult.Success,
		"duration", taskResult.Duration)

	return taskResult, nil
}

func (e *taskExecutor) prepareParameters(task *agent.Task, executionContext taskExecutionContext) map[string]interface{} {
	params := map[string]interface{}{
		"action": "create_session",
		"prompt": task.Prompt,
	}

	if executionContext.goal != nil && executionContext.goal.Context.SourceID != "" {
		params["source_id"] = executionContext.goal.Context.SourceID
	}

	if executionContext.projectContext != nil {
		params["project_path"] = executionContext.projectContext.ProjectPath
	}

	return params
}

func (e *taskExecutor) setValidator(validator *ConstraintValidator) {
	e.validator = validator
}

func (e *taskExecutor) validateChanges(changes []agent.Change) {
	if e.validator == nil || len(changes) == 0 {
		return
	}
	if err := e.validator.ValidateChanges(changes); err != nil {
		e.logger.Warn("agent.act.constraint_violation", "error", err)
	}
}

func (e *taskExecutor) recordToolInvocation(toolName string, success bool, duration time.Duration) {
	if e.telemetry != nil {
		e.telemetry.RecordToolInvocation(toolName, success, duration)
	}
}

func (e *taskExecutor) recordTaskResult(toolName string, success bool, duration time.Duration) {
	if e.telemetry == nil {
		return
	}
	e.telemetry.RecordToolInvocation(toolName, success, duration)
	e.telemetry.RecordTask(true, success, !success, duration)
}
