package tools

import (
	"context"
	"fmt"

	"jules-automation/internal/automation"
	"jules-automation/internal/templates"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTemplateTools registers all template-related MCP tools
func RegisterTemplateTools(server *mcp.Server, templateManager *templates.Manager, automationEngine *automation.Engine) {
	// Execute Template Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "execute_template",
		Description: "Execute a template on a project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ExecuteTemplateInput) (*mcp.CallToolResult, ExecuteTemplateOutput, error) {
		return executeTemplate(ctx, req, input, automationEngine)
	})

	// List Templates Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_templates",
		Description: "List available automation templates",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListTemplatesInput) (*mcp.CallToolResult, ListTemplatesOutput, error) {
		return listTemplates(ctx, req, input, templateManager)
	})

	// Search Templates Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_templates",
		Description: "Search templates by query",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchTemplatesInput) (*mcp.CallToolResult, SearchTemplatesOutput, error) {
		return searchTemplates(ctx, req, input, templateManager)
	})

	// Create Template Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_template",
		Description: "Create a new custom template",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateTemplateInput) (*mcp.CallToolResult, CreateTemplateOutput, error) {
		return createTemplate(ctx, req, input, templateManager)
	})
}

// ExecuteTemplateInput represents input for execute_template tool
type ExecuteTemplateInput struct {
	TemplateName string            `json:"template_name" jsonschema:"Name of the template to execute"`
	ProjectPath  string            `json:"project_path" jsonschema:"Path to the project directory"`
	CustomParams map[string]string `json:"custom_params,omitempty" jsonschema:"Custom parameters for the template"`
}

// ExecuteTemplateOutput represents output for execute_template tool
type ExecuteTemplateOutput struct {
	TemplateName    string   `json:"template_name"`
	ProjectPath     string   `json:"project_path"`
	Duration        string   `json:"duration"`
	Success         bool     `json:"success"`
	TasksExecuted   int      `json:"tasks_executed"`
	OutputFiles     []string `json:"output_files"`
	Recommendations []string `json:"recommendations"`
}

// executeTemplate executes a template on a project
func executeTemplate(ctx context.Context, req *mcp.CallToolRequest, input ExecuteTemplateInput, engine *automation.Engine) (
	*mcp.CallToolResult,
	ExecuteTemplateOutput,
	error,
) {
	result, err := engine.ExecuteTemplate(ctx, input.TemplateName, input.CustomParams)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to execute template: %v", err)},
			},
		}, ExecuteTemplateOutput{}, err
	}

	output := ExecuteTemplateOutput{
		TemplateName:    result.TemplateName,
		ProjectPath:     result.ProjectPath,
		Duration:        result.Duration.String(),
		Success:         result.Success,
		TasksExecuted:   len(result.TasksExecuted),
		OutputFiles:     result.OutputFiles,
		Recommendations: result.Recommendations,
	}

	return nil, output, nil
}

// ListTemplatesInput represents input for list_templates tool
type ListTemplatesInput struct {
	Category string `json:"category,omitempty" jsonschema:"Filter templates by category"`
}

// ListTemplatesOutput represents output for list_templates tool
type ListTemplatesOutput struct {
	Templates []templates.RegistryTemplate `json:"templates"`
}

// listTemplates lists available templates
func listTemplates(ctx context.Context, req *mcp.CallToolRequest, input ListTemplatesInput, manager *templates.Manager) (
	*mcp.CallToolResult,
	ListTemplatesOutput,
	error,
) {
	var templateList []templates.RegistryTemplate

	if input.Category != "" {
		templateList = manager.ListTemplatesByCategory(input.Category)
	} else {
		templateList = manager.ListTemplates()
	}

	output := ListTemplatesOutput{
		Templates: templateList,
	}

	return nil, output, nil
}

// SearchTemplatesInput represents input for search_templates tool
type SearchTemplatesInput struct {
	Query string `json:"query" jsonschema:"Search query"`
}

// SearchTemplatesOutput represents output for search_templates tool
type SearchTemplatesOutput struct {
	Templates []templates.RegistryTemplate `json:"templates"`
}

// searchTemplates searches templates by query
func searchTemplates(ctx context.Context, req *mcp.CallToolRequest, input SearchTemplatesInput, manager *templates.Manager) (
	*mcp.CallToolResult,
	SearchTemplatesOutput,
	error,
) {
	templateList := manager.SearchTemplates(input.Query)

	output := SearchTemplatesOutput{
		Templates: templateList,
	}

	return nil, output, nil
}

// CreateTemplateInput represents input for create_template tool
type CreateTemplateInput struct {
	Name        string `json:"name" jsonschema:"Template name"`
	Category    string `json:"category" jsonschema:"Template category"`
	Description string `json:"description" jsonschema:"Template description"`
}

// CreateTemplateOutput represents output for create_template tool
type CreateTemplateOutput struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// createTemplate creates a new custom template
func createTemplate(ctx context.Context, req *mcp.CallToolRequest, input CreateTemplateInput, manager *templates.Manager) (
	*mcp.CallToolResult,
	CreateTemplateOutput,
	error,
) {
	template, err := manager.CreateTemplate(input.Name, input.Category, input.Description)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create template: %v", err)},
			},
		}, CreateTemplateOutput{}, err
	}

	if err := manager.SaveTemplate(template); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to save template: %v", err)},
			},
		}, CreateTemplateOutput{}, err
	}

	output := CreateTemplateOutput{
		Name:        template.Metadata.Name,
		Category:    template.Metadata.Category,
		Description: template.Metadata.Description,
		Version:     template.Metadata.Version,
	}

	return nil, output, nil
}
