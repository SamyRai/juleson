package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/SamyRai/juleson/internal/orchestrator"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running tests...")
		ctx := context.Background()

		if err := svc.Test(ctx, orchestrator.TestOptions{
			Verbose: true,
			Race:    true,
		}); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Tests completed")
	},
}

var testShortCmd = &cobra.Command{
	Use:   "test-short",
	Short: "Run short tests",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running short tests...")
		ctx := context.Background()

		if err := svc.Test(ctx, orchestrator.TestOptions{
			Verbose: true,
			Short:   true,
		}); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Short tests completed")
	},
}

var coverageCmd = &cobra.Command{
	Use:   "coverage",
	Short: "Generate test coverage report",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generating coverage report...")
		ctx := context.Background()

		if err := svc.Coverage(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Coverage report generated successfully")
	},
}
