package core

import (
	"context"
	"fmt"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/logger"

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
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts), jules.WithDebugLog(cfg.Jules.DebugLog), jules.WithLogger(logger.New(logger.Config{Debug: cfg.Jules.DebugLog})))

	ctx := context.Background()

	response, err := julesClient.Sources().List(ctx, &jules.ListSourcesOptions{PageSize: 100, Filter: filter})
	if err != nil {
		return fmt.Errorf("failed to list sources: %w", err)
	}

	fmt.Print(FormatSourcesList(response))
	return nil
}

// FormatSourcesList formats the SourcesResponse into a human-readable string.
func FormatSourcesList(response *jules.SourcesResponse) string {
	sources := response.Sources
	if sources == nil {
		sources = []jules.Source{}
	}

	var output string
	output += fmt.Sprintf("📚 Connected Sources (%d total)\n\n", len(sources))

	if len(sources) == 0 {
		output += "No sources connected. Connect repositories via the Jules web UI at https://jules.google.com\n"
		return output
	}

	for i, source := range sources {
		output += fmt.Sprintf("%d. %s\n", i+1, source.Name)

		if source.GithubRepo != nil {
			owner := source.GithubRepo.Owner
			repo := source.GithubRepo.Repo
			if owner == "" {
				owner = "unknown"
			}
			if repo == "" {
				repo = "unknown"
			}
			output += fmt.Sprintf("   📁 Repository: %s/%s\n", owner, repo)

			if source.GithubRepo.DefaultBranch != nil {
				output += fmt.Sprintf("   🌿 Default Branch: %s\n", source.GithubRepo.DefaultBranch.DisplayName)
			}

			if len(source.GithubRepo.Branches) > 0 {
				var branchNames []string
				for _, branch := range source.GithubRepo.Branches {
					if branch.DisplayName != "" {
						branchNames = append(branchNames, branch.DisplayName)
					}
				}

				if len(branchNames) > 0 {
					output += fmt.Sprintf("   🌳 Branches: %s", branchNames[0])
					for j := 1; j < len(branchNames); j++ {
						output += fmt.Sprintf(", %s", branchNames[j])
					}
					output += "\n"
				}
			}
		}
		output += "\n"
	}

	if response.NextPageToken != "" {
		output += "💡 More sources available. Use pagination for full list.\n"
	}

	return output
}

// getSource gets details for a specific source
func getSource(cfg *config.Config, sourceID string) error {
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts), jules.WithDebugLog(cfg.Jules.DebugLog), jules.WithLogger(logger.New(logger.Config{Debug: cfg.Jules.DebugLog})))

	ctx := context.Background()

	source, err := julesClient.Sources().Get(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get source: %w", err)
	}

	fmt.Print(FormatSourceDetails(source))
	return nil
}

// FormatSourceDetails formats a single Source into a human-readable string.
func FormatSourceDetails(source *jules.Source) string {
	var output string
	output += "📚 Source Details\n"
	output += fmt.Sprintf("Name: %s\n", source.Name)
	output += fmt.Sprintf("ID: %s\n", source.ID)

	if source.GithubRepo != nil {
		output += "\n📁 GitHub Repository:\n"
		output += fmt.Sprintf("  Owner: %s\n", source.GithubRepo.Owner)
		output += fmt.Sprintf("  Repository: %s\n", source.GithubRepo.Repo)

		if source.GithubRepo.DefaultBranch != nil {
			output += fmt.Sprintf("  Default Branch: %s\n", source.GithubRepo.DefaultBranch.DisplayName)
		}

		if len(source.GithubRepo.Branches) > 0 {
			output += "  Available Branches:\n"
			for _, branch := range source.GithubRepo.Branches {
				if branch.DisplayName != "" {
					output += fmt.Sprintf("    • %s\n", branch.DisplayName)
				}
			}
		}
	}

	return output
}
