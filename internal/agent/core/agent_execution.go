package core

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
)

// act executes tasks using appropriate tools.
func (a *CoreAgent) act(ctx context.Context) error {
	a.logger.Info("agent.act.start", "tasks", len(a.currentPlan))

	if len(a.currentPlan) == 0 {
		a.setState(agent.StateReviewing)
		return nil
	}

	var currentTask *agent.Task
	for i := range a.currentPlan {
		if a.currentPlan[i].State == agent.TaskStatePending {
			currentTask = &a.currentPlan[i]
			currentTask.State = agent.TaskStateInProgress
			break
		}
	}

	if currentTask == nil {
		a.setState(agent.StateExecuting)
		return nil
	}

	a.logger.Info("agent.act.task", "task_id", currentTask.ID, "name", currentTask.Name)

	taskResult, err := a.executeSingleTask(ctx, currentTask)
	if err != nil {
		currentTask.State = agent.TaskStateFailed
		return err
	}

	currentTask.State = agent.TaskStateComplete
	currentTask.Result = taskResult

	a.setState(agent.StateExecuting)
	return nil
}

// executeSingleTask executes a single task using the appropriate tool.
func (a *CoreAgent) executeSingleTask(ctx context.Context, task *agent.Task) (*agent.TaskResult, error) {
	matchingTools := a.toolRegistry.FindForTask(*task)
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

	if a.dryRun {
		a.logger.Info("agent.act.dry_run", "task_id", task.ID)
		taskResult.Success = true
		taskResult.Duration = time.Second

		if a.telemetry != nil {
			a.telemetry.RecordToolInvocation(tool.Name(), true, time.Second)
		}

		return taskResult, nil
	}

	params := a.prepareToolParameters(task)

	startTime := time.Now()
	toolResult, err := tool.Execute(ctx, params)
	taskResult.Duration = time.Since(startTime)

	if err != nil {
		taskResult.Error = err
		a.logger.Error("agent.act.tool_failed", "error", err)

		if a.telemetry != nil {
			a.telemetry.RecordToolInvocation(tool.Name(), false, taskResult.Duration)
		}

		return taskResult, err
	}

	taskResult.Success = toolResult.Success
	taskResult.Changes = toolResult.Changes

	if a.validator != nil && len(taskResult.Changes) > 0 {
		if err := a.validator.ValidateChanges(taskResult.Changes); err != nil {
			a.logger.Warn("agent.act.constraint_violation", "error", err)
		}
	}

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

// prepareToolParameters prepares parameters for tool execution.
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
