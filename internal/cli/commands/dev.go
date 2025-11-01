package commands

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/SamyRai/juleson/internal/build"
	"github.com/spf13/cobra"
)

// NewDevCommand creates the dev command for developer tools
func NewDevCommand() *cobra.Command {
	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Developer tools and build commands",
		Long:  "Comprehensive developer tools for building, testing, and maintaining Juleson",
	}

	devCmd.AddCommand(newBuildCommand())
	devCmd.AddCommand(newTestCommand())
	devCmd.AddCommand(newLintCommand())
	devCmd.AddCommand(newFormatCommand())
	devCmd.AddCommand(newCleanCommand())
	devCmd.AddCommand(newModCommand())
	devCmd.AddCommand(newCheckCommand())
	devCmd.AddCommand(newInstallCommand())
	devCmd.AddCommand(newReleaseCommand())

	return devCmd
}

// newBuildCommand creates the build command
func newBuildCommand() *cobra.Command {
	var (
		all     bool
		cli     bool
		mcp     bool
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

			// Determine what to build
			buildCLI := cli || all || (!cli && !mcp)
			buildMCP := mcp || all || (!cli && !mcp)

			var results []*build.BuildResult

			// Build CLI
			if buildCLI {
				fmt.Println("🔨 Building Juleson CLI...")
				config := build.DefaultConfig("juleson", "./cmd/juleson")
				config.Version = version
				config.GOOS = goos
				config.GOARCH = goarch
				config.Race = race

				if version != "" && version != "dev" {
					config.LDFlags = append(config.LDFlags, fmt.Sprintf("-X main.version=%s", version))
				}

				builder := build.NewBuilder(config)
				result := builder.BuildWithResult(ctx)
				results = append(results, result)

				if result.Success {
					fmt.Printf("✅ %s\n", result.String())
				} else {
					fmt.Printf("❌ %s\n", result.String())
				}
			}

			// Build MCP
			if buildMCP {
				fmt.Println("🔨 Building Juleson MCP Server...")
				config := build.DefaultConfig("juleson-mcp", "./cmd/jules-mcp")
				config.Version = version
				config.GOOS = goos
				config.GOARCH = goarch
				config.Race = race

				if version != "" && version != "dev" {
					config.LDFlags = append(config.LDFlags, fmt.Sprintf("-X main.version=%s", version))
				}

				builder := build.NewBuilder(config)
				result := builder.BuildWithResult(ctx)
				results = append(results, result)

				if result.Success {
					fmt.Printf("✅ %s\n", result.String())
				} else {
					fmt.Printf("❌ %s\n", result.String())
				}
			}

			// Print summary
			fmt.Println("\n📊 Build Summary:")
			successCount := 0
			totalDuration := time.Duration(0)
			totalSize := int64(0)

			for _, result := range results {
				if result.Success {
					successCount++
					totalDuration += result.Duration
					totalSize += result.OutputSize
				}
			}

			fmt.Printf("  Successful: %d/%d\n", successCount, len(results))
			fmt.Printf("  Total Time: %v\n", totalDuration)
			fmt.Printf("  Total Size: %.2f MB\n", float64(totalSize)/(1024*1024))

			if successCount < len(results) {
				return fmt.Errorf("some builds failed")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Build all binaries")
	cmd.Flags().BoolVar(&cli, "cli", false, "Build CLI only")
	cmd.Flags().BoolVar(&mcp, "mcp", false, "Build MCP server only")
	cmd.Flags().BoolVar(&race, "race", false, "Enable race detection")
	cmd.Flags().StringVar(&version, "version", "dev", "Version to embed in binaries")
	cmd.Flags().StringVar(&goos, "goos", runtime.GOOS, "Target operating system")
	cmd.Flags().StringVar(&goarch, "goarch", runtime.GOARCH, "Target architecture")

	return cmd
}

// newTestCommand creates the test command
func newTestCommand() *cobra.Command {
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

			fmt.Println("🧪 Running tests...")

			config := build.DefaultTestConfig()
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

			tester := build.NewTester(config)
			result := tester.TestWithResult(ctx)

			if result.Success {
				fmt.Printf("\n✅ %s\n", result.String())
			} else {
				fmt.Printf("\n❌ %s\n", result.String())
				return result.Error
			}

			// Generate HTML coverage report if requested
			if cover && coverProfile != "" {
				htmlPath := "coverage.html"
				fmt.Printf("\n📊 Generating HTML coverage report...\n")
				if err := tester.GenerateCoverageHTML(ctx, htmlPath); err != nil {
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

// newLintCommand creates the lint command
func newLintCommand() *cobra.Command {
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

			fmt.Println("🔍 Running linters...")

			config := build.DefaultLintConfig()
			config.FixMode = fix
			config.Verbose = verbose
			config.Fast = fast
			config.Timeout = timeout

			if len(args) > 0 {
				config.Packages = args
			}

			linter := build.NewLinter(config)
			result := linter.LintWithResult(ctx)

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

// newFormatCommand creates the format command
func newFormatCommand() *cobra.Command {
	var useGofumpt bool

	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "Format code",
		Long:  "Format Go code using go fmt or gofumpt",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			fmt.Println("✨ Formatting code...")

			formatter := build.NewFormatter()

			var err error
			if useGofumpt {
				err = formatter.FormatWithGofumpt(ctx, args...)
			} else {
				err = formatter.Format(ctx, args...)
			}

			if err != nil {
				fmt.Printf("❌ Format failed: %v\n", err)
				return err
			}

			fmt.Println("✅ Code formatted successfully")
			return nil
		},
	}

	cmd.Flags().BoolVar(&useGofumpt, "gofumpt", false, "Use gofumpt instead of go fmt")

	return cmd
}

// newCleanCommand creates the clean command
func newCleanCommand() *cobra.Command {
	var (
		all       bool
		cache     bool
		modCache  bool
		testCache bool
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean build artifacts",
		Long:  "Clean build artifacts, caches, and generated files",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			fmt.Println("🧹 Cleaning...")

			cleaner := build.NewCleaner("bin", []string{"coverage.out", "coverage.html"})

			if all {
				if err := cleaner.CleanAll(ctx); err != nil {
					fmt.Printf("❌ Clean failed: %v\n", err)
					return err
				}
			} else {
				if err := cleaner.Clean(ctx); err != nil {
					fmt.Printf("❌ Clean failed: %v\n", err)
					return err
				}

				if cache {
					if err := cleaner.CleanCache(ctx); err != nil {
						fmt.Printf("⚠️  Cache clean failed: %v\n", err)
					}
				}

				if modCache {
					if err := cleaner.CleanModCache(ctx); err != nil {
						fmt.Printf("⚠️  Module cache clean failed: %v\n", err)
					}
				}

				if testCache {
					if err := cleaner.CleanTestCache(ctx); err != nil {
						fmt.Printf("⚠️  Test cache clean failed: %v\n", err)
					}
				}
			}

			fmt.Println("✅ Cleaned successfully")
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Clean everything including caches")
	cmd.Flags().BoolVar(&cache, "cache", false, "Clean build cache")
	cmd.Flags().BoolVar(&modCache, "modcache", false, "Clean module cache")
	cmd.Flags().BoolVar(&testCache, "testcache", false, "Clean test cache")

	return cmd
}

// newModCommand creates the mod command
func newModCommand() *cobra.Command {
	modCmd := &cobra.Command{
		Use:   "mod",
		Short: "Module maintenance",
		Long:  "Go module maintenance commands",
	}

	modCmd.AddCommand(&cobra.Command{
		Use:   "tidy",
		Short: "Tidy dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("📦 Tidying dependencies...")
			manager := build.NewModuleManager()
			if err := manager.Tidy(context.Background()); err != nil {
				return err
			}
			fmt.Println("✅ Dependencies tidied")
			return nil
		},
	})

	modCmd.AddCommand(&cobra.Command{
		Use:   "download",
		Short: "Download dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("📥 Downloading dependencies...")
			manager := build.NewModuleManager()
			if err := manager.Download(context.Background()); err != nil {
				return err
			}
			fmt.Println("✅ Dependencies downloaded")
			return nil
		},
	})

	modCmd.AddCommand(&cobra.Command{
		Use:   "verify",
		Short: "Verify dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🔍 Verifying dependencies...")
			manager := build.NewModuleManager()
			if err := manager.Verify(context.Background()); err != nil {
				return err
			}
			fmt.Println("✅ Dependencies verified")
			return nil
		},
	})

	modCmd.AddCommand(&cobra.Command{
		Use:   "graph",
		Short: "Print dependency graph",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := build.NewModuleManager()
			return manager.Graph(context.Background())
		},
	})

	return modCmd
}

