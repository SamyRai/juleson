package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/spf13/cobra"
)

// Artifacts commands

func newActionsArtifactsCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "artifacts",
		Aliases: []string{"artifact"},
		Short:   "Manage workflow artifacts",
		Long:    "List, download, and delete workflow artifacts",
	}

	cmd.AddCommand(
		newActionsArtifactsListCommand(cfg),
		newActionsArtifactsDownloadCommand(cfg),
		newActionsArtifactsDeleteCommand(cfg),
	)

	return cmd
}
func newActionsArtifactsListCommand(cfg *config.Config) *cobra.Command {
	var repo string
	var runID int64

	cmd := &cobra.Command{
		Use:   "list [owner/repo]",
		Short: "List artifacts",
		Long:  "List artifacts for a repository or workflow run",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, args, repo)
			if err != nil {
				return err
			}

			artifacts, err := client.Actions.ListArtifacts(ctx, owner, repoName, runID)
			if err != nil {
				return fmt.Errorf("failed to list artifacts: %w", err)
			}

			if len(artifacts) == 0 {
				fmt.Println("No artifacts found")
				return nil
			}

			fmt.Printf("Artifacts:\n\n")
			for _, artifact := range artifacts {
				fmt.Printf("  📦 %s\n", artifact.GetName())
				fmt.Printf("     ID: %d\n", artifact.GetID())
				fmt.Printf("     Size: %.2f MB\n", float64(artifact.GetSizeInBytes())/(1024*1024))
				if artifact.GetExpired() {
					fmt.Printf("     Status: ⚠️  Expired\n")
				} else {
					fmt.Printf("     Expires: %s\n", artifact.GetExpiresAt().Format("2006-01-02"))
				}
				fmt.Printf("     Created: %s\n", artifact.GetCreatedAt().Format("2006-01-02T15:04:05Z"))
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")
	cmd.Flags().Int64Var(&runID, "run-id", 0, "Filter by workflow run ID")

	return cmd
}
func newActionsArtifactsDownloadCommand(cfg *config.Config) *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "download <artifact-id> [owner/repo]",
		Short: "Get artifact download URL",
		Long:  "Get the download URL for a workflow artifact",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			artifactID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid artifact ID: %w", err)
			}

			repoArgs := args[1:]
			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, repoArgs, repo)
			if err != nil {
				return err
			}

			url, err := client.Actions.DownloadArtifact(ctx, owner, repoName, artifactID)
			if err != nil {
				return fmt.Errorf("failed to get artifact download URL: %w", err)
			}

			fmt.Printf("Artifact Download URL: %s\n", url)
			fmt.Println("\nDownload with:")
			fmt.Printf("  curl -L '%s' -o artifact.zip\n", url)

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")

	return cmd
}
func newActionsArtifactsDeleteCommand(cfg *config.Config) *cobra.Command {
	var repo string
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <artifact-id> [owner/repo]",
		Short: "Delete an artifact",
		Long:  "Delete a workflow artifact",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			artifactID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid artifact ID: %w", err)
			}

			repoArgs := args[1:]
			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, repoArgs, repo)
			if err != nil {
				return err
			}

			if !force {
				fmt.Printf("Delete artifact #%d? (y/N): ", artifactID)
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			err = client.Actions.DeleteArtifact(ctx, owner, repoName, artifactID)
			if err != nil {
				return fmt.Errorf("failed to delete artifact: %w", err)
			}

			fmt.Printf("✅ Artifact #%d deleted\n", artifactID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}
