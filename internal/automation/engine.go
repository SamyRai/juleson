package automation

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/analyzer"
	"github.com/SamyRai/juleson/internal/jules"
	"github.com/SamyRai/juleson/internal/templates"
)

// Default configuration constants for automation engine
const (
	DefaultSourceListLimit  = 100
	DefaultSessionListLimit = 10
	DefaultBranchName       = "main"
	DefaultAutomationMode   = "AUTO_CREATE_PR"
	DefaultFilePermissions  = 0644
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
	if projectPath == "" {
		return nil, fmt.Errorf("project path cannot be empty")
	}

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
	if templateName == "" {
		return nil, fmt.Errorf("template name cannot be empty")
	}
	if e.projectPath == "" {
		return nil, fmt.Errorf("project must be analyzed before executing templates")
	}

	fmt.Fprintf(os.Stderr, "\n🚀 Starting template execution: %s\n", templateName)
	fmt.Fprintf(os.Stderr, "📁 Project path: %s\n\n", e.projectPath)

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
	fmt.Fprintf(os.Stderr, "\n🔧 Executing task: %s (%s)\n", task.Name, task.Type)

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
		return result, fmt.Errorf("failed to process prompt: %w", err)
	}

	// Get available sources from Jules API
	fmt.Fprintf(os.Stderr, "🔍 Fetching available sources from Jules API...\n")
	sources, err := e.julesClient.ListSources(ctx, DefaultSourceListLimit)
	if err != nil {
		result.Error = fmt.Sprintf("failed to list sources: %v", err)
		result.Success = false
		return result, fmt.Errorf("failed to list sources: %w", err)
	}
	fmt.Fprintf(os.Stderr, "✅ Found %d sources\n", len(sources))

	if len(sources) == 0 {
		err := fmt.Errorf("no sources available - connect a repository via Jules web UI first")
		result.Error = err.Error()
		result.Success = false
		return result, err
	}

	// Match current git repository to a Jules source
	fmt.Fprintf(os.Stderr, "🔍 Matching git repository to Jules source...\n")
	source, err := e.matchGitRepoToSource(sources)
	if err != nil {
		result.Error = fmt.Sprintf("failed to match repository to source: %v", err)
		result.Success = false
		return result, fmt.Errorf("failed to match repository to source: %w", err)
	}
	fmt.Fprintf(os.Stderr, "✅ Matched source: %s\n", source.Name)

	// Detect current branch if it's a git repo
	branch := DefaultBranchName
	gitAnalyzer := analyzer.NewGitAnalyzer()
	if currentBranch, err := gitAnalyzer.GetBranch(e.projectPath); err == nil && currentBranch != "" {
		branch = currentBranch
	}
	fmt.Fprintf(os.Stderr, "🌿 Using branch: %s\n", branch)

	// Check for existing active sessions with similar title
	fmt.Fprintf(os.Stderr, "🔍 Checking for existing active sessions...\n")
	existingSessions, err := e.julesClient.ListSessions(ctx, DefaultSessionListLimit)
	if err == nil {
		taskTitle := fmt.Sprintf("Execute %s task: %s", task.Type, task.Description)
		for _, existingSession := range existingSessions {
			if existingSession.Title == taskTitle &&
				(existingSession.State == "PLANNING" || existingSession.State == "IN_PROGRESS") {
				fmt.Fprintf(os.Stderr, "♻️  Found existing active session: %s\n", existingSession.ID)
				fmt.Fprintf(os.Stderr, "   State: %s\n", existingSession.State)
				fmt.Fprintf(os.Stderr, "   URL: %s\n", existingSession.URL)
				fmt.Fprintf(os.Stderr, "   Reusing existing session instead of creating a new one.\n")

				result.JulesSessionID = existingSession.ID
				result.Output = fmt.Sprintf("Reused existing session: %s", existingSession.URL)
				result.Success = true
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)

				// Note: Activities will be available as the session progresses
				// User can monitor via the Jules web UI
				return result, nil
			}
		}
	}

	// Create Jules session with comprehensive error handling
	result, err = e.createJulesSession(ctx, prompt, task, source, branch)
	if err != nil {
		return result, fmt.Errorf("failed to create Jules session: %w", err)
	}

	return result, nil
}

