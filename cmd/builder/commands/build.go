package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build all binaries",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building all binaries...")
		ctx := context.Background()

		if err := svc.BuildAll(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("All binaries built successfully")
	},
}

var buildCLICmd = &cobra.Command{
	Use:   "build-cli",
	Short: "Build CLI binary",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building CLI binary...")
		ctx := context.Background()

		if err := svc.BuildCLI(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("CLI binary built successfully")
	},
}
