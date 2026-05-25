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

const (
	templateLoadNodeName     = "template.loadTemplate"
	templateAnalyzeNodeName  = "template.analyzeProject"
	templateRenderNodeName   = "template.renderAndOrder"
	templateExecuteNodeName  = "template.executeTasks"
	templateWriteOutputsName = "template.writeOutputs"
	templateCompleteNodeName = "template.complete"
)

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
	state := &appRunState{
		goal: domain.Goal{
			ID:          templateName,
			Description: templateName,
			Context: domain.GoalContext{
				ProjectPath: projectPath,
				Values:      values,
			},
		},
		project:   &domain.ProjectContext{ProjectPath: projectPath, Values: copyValues(values)},
		values:    copyValues(values),
		startedAt: start,
	}

	graph, err := newAgentGraph(templateLoadNodeName, map[string]graphNode{
		templateLoadNodeName:     r.templateLoadNode,
		templateAnalyzeNodeName:  r.templateAnalyzeNode,
		templateRenderNodeName:   r.templateRenderNode,
		templateExecuteNodeName:  r.templateExecuteNode,
		templateWriteOutputsName: r.templateWriteOutputsNode,
		templateCompleteNodeName: r.templateCompleteNode,
	})
	if err != nil {
		return nil, nil, err
	}

	if err := graph.run(ctx, state); err != nil {
		return state.result, state.outputs, err
	}
	return state.result, state.outputs, nil
}

func (r *TemplateRunner) templateLoadNode(ctx context.Context, state *appRunState) (string, error) {
	template, err := r.deps.TemplateStore.LoadTemplate(ctx, state.goal.ID)
	if err != nil {
		return "", fmt.Errorf("load template: %w", err)
	}
	state.template = template
	state.outputFiles = template.OutputFiles
	state.goal.ID = template.Name
	state.goal.Description = template.Description
	return templateAnalyzeNodeName, nil
}

func (r *TemplateRunner) templateAnalyzeNode(ctx context.Context, state *appRunState) (string, error) {
	if r.deps.ProjectAnalyzer != nil {
		project, err := r.deps.ProjectAnalyzer.AnalyzeProject(ctx, state.goal.Context.ProjectPath)
		if err != nil {
			return "", fmt.Errorf("analyze project: %w", err)
		}
		state.project = project
	}
	if state.project == nil {
		state.project = &domain.ProjectContext{ProjectPath: state.goal.Context.ProjectPath}
	}
	if state.project.Values == nil {
		state.project.Values = map[string]string{}
	}
	for key, value := range state.values {
		state.project.Values[key] = value
	}
	return templateRenderNodeName, nil
}

func (r *TemplateRunner) templateRenderNode(ctx context.Context, state *appRunState) (string, error) {
	tasks, err := r.renderTasks(ctx, state.template.Tasks, state.project.Values)
	if err != nil {
		return "", err
	}
	tasks, err = r.scheduler.Order(tasks)
	if err != nil {
		return "", err
	}
	state.ordered = tasks
	state.plan = &domain.Plan{
		ID:        state.template.Name,
		Goal:      state.goal,
		Tasks:     tasks,
		CreatedAt: state.startedAt,
	}
	state.result = &domain.Result{
		Goal:  state.goal,
		State: domain.StateExecuting,
		Plan:  state.plan,
		Tasks: make([]domain.TaskResult, 0, len(tasks)),
	}
	state.execution = domain.ExecutionContext{
		Goal:      state.goal,
		Project:   state.project,
		Plan:      state.plan,
		StartedAt: state.startedAt,
		Values:    state.project.Values,
	}
	return templateExecuteNodeName, nil
}

func (r *TemplateRunner) templateExecuteNode(ctx context.Context, state *appRunState) (string, error) {
	for i, task := range state.ordered {
		if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
			State:          domain.StateExecuting,
			CurrentTask:    task.Name,
			CompletedTasks: i,
			TotalTasks:     len(state.ordered),
			Progress:       percent(i, len(state.ordered)),
			Message:        "Executing template task",
			Timestamp:      r.clock.Now(),
		}); err != nil {
			return "", err
		}
		taskResult, err := r.deps.TaskExecutor.ExecuteTask(ctx, task, state.execution)
		if taskResult != nil {
			state.result.Tasks = append(state.result.Tasks, *taskResult)
			state.result.Artifacts = append(state.result.Artifacts, taskResult.Artifacts...)
		}
		if err != nil {
			state.result.State = domain.StateFailed
			state.result.Error = err
			state.result.Duration = r.clock.Now().Sub(state.startedAt)
			return "", err
		}
		state.execution.Completed = append([]domain.TaskResult(nil), state.result.Tasks...)
	}
	state.result.State = domain.StateComplete
	state.result.Success = true
	state.result.Duration = r.clock.Now().Sub(state.startedAt)
	return templateWriteOutputsName, nil
}

func (r *TemplateRunner) templateWriteOutputsNode(ctx context.Context, state *appRunState) (string, error) {
	if r.deps.OutputWriter == nil {
		return templateCompleteNodeName, nil
	}
	template := *state.template
	template.OutputFiles = state.outputFiles
	outputs, err := r.deps.OutputWriter.WriteOutputs(ctx, template, *state.result)
	state.outputs = outputs
	if err != nil {
		state.result.Error = err
		return "", err
	}
	return templateCompleteNodeName, nil
}

func (r *TemplateRunner) templateCompleteNode(ctx context.Context, state *appRunState) (string, error) {
	return graphEndNode, nil
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
