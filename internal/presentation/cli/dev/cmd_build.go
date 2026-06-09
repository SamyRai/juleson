package dev

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/SamyRai/juleson/internal/logger"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
	"github.com/SamyRai/juleson/pkg/builder"
	"github.com/spf13/cobra"
	"log/slog"
)

func (h *CommandHandler) buildBinaries(ctx context.Context, target, version, goos, goarch string, race bool) (*builder.BuildSummary, error) {
	summary, err := h.svc.BuildWithResults(ctx, builder.BuildOptions{
		Target:  target,
		Version: version,
		GOOS:    goos,
		GOARCH:  goarch,
		Race:    race,
	})
	for _, result := range summary.Results {
		fmt.Printf("🔨 Building %s...\n", result.Name)
		if result.Success {
			fmt.Printf("✅ %s\n", result.String())
		} else {
			fmt.Printf("❌ %s\n", result.String())
		}
	}
	return summary, err
}

func (h *CommandHandler) BuildCmd() *cobra.Command {
	var (
		all     bool
		cli     bool
		alias   bool
		race    bool
		version string
		goos    string
		goarch  string
	)

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build binaries",
		Long:  "Build Juleson binaries with various options",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			target := "all"
			if cli {
				target = "cli"
			}
			if alias {
				target = "alias"
			}
			if all {
				target = "all"
			}

			summary, err := h.buildBinaries(ctx, target, version, goos, goarch, race)
			if err != nil {
				return err
			}

			fmt.Println("\n📊 Build Summary:")
			fmt.Printf("  Successful: %d/%d\n", summary.SuccessCount, len(summary.Results))
			fmt.Printf("  Total Time: %v\n", summary.TotalDuration)
			fmt.Printf("  Total Size: %.2f MB\n", float64(summary.TotalSize)/(1024*1024))

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Build all binaries")
	cmd.Flags().BoolVar(&cli, "cli", false, "Build juleson only")
	cmd.Flags().BoolVar(&alias, "alias", false, "Build jsn alias only")
	cmd.Flags().BoolVar(&race, "race", false, "Enable race detection")
	cmd.Flags().StringVar(&version, "version", "dev", "Version to embed in binaries")
	cmd.Flags().StringVar(&goos, "goos", runtime.GOOS, "Target operating system")
	cmd.Flags().StringVar(&goarch, "goarch", runtime.GOARCH, "Target architecture")

	return cmd
}

func (h *CommandHandler) InstallCmd() *cobra.Command {
	var (
		installPath string
		skipChecks  bool
		skipLint    bool
		skipTests   bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install binaries to $GOPATH/bin",
		Long:  "Build and install binaries to $GOPATH/bin or custom directory. Runs quality checks by default.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if !skipChecks {
				slog.Info("Running quality checks...")
				slog.Info("Formatting code...")
				if !skipLint {
					slog.Info("Running linters...")
				} else {
					slog.Debug("Skipping linters (--skip-lint flag used)")
				}

				testConfig := builder.DefaultTestConfig()
				testConfig.Cover = true
				testConfig.CoverProfile = "coverage.out"
				if !skipTests {
					slog.Info("Running tests...")
				} else {
					slog.Debug("Skipping tests (--skip-tests flag used)")
				}

				summary, err := h.svc.RunQualityChecks(ctx, builder.QualityOptions{
					Format:     true,
					Lint:       !skipLint,
					Test:       !skipTests,
					TestConfig: testConfig,
				})
				if err != nil {
					return err
				}
				fmt.Printf("  ✅ Completed checks: %s\n", strings.Join(summary.Checks, ", "))
				if summary.TestResult != nil {
					fmt.Printf("  ✅ %s\n", summary.TestResult.String())
				}
				logger.Success(slog.Default(), "All quality checks passed!")
			} else {
				slog.Debug("Skipping quality checks (--skip-checks flag used)")
			}

			slog.Info("Building binaries...")
			_, err := h.buildBinaries(ctx, "all", "dev", runtime.GOOS, runtime.GOARCH, false)
			if err != nil {
				return err
			}

			var result *builder.InstallResult
			if installPath != "" {
				fmt.Printf("📦 Installing to %s...", installPath)
				result, err = h.svc.InstallWithResult(ctx, builder.InstallOptions{Path: installPath, SkipBuild: true})
			} else {
				fmt.Printf("📦 Installing...")
				result, err = h.svc.InstallWithResult(ctx, builder.InstallOptions{SkipBuild: true})
			}

			if err != nil {
				return fmt.Errorf("installation failed: %w", err)
			}

			fmt.Println("\n✅ Installation successful!")
			fmt.Printf("   Install directory: %s\n", result.InstallDir)
			slog.Debug("Installed binaries:")
			for _, binary := range result.Installed {
				fmt.Printf("   - %s\n", binary)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&installPath, "path", "", "Custom installation directory")
	cmd.Flags().BoolVar(&skipChecks, "skip-checks", false, "Skip quality checks before installation")
	cmd.Flags().BoolVar(&skipLint, "skip-lint", false, "Skip linting during quality checks")
	cmd.Flags().BoolVar(&skipTests, "skip-tests", false, "Skip tests during quality checks")

	return cmd
}

func (h *CommandHandler) ReleaseCmd() *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "release",
		Short: "Build release binaries for all platforms",
		Long:  "Build release binaries for Linux, macOS, and Windows",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if version == "" {
				return fmt.Errorf("version is required (use --version flag)")
			}

			fmt.Printf("🚀 Building release %s...\n\n", version)
			summary, err := h.svc.ReleaseWithResults(ctx, version)
			for _, result := range summary.Results {
				if result.Success {
					fmt.Printf("✅ %s\n", result.String())
				} else {
					fmt.Printf("❌ %s\n", result.String())
				}
			}

			fmt.Printf("\n📊 Release Summary:\n")
			fmt.Printf("  Success: %d\n", summary.SuccessCount)
			fmt.Printf("  Failed: %d\n", len(summary.Results)-summary.SuccessCount)

			if err != nil {
				return err
			}

			fmt.Println("\n🎉 Release build complete!")
			return nil
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Release version (required)")
	core.MustMarkFlagRequired(cmd, "version")

	return cmd
}
