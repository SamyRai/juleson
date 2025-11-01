package commands

import (
	"fmt"

	"jules-automation/internal/automation"

	"github.com/spf13/cobra"
)

// NewAnalyzeCommand creates the analyze command
func NewAnalyzeCommand(initializeEngine func() (*automation.Engine, error), displayProjectAnalysis func(*automation.ProjectContext)) *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [project-path]",
		Short: "Analyze project structure and context",
		Long:  "Analyze the project structure, dependencies, and create context for automation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := args[0]

			// Initialize automation engine
			engine, err := initializeEngine()
			if err != nil {
				return fmt.Errorf("failed to initialize automation engine: %w", err)
			}

			// Analyze project
			context, err := engine.AnalyzeProject(projectPath)
			if err != nil {
				return fmt.Errorf("failed to analyze project: %w", err)
			}

			// Display analysis results
			displayProjectAnalysis(context)

			return nil
		},
	}
}
