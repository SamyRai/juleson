package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/internal/services"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterProjectTools registers all project-related MCP tools
func RegisterProjectTools(server *mcp.Server, container *services.Container) {
	// Analyze Project Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "analyze_project",
		Description: "Analyze project structure and create context for automation",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeProjectInput) (*mcp.CallToolResult, AnalyzeProjectOutput, error) {
		return analyzeProject(ctx, req, input, container)
	})

	// Sync Project Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sync_project",
		Description: "Sync project with remote repository",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SyncProjectInput) (*mcp.CallToolResult, SyncProjectOutput, error) {
		return syncProject(ctx, req, input)
	})
}

// AnalyzeProjectInput represents input for analyze_project tool
type AnalyzeProjectInput struct {
	ProjectPath string `json:"project_path" jsonschema:"Path to the project directory"`
}

// AnalyzeProjectOutput represents output for analyze_project tool
type AnalyzeProjectOutput struct {
	ProjectName   string            `json:"project_name"`
	ProjectType   string            `json:"project_type"`
	Languages     []string          `json:"languages"`
	Frameworks    []string          `json:"frameworks"`
	Architecture  string            `json:"architecture"`
	Complexity    string            `json:"complexity"`
	Dependencies  map[string]string `json:"dependencies"`
	FileStructure map[string]int    `json:"file_structure"`
	TestCoverage  float64           `json:"test_coverage"`
	GitStatus     string            `json:"git_status"`
	CodeQuality   *CodeQualityInfo  `json:"code_quality,omitempty"`
}

// CodeQualityInfo represents code quality metrics for MCP output
type CodeQualityInfo struct {
	TestCoverage      float64 `json:"test_coverage"`
	CodeComplexity    float64 `json:"code_complexity"`
	Maintainability   float64 `json:"maintainability"`
	DuplicationRate   float64 `json:"duplication_rate"`
	SecurityIssues    int     `json:"security_issues"`
	CodeSmells        int     `json:"code_smells"`
	PerformanceIssues int     `json:"performance_issues"`
}

// analyzeProject analyzes a project and returns context
func analyzeProject(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeProjectInput, container *services.Container) (
	*mcp.CallToolResult,
	AnalyzeProjectOutput,
	error,
) {
	engine, err := container.AutomationEngine()
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to initialize automation engine: %v", err)},
			},
		}, AnalyzeProjectOutput{}, err
	}

	context, err := engine.AnalyzeProject(input.ProjectPath)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to analyze project: %v", err)},
			},
		}, AnalyzeProjectOutput{}, err
	}

	output := AnalyzeProjectOutput{
		ProjectName:   context.ProjectName,
		ProjectType:   context.ProjectType,
		Languages:     context.Languages,
		Frameworks:    context.Frameworks,
		Architecture:  context.Architecture,
		Complexity:    context.Complexity,
		Dependencies:  context.Dependencies,
		FileStructure: context.FileStructure,
		TestCoverage:  context.TestCoverage,
		GitStatus:     context.GitStatus,
	}

	// Add code quality information if available
	if context.CodeQuality != nil {
		output.CodeQuality = &CodeQualityInfo{
			TestCoverage:      context.CodeQuality.TestCoverage,
			CodeComplexity:    context.CodeQuality.CodeComplexity,
			Maintainability:   context.CodeQuality.Maintainability,
			DuplicationRate:   context.CodeQuality.DuplicationRate,
			SecurityIssues:    len(context.CodeQuality.SecurityIssues),
			CodeSmells:        len(context.CodeQuality.CodeSmells),
			PerformanceIssues: len(context.CodeQuality.PerformanceIssues),
		}
	}

	return nil, output, nil
}

// SyncProjectInput represents input for sync_project tool
type SyncProjectInput struct {
	ProjectPath string `json:"project_path" jsonschema:"Path to the project directory"`
	Remote      string `json:"remote" jsonschema:"Remote repository name"`
}

// SyncProjectOutput represents output for sync_project tool
type SyncProjectOutput struct {
	ProjectPath string `json:"project_path"`
	Remote      string `json:"remote"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// syncProject syncs project with remote repository
func syncProject(ctx context.Context, req *mcp.CallToolRequest, input SyncProjectInput) (
	*mcp.CallToolResult,
	SyncProjectOutput,
	error,
) {
	if err := julesops.SyncGitRepository(ctx, julesops.GitSyncOptions{
		ProjectPath: input.ProjectPath,
		Remote:      input.Remote,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}); err != nil {
		return nil, SyncProjectOutput{}, err
	}

	output := SyncProjectOutput{
		ProjectPath: input.ProjectPath,
		Remote:      input.Remote,
		Status:      "synced",
		Message:     "Project synced successfully",
	}

	return nil, output, nil
}
