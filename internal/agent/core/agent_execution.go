package core

import (
	"context"

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
	return a.executor.execute(ctx, task, a.taskExecutionContext())
}

// prepareToolParameters prepares parameters for tool execution.
func (a *CoreAgent) prepareToolParameters(task *agent.Task) map[string]interface{} {
	return a.executor.prepareParameters(task, a.taskExecutionContext())
}

func (a *CoreAgent) taskExecutionContext() taskExecutionContext {
	return taskExecutionContext{
		goal:           a.currentGoal,
		projectContext: a.projectContext,
	}
}
