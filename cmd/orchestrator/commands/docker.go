package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var dockerBuildCmd = &cobra.Command{
	Use:   "docker-build",
	Short: "Build Docker image",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building Docker image...")
		ctx := context.Background()

		if err := svc.DockerBuild(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Docker image built successfully")
	},
}

var dockerRunCmd = &cobra.Command{
	Use:   "docker-run [args...]",
	Short: "Run Docker container",
	Long:  `Run Docker container with optional arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running Docker container...")
		ctx := context.Background()

		if err := svc.DockerRun(ctx, args); err != nil {
			log.Fatal(err)
		}
	},
}

var dockerRunCLICmd = &cobra.Command{
	Use:   "docker-run-cli [args...]",
	Short: "Run CLI in Docker container",
	Long:  `Run the CLI binary in a Docker container with optional arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running CLI in Docker container...")
		ctx := context.Background()

		if err := svc.DockerRunCLI(ctx, args); err != nil {
			log.Fatal(err)
		}
	},
}

var dockerRunMCPCmd = &cobra.Command{
	Use:   "docker-run-mcp",
	Short: "Run MCP server in Docker container",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running MCP server in Docker container...")
		ctx := context.Background()

		if err := svc.DockerRunMCP(ctx); err != nil {
			log.Fatal(err)
		}
	},
}

var dockerPushCmd = &cobra.Command{
	Use:   "docker-push",
	Short: "Push Docker image to registry",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Pushing Docker image...")
		ctx := context.Background()

		if err := svc.DockerPush(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Docker image pushed successfully")
	},
}

var dockerComposeUpCmd = &cobra.Command{
	Use:   "docker-compose-up",
	Short: "Start services with docker-compose",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting services with docker-compose...")
		ctx := context.Background()

		if err := svc.DockerComposeUp(ctx); err != nil {
			log.Fatal(err)
		}
	},
}

var dockerComposeDownCmd = &cobra.Command{
	Use:   "docker-compose-down",
	Short: "Stop services with docker-compose",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping services with docker-compose...")
		ctx := context.Background()

		if err := svc.DockerComposeDown(ctx); err != nil {
			log.Fatal(err)
		}
	},
}

var dockerCleanCmd = &cobra.Command{
	Use:   "docker-clean",
	Short: "Clean Docker artifacts",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Cleaning Docker artifacts...")
		ctx := context.Background()

		if err := svc.DockerClean(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Docker cleanup completed")
	},
}
