package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/internal/templates"
)

type TemplateStoreAdapter struct {
	manager *templates.Manager
}

func NewTemplateStoreAdapter(manager *templates.Manager) *TemplateStoreAdapter {
	return &TemplateStoreAdapter{manager: manager}
}

func (a *TemplateStoreAdapter) LoadTemplate(ctx context.Context, name string) (*domain.Template, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if a.manager == nil {
		return nil, fmt.Errorf("template manager is required")
	}
	template, err := a.manager.LoadTemplate(name)
	if err != nil {
		return nil, err
	}
	return templateToDomain(template), nil
}

func templateToDomain(template *templates.Template) *domain.Template {
	if template == nil {
		return nil
	}
	tasks := make([]domain.Task, 0, len(template.Tasks))
	for _, task := range template.Tasks {
		tasks = append(tasks, domain.Task{
			ID:               task.Name,
			Name:             task.Name,
			Type:             task.Type,
			Description:      task.Description,
			Prompt:           task.JulesPrompt,
			Dependencies:     append([]string(nil), task.DependsOn...),
			Context:          copyStringMap(task.ContextVars),
			RequiresApproval: task.RequiresApproval,
		})
	}
	outputs := make([]domain.OutputFile, 0, len(template.Output.Files))
	for _, output := range template.Output.Files {
		outputs = append(outputs, domain.OutputFile{
			Path:     output.Path,
			Template: output.Template,
		})
	}
	return &domain.Template{
		Name:        template.Metadata.Name,
		Description: template.Metadata.Description,
		Tasks:       tasks,
		OutputFiles: outputs,
		Metadata: map[string]string{
			"version":  template.Metadata.Version,
			"author":   template.Metadata.Author,
			"category": template.Metadata.Category,
		},
	}
}

type PromptRendererAdapter struct{}

func NewPromptRendererAdapter() *PromptRendererAdapter {
	return &PromptRendererAdapter{}
}

func (PromptRendererAdapter) RenderPrompt(ctx context.Context, template string, values map[string]string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	rendered := template
	for key, value := range values {
		rendered = strings.ReplaceAll(rendered, "{{"+key+"}}", value)
		rendered = strings.ReplaceAll(rendered, "{{."+key+"}}", value)
	}
	return rendered, nil
}
