package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Run linters",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running linters...")
		ctx := context.Background()

		if err := svc.Lint(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Linting completed")
	},
}

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Formatting code...")
		ctx := context.Background()

		if err := svc.Format(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Code formatting completed")
	},
}
