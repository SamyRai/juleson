package orchestrator

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/SamyRai/juleson/internal/build"
)

type BuildOptions struct {
	Target  string
	Version string
	Race    bool
	GOOS    string
	GOARCH  string
}

type BuildSummary struct {
	Target        string
	Results       []*build.BuildResult
	SuccessCount  int
	TotalDuration time.Duration
	TotalSize     int64
}

func (s *Service) BuildWithResults(ctx context.Context, options BuildOptions) (*BuildSummary, error) {
	if options.Target == "" {
		options.Target = "all"
	}
	if options.Version == "" {
		options.Version = "dev"
	}
	if options.GOOS == "" {
		options.GOOS = runtime.GOOS
	}
	if options.GOARCH == "" {
		options.GOARCH = runtime.GOARCH
	}

	binaries := s.selectedBinaries(options.Target)
	summary := &BuildSummary{
		Target:  options.Target,
		Results: make([]*build.BuildResult, 0, len(binaries)),
	}

	for _, binary := range binaries {
		config := build.DefaultConfig(binary.name, binary.path)
		config.Version = options.Version
		config.GOOS = options.GOOS
		config.GOARCH = options.GOARCH
		config.Race = options.Race
		if options.Version != "" && options.Version != "dev" {
			config.LDFlags = append(config.LDFlags, fmt.Sprintf("-X main.version=%s", options.Version))
		}

		result := build.NewBuilder(config).BuildWithResult(ctx)
		summary.Results = append(summary.Results, result)
		if result.Success {
			summary.SuccessCount++
			summary.TotalDuration += result.Duration
			summary.TotalSize += result.OutputSize
		}
	}

	if summary.SuccessCount < len(summary.Results) {
		return summary, fmt.Errorf("some builds failed")
	}
	return summary, nil
}

func (s *Service) RunTestsWithResult(ctx context.Context, config build.TestConfig) *build.TestResult {
	return build.NewTester(config).TestWithResult(ctx)
}

func DefaultTestConfig() build.TestConfig {
	return build.DefaultTestConfig()
}

func (s *Service) GenerateCoverageHTML(ctx context.Context, config build.TestConfig, outputPath string) error {
	return build.NewTester(config).GenerateCoverageHTML(ctx, outputPath)
}

func (s *Service) LintWithResult(ctx context.Context, config build.LintConfig) *build.LintResult {
	return build.NewLinter(config).LintWithResult(ctx)
}

func DefaultLintConfig() build.LintConfig {
	return build.DefaultLintConfig()
}

func (s *Service) FormatCode(ctx context.Context, useGofumpt bool, paths ...string) error {
	formatter := build.NewFormatter()
	if useGofumpt {
		return formatter.FormatWithGofumpt(ctx, paths...)
	}
	return formatter.Format(ctx, paths...)
}

type CleanOptions struct {
	All       bool
	Cache     bool
	ModCache  bool
	TestCache bool
}

func (s *Service) CleanArtifacts(ctx context.Context, options CleanOptions) ([]string, error) {
	cleaner := build.NewCleaner(s.config.BinDir, []string{s.config.CoverageFile, s.config.CoverageHTML})
	cleaned := make([]string, 0, 4)

	if options.All {
		if err := cleaner.CleanAll(ctx); err != nil {
			return cleaned, err
		}
		return []string{"artifacts", "build cache", "module cache", "test cache"}, nil
	}

	if err := cleaner.Clean(ctx); err != nil {
		return cleaned, err
	}
	cleaned = append(cleaned, "artifacts")

	if options.Cache {
		if err := cleaner.CleanCache(ctx); err != nil {
			return cleaned, err
		}
		cleaned = append(cleaned, "build cache")
	}
	if options.ModCache {
		if err := cleaner.CleanModCache(ctx); err != nil {
			return cleaned, err
		}
		cleaned = append(cleaned, "module cache")
	}
	if options.TestCache {
		if err := cleaner.CleanTestCache(ctx); err != nil {
			return cleaned, err
		}
		cleaned = append(cleaned, "test cache")
	}

	return cleaned, nil
}

func (s *Service) RunModuleMaintenance(ctx context.Context, operation string, packages ...string) error {
	manager := build.NewModuleManager()
	switch operation {
	case "tidy":
		return manager.Tidy(ctx)
	case "download":
		return manager.Download(ctx)
	case "verify":
		return manager.Verify(ctx)
	case "vendor":
		return manager.Vendor(ctx)
	case "graph":
		return manager.Graph(ctx)
	case "why":
		return manager.Why(ctx, packages...)
	default:
		return fmt.Errorf("unknown module operation: %s", operation)
	}
}

type InstallOptions struct {
	Path       string
	SkipBuild  bool
	SkipChecks bool
	SkipLint   bool
	SkipTests  bool
}

type InstallResult = build.InstallResult

func (s *Service) InstallWithResult(ctx context.Context, options InstallOptions) (*build.InstallResult, error) {
	if !options.SkipBuild {
		if _, err := s.BuildWithResults(ctx, BuildOptions{Target: "all", Version: "dev"}); err != nil {
			return nil, err
		}
	}

	installer := build.NewInstaller(s.config.BinDir, []string{s.config.BinaryCLI, s.config.BinaryMCP})
	if options.Path != "" {
		return installer.InstallTo(ctx, options.Path)
	}
	return installer.Install(ctx)
}

func (s *Service) ReleaseWithResults(ctx context.Context, version string) (*BuildSummary, error) {
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

	summary := &BuildSummary{Target: "release"}
	for _, platform := range platforms {
		for _, binary := range s.selectedBinaries("all") {
			config := build.DefaultConfig(binary.name, binary.path)
			config.Version = version
			config.GOOS = platform.goos
			config.GOARCH = platform.goarch
			config.OutputDir = fmt.Sprintf("dist/%s-%s-%s", binary.name, platform.goos, platform.goarch)
			config.LDFlags = append(config.LDFlags, fmt.Sprintf("-X main.version=%s", version))

			result := build.NewBuilder(config).BuildWithResult(ctx)
			summary.Results = append(summary.Results, result)
			if result.Success {
				summary.SuccessCount++
				summary.TotalDuration += result.Duration
				summary.TotalSize += result.OutputSize
			}
		}
	}

	if summary.SuccessCount < len(summary.Results) {
		return summary, fmt.Errorf("some builds failed")
	}
	return summary, nil
}

type binaryTarget struct {
	name string
	path string
}

func (s *Service) selectedBinaries(target string) []binaryTarget {
	switch target {
	case "cli":
		return []binaryTarget{{name: s.config.BinaryCLI, path: "./" + s.config.CmdCLIDir}}
	case "mcp":
		return []binaryTarget{{name: s.config.BinaryMCP, path: "./" + s.config.CmdMCPDir}}
	default:
		return []binaryTarget{
			{name: s.config.BinaryCLI, path: "./" + s.config.CmdCLIDir},
			{name: s.config.BinaryMCP, path: "./" + s.config.CmdMCPDir},
		}
	}
}
