package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command
func NewInitCommand(generateConfig func(string) string) *cobra.Command {
	return &cobra.Command{
		Use:   "init [project-path]",
		Short: "Initialize a new project for Jules automation",
		Long:  "Initialize a new project directory with Jules automation configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := args[0]

			// Create project directory
			if err := os.MkdirAll(projectPath, 0755); err != nil {
				return fmt.Errorf("failed to create project directory: %w", err)
			}

			// Create Jules automation config
			configPath := filepath.Join(projectPath, "juleson.yaml")
			configContent := generateConfig(projectPath)

			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				return fmt.Errorf("failed to create project config: %w", err)
			}

			fmt.Printf("✅ Initialized Jules automation project at: %s\n", projectPath)
			fmt.Printf("📝 Configuration file created: %s\n", configPath)

			return nil
		},
	}
}
