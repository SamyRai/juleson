package actions

import (
	"context"
	"fmt"
	"strconv"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/spf13/cobra"
)

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
