package automation

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/analyzer"
	"github.com/SamyRai/juleson/internal/jules"
	"github.com/SamyRai/juleson/internal/templates"
)

// Engine represents the automation engine
type Engine struct {
	julesClient     *jules.Client
	templateManager *templates.Manager
	projectAnalyzer *analyzer.ProjectAnalyzer
	projectPath     string
	context         *analyzer.ProjectContext
}

// ExecutionResult represents the result of template execution
type ExecutionResult struct {
	TemplateName    string                 `json:"template_name"`
	ProjectPath     string                 `json:"project_path"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Duration        time.Duration          `json:"duration"`
	TasksExecuted   []TaskExecutionResult  `json:"tasks_executed"`
	Success         bool                   `json:"success"`
	Error           string                 `json:"error,omitempty"`
	OutputFiles     []string               `json:"output_files"`
	Recommendations []string               `json:"recommendations"`
	Metrics         map[string]interface{} `json:"metrics"`
}

// TaskExecutionResult represents the result of a single task execution
type TaskExecutionResult struct {
	TaskName       string                 `json:"task_name"`
	TaskType       string                 `json:"task_type"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        time.Time              `json:"end_time"`
	Duration       time.Duration          `json:"duration"`
	Success        bool                   `json:"success"`
	Error          string                 `json:"error,omitempty"`
	JulesSessionID string                 `json:"jules_session_id,omitempty"`
	Activities     []jules.Activity       `json:"activities,omitempty"`
	Output         string                 `json:"output,omitempty"`
	Metrics        map[string]interface{} `json:"metrics"`
}

// NewEngine creates a new automation engine
func NewEngine(julesClient *jules.Client, templateManager *templates.Manager) *Engine {
	return &Engine{
		julesClient:     julesClient,
		templateManager: templateManager,
		projectAnalyzer: analyzer.NewProjectAnalyzer(),
	}
}

// AnalyzeProject analyzes the project and creates context
func (e *Engine) AnalyzeProject(projectPath string) (*analyzer.ProjectContext, error) {
	context, err := e.projectAnalyzer.Analyze(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project: %w", err)
	}

	e.projectPath = projectPath
	e.context = context

	return context, nil
}

