package commands

import (
	"github.com/SamyRai/juleson/internal/orchestrator"
	"github.com/spf13/cobra"
)

var (
	svc *orchestrator.Service
)

// AddCommands adds all commands to the root command
func AddCommands(rootCmd *cobra.Command, ver, build, commit string) {
	// Initialize the orchestrator service
	config := orchestrator.DefaultConfig(ver, build, commit)
	svc = orchestrator.NewService(config)

	// All command
	rootCmd.AddCommand(allCmd)

	// Build commands
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(buildCLICmd)
	rootCmd.AddCommand(buildMCPCmd)

	// Clean command
	rootCmd.AddCommand(cleanCmd)

	// Test commands
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(testShortCmd)
	rootCmd.AddCommand(coverageCmd)

	// Code quality commands
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(fmtCmd)

	// Dependency commands
	rootCmd.AddCommand(depsCmd)
	rootCmd.AddCommand(tidyCmd)

	// Install command
	rootCmd.AddCommand(installCmd)

	// Run commands
	rootCmd.AddCommand(runCLICmd)
	rootCmd.AddCommand(runMCPCmd)

	// Development command
	rootCmd.AddCommand(devCmd)

	// Check command
	rootCmd.AddCommand(checkCmd)

	// Docker commands
	rootCmd.AddCommand(dockerBuildCmd)
	rootCmd.AddCommand(dockerRunCmd)
	rootCmd.AddCommand(dockerRunCLICmd)
	rootCmd.AddCommand(dockerRunMCPCmd)
	rootCmd.AddCommand(dockerPushCmd)
	rootCmd.AddCommand(dockerComposeUpCmd)
	rootCmd.AddCommand(dockerComposeDownCmd)
	rootCmd.AddCommand(dockerCleanCmd)

	// Version command
	rootCmd.AddCommand(versionCmd)
}
