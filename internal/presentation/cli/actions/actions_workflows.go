package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/spf13/cobra"
)

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

			inputsMap := parseWorkflowDispatchInputs(inputs)

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

func parseWorkflowDispatchInputs(inputs []string) map[string]interface{} {
	inputsMap := make(map[string]interface{})
	for _, input := range inputs {
		parts := strings.SplitN(input, "=", 2)
		if len(parts) == 2 {
			inputsMap[parts[0]] = parts[1]
		}
	}
	return inputsMap
}
