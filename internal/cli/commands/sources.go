package commands

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/jules"

	"github.com/spf13/cobra"
)

// NewSourcesCommand creates the sources command
func NewSourcesCommand(cfg *config.Config) *cobra.Command {
	sourcesCmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage Jules sources",
		Long:  "List and manage connected Jules sources (GitHub repositories)",
	}

	// List sources
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all connected sources",
		Long: `List all GitHub repositories connected to Jules.

Examples:
  # List all sources
  juleson sources list

  # Filter by exact source name
  juleson sources list --filter "name=sources/github/SamyRai/juleson"

  # Filter multiple sources (OR condition)
  juleson sources list --filter "name=sources/github/SamyRai/juleson OR name=sources/github/SamyRai/juleson-test"

  # For advanced filtering, use grep
  juleson sources list | grep juleson`,
		RunE: func(cmd *cobra.Command, args []string) error {
			filter, _ := cmd.Flags().GetString("filter")
			return listSources(cfg, filter)
		},
	}
	listCmd.Flags().StringP("filter", "f", "", "Filter sources by exact name (e.g., 'name=sources/github/owner/repo')")

	sourcesCmd.AddCommand(listCmd)

	// Get source
	sourcesCmd.AddCommand(&cobra.Command{
		Use:   "get [source-id]",
		Short: "Get details for a specific source",
		Long:  "Get detailed information about a specific connected source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return getSource(cfg, args[0])
		},
	})

	return sourcesCmd
}

// listSources lists all connected sources
func listSources(cfg *config.Config, filter string) error {
	julesClient := jules.NewClient(
		cfg.Jules.APIKey,
		cfg.Jules.BaseURL,
		cfg.Jules.Timeout,
		cfg.Jules.RetryAttempts,
	)

	ctx := context.Background()

	response, err := julesClient.ListSourcesWithPagination(ctx, 100, "", filter)
	if err != nil {
		return fmt.Errorf("failed to list sources: %w", err)
	}

	sources := response.Sources
	if sources == nil {
		sources = []jules.Source{}
	}

	fmt.Printf("📚 Connected Sources (%d total)\n\n", len(sources))

	if len(sources) == 0 {
		fmt.Println("No sources connected. Connect repositories via the Jules web UI at https://jules.google.com")
		return nil
	}

	for i, source := range sources {
		fmt.Printf("%d. %s\n", i+1, source.Name)

		if source.GithubRepo != nil {
			owner := source.GithubRepo.Owner
			repo := source.GithubRepo.Repo
			if owner == "" {
				owner = "unknown"
			}
			if repo == "" {
				repo = "unknown"
			}
			fmt.Printf("   📁 Repository: %s/%s\n", owner, repo)

			if source.GithubRepo.DefaultBranch != nil {
				fmt.Printf("   🌿 Default Branch: %s\n", source.GithubRepo.DefaultBranch.DisplayName)
			}

			if len(source.GithubRepo.Branches) > 0 {
				// Collect non-empty branch names
				var branchNames []string
				for _, branch := range source.GithubRepo.Branches {
					if branch.DisplayName != "" {
						branchNames = append(branchNames, branch.DisplayName)
					}
				}

				if len(branchNames) > 0 {
					fmt.Printf("   🌳 Branches: %s", branchNames[0])
					for j := 1; j < len(branchNames); j++ {
						fmt.Printf(", %s", branchNames[j])
					}
					fmt.Println()
				}
			}
		}
		fmt.Println()
	}

	if response.NextPageToken != "" {
		fmt.Printf("💡 More sources available. Use pagination for full list.\n")
	}

	return nil
}

// getSource gets details for a specific source
func getSource(cfg *config.Config, sourceID string) error {
	julesClient := jules.NewClient(cfg.Jules.APIKey, cfg.Jules.BaseURL, cfg.Jules.Timeout, cfg.Jules.RetryAttempts)

	ctx := context.Background()

	source, err := julesClient.GetSource(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get source: %w", err)
	}

	fmt.Printf("📚 Source Details\n")
	fmt.Printf("Name: %s\n", source.Name)
	fmt.Printf("ID: %s\n", source.ID)

	if source.GithubRepo != nil {
		fmt.Printf("\n📁 GitHub Repository:\n")
		fmt.Printf("  Owner: %s\n", source.GithubRepo.Owner)
		fmt.Printf("  Repository: %s\n", source.GithubRepo.Repo)

		if source.GithubRepo.DefaultBranch != nil {
			fmt.Printf("  Default Branch: %s\n", source.GithubRepo.DefaultBranch.DisplayName)
		}

		if len(source.GithubRepo.Branches) > 0 {
			fmt.Printf("  Available Branches:\n")
			for _, branch := range source.GithubRepo.Branches {
				if branch.DisplayName != "" {
					fmt.Printf("    • %s\n", branch.DisplayName)
				}
			}
		}
	}

	return nil
}
