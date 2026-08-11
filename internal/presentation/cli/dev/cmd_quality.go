package dev

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/logger"
	"github.com/SamyRai/juleson/pkg/builder"
	"github.com/spf13/cobra"
)

func (h *CommandHandler) TestCmd() *cobra.Command {
	var (
		verbose      bool
		race         bool
		cover        bool
		coverProfile string
		short        bool
		timeout      string
		parallel     int
		run          string
		skip         string
		failFast     bool
		shuffle      string
	)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests",
		Long:  "Run tests with various options and generate coverage reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			slog.Info("Running tests...")

			config := builder.DefaultTestConfig()
			config.Verbose = verbose
			config.Race = race
			config.Cover = cover
			config.CoverProfile = coverProfile
			config.Short = short
			config.RunPattern = run
			config.SkipPattern = skip
			config.FailFast = failFast
			config.Shuffle = shuffle

			if parallel > 0 {
				config.Parallel = parallel
			}

			if timeout != "" {
				duration, err := time.ParseDuration(timeout)
				if err != nil {
					return fmt.Errorf("invalid timeout: %w", err)
				}
				config.Timeout = duration
			}

			if len(args) > 0 {
				config.Packages = args
			}

			result := h.svc.RunTestsWithResult(ctx, config)
			if result.Success {
				fmt.Printf("\n✅ %s\n", result.String())
			} else {
				fmt.Printf("\n❌ %s\n", result.String())
				return result.Error
			}

			if cover && coverProfile != "" {
				htmlPath := "coverage.html"
				fmt.Printf("\n📊 Generating HTML coverage report...\n")
				if err := h.svc.GenerateCoverageHTML(ctx, config, htmlPath); err != nil {
					fmt.Printf("⚠️  Failed to generate HTML report: %v\n", err)
				} else {
					fmt.Printf("✅ Coverage report: %s\n", htmlPath)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().BoolVar(&race, "race", true, "Enable race detection")
	cmd.Flags().BoolVar(&cover, "cover", false, "Enable coverage")
	cmd.Flags().StringVar(&coverProfile, "coverprofile", "", "Coverage profile output file")
	cmd.Flags().BoolVar(&short, "short", false, "Run short tests only")
	cmd.Flags().StringVar(&timeout, "timeout", "10m", "Test timeout")
	cmd.Flags().IntVarP(&parallel, "parallel", "p", 0, "Number of parallel tests")
	cmd.Flags().StringVar(&run, "run", "", "Run only tests matching pattern")
	cmd.Flags().StringVar(&skip, "skip", "", "Skip tests matching pattern")
	cmd.Flags().BoolVar(&failFast, "failfast", false, "Stop on first test failure")
	cmd.Flags().StringVar(&shuffle, "shuffle", "", "Randomize test execution order")

	return cmd
}

func (h *CommandHandler) LintCmd() *cobra.Command {
	var (
		fix     bool
		verbose bool
		fast    bool
		timeout string
	)

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Run linters",
		Long:  "Run go vet and golangci-lint to check code quality",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			slog.Info("Running linters...")

			config := builder.DefaultLintConfig()
			config.FixMode = fix
			config.Verbose = verbose
			config.Fast = fast
			config.Timeout = timeout

			if len(args) > 0 {
				config.Packages = args
			}

			result := h.svc.LintWithResult(ctx, config)
			if result.Success {
				fmt.Printf("\n✅ %s\n", result.String())
			} else {
				fmt.Printf("\n❌ %s\n", result.String())
				return result.Error
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "Automatically fix issues")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().BoolVar(&fast, "fast", false, "Fast mode (fewer linters)")
	cmd.Flags().StringVar(&timeout, "timeout", "5m", "Lint timeout")

	return cmd
}

func (h *CommandHandler) FormatCmd() *cobra.Command {
	var useGofumpt bool

	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "Format code",
		Long:  "Format Go code using go fmt or gofumpt",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			slog.Info("Formatting code...")

			if err := h.svc.FormatCode(ctx, useGofumpt, args...); err != nil {
				fmt.Printf("❌ Format failed: %v\n", err)
				return err
			}

			logger.Success(slog.Default(), "Code formatted successfully")
			return nil
		},
	}

	cmd.Flags().BoolVar(&useGofumpt, "gofumpt", false, "Use gofumpt instead of go fmt")

	return cmd
}

func (h *CommandHandler) CheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Run all quality checks",
		Long:  "Run formatting, linting, and tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			slog.Info("Formatting code...")
			fmt.Println("\n🔍 Running linters...")
			fmt.Println("\n🧪 Running tests...")
			config := builder.DefaultTestConfig()
			config.Cover = true
			config.CoverProfile = "coverage.out"
			fmt.Println("\n🔨 Building binaries...")

			summary, err := h.svc.RunQualityChecks(ctx, builder.QualityOptions{
				Format:     true,
				Lint:       true,
				Test:       true,
				TestConfig: config,
				Build:      true,
			})
			if err != nil {
				return err
			}
			fmt.Printf("✅ Completed checks: %s\n", strings.Join(summary.Checks, ", "))
			if summary.TestResult != nil {
				fmt.Printf("✅ %s\n", summary.TestResult.String())
			}

			fmt.Println("\n🎉 All checks passed!")
			return nil
		},
	}
}