// newCheckCommand creates the check command
func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Run all quality checks",
		Long:  "Run formatting, linting, and tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Format
			fmt.Println("✨ Formatting code...")
			formatter := build.NewFormatter()
			if err := formatter.Format(ctx); err != nil {
				return fmt.Errorf("format failed: %w", err)
			}
			fmt.Println("✅ Code formatted")

			// Lint
			fmt.Println("\n🔍 Running linters...")
			linter := build.NewLinter(build.DefaultLintConfig())
			if err := linter.Lint(ctx); err != nil {
				return fmt.Errorf("lint failed: %w", err)
			}
			fmt.Println("✅ Linting passed")

			// Test
			fmt.Println("\n🧪 Running tests...")
			config := build.DefaultTestConfig()
			config.Cover = true
			config.CoverProfile = "coverage.out"

			tester := build.NewTester(config)
			result := tester.TestWithResult(ctx)

			if !result.Success {
				return fmt.Errorf("tests failed: %w", result.Error)
			}
			fmt.Printf("✅ %s\n", result.String())

			fmt.Println("\n🎉 All checks passed!")
			return nil
		},
	}
}

// newInstallCommand creates the install command
func newInstallCommand() *cobra.Command {
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

			// Run quality checks unless skipped
			if !skipChecks {
				fmt.Println("🔍 Running quality checks...")

				// Format
				fmt.Println("  ✨ Formatting code...")
				formatter := build.NewFormatter()
				if err := formatter.Format(ctx); err != nil {
					return fmt.Errorf("format failed: %w", err)
				}
				fmt.Println("  ✅ Code formatted")

				// Lint (unless skipped)
				if !skipLint {
					fmt.Println("  🔍 Running linters...")
					linter := build.NewLinter(build.DefaultLintConfig())
					if err := linter.Lint(ctx); err != nil {
						return fmt.Errorf("lint failed: %w", err)
					}
					fmt.Println("  ✅ Linting passed")
				} else {
					fmt.Println("  ⏭️  Skipping linters (--skip-lint flag used)")
				}

				// Test (unless skipped)
				if !skipTests {
					fmt.Println("  🧪 Running tests...")
					config := build.DefaultTestConfig()
					config.Cover = true
					config.CoverProfile = "coverage.out"

					tester := build.NewTester(config)
					result := tester.TestWithResult(ctx)

					if !result.Success {
						return fmt.Errorf("tests failed: %w", result.Error)
					}
					fmt.Printf("  ✅ %s", result.String())
				} else {
					fmt.Println("  ⏭️  Skipping tests (--skip-tests flag used)")
				}

				fmt.Println("🎉 All quality checks passed!")
			} else {
				fmt.Println("⏭️  Skipping quality checks (--skip-checks flag used)")
			}

			// Build both binaries
			fmt.Println("🔨 Building binaries...")

			// Build CLI
			cliConfig := build.DefaultConfig("juleson", "./cmd/juleson")
			cliBuilder := build.NewBuilder(cliConfig)
			if err := cliBuilder.Build(ctx); err != nil {
				return fmt.Errorf("failed to build CLI: %w", err)
			}
			fmt.Println("✅ CLI built")

			// Build MCP
			mcpConfig := build.DefaultConfig("juleson-mcp", "./cmd/jules-mcp")
			mcpBuilder := build.NewBuilder(mcpConfig)
			if err := mcpBuilder.Build(ctx); err != nil {
				return fmt.Errorf("failed to build MCP: %w", err)
			}
			fmt.Println("✅ MCP server built")

			// Install binaries
			installer := build.NewInstaller("bin", []string{"juleson", "juleson-mcp"})

			var result *build.InstallResult
			var err error

			if installPath != "" {
				fmt.Printf("📦 Installing to %s...", installPath)
				result, err = installer.InstallTo(ctx, installPath)
			} else {
				defaultPath, pathErr := installer.GetInstallPath()
				if pathErr != nil {
					return fmt.Errorf("failed to determine install path: %w", pathErr)
				}
				fmt.Printf("📦 Installing to %s...", defaultPath)
				result, err = installer.Install(ctx)
			}

			if err != nil {
				return fmt.Errorf("installation failed: %w", err)
			}

			// Display results
			fmt.Println("\n✅ Installation successful!")
			fmt.Printf("   Install directory: %s\n", result.InstallDir)
			fmt.Println("   Installed binaries:")
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

// newReleaseCommand creates the release command
func newReleaseCommand() *cobra.Command {
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

			fmt.Printf("🚀 Building release %s for %d platforms...\n\n", version, len(platforms)*2)

			successCount := 0
			failCount := 0

			for _, platform := range platforms {
				for _, binary := range []struct {
					name string
					path string
				}{
					{"juleson", "./cmd/juleson"},
					{"juleson-mcp", "./cmd/jules-mcp"},
				} {
					fmt.Printf("Building %s for %s/%s...\n", binary.name, platform.goos, platform.goarch)

					config := build.DefaultConfig(binary.name, binary.path)
					config.Version = version
					config.GOOS = platform.goos
					config.GOARCH = platform.goarch
					config.OutputDir = fmt.Sprintf("dist/%s-%s-%s", binary.name, platform.goos, platform.goarch)
					config.LDFlags = append(config.LDFlags, fmt.Sprintf("-X main.version=%s", version))

					builder := build.NewBuilder(config)
					result := builder.BuildWithResult(ctx)

					if result.Success {
						fmt.Printf("  ✅ %s\n", result.String())
						successCount++
					} else {
						fmt.Printf("  ❌ Failed: %v\n", result.Error)
						failCount++
					}
				}
				fmt.Println()
			}

			fmt.Printf("\n📊 Release Summary:\n")
			fmt.Printf("  Success: %d\n", successCount)
			fmt.Printf("  Failed: %d\n", failCount)

			if failCount > 0 {
				return fmt.Errorf("some builds failed")
			}

			fmt.Println("\n🎉 Release build complete!")
			return nil
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Release version (required)")
	cmd.MarkFlagRequired("version")

	return cmd
}
