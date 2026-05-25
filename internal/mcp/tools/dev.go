package tools

import (
	"context"
	"fmt"
	"runtime"

	"github.com/SamyRai/juleson/internal/orchestrator"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterDevTools registers developer tool MCP tools
func RegisterDevTools(server *mcp.Server) {
	// Build tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "build_project",
		Description: "Build Juleson binaries with various options",
	}, buildProjectHandler)

	// Test tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "run_tests",
		Description: "Run tests with coverage and various options",
	}, runTestsHandler)

	// Lint tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "lint_code",
		Description: "Run linters to check code quality",
	}, lintCodeHandler)

	// Format tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "format_code",
		Description: "Format Go code",
	}, formatCodeHandler)

	// Clean tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "clean_artifacts",
		Description: "Clean build artifacts and caches",
	}, cleanArtifactsHandler)

	// Quality check tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "quality_check",
		Description: "Run all quality checks (format, lint, test)",
	}, qualityCheckHandler)

	// Module maintenance tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "module_maintenance",
		Description: "Go module maintenance operations",
	}, moduleMaintenanceHandler)

	// Release build tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "build_release",
		Description: "Build release binaries for all platforms",
	}, buildReleaseHandler)
}

// Input/Output types

type BuildProjectInput struct {
	Target  string `json:"target" jsonschema:"Build target: 'cli', 'mcp', or 'all' (default: all)"`
	Version string `json:"version" jsonschema:"Version to embed in binaries (default: dev)"`
	Race    bool   `json:"race" jsonschema:"Enable race detection (default: false)"`
	GOOS    string `json:"goos" jsonschema:"Target operating system (default: current)"`
	GOARCH  string `json:"goarch" jsonschema:"Target architecture (default: current)"`
}

type BuildProjectOutput struct {
	Target  string   `json:"target"`
	Results []string `json:"results"`
	Summary string   `json:"summary"`
}

type RunTestsInput struct {
	Verbose  bool     `json:"verbose" jsonschema:"Verbose output (default: true)"`
	Race     bool     `json:"race" jsonschema:"Enable race detection (default: true)"`
	Cover    bool     `json:"cover" jsonschema:"Enable coverage reporting (default: false)"`
	Short    bool     `json:"short" jsonschema:"Run short tests only (default: false)"`
	Packages []string `json:"packages" jsonschema:"Specific packages to test"`
}

type RunTestsOutput struct {
	Success     bool    `json:"success"`
	Duration    string  `json:"duration"`
	TestsPassed int     `json:"tests_passed"`
	TestsFailed int     `json:"tests_failed"`
	Coverage    float64 `json:"coverage,omitempty"`
	Summary     string  `json:"summary"`
}

type LintCodeInput struct {
	Fix     bool `json:"fix" jsonschema:"Automatically fix issues (default: false)"`
	Verbose bool `json:"verbose" jsonschema:"Verbose output (default: false)"`
}

type LintCodeOutput struct {
	Success bool   `json:"success"`
	Issues  int    `json:"issues"`
	Summary string `json:"summary"`
}

type FormatCodeInput struct {
	UseGofumpt bool `json:"use_gofumpt" jsonschema:"Use gofumpt instead of go fmt (default: false)"`
}

type FormatCodeOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CleanArtifactsInput struct {
	All      bool `json:"all" jsonschema:"Clean everything including caches (default: false)"`
	Cache    bool `json:"cache" jsonschema:"Clean build cache (default: false)"`
	ModCache bool `json:"modcache" jsonschema:"Clean module cache (default: false)"`
}

type CleanArtifactsOutput struct {
	Success bool     `json:"success"`
	Cleaned []string `json:"cleaned"`
	Summary string   `json:"summary"`
}

type QualityCheckInput struct{}

type QualityCheckOutput struct {
	Success  bool     `json:"success"`
	Checks   []string `json:"checks"`
	Coverage float64  `json:"coverage,omitempty"`
	Summary  string   `json:"summary"`
}

