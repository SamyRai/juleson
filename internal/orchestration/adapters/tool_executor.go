package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
	"github.com/SamyRai/juleson/internal/agent/tools"
	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

type ToolExecutorAdapter struct {
	registry tools.ToolRegistry
}

func NewToolExecutorAdapter(registry tools.ToolRegistry) *ToolExecutorAdapter {
	return &ToolExecutorAdapter{registry: registry}
}

func (a *ToolExecutorAdapter) ExecuteTask(ctx context.Context, task domain.Task, execution domain.ExecutionContext) (*domain.TaskResult, error) {
	if a.registry == nil {
		return nil, fmt.Errorf("tool registry is required")
	}
	start := time.Now()
	agentTask := taskToAgent(task)
	matches := a.registry.FindForTask(agentTask)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no tool can handle task %q", task.Name)
	}
	tool := matches[0]
	result, err := tool.Execute(ctx, taskContext(task, execution))
	if result == nil {
		result = &tools.ToolResult{}
	}
	converted := &domain.TaskResult{
		TaskID:    task.ID,
		TaskName:  task.Name,
		TaskType:  task.Type,
		Tool:      tool.Name(),
		Success:   result.Success && err == nil,
		StartTime: start,
		EndTime:   time.Now(),
		Changes:   changesToDomain(result.Changes),
		Artifacts: artifactsToDomain(result.Artifacts),
		Error:     err,
		Metrics:   map[string]any{},
	}
	converted.Duration = converted.EndTime.Sub(converted.StartTime)
	if result.Error != nil && err == nil {
		converted.Error = result.Error
		err = result.Error
	}
	if result.Output != nil {
		converted.Output = fmt.Sprint(result.Output)
	}
	for key, value := range result.Metadata {
		converted.Metrics[key] = value
	}
	return converted, err
}

func taskToAgent(task domain.Task) agent.Task {
	return agent.Task{
		ID:           task.ID,
		Name:         task.Name,
		Description:  task.Description,
		Prompt:       task.Prompt,
		Priority:     agent.Priority(task.Priority),
		Dependencies: append([]string(nil), task.Dependencies...),
		Tool:         task.Tool,
		Context:      stringMapToAny(task.Context),
		State:        agent.TaskState(task.State),
	}
}

func taskContext(task domain.Task, execution domain.ExecutionContext) map[string]any {
	values := stringMapToAny(task.Context)
	values["prompt"] = task.Prompt
	values["description"] = task.Description
	values["project_path"] = execution.Goal.Context.ProjectPath
	if execution.Project != nil {
		values["project_name"] = execution.Project.ProjectName
		values["architecture"] = execution.Project.Architecture
	}
	return values
}

func stringMapToAny(values map[string]string) map[string]any {
	converted := make(map[string]any, len(values))
	for key, value := range values {
		converted[key] = value
	}
	return converted
}

func changesToDomain(changes []agent.Change) []domain.Change {
	converted := make([]domain.Change, 0, len(changes))
	for _, change := range changes {
		converted = append(converted, domain.Change{
			FilePath:    change.FilePath,
			Type:        domain.ChangeType(change.Type),
			Additions:   change.Additions,
			Deletions:   change.Deletions,
			Patch:       change.Patch,
			Description: change.Description,
		})
	}
	return converted
}

func artifactsToDomain(artifacts []agent.Artifact) []domain.Artifact {
	converted := make([]domain.Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		metadata := make(map[string]string, len(artifact.Metadata))
		for key, value := range artifact.Metadata {
			metadata[key] = fmt.Sprint(value)
		}
		converted = append(converted, domain.Artifact{
			Type:        domain.ArtifactType(artifact.Type),
			Path:        artifact.Path,
			Content:     artifact.Content,
			Description: artifact.Description,
			Metadata:    metadata,
		})
	}
	return converted
}
