package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var runCLICmd = &cobra.Command{
	Use:   "run-cli [args...]",
	Short: "Run CLI binary",
	Long:  `Run the CLI binary with optional arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running CLI...")
		ctx := context.Background()

		if err := svc.RunCLI(ctx, args); err != nil {
			log.Fatal(err)
		}
	},
}

var runMCPCmd = &cobra.Command{
	Use:   "run-mcp",
	Short: "Run MCP server binary",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running MCP server...")
		ctx := context.Background()

		if err := svc.RunMCP(ctx); err != nil {
			log.Fatal(err)
		}
	},
}
