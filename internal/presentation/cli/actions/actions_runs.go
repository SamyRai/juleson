package actions

import (
	"context"
	"fmt"
	"strconv"

	"github.com/SamyRai/juleson/internal/config"
	gh "github.com/google/go-github/v76/github"
	"github.com/spf13/cobra"
)

// Runs commands

func newActionsRunsCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "runs",
		Aliases: []string{"run"},
		Short:   "Manage workflow runs",
		Long:    "List, get, rerun, and cancel workflow runs",
	}

	cmd.AddCommand(
		newActionsRunsListCommand(cfg),
		newActionsRunsGetCommand(cfg),
		newActionsRunsRerunCommand(cfg),
		newActionsRunsCancelCommand(cfg),
		newActionsRunsLogsCommand(cfg),
	)

	return cmd
}
func newActionsRunsListCommand(cfg *config.Config) *cobra.Command {
	var repo string
	var workflow string
	var status string
	var branch string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [owner/repo]",
		Short: "List workflow runs",
		Long:  "List workflow runs for a repository or specific workflow",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, args, repo)
			if err != nil {
				return err
			}

			opts := &gh.ListWorkflowRunsOptions{
				Status: status,
				Branch: branch,
				ListOptions: gh.ListOptions{
					PerPage: limit,
				},
			}

			runs, err := client.Actions.ListWorkflowRuns(ctx, owner, repoName, workflow, opts)
			if err != nil {
				return fmt.Errorf("failed to list workflow runs: %w", err)
			}

			if len(runs) == 0 {
				fmt.Println("No workflow runs found")
				return nil
			}

			fmt.Printf("Workflow Runs in %s/%s:\n\n", owner, repoName)
			for _, run := range runs {
				statusIcon := getStatusIcon(run.Status, run.Conclusion)
				fmt.Printf("  %s Run #%d: %s\n", statusIcon, run.RunNumber, run.Name)
				fmt.Printf("     ID: %d\n", run.ID)
				fmt.Printf("     Branch: %s\n", run.HeadBranch)
				fmt.Printf("     Status: %s", run.Status)
				if run.Conclusion != "" {
					fmt.Printf(" (%s)", run.Conclusion)
				}
				fmt.Println()
				fmt.Printf("     Actor: %s\n", run.Actor)
				fmt.Printf("     Event: %s\n", run.Event)
				fmt.Printf("     Created: %s\n", run.CreatedAt)
				fmt.Printf("     URL: %s\n", run.URL)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")
	cmd.Flags().StringVarP(&workflow, "workflow", "w", "", "Filter by workflow ID or filename")
	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status (queued, in_progress, completed)")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Filter by branch")
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Maximum number of runs to show")

	return cmd
}
func newActionsRunsGetCommand(cfg *config.Config) *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "get <run-id> [owner/repo]",
		Short: "Get workflow run details",
		Long:  "Get details for a specific workflow run",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			runID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid run ID: %w", err)
			}

			repoArgs := args[1:]
			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, repoArgs, repo)
			if err != nil {
				return err
			}

			run, err := client.Actions.GetWorkflowRun(ctx, owner, repoName, runID)
			if err != nil {
				return fmt.Errorf("failed to get workflow run: %w", err)
			}

			statusIcon := getStatusIcon(run.Status, run.Conclusion)
			fmt.Printf("Workflow Run Details:\n\n")
			fmt.Printf("  %s %s (Run #%d)\n\n", statusIcon, run.Name, run.RunNumber)
			fmt.Printf("  ID: %d\n", run.ID)
			fmt.Printf("  Workflow ID: %d\n", run.WorkflowID)
			fmt.Printf("  Branch: %s\n", run.HeadBranch)
			fmt.Printf("  Status: %s", run.Status)
			if run.Conclusion != "" {
				fmt.Printf(" (%s)", run.Conclusion)
			}
			fmt.Println()
			fmt.Printf("  Actor: %s\n", run.Actor)
			fmt.Printf("  Event: %s\n", run.Event)
			fmt.Printf("  Attempt: %d\n", run.RunAttempt)
			fmt.Printf("  Created: %s\n", run.CreatedAt)
			fmt.Printf("  Updated: %s\n", run.UpdatedAt)
			if run.RunStartedAt != "" {
				fmt.Printf("  Started: %s\n", run.RunStartedAt)
			}
			fmt.Printf("  URL: %s\n", run.URL)

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")

	return cmd
}
func newActionsRunsRerunCommand(cfg *config.Config) *cobra.Command {
	var repo string
	var failedOnly bool

	cmd := &cobra.Command{
		Use:   "rerun <run-id> [owner/repo]",
		Short: "Re-run a workflow run",
		Long:  "Re-run a workflow run or just the failed jobs",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			runID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid run ID: %w", err)
			}

			repoArgs := args[1:]
			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, repoArgs, repo)
			if err != nil {
				return err
			}

			if failedOnly {
				err = client.Actions.RerunFailedJobs(ctx, owner, repoName, runID)
				if err != nil {
					return fmt.Errorf("failed to rerun failed jobs: %w", err)
				}
				fmt.Printf("✅ Rerunning failed jobs for run #%d\n", runID)
			} else {
				err = client.Actions.RerunWorkflow(ctx, owner, repoName, runID)
				if err != nil {
					return fmt.Errorf("failed to rerun workflow: %w", err)
				}
				fmt.Printf("✅ Rerunning workflow run #%d\n", runID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")
	cmd.Flags().BoolVarP(&failedOnly, "failed-only", "f", false, "Rerun only failed jobs")

	return cmd
}
func newActionsRunsCancelCommand(cfg *config.Config) *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "cancel <run-id> [owner/repo]",
		Short: "Cancel a workflow run",
		Long:  "Cancel a running workflow",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			runID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid run ID: %w", err)
			}

			repoArgs := args[1:]
			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, repoArgs, repo)
			if err != nil {
				return err
			}

			err = client.Actions.CancelWorkflow(ctx, owner, repoName, runID)
			if err != nil {
				return fmt.Errorf("failed to cancel workflow: %w", err)
			}

			fmt.Printf("✅ Workflow run #%d cancelled\n", runID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")

	return cmd
}
func newActionsRunsLogsCommand(cfg *config.Config) *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "logs <run-id> [owner/repo]",
		Short: "Get workflow run logs URL",
		Long:  "Get the download URL for workflow run logs",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			runID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid run ID: %w", err)
			}

			repoArgs := args[1:]
			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, repoArgs, repo)
			if err != nil {
				return err
			}

			url, err := client.Actions.DownloadWorkflowLogs(ctx, owner, repoName, runID)
			if err != nil {
				return fmt.Errorf("failed to get logs URL: %w", err)
			}

			fmt.Printf("Logs URL: %s\n", url)
			fmt.Println("\nDownload logs with:")
			fmt.Printf("  curl -L '%s' -o logs.zip\n", url)

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")

	return cmd
}