type ModuleMaintenanceInput struct {
	Operation string   `json:"operation" jsonschema:"Operation: tidy, download, verify, vendor, graph, why"`
	Packages  []string `json:"packages,omitempty" jsonschema:"Packages for the why operation"`
}

type ModuleMaintenanceOutput struct {
	Success   bool   `json:"success"`
	Operation string `json:"operation"`
	Message   string `json:"message"`
}

type BuildReleaseInput struct {
	Version string `json:"version" jsonschema:"Release version (e.g., v1.0.0)"`
}

type BuildReleaseOutput struct {
	Success          bool     `json:"success"`
	Version          string   `json:"version"`
	TotalBuilds      int      `json:"total_builds"`
	SuccessfulBuilds int      `json:"successful_builds"`
	Platforms        []string `json:"platforms"`
	Summary          string   `json:"summary"`
}

// Handler functions

func devToolOrchestrator() *orchestrator.Service {
	return orchestrator.NewService(orchestrator.DefaultConfig("dev", "", ""))
}

func buildProjectHandler(ctx context.Context, req *mcp.CallToolRequest, input BuildProjectInput) (
	*mcp.CallToolResult,
	BuildProjectOutput,
	error,
) {
	// Set defaults
	if input.Target == "" {
		input.Target = "all"
	}
	if input.Version == "" {
		input.Version = "dev"
	}
	if input.GOOS == "" {
		input.GOOS = runtime.GOOS
	}
	if input.GOARCH == "" {
		input.GOARCH = runtime.GOARCH
	}

	summary, err := devToolOrchestrator().BuildWithResults(ctx, orchestrator.BuildOptions{
		Target:  input.Target,
		Version: input.Version,
		Race:    input.Race,
		GOOS:    input.GOOS,
		GOARCH:  input.GOARCH,
	})
	results := make([]string, 0, len(summary.Results))
	for _, result := range summary.Results {
		label := result.Name
		if result.Success {
			results = append(results, fmt.Sprintf("✅ %s: %s", label, result.String()))
		} else {
			results = append(results, fmt.Sprintf("❌ %s: %s", label, result.String()))
		}
	}

	output := BuildProjectOutput{
		Target:  input.Target,
		Results: results,
		Summary: fmt.Sprintf("Build Summary: %d/%d successful", summary.SuccessCount, len(summary.Results)),
	}

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Build failed: %s", output.Summary)},
			},
		}, output, nil
	}

	return nil, output, nil
}

func runTestsHandler(ctx context.Context, req *mcp.CallToolRequest, input RunTestsInput) (
	*mcp.CallToolResult,
	RunTestsOutput,
	error,
) {
	config := orchestrator.DefaultTestConfig()
	config.Verbose = input.Verbose
	config.Race = input.Race
	config.Cover = input.Cover
	config.Short = input.Short

	if len(input.Packages) > 0 {
		config.Packages = input.Packages
	}

	if input.Cover {
		config.CoverProfile = "coverage.out"
	}

	result := devToolOrchestrator().RunTestsWithResult(ctx, config)

	output := RunTestsOutput{
		Success:  result.Success,
		Duration: result.Duration.String(),
		Summary:  result.String(),
	}

	if !result.Success {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Tests failed: %s", result.String())},
			},
		}, output, nil
	}

	return nil, output, nil
}

func lintCodeHandler(ctx context.Context, req *mcp.CallToolRequest, input LintCodeInput) (
	*mcp.CallToolResult,
	LintCodeOutput,
	error,
) {
	config := orchestrator.DefaultLintConfig()
	config.FixMode = input.Fix
	config.Verbose = input.Verbose

	result := devToolOrchestrator().LintWithResult(ctx, config)

	output := LintCodeOutput{
		Success: result.Success,
		Summary: result.String(),
	}

	if !result.Success {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Linting failed: %s", result.String())},
			},
		}, output, nil
	}

	return nil, output, nil
}

