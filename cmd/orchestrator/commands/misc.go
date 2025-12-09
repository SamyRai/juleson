package commands

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/SamyRai/juleson/internal/orchestrator"
	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all tasks (clean, lint, test, build)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running all tasks...")
		ctx := context.Background()

		// Clean
		if err := svc.Clean(ctx); err != nil {
			log.Fatal("Clean failed:", err)
		}

		// Lint
		if err := svc.Lint(ctx); err != nil {
			log.Fatal("Lint failed:", err)
		}

		// Test
		if err := svc.Test(ctx, orchestrator.TestOptions{
			Verbose: true,
			Race:    true,
		}); err != nil {
			log.Fatal("Test failed:", err)
		}

		// Build
		if err := svc.BuildAll(ctx); err != nil {
			log.Fatal("Build failed:", err)
		}

		fmt.Println("All tasks completed successfully")
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean build artifacts",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Cleaning build artifacts...")
		ctx := context.Background()

		if err := svc.Clean(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Clean completed")
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install binaries to GOPATH/bin",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Installing binaries...")
		ctx := context.Background()

		// Install to default GOPATH/bin (empty string means use GOPATH)
		if err := svc.Install(ctx, ""); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Binaries installed successfully")
	},
}

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Development mode with live reload",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting development mode...")
		ctx := context.Background()

		if err := svc.StartDev(ctx); err != nil {
			log.Fatal(err)
		}
	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run all checks (lint, test, build)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running all checks...")
		ctx := context.Background()

		if err := svc.RunAllChecks(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("All checks passed")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		info := svc.GetVersion()
		fmt.Printf("Orchestrator CLI v%s\n", info.Version)
		fmt.Printf("Build Date: %s\n", info.BuildDate)
		fmt.Printf("Git Commit: %s\n", info.GitCommit)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
