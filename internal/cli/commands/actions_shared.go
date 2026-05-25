package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/github"
	"github.com/SamyRai/juleson/pkg/jules"
)

// Helper functions

func getGitHubClientAndRepo(ctx context.Context, cfg *config.Config, args []string, repoFlag string) (*github.Client, string, string, error) {
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	client := github.NewClient(cfg.GitHub.Token, julesClient)
	if client == nil {
		return nil, "", "", fmt.Errorf("GitHub token not configured. Run 'juleson github login' first")
	}

	var owner, repo string

	// Try to parse from arguments
	if len(args) > 0 {
		parts := strings.Split(args[0], "/")
		if len(parts) != 2 {
			return nil, "", "", fmt.Errorf("invalid repository format. Use 'owner/repo'")
		}
		owner = parts[0]
		repo = parts[1]
	} else if repoFlag != "" {
		// Try repo flag
		parts := strings.Split(repoFlag, "/")
		if len(parts) != 2 {
			return nil, "", "", fmt.Errorf("invalid repository format. Use 'owner/repo'")
		}
		owner = parts[0]
		repo = parts[1]
	} else {
		// Try to discover from current directory
		currentRepo, err := client.Repositories.DiscoverCurrentRepo(ctx)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to detect repository. Specify repository with 'owner/repo' or use --repo flag: %w", err)
		}
		owner = currentRepo.Owner
		repo = currentRepo.Name
	}

	return client, owner, repo, nil
}
func getStatusIcon(status, conclusion string) string {
	switch status {
	case "completed":
		switch conclusion {
		case "success":
			return "✅"
		case "failure":
			return "❌"
		case "cancelled":
			return "🚫"
		case "skipped":
			return "⏭️"
		default:
			return "✔️"
		}
	case "in_progress":
		return "🔄"
	case "queued":
		return "⏳"
	default:
		return "❓"
	}
}
