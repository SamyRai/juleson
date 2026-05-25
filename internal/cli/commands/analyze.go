package commands

import (
	"fmt"

	"github.com/SamyRai/juleson/internal/analyzer"

	"github.com/spf13/cobra"
)

// NewAnalyzeCommand creates the analyze command
func NewAnalyzeCommand(analyzeProject func(string) (*analyzer.ProjectContext, error), displayProjectAnalysis func(*analyzer.ProjectContext)) *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [project-path]",
		Short: "Analyze project structure and context",
		Long:  "Analyze the project structure, dependencies, and create context for automation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := args[0]

			context, err := analyzeProject(projectPath)
			if err != nil {
				return fmt.Errorf("failed to analyze project: %w", err)
			}

			displayProjectAnalysis(context)
			return nil
		},
	}
}
