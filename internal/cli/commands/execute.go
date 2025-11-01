package commands

import (
	"fmt"
	"strings"

	"jules-automation/internal/automation"

	"github.com/spf13/cobra"
)

// NewExecuteCommand creates the execute command
func NewExecuteCommand(initializeEngine func() (*automation.Engine, error), displayExecutionResult func(*automation.ExecutionResult)) *cobra.Command {
	executeCmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute automation tasks",
		Long:  "Execute templates and automation tasks on projects",
	}

	// Execute template
	executeCmd.AddCommand(&cobra.Command{
		Use:   "template [template-name] [project-path]",
		Short: "Execute a template on a project",
		Long:  "Execute a specific template on a project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			projectPath := args[1]

			// Initialize automation engine
			engine, err := initializeEngine()
			if err != nil {
				return fmt.Errorf("failed to initialize automation engine: %w", err)
			}

			// Analyze project first
			_, err = engine.AnalyzeProject(projectPath)
			if err != nil {
				return fmt.Errorf("failed to analyze project: %w", err)
			}

			// Execute template
			result, err := engine.ExecuteTemplate(cmd.Context(), templateName, make(map[string]string))
			if err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}

			// Display results
			displayExecutionResult(result)

			return nil
		},
	})

	// Execute with custom parameters
	executeCmd.AddCommand(&cobra.Command{
		Use:   "template-with-params [template-name] [project-path] [key=value]...",
		Short: "Execute template with custom parameters",
		Long:  "Execute a template with custom parameters",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			projectPath := args[1]

			// Parse custom parameters
			customParams := make(map[string]string)
			for i := 2; i < len(args); i++ {
				parts := strings.SplitN(args[i], "=", 2)
				if len(parts) == 2 {
					customParams[parts[0]] = parts[1]
				}
			}

			// Initialize automation engine
			engine, err := initializeEngine()
			if err != nil {
				return fmt.Errorf("failed to initialize automation engine: %w", err)
			}

			// Analyze project first
			_, err = engine.AnalyzeProject(projectPath)
			if err != nil {
				return fmt.Errorf("failed to analyze project: %w", err)
			}

			// Execute template with custom parameters
			result, err := engine.ExecuteTemplate(cmd.Context(), templateName, customParams)
			if err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}

			// Display results
			displayExecutionResult(result)

			return nil
		},
	})

	return executeCmd
}
