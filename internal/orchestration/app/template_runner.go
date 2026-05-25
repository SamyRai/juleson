package app

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/internal/orchestration/ports"
)

type TemplateRunnerDeps struct {
	ProjectAnalyzer ports.ProjectAnalyzer
	TemplateStore   ports.TemplateStore
	PromptRenderer  ports.PromptRenderer
	TaskExecutor    ports.TaskExecutor
	OutputWriter    ports.OutputWriter
	ProgressSink    ports.ProgressSink
	Clock           ports.Clock
}

type TemplateRunner struct {
	deps      TemplateRunnerDeps
	clock     ports.Clock
	scheduler taskScheduler
}

func NewTemplateRunner(deps TemplateRunnerDeps) *TemplateRunner {
	return &TemplateRunner{
		deps:  deps,
		clock: clockOrDefault(deps.Clock),
	}
}

func (r *TemplateRunner) Run(ctx context.Context, templateName, projectPath string, values map[string]string) (*domain.Result, []string, error) {
	if templateName == "" {
		return nil, nil, fmt.Errorf("template name cannot be empty")
	}
	if projectPath == "" {
		return nil, nil, fmt.Errorf("project path cannot be empty")
	}
	if r.deps.TemplateStore == nil {
		return nil, nil, fmt.Errorf("template store is required")
	}
	if r.deps.TaskExecutor == nil {
		return nil, nil, fmt.Errorf("task executor is required")
	}

	start := r.clock.Now()
	template, err := r.deps.TemplateStore.LoadTemplate(ctx, templateName)
	if err != nil {
		return nil, nil, fmt.Errorf("load template: %w", err)
	}

	project := &domain.ProjectContext{ProjectPath: projectPath, Values: copyValues(values)}
	if r.deps.ProjectAnalyzer != nil {
		project, err = r.deps.ProjectAnalyzer.AnalyzeProject(ctx, projectPath)
		if err != nil {
			return nil, nil, fmt.Errorf("analyze project: %w", err)
		}
		if project.Values == nil {
			project.Values = map[string]string{}
		}
		for key, value := range values {
			project.Values[key] = value
		}
	}

	tasks, err := r.renderTasks(ctx, template.Tasks, project.Values)
	if err != nil {
		return nil, nil, err
	}
	tasks, err = r.scheduler.Order(tasks)
	if err != nil {
		return nil, nil, err
	}

	goal := domain.Goal{
		ID:          template.Name,
		Description: template.Description,
		Context: domain.GoalContext{
			ProjectPath: projectPath,
			Values:      values,
		},
	}
	result := &domain.Result{
		Goal:  goal,
		State: domain.StateExecuting,
		Plan: &domain.Plan{
			ID:        template.Name,
			Goal:      goal,
			Tasks:     tasks,
			CreatedAt: start,
		},
		Tasks: make([]domain.TaskResult, 0, len(tasks)),
	}

	execution := domain.ExecutionContext{
		Goal:      goal,
		Project:   project,
		Plan:      result.Plan,
		StartedAt: start,
		Values:    project.Values,
	}

	for i, task := range tasks {
		if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
			State:          domain.StateExecuting,
			CurrentTask:    task.Name,
			CompletedTasks: i,
			TotalTasks:     len(tasks),
			Progress:       percent(i, len(tasks)),
			Message:        "Executing template task",
			Timestamp:      r.clock.Now(),
		}); err != nil {
			return result, nil, err
		}
		taskResult, err := r.deps.TaskExecutor.ExecuteTask(ctx, task, execution)
		if taskResult != nil {
			result.Tasks = append(result.Tasks, *taskResult)
			result.Artifacts = append(result.Artifacts, taskResult.Artifacts...)
		}
		if err != nil {
			result.State = domain.StateFailed
			result.Error = err
			result.Duration = r.clock.Now().Sub(start)
			return result, nil, err
		}
		execution.Completed = append([]domain.TaskResult(nil), result.Tasks...)
	}

	result.State = domain.StateComplete
	result.Success = true
	result.Duration = r.clock.Now().Sub(start)

	var outputs []string
	if r.deps.OutputWriter != nil {
		outputs, err = r.deps.OutputWriter.WriteOutputs(ctx, *template, *result)
		if err != nil {
			result.Error = err
			return result, outputs, err
		}
	}
	return result, outputs, nil
}

func (r *TemplateRunner) renderTasks(ctx context.Context, tasks []domain.Task, values map[string]string) ([]domain.Task, error) {
	rendered := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		next := task
		if r.deps.PromptRenderer != nil {
			prompt, err := r.deps.PromptRenderer.RenderPrompt(ctx, task.Prompt, values)
			if err != nil {
				return nil, fmt.Errorf("render prompt for task %q: %w", task.Name, err)
			}
			next.Prompt = prompt
		}
		rendered = append(rendered, next)
	}
	return rendered, nil
}

func copyValues(values map[string]string) map[string]string {
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}
