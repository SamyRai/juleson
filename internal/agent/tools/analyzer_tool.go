package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
	"github.com/SamyRai/juleson/internal/analyzer"
)

// AnalyzerTool provides project analysis capabilities to the agent
type AnalyzerTool struct {
	analyzer *analyzer.ProjectAnalyzer
}

// NewAnalyzerTool creates a new analyzer tool
func NewAnalyzerTool() *AnalyzerTool {
	return &AnalyzerTool{
		analyzer: analyzer.NewProjectAnalyzer(),
	}
}

// Name returns the tool name
func (a *AnalyzerTool) Name() string {
	return "analyzer"
}

// Description returns what this tool does
func (a *AnalyzerTool) Description() string {
	return "Analyze project structure, languages, frameworks, dependencies, and code quality metrics"
}

// Parameters returns tool parameters
func (a *AnalyzerTool) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "action",
			Description: "Action to perform: analyze_project",
			Type:        ParameterTypeString,
			Required:    true,
		},
		{
			Name:        "project_path",
			Description: "Path to the project directory to analyze",
			Type:        ParameterTypeString,
			Required:    true,
		},
	}
}

// Execute runs the analyzer tool
func (a *AnalyzerTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	action, ok := params["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action parameter is required")
	}

	switch action {
	case "analyze_project":
		return a.analyzeProject(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// RequiresApproval returns whether this tool needs approval
func (a *AnalyzerTool) RequiresApproval() bool {
	return false // Analysis is safe and doesn't modify anything
}

// CanHandle returns whether this tool can handle a task
func (a *AnalyzerTool) CanHandle(task agent.Task) bool {
	// Can handle analysis-related tasks
	return task.Tool == "analyzer" ||
		containsString(task.Description, "analyze") ||
		containsString(task.Description, "examine") ||
		containsString(task.Description, "inspect") ||
		containsString(task.Description, "review") ||
		containsString(task.Prompt, "analyze") ||
		containsString(task.Prompt, "examine") ||
		containsString(task.Prompt, "inspect") ||
		containsString(task.Prompt, "review")
}

// analyzeProject analyzes a project and returns detailed context
func (a *AnalyzerTool) analyzeProject(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	projectPath, ok := params["project_path"].(string)
	if !ok {
		return nil, fmt.Errorf("project_path parameter is required")
	}

	// Perform analysis
	context, err := a.analyzer.Analyze(projectPath)
	if err != nil {
		return &ToolResult{
			Success:  false,
			Error:    err,
			Duration: time.Since(time.Now()).Milliseconds(),
		}, err
	}

	// Convert to tool result format
	output := map[string]interface{}{
		"project_name":   context.ProjectName,
		"project_type":   context.ProjectType,
		"languages":      context.Languages,
		"frameworks":     context.Frameworks,
		"architecture":   context.Architecture,
		"complexity":     context.Complexity,
		"dependencies":   context.Dependencies,
		"file_structure": context.FileStructure,
		"git_status":     context.GitStatus,
	}

	// Add code quality information if available
	if context.CodeQuality != nil {
		output["code_quality"] = map[string]interface{}{
			"test_coverage":      context.CodeQuality.TestCoverage,
			"code_complexity":    context.CodeQuality.CodeComplexity,
			"maintainability":    context.CodeQuality.Maintainability,
			"duplication_rate":   context.CodeQuality.DuplicationRate,
			"security_issues":    len(context.CodeQuality.SecurityIssues),
			"code_smells":        len(context.CodeQuality.CodeSmells),
			"performance_issues": len(context.CodeQuality.PerformanceIssues),
		}
	}

	result := &ToolResult{
		Success:  true,
		Duration: time.Since(time.Now()).Milliseconds(),
		Output:   output,
	}

	return result, nil
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