func formatCodeHandler(ctx context.Context, req *mcp.CallToolRequest, input FormatCodeInput) (
	*mcp.CallToolResult,
	FormatCodeOutput,
	error,
) {
	if err := devToolOrchestrator().FormatCode(ctx, input.UseGofumpt); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Format failed: %v", err)},
			},
		}, FormatCodeOutput{Success: false, Message: err.Error()}, nil
	}

	return nil, FormatCodeOutput{Success: true, Message: "Code formatted successfully"}, nil
}

func cleanArtifactsHandler(ctx context.Context, req *mcp.CallToolRequest, input CleanArtifactsInput) (
	*mcp.CallToolResult,
	CleanArtifactsOutput,
	error,
) {
	cleaned, err := devToolOrchestrator().CleanArtifacts(ctx, orchestrator.CleanOptions{
		All:      input.All,
		Cache:    input.Cache,
		ModCache: input.ModCache,
	})
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Clean failed: %v", err)},
			},
		}, CleanArtifactsOutput{Success: false}, nil
	}

	return nil, CleanArtifactsOutput{
		Success: true,
		Cleaned: cleaned,
		Summary: fmt.Sprintf("Cleaned: %v", cleaned),
	}, nil
}

func qualityCheckHandler(ctx context.Context, req *mcp.CallToolRequest, input QualityCheckInput) (
	*mcp.CallToolResult,
	QualityCheckOutput,
	error,
) {
	service := devToolOrchestrator()
	config := orchestrator.DefaultTestConfig()
	config.Cover = true
	config.CoverProfile = "coverage.out"

	summary, err := service.RunQualityChecks(ctx, orchestrator.QualityOptions{
		Format:     true,
		Lint:       true,
		Test:       true,
		TestConfig: config,
		Build:      true,
	})
	checks := make([]string, 0, len(summary.Checks))
	for _, check := range summary.Checks {
		checks = append(checks, "✅ "+check)
	}
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, QualityCheckOutput{Success: false, Checks: checks}, nil
	}

	return nil, QualityCheckOutput{
		Success: true,
		Checks:  checks,
		Summary: "All quality checks passed",
	}, nil
}

func moduleMaintenanceHandler(ctx context.Context, req *mcp.CallToolRequest, input ModuleMaintenanceInput) (
	*mcp.CallToolResult,
	ModuleMaintenanceOutput,
	error,
) {
	if input.Operation != "tidy" && input.Operation != "download" && input.Operation != "verify" &&
		input.Operation != "vendor" && input.Operation != "graph" && input.Operation != "why" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Unknown operation: %s", input.Operation)},
			},
		}, ModuleMaintenanceOutput{Success: false, Operation: input.Operation}, nil
	}

	err := devToolOrchestrator().RunModuleMaintenance(ctx, input.Operation, input.Packages...)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Operation failed: %v", err)},
			},
		}, ModuleMaintenanceOutput{Success: false, Operation: input.Operation, Message: err.Error()}, nil
	}

	return nil, ModuleMaintenanceOutput{
		Success:   true,
		Operation: input.Operation,
		Message:   fmt.Sprintf("%s completed successfully", input.Operation),
	}, nil
}

func buildReleaseHandler(ctx context.Context, req *mcp.CallToolRequest, input BuildReleaseInput) (
	*mcp.CallToolResult,
	BuildReleaseOutput,
	error,
) {
	if input.Version == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Version is required"},
			},
		}, BuildReleaseOutput{}, nil
	}

	summary, err := devToolOrchestrator().ReleaseWithResults(ctx, input.Version)
	var builtPlatforms []string
	for _, result := range summary.Results {
		if result.Success {
			builtPlatforms = append(builtPlatforms, result.OutputPath)
		}
	}

	output := BuildReleaseOutput{
		Success:          summary.SuccessCount == len(summary.Results),
		Version:          input.Version,
		TotalBuilds:      len(summary.Results),
		SuccessfulBuilds: summary.SuccessCount,
		Platforms:        builtPlatforms,
		Summary:          fmt.Sprintf("Release %s: %d/%d builds successful", input.Version, summary.SuccessCount, len(summary.Results)),
	}

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Release build incomplete: %s", output.Summary)},
			},
		}, output, nil
	}

	return nil, output, nil
}
