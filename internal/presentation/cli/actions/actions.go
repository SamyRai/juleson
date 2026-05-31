package actions

import (
	"github.com/SamyRai/juleson/internal/config"
	"github.com/spf13/cobra"
)

// NewActionsCommand creates the actions command
func NewActionsCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "actions",
		Short: "Manage GitHub Actions workflows and runs",
		Long: `Manage GitHub Actions workflows, runs, jobs, artifacts, and caches.

This command provides comprehensive GitHub Actions management including:
- Workflow management (list, get, trigger)
- Workflow run monitoring (list, get, rerun, cancel)
- Job management (list, get, rerun, logs)
- Artifact management (list, download, delete)
- Cache management (list, delete)`,
	}

	cmd.AddCommand(
		newActionsWorkflowsCommand(cfg),
		newActionsRunsCommand(cfg),
		newActionsJobsCommand(cfg),
		newActionsArtifactsCommand(cfg),
		newActionsCacheCommand(cfg),
	)

	return cmd
}
