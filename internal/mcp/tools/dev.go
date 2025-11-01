package tools

import (
	"context"
	"fmt"
	"runtime"

	"github.com/SamyRai/juleson/internal/build"
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
	Operation string `json:"operation" jsonschema:"Operation: tidy, download, verify, graph"`
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

	var results []string
	buildCLI := input.Target == "cli" || input.Target == "all"
	buildMCP := input.Target == "mcp" || input.Target == "all"

	successCount := 0
	totalCount := 0

	// Build CLI
	if buildCLI {
		totalCount++
		config := build.DefaultConfig("juleson", "./cmd/juleson")
		config.Version = input.Version
		config.GOOS = input.GOOS
		config.GOARCH = input.GOARCH
		config.Race = input.Race

		if input.Version != "dev" {
			config.LDFlags = append(config.LDFlags, fmt.Sprintf("-X main.version=%s", input.Version))
		}

		builder := build.NewBuilder(config)
		result := builder.BuildWithResult(ctx)

		if result.Success {
			results = append(results, fmt.Sprintf("✅ CLI: %s", result.String()))
			successCount++
		} else {
			results = append(results, fmt.Sprintf("❌ CLI: %s", result.String()))
		}
	}

	// Build MCP
	if buildMCP {
		totalCount++
		config := build.DefaultConfig("juleson-mcp", "./cmd/jules-mcp")
		config.Version = input.Version
		config.GOOS = input.GOOS
		config.GOARCH = input.GOARCH
		config.Race = input.Race

		if input.Version != "dev" {
			config.LDFlags = append(config.LDFlags, fmt.Sprintf("-X main.version=%s", input.Version))
		}

		builder := build.NewBuilder(config)
		result := builder.BuildWithResult(ctx)

		if result.Success {
			results = append(results, fmt.Sprintf("✅ MCP: %s", result.String()))
			successCount++
		} else {
			results = append(results, fmt.Sprintf("❌ MCP: %s", result.String()))
		}
	}

	output := BuildProjectOutput{
		Target:  input.Target,
		Results: results,
		Summary: fmt.Sprintf("Build Summary: %d/%d successful", successCount, totalCount),
	}

	if successCount < totalCount {
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
	config := build.DefaultTestConfig()
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

	tester := build.NewTester(config)
	result := tester.TestWithResult(ctx)

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
	config := build.DefaultLintConfig()
	config.FixMode = input.Fix
	config.Verbose = input.Verbose

	linter := build.NewLinter(config)
	result := linter.LintWithResult(ctx)

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
	formatter := build.NewFormatter()

	var err error
	if input.UseGofumpt {
		err = formatter.FormatWithGofumpt(ctx)
	} else {
		err = formatter.Format(ctx)
	}

	if err != nil {
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
	cleaner := build.NewCleaner("bin", []string{"coverage.out", "coverage.html"})
	var cleaned []string

	if input.All {
		if err := cleaner.CleanAll(ctx); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Clean failed: %v", err)},
				},
			}, CleanArtifactsOutput{Success: false}, nil
		}
		cleaned = []string{"artifacts", "build cache", "module cache"}
	} else {
		if err := cleaner.Clean(ctx); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Clean failed: %v", err)},
				},
			}, CleanArtifactsOutput{Success: false}, nil
		}
		cleaned = append(cleaned, "artifacts")

		if input.Cache {
			if err := cleaner.CleanCache(ctx); err == nil {
				cleaned = append(cleaned, "build cache")
			}
		}

		if input.ModCache {
			if err := cleaner.CleanModCache(ctx); err == nil {
				cleaned = append(cleaned, "module cache")
			}
		}
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
	var checks []string

	// Format
	formatter := build.NewFormatter()
	if err := formatter.Format(ctx); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Format failed: %v", err)},
			},
		}, QualityCheckOutput{Success: false, Checks: checks}, nil
	}
	checks = append(checks, "✅ Format")

	// Lint
	linter := build.NewLinter(build.DefaultLintConfig())
	if err := linter.Lint(ctx); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Lint failed: %v", err)},
			},
		}, QualityCheckOutput{Success: false, Checks: checks}, nil
	}
	checks = append(checks, "✅ Lint")

	// Test
	config := build.DefaultTestConfig()
	config.Cover = true
	config.CoverProfile = "coverage.out"

	tester := build.NewTester(config)
	result := tester.TestWithResult(ctx)

	if !result.Success {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Tests failed: %v", result.Error)},
			},
		}, QualityCheckOutput{Success: false, Checks: checks}, nil
	}
	checks = append(checks, "✅ Tests")

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
	manager := build.NewModuleManager()

	var err error
	switch input.Operation {
	case "tidy":
		err = manager.Tidy(ctx)
	case "download":
		err = manager.Download(ctx)
	case "verify":
		err = manager.Verify(ctx)
	case "graph":
		err = manager.Graph(ctx)
	default:
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Unknown operation: %s", input.Operation)},
			},
		}, ModuleMaintenanceOutput{Success: false, Operation: input.Operation}, nil
	}

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

	platforms := []struct {
		goos   string
		goarch string
	}{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	var builtPlatforms []string
	successCount := 0
	totalCount := 0

	for _, platform := range platforms {
		for _, binary := range []struct {
			name string
			path string
		}{
			{"juleson", "./cmd/juleson"},
			{"juleson-mcp", "./cmd/jules-mcp"},
		} {
			totalCount++

			config := build.DefaultConfig(binary.name, binary.path)
			config.Version = input.Version
			config.GOOS = platform.goos
			config.GOARCH = platform.goarch
			config.OutputDir = fmt.Sprintf("dist/%s-%s-%s", binary.name, platform.goos, platform.goarch)
			config.LDFlags = append(config.LDFlags, fmt.Sprintf("-X main.version=%s", input.Version))

			builder := build.NewBuilder(config)
			result := builder.BuildWithResult(ctx)

			if result.Success {
				builtPlatforms = append(builtPlatforms, fmt.Sprintf("%s-%s/%s", platform.goos, platform.goarch, binary.name))
				successCount++
			}
		}
	}

	output := BuildReleaseOutput{
		Success:          successCount == totalCount,
		Version:          input.Version,
		TotalBuilds:      totalCount,
		SuccessfulBuilds: successCount,
		Platforms:        builtPlatforms,
		Summary:          fmt.Sprintf("Release %s: %d/%d builds successful", input.Version, successCount, totalCount),
	}

	if successCount < totalCount {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Release build incomplete: %s", output.Summary)},
			},
		}, output, nil
	}

	return nil, output, nil
}