// createJulesSession creates a new Jules session with proper error handling
func (e *Engine) createJulesSession(ctx context.Context, prompt string, task templates.TemplateTask, source *jules.Source, branch string) (*TaskExecutionResult, error) {
	result := &TaskExecutionResult{
		TaskName:  task.Name,
		TaskType:  task.Type,
		StartTime: time.Now(),
		Metrics:   make(map[string]interface{}),
	}

	fmt.Fprintf(os.Stderr, "🚀 Creating new Jules session...\n")
	fmt.Fprintf(os.Stderr, "   Prompt: %s\n", prompt)
	fmt.Fprintf(os.Stderr, "   Source: %s\n", source.Name)
	fmt.Fprintf(os.Stderr, "   Branch: %s\n", branch)

	sessionReq := &jules.CreateSessionRequest{
		Prompt: prompt,
		Title:  fmt.Sprintf("Execute %s task: %s", task.Type, task.Description),
		SourceContext: &jules.SourceContext{
			Source: source.Name, // Use proper source identifier (sources/github/owner/repo)
			GithubRepoContext: &jules.GithubRepoContext{
				StartingBranch: branch,
			},
		},
		RequirePlanApproval: task.RequiresApproval,
		AutomationMode:      DefaultAutomationMode,
	}

	session, err := e.julesClient.CreateSession(ctx, sessionReq)
	if err != nil {
		result.Error = err.Error()
		result.Success = false
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, fmt.Errorf("create session API call failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✅ Session created successfully:\n")
	fmt.Fprintf(os.Stderr, "   ID: %s\n", session.ID)
	fmt.Fprintf(os.Stderr, "   Name: %s\n", session.Name)
	fmt.Fprintf(os.Stderr, "   URL: %s\n", session.URL)
	fmt.Fprintf(os.Stderr, "   State: %s\n", session.State)
	fmt.Fprintf(os.Stderr, "\n💡 Monitor progress at: %s\n", session.URL)
	fmt.Fprintf(os.Stderr, "💡 The session will run asynchronously. Activities and results will be available as Jules works.\n")

	result.JulesSessionID = session.ID
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
	return os.WriteFile(filePath, []byte(content), DefaultFilePermissions)
}

// matchGitRepoToSource matches the current git repository to a Jules source
func (e *Engine) matchGitRepoToSource(sources []jules.Source) (*jules.Source, error) {
	// Get git remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = e.projectPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git remote: %w (ensure project is a git repository)", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Parse owner and repo from remote URL
	// Supports: https://github.com/owner/repo.git, git@github.com:owner/repo.git
	var owner, repo string

	if strings.Contains(remoteURL, "github.com") {
		// Remove .git suffix if present
		remoteURL = strings.TrimSuffix(remoteURL, ".git")

		// Extract from HTTPS URL: https://github.com/owner/repo
		if strings.HasPrefix(remoteURL, "https://") {
			parts := strings.Split(remoteURL, "/")
			if len(parts) >= 2 {
				repo = parts[len(parts)-1]
				owner = parts[len(parts)-2]
			}
		} else if strings.HasPrefix(remoteURL, "git@") {
			// Extract from SSH URL: git@github.com:owner/repo
			parts := strings.Split(remoteURL, ":")
			if len(parts) == 2 {
				pathParts := strings.Split(parts[1], "/")
				if len(pathParts) == 2 {
					owner = pathParts[0]
					repo = pathParts[1]
				}
			}
		}
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("failed to parse GitHub owner/repo from remote URL: %s", remoteURL)
	}

	// Match against available sources
	expectedSourceName := fmt.Sprintf("sources/github/%s/%s", owner, repo)
	for i := range sources {
		if sources[i].Name == expectedSourceName {
			return &sources[i], nil
		}
	}

	return nil, fmt.Errorf("repository %s/%s not found in connected sources - connect it via Jules web UI at https://jules.google.com", owner, repo)
}