// ExecuteTemplate executes a template on the project
func (e *Engine) ExecuteTemplate(ctx context.Context, templateName string, customParams map[string]string) (*ExecutionResult, error) {
	// Load template
	template, err := e.templateManager.LoadTemplate(templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Merge custom parameters
	if e.context != nil {
		for k, v := range customParams {
			e.context.CustomParams[k] = v
		}
	}

	// Create execution result
	result := &ExecutionResult{
		TemplateName:    templateName,
		ProjectPath:     e.projectPath,
		StartTime:       time.Now(),
		TasksExecuted:   make([]TaskExecutionResult, 0),
		OutputFiles:     make([]string, 0),
		Recommendations: make([]string, 0),
		Metrics:         make(map[string]interface{}),
	}

	// Execute tasks in dependency order
	taskResults, err := e.executeTasks(ctx, template.Tasks)
	if err != nil {
		result.Error = err.Error()
		result.Success = false
	} else {
		result.TasksExecuted = taskResults
		result.Success = true
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Generate output files
	if err := e.generateOutputFiles(template, result); err != nil {
		result.Error = fmt.Sprintf("%s; output generation failed: %v", result.Error, err)
	}

	return result, nil
}

// executeTasks executes template tasks in dependency order
func (e *Engine) executeTasks(ctx context.Context, tasks []templates.TemplateTask) ([]TaskExecutionResult, error) {
	var results []TaskExecutionResult
	executed := make(map[string]bool)

	// Execute tasks in dependency order
	for len(executed) < len(tasks) {
		progress := false

		for _, task := range tasks {
			if executed[task.Name] {
				continue
			}

			// Check if dependencies are satisfied
			depsSatisfied := true
			for _, dep := range task.DependsOn {
				if !executed[dep] {
					depsSatisfied = false
					break
				}
			}

			if !depsSatisfied {
				continue
			}

			// Execute task
			result, err := e.executeTask(ctx, task)
			if err != nil {
				return results, fmt.Errorf("task '%s' failed: %w", task.Name, err)
			}

			results = append(results, *result)
			executed[task.Name] = true
			progress = true
		}

		if !progress {
			return results, fmt.Errorf("circular dependency detected in tasks")
		}
	}

	return results, nil
}

// executeTask executes a single task
func (e *Engine) executeTask(ctx context.Context, task templates.TemplateTask) (*TaskExecutionResult, error) {
	result := &TaskExecutionResult{
		TaskName:  task.Name,
		TaskType:  task.Type,
		StartTime: time.Now(),
		Metrics:   make(map[string]interface{}),
	}

	// Process Jules prompt with context variables
	prompt, err := e.processPrompt(task.JulesPrompt, task.ContextVars)
	if err != nil {
		result.Error = err.Error()
		result.Success = false
		return result, err
	}

	// Create Jules session
	sessionReq := &jules.CreateSessionRequest{
		Prompt: prompt,
		Title:  fmt.Sprintf("Execute %s task: %s", task.Type, task.Description),
		SourceContext: &jules.SourceContext{
			Source: e.projectPath,
		},
		RequirePlanApproval: task.RequiresApproval,
		AutomationMode:      "AUTO_CREATE_PR",
	}

	session, err := e.julesClient.CreateSession(ctx, sessionReq)
	if err != nil {
		result.Error = err.Error()
		result.Success = false
		return result, err
	}

	result.JulesSessionID = session.ID

	// Wait for session to complete or timeout
	// In a real implementation, you might poll the session status
	// For now, we'll just fetch the activities

	// Get activities
	activities, err := e.julesClient.ListActivities(ctx, session.ID, 100)
	if err != nil {
		result.Error = err.Error()
		result.Success = false
		return result, err
	}

	result.Activities = activities
	result.Output = fmt.Sprintf("Session created: %s", session.URL)
	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// processPrompt processes a Jules prompt with context variables
func (e *Engine) processPrompt(prompt string, contextVars map[string]string) (string, error) {
	processed := prompt

	// Replace context variables
	for key := range contextVars {
		placeholder := fmt.Sprintf("{{.%s}}", key)

		// Get actual value from context
		actualValue := e.getContextValue(key)
		processed = strings.ReplaceAll(processed, placeholder, actualValue)
	}

	// Replace built-in variables
	builtins := map[string]string{
		"ProjectPath": e.projectPath,
		"ProjectName": e.context.ProjectName,
		"ProjectType": e.context.ProjectType,
		"Timestamp":   time.Now().Format(time.RFC3339),
	}

	for key, value := range builtins {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		processed = strings.ReplaceAll(processed, placeholder, value)
	}

	return processed, nil
}

// getContextValue gets a value from the project context
func (e *Engine) getContextValue(key string) string {
	if e.context == nil {
		return ""
	}

	switch key {
	case "ProjectPath":
		return e.context.ProjectPath
	case "ProjectName":
		return e.context.ProjectName
	case "ProjectType":
		return e.context.ProjectType
	case "Languages":
		return strings.Join(e.context.Languages, ", ")
	case "Frameworks":
		return strings.Join(e.context.Frameworks, ", ")
	case "Architecture":
		return e.context.Architecture
	case "Complexity":
		return e.context.Complexity
	case "GitStatus":
		return e.context.GitStatus
	default:
		// Check custom parameters
		if value, exists := e.context.CustomParams[key]; exists {
			return value
		}
		return ""
	}
}

// generateOutputFiles generates output files based on template configuration
func (e *Engine) generateOutputFiles(template *templates.Template, result *ExecutionResult) error {
	for _, outputFile := range template.Output.Files {
		// Process file path with context variables
		filePath, err := e.processPrompt(outputFile.Path, make(map[string]string))
		if err != nil {
			return fmt.Errorf("failed to process output file path: %w", err)
		}

		// Generate file content based on template
		content, err := e.generateFileContent(outputFile.Template, result)
		if err != nil {
			return fmt.Errorf("failed to generate file content: %w", err)
		}

		// Write file
		if err := e.writeFile(filePath, content); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		result.OutputFiles = append(result.OutputFiles, filePath)
	}

	return nil
}

// generateFileContent generates markdown report content
func (e *Engine) generateFileContent(templateName string, result *ExecutionResult) (string, error) {
	content := fmt.Sprintf(`# %s Execution Report

## Summary
- Template: %s
- Project: %s
- Duration: %v
- Success: %t

## Tasks Executed
`, templateName, result.TemplateName, result.ProjectPath, result.Duration, result.Success)

	for _, task := range result.TasksExecuted {
		content += fmt.Sprintf("- %s (%s): %t\n", task.TaskName, task.TaskType, task.Success)
	}

	return content, nil
}

// writeFile writes content to a file
func (e *Engine) writeFile(filePath string, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
