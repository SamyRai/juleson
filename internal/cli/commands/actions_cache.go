package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/spf13/cobra"
)

// Cache commands

func newActionsCacheCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cache",
		Aliases: []string{"caches"},
		Short:   "Manage GitHub Actions caches",
		Long:    "List and delete GitHub Actions caches",
	}

	cmd.AddCommand(
		newActionsCacheListCommand(cfg),
		newActionsCacheDeleteCommand(cfg),
	)

	return cmd
}
func newActionsCacheListCommand(cfg *config.Config) *cobra.Command {
	var repo string
	var ref string

	cmd := &cobra.Command{
		Use:   "list [owner/repo]",
		Short: "List caches",
		Long:  "List GitHub Actions caches for a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, args, repo)
			if err != nil {
				return err
			}

			var refPtr *string
			if ref != "" {
				refPtr = &ref
			}

			caches, err := client.Actions.ListCaches(ctx, owner, repoName, refPtr)
			if err != nil {
				return fmt.Errorf("failed to list caches: %w", err)
			}

			if len(caches) == 0 {
				fmt.Println("No caches found")
				return nil
			}

			fmt.Printf("Caches:\n\n")
			for _, cache := range caches {
				fmt.Printf("  🗄️  %s\n", cache.GetKey())
				fmt.Printf("     ID: %d\n", cache.GetID())
				fmt.Printf("     Ref: %s\n", cache.GetRef())
				fmt.Printf("     Size: %.2f MB\n", float64(cache.GetSizeInBytes())/(1024*1024))
				fmt.Printf("     Last accessed: %s\n", cache.GetLastAccessedAt().Format("2006-01-02T15:04:05Z"))
				fmt.Printf("     Created: %s\n", cache.GetCreatedAt().Format("2006-01-02T15:04:05Z"))
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")
	cmd.Flags().StringVar(&ref, "ref", "", "Filter by Git reference (branch/tag)")

	return cmd
}
func newActionsCacheDeleteCommand(cfg *config.Config) *cobra.Command {
	var repo string
	var key string
	var id int64
	var ref string
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [owner/repo]",
		Short: "Delete caches",
		Long:  "Delete caches by key or ID",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, args, repo)
			if err != nil {
				return err
			}

			if key == "" && id == 0 {
				return fmt.Errorf("either --key or --id must be specified")
			}

			if !force {
				if key != "" {
					fmt.Printf("Delete caches with key '%s'? (y/N): ", key)
				} else {
					fmt.Printf("Delete cache #%d? (y/N): ", id)
				}
				var response string
				if err := scanPromptValue(&response); err != nil {
					return fmt.Errorf("failed to read response: %w", err)
				}
				if strings.ToLower(response) != "y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			if key != "" {
				var refPtr *string
				if ref != "" {
					refPtr = &ref
				}
				err = client.Actions.DeleteCachesByKey(ctx, owner, repoName, key, refPtr)
				if err != nil {
					return fmt.Errorf("failed to delete caches: %w", err)
				}
				fmt.Printf("✅ Caches with key '%s' deleted\n", key)
			} else {
				err = client.Actions.DeleteCacheByID(ctx, owner, repoName, id)
				if err != nil {
					return fmt.Errorf("failed to delete cache: %w", err)
				}
				fmt.Printf("✅ Cache #%d deleted\n", id)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")
	cmd.Flags().StringVar(&key, "key", "", "Cache key")
	cmd.Flags().Int64Var(&id, "id", 0, "Cache ID")
	cmd.Flags().StringVar(&ref, "ref", "", "Git reference (when using --key)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}
