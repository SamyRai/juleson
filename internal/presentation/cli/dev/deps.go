package dev

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/SamyRai/juleson/internal/intelligence"
	"github.com/spf13/cobra"
)

// newDepsCommand creates the deps command.
func newDepsCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "deps [path]",
		Short: "Analyze and visualize dependencies",
		Long:  "Extract internal package dependencies and render them as a Mermaid graph",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			slog.Info("Analyzing dependencies...")

			graph, err := intelligence.AnalyzeDependencies(context.Background(), path)
			if err != nil {
				return fmt.Errorf("dependency analysis failed: %w", err)
			}

			if format == "mermaid" {
				fmt.Println(intelligence.RenderMermaid(graph))
			} else {
				fmt.Println("Error: Only 'mermaid' format is supported right now.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "mermaid", "Output format (mermaid)")

	return cmd
}
