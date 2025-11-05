package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/github"
	"github.com/SamyRai/juleson/internal/jules"
	gh "github.com/google/go-github/v76/github"
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

// Workflows commands

func newActionsWorkflowsCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workflows",
		Aliases: []string{"wf", "workflow"},
		Short:   "Manage GitHub Actions workflows",
		Long:    "List, get, and trigger GitHub Actions workflows",
	}

	cmd.AddCommand(
		newActionsWorkflowsListCommand(cfg),
		newActionsWorkflowsGetCommand(cfg),
		newActionsWorkflowsTriggerCommand(cfg),
	)

	return cmd
}

func newActionsWorkflowsListCommand(cfg *config.Config) *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "list [owner/repo]",
		Short: "List workflows in a repository",
		Long:  "List all GitHub Actions workflows in a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, args, repo)
			if err != nil {
				return err
			}

			workflows, err := client.Actions.ListWorkflows(ctx, owner, repoName)
			if err != nil {
				return fmt.Errorf("failed to list workflows: %w", err)
			}

			if len(workflows) == 0 {
				fmt.Println("No workflows found")
				return nil
			}

			fmt.Printf("Workflows in %s/%s:\n\n", owner, repoName)
			for _, wf := range workflows {
				fmt.Printf("  ID: %d\n", wf.ID)
				fmt.Printf("  Name: %s\n", wf.Name)
				fmt.Printf("  Path: %s\n", wf.Path)
				fmt.Printf("  State: %s\n", wf.State)
				fmt.Printf("  URL: %s\n", wf.URL)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")

	return cmd
}

func newActionsWorkflowsGetCommand(cfg *config.Config) *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "get <workflow-id-or-file> [owner/repo]",
		Short: "Get workflow details",
		Long:  "Get details for a specific workflow by ID or filename",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			workflowIDOrFile := args[0]

			repoArgs := args[1:]
			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, repoArgs, repo)
			if err != nil {
				return err
			}

			wf, err := client.Actions.GetWorkflow(ctx, owner, repoName, workflowIDOrFile)
			if err != nil {
				return fmt.Errorf("failed to get workflow: %w", err)
			}

			fmt.Printf("Workflow Details:\n\n")
			fmt.Printf("  ID: %d\n", wf.ID)
			fmt.Printf("  Name: %s\n", wf.Name)
			fmt.Printf("  Path: %s\n", wf.Path)
			fmt.Printf("  State: %s\n", wf.State)
			fmt.Printf("  Created: %s\n", wf.CreatedAt)
			fmt.Printf("  Updated: %s\n", wf.UpdatedAt)
			fmt.Printf("  URL: %s\n", wf.URL)
			fmt.Printf("  Badge: %s\n", wf.BadgeURL)

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")

	return cmd
}

func newActionsWorkflowsTriggerCommand(cfg *config.Config) *cobra.Command {
	var repo string
	var ref string
	var inputs []string

	cmd := &cobra.Command{
		Use:   "trigger <workflow-id-or-file> [owner/repo]",
		Short: "Trigger a workflow dispatch event",
		Long:  "Manually trigger a workflow using workflow_dispatch event",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			workflowIDOrFile := args[0]

			repoArgs := args[1:]
			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, repoArgs, repo)
			if err != nil {
				return err
			}

			// Parse inputs from key=value format
			inputsMap := make(map[string]interface{})
			for _, input := range inputs {
				parts := strings.SplitN(input, "=", 2)
				if len(parts) == 2 {
					inputsMap[parts[0]] = parts[1]
				}
			}

			// Default to main branch if not specified
			if ref == "" {
				ghRepo, _, err := client.Client.Repositories.Get(ctx, owner, repoName)
				if err != nil {
					return fmt.Errorf("failed to get repository: %w", err)
				}
				ref = ghRepo.GetDefaultBranch()
			}

			err = client.Actions.TriggerWorkflow(ctx, owner, repoName, workflowIDOrFile, ref, inputsMap)
			if err != nil {
				return fmt.Errorf("failed to trigger workflow: %w", err)
			}

			fmt.Printf("✅ Workflow triggered successfully on branch '%s'\n", ref)

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")
	cmd.Flags().StringVar(&ref, "ref", "", "Git reference (branch or tag)")
	cmd.Flags().StringArrayVarP(&inputs, "input", "i", []string{}, "Workflow inputs (key=value format)")

	return cmd
}

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

// Jobs commands

func newActionsJobsCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jobs",
		Aliases: []string{"job"},
		Short:   "Manage workflow jobs",
		Long:    "List, get, and rerun workflow jobs",
	}

	cmd.AddCommand(
		newActionsJobsListCommand(cfg),
		newActionsJobsGetCommand(cfg),
		newActionsJobsRerunCommand(cfg),
		newActionsJobsLogsCommand(cfg),
	)

	return cmd
}

