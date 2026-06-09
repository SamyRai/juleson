package dev

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/SamyRai/juleson/internal/intelligence"
	"github.com/spf13/cobra"
)

// newCheckComplexityCommand creates the check-complexity command.
func newCheckComplexityCommand() *cobra.Command {
	var threshold int

	cmd := &cobra.Command{
		Use:   "check-complexity [path]",
		Short: "Analyze AST complexity",
		Long:  "Calculate cyclomatic complexity of Go functions in the given path",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			slog.Info("Analyzing code complexity...")

			results, err := intelligence.AnalyzeComplexity(context.Background(), path)
			if err != nil {
				return fmt.Errorf("complexity analysis failed: %w", err)
			}

			fmt.Println("\n📊 Complexity Report (Functions exceeding threshold):")
			fmt.Println("--------------------------------------------------")

			count := 0
			for _, res := range results {
				if res.Complexity >= threshold {
					count++
					fmt.Printf("%-40s %-20s Complexity: %d\n", res.FuncName, fmt.Sprintf("(%s:%d)", res.FileName, res.Line), res.Complexity)
				}
			}

			if count == 0 {
				fmt.Printf("✅ No functions exceed the complexity threshold of %d\n", threshold)
			} else {
				fmt.Printf("\n⚠️  Found %d function(s) exceeding complexity threshold of %d\n", count, threshold)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&threshold, "threshold", 10, "Minimum complexity score to report")

	return cmd
}
