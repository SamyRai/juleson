package orchestrator

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/build"
)

type QualityOptions struct {
	Format       bool
	UseGofumpt   bool
	FormatPaths  []string
	Lint         bool
	LintConfig   build.LintConfig
	Test         bool
	TestConfig   build.TestConfig
	Build        bool
	BuildOptions BuildOptions
}

type QualitySummary struct {
	Checks       []string
	LintResult   *build.LintResult
	TestResult   *build.TestResult
	BuildSummary *BuildSummary
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

func (s *Service) RunQualityChecks(ctx context.Context, options QualityOptions) (*QualitySummary, error) {
	summary := &QualitySummary{Checks: make([]string, 0, 4)}

	if options.Format {
		if err := s.FormatCode(ctx, options.UseGofumpt, options.FormatPaths...); err != nil {
			return summary, fmt.Errorf("format failed: %w", err)
		}
		summary.Checks = append(summary.Checks, "Format")
	}

	if options.Lint {
		config := normalizeLintConfig(options.LintConfig)
		result := s.LintWithResult(ctx, config)
		summary.LintResult = result
		if !result.Success {
			return summary, fmt.Errorf("lint failed: %w", result.Error)
		}
		summary.Checks = append(summary.Checks, "Lint")
	}

	if options.Test {
		config := normalizeTestConfig(options.TestConfig)
		result := s.RunTestsWithResult(ctx, config)
		summary.TestResult = result
		if !result.Success {
			return summary, fmt.Errorf("tests failed: %w", result.Error)
		}
		summary.Checks = append(summary.Checks, "Tests")
	}

	if options.Build {
		buildOptions := options.BuildOptions
		if buildOptions.Target == "" {
			buildOptions.Target = "all"
		}
		if buildOptions.Version == "" {
			buildOptions.Version = "dev"
		}
		buildSummary, err := s.BuildWithResults(ctx, buildOptions)
		summary.BuildSummary = buildSummary
		if err != nil {
			return summary, fmt.Errorf("build failed: %w", err)
		}
		summary.Checks = append(summary.Checks, "Build")
	}

	return summary, nil
}

func normalizeLintConfig(config build.LintConfig) build.LintConfig {
	defaults := DefaultLintConfig()
	if len(config.Packages) == 0 {
		config.Packages = defaults.Packages
	}
	return config
}

func normalizeTestConfig(config build.TestConfig) build.TestConfig {
	defaults := DefaultTestConfig()
	if len(config.Packages) == 0 {
		config.Packages = defaults.Packages
	}
	if config.Timeout == 0 {
		config.Timeout = defaults.Timeout
	}
	return config
}