func newActionsJobsListCommand(cfg *config.Config) *cobra.Command {
	var repo string
	var filter string

	cmd := &cobra.Command{
		Use:   "list <run-id> [owner/repo]",
		Short: "List jobs for a workflow run",
		Long:  "List all jobs for a specific workflow run",
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

			jobs, err := client.Actions.ListWorkflowJobs(ctx, owner, repoName, runID, filter)
			if err != nil {
				return fmt.Errorf("failed to list jobs: %w", err)
			}

			if len(jobs) == 0 {
				fmt.Println("No jobs found")
				return nil
			}

			fmt.Printf("Jobs for Run #%d:\n\n", runID)
			for _, job := range jobs {
				statusIcon := getStatusIcon(job.Status, job.Conclusion)
				fmt.Printf("  %s %s\n", statusIcon, job.Name)
				fmt.Printf("     ID: %d\n", job.ID)
				fmt.Printf("     Status: %s", job.Status)
				if job.Conclusion != "" {
					fmt.Printf(" (%s)", job.Conclusion)
				}
				fmt.Println()
				if job.RunnerName != "" {
					fmt.Printf("     Runner: %s\n", job.RunnerName)
				}
				if job.StartedAt != "" {
					fmt.Printf("     Started: %s\n", job.StartedAt)
				}
				if job.CompletedAt != "" {
					fmt.Printf("     Completed: %s\n", job.CompletedAt)
				}
				fmt.Printf("     URL: %s\n", job.URL)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")
	cmd.Flags().StringVarP(&filter, "filter", "f", "latest", "Filter jobs (latest, all)")

	return cmd
}

func newActionsJobsGetCommand(cfg *config.Config) *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "get <job-id> [owner/repo]",
		Short: "Get job details",
		Long:  "Get details for a specific job",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			jobID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid job ID: %w", err)
			}

			repoArgs := args[1:]
			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, repoArgs, repo)
			if err != nil {
				return err
			}

			job, err := client.Actions.GetWorkflowJob(ctx, owner, repoName, jobID)
			if err != nil {
				return fmt.Errorf("failed to get job: %w", err)
			}

			statusIcon := getStatusIcon(job.Status, job.Conclusion)
			fmt.Printf("Job Details:\n\n")
			fmt.Printf("  %s %s\n\n", statusIcon, job.Name)
			fmt.Printf("  ID: %d\n", job.ID)
			fmt.Printf("  Run ID: %d\n", job.RunID)
			fmt.Printf("  Status: %s", job.Status)
			if job.Conclusion != "" {
				fmt.Printf(" (%s)", job.Conclusion)
			}
			fmt.Println()
			if job.RunnerName != "" {
				fmt.Printf("  Runner: %s\n", job.RunnerName)
			}
			if job.StartedAt != "" {
				fmt.Printf("  Started: %s\n", job.StartedAt)
			}
			if job.CompletedAt != "" {
				fmt.Printf("  Completed: %s\n", job.CompletedAt)
			}
			fmt.Printf("  URL: %s\n", job.URL)

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")

	return cmd
}

func newActionsJobsRerunCommand(cfg *config.Config) *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "rerun <job-id> [owner/repo]",
		Short: "Re-run a job",
		Long:  "Re-run a specific job and its dependent jobs",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			jobID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid job ID: %w", err)
			}

			repoArgs := args[1:]
			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, repoArgs, repo)
			if err != nil {
				return err
			}

			err = client.Actions.RerunJob(ctx, owner, repoName, jobID)
			if err != nil {
				return fmt.Errorf("failed to rerun job: %w", err)
			}

			fmt.Printf("✅ Job #%d re-run started\n", jobID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")

	return cmd
}

func newActionsJobsLogsCommand(cfg *config.Config) *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "logs <job-id> [owner/repo]",
		Short: "Get job logs URL",
		Long:  "Get the download URL for job logs",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			jobID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid job ID: %w", err)
			}

			repoArgs := args[1:]
			client, owner, repoName, err := getGitHubClientAndRepo(ctx, cfg, repoArgs, repo)
			if err != nil {
				return err
			}

			url, err := client.Actions.DownloadJobLogs(ctx, owner, repoName, jobID)
			if err != nil {
				return fmt.Errorf("failed to get logs URL: %w", err)
			}

			fmt.Printf("Logs URL: %s\n", url)
			fmt.Println("\nView logs with:")
			fmt.Printf("  curl -L '%s'\n", url)

			return nil
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository (owner/repo format)")

	return cmd
}

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
				fmt.Scanln(&response)
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

// Helper functions

func getGitHubClientAndRepo(ctx context.Context, cfg *config.Config, args []string, repoFlag string) (*github.Client, string, string, error) {
	julesClient := jules.NewClient(cfg.Jules.APIKey, cfg.Jules.BaseURL, cfg.Jules.Timeout, cfg.Jules.RetryAttempts)

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
