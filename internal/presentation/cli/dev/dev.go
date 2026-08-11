package dev

import (
	"github.com/SamyRai/juleson/pkg/builder"
	"github.com/spf13/cobra"
)

// NewDevCommand creates the dev command for developer tools.
func NewDevCommand() *cobra.Command {
	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Developer tools and build commands",
		Long:  "Comprehensive developer tools for building, testing, and maintaining Juleson",
	}

	// We pass a default dev builder to the commands here
	svc := builder.NewService(builder.DefaultConfig("dev", "", ""))
	handler := NewCommandHandler(svc)

	devCmd.AddCommand(handler.BuildCmd())
	devCmd.AddCommand(handler.TestCmd())
	devCmd.AddCommand(handler.LintCmd())
	devCmd.AddCommand(handler.FormatCmd())
	devCmd.AddCommand(handler.CleanCmd())
	devCmd.AddCommand(handler.ModCmd())
	devCmd.AddCommand(handler.CheckCmd())
	devCmd.AddCommand(handler.InstallCmd())
	devCmd.AddCommand(handler.ReleaseCmd())

	// Add existing commands from complexity.go and deps.go which are un-refactored
	devCmd.AddCommand(newCheckComplexityCommand())
	devCmd.AddCommand(newDepsCommand())

	return devCmd
}
