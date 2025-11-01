package automation

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"jules-automation/internal/jules"
	"jules-automation/internal/templates"
)

// Engine represents the automation engine
type Engine struct {
	julesClient     *jules.Client
	templateManager *templates.Manager
	projectPath     string
	context         *ProjectContext
}

// ProjectContext contains project analysis context
type ProjectContext struct {
	ProjectPath   string            `json:"project_path"`
	ProjectName   string            `json:"project_name"`
	ProjectType   string            `json:"project_type"`
	Languages     []string          `json:"languages"`
	Frameworks    []string          `json:"frameworks"`
	Dependencies  map[string]string `json:"dependencies"`
	FileStructure map[string]int    `json:"file_structure"`
	TestCoverage  float64           `json:"test_coverage"`
	Architecture  string            `json:"architecture"`
	Complexity    string            `json:"complexity"`
	LastModified  time.Time         `json:"last_modified"`
	GitStatus     string            `json:"git_status"`
	CustomParams  map[string]string `json:"custom_params"`
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
	}
}

// AnalyzeProject analyzes the project and creates context
func (e *Engine) AnalyzeProject(projectPath string) (*ProjectContext, error) {
	// Extract project name from path
	projectName := filepath.Base(projectPath)

	// Analyze project structure
	fileStructure, err := e.analyzeFileStructure(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze file structure: %w", err)
	}

	// Detect languages and frameworks
	languages, frameworks, err := e.detectLanguagesAndFrameworks(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect languages and frameworks: %w", err)
	}

	// Analyze dependencies
	dependencies, err := e.analyzeDependencies(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze dependencies: %w", err)
	}

	// Detect project type
	projectType := e.detectProjectType(languages, frameworks)

	// Analyze architecture
	architecture := e.analyzeArchitecture(projectPath, fileStructure)

	// Calculate complexity
	complexity := e.calculateComplexity(fileStructure, dependencies)

	// Get git status
	gitStatus := e.getGitStatus(projectPath)

	context := &ProjectContext{
		ProjectPath:   projectPath,
		ProjectName:   projectName,
		ProjectType:   projectType,
		Languages:     languages,
		Frameworks:    frameworks,
		Dependencies:  dependencies,
		FileStructure: fileStructure,
		TestCoverage:  0.0, // TODO: Calculate actual test coverage
		Architecture:  architecture,
		Complexity:    complexity,
		LastModified:  time.Now(),
		GitStatus:     gitStatus,
		CustomParams:  make(map[string]string),
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

// Helper methods for project analysis (simplified implementations)

func (e *Engine) analyzeFileStructure(projectPath string) (map[string]int, error) {
	// Simplified implementation - count files by extension
	structure := make(map[string]int)
	// TODO: Implement actual file structure analysis
	return structure, nil
}

func (e *Engine) detectLanguagesAndFrameworks(projectPath string) ([]string, []string, error) {
	// Simplified implementation
	languages := []string{"go"} // Default to Go for now
	frameworks := []string{}
	// TODO: Implement actual language and framework detection
	return languages, frameworks, nil
}

func (e *Engine) analyzeDependencies(projectPath string) (map[string]string, error) {
	// Simplified implementation
	dependencies := make(map[string]string)
	// TODO: Implement actual dependency analysis
	return dependencies, nil
}

func (e *Engine) detectProjectType(languages []string, frameworks []string) string {
	if len(languages) == 0 {
		return "unknown"
	}
	return languages[0] // Simplified - use first language
}

func (e *Engine) analyzeArchitecture(projectPath string, fileStructure map[string]int) string {
	// Simplified implementation
	return "monolithic" // Default architecture
}

func (e *Engine) calculateComplexity(fileStructure map[string]int, dependencies map[string]string) string {
	// Simplified implementation
	return "medium" // Default complexity
}

func (e *Engine) getGitStatus(projectPath string) string {
	// Simplified implementation
	return "clean" // Default git status
}

func (e *Engine) generateFileContent(templateName string, result *ExecutionResult) (string, error) {
	// Simplified implementation - generate basic markdown report
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

func (e *Engine) writeFile(filePath string, content string) error {
	// TODO: Implement file writing with proper error handling
	return nil
}
