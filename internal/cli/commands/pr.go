package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	ghclient "github.com/SamyRai/juleson/internal/github"
	"github.com/SamyRai/juleson/internal/jules"
	"github.com/google/go-github/v76/github"
	"github.com/spf13/cobra"
)

// prCmd represents the pr command
var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Manage pull requests created by Jules sessions",
	Long: `Manage pull requests that were created by Jules automation sessions.
This command provides functionality to list, view, merge, and review
pull requests created by Jules sessions.`,
}

// prListCmd represents the pr list command
var prListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pull requests from Jules sessions",
	Long: `List all pull requests that were created by Jules sessions.
Shows PR details including status, repository, and merge status.`,
	RunE: runPRList,
}

// prGetCmd represents the pr get command
var prGetCmd = &cobra.Command{
	Use:   "get <session-id>",
	Short: "Get details of a pull request from a Jules session",
	Long: `Get detailed information about a pull request created by a specific Jules session.
Shows PR title, description, status, reviewers, and CI checks.`,
	Args: cobra.ExactArgs(1),
	RunE: runPRGet,
}

// prMergeCmd represents the pr merge command
var prMergeCmd = &cobra.Command{
	Use:   "merge <session-id>",
	Short: "Merge a pull request from a Jules session",
	Long: `Merge a pull request that was created by a Jules session.
Supports different merge methods: merge, squash, or rebase.`,
	Args: cobra.ExactArgs(1),
	RunE: runPRMerge,
}

// prDiffCmd represents the pr diff command
var prDiffCmd = &cobra.Command{
	Use:   "diff <session-id>",
	Short: "Show the diff of a pull request from a Jules session",
	Long: `Display the git diff of changes in a pull request created by a Jules session.
Shows the actual code changes that would be merged.`,
	Args: cobra.ExactArgs(1),
	RunE: runPRDiff,
}

var (
	prListLimit   int
	prMergeMethod string
	prMergeCommit string
)

func init() {
	// Add subcommands to pr
	prCmd.AddCommand(prListCmd)
	prCmd.AddCommand(prGetCmd)
	prCmd.AddCommand(prMergeCmd)
	prCmd.AddCommand(prDiffCmd)

	// Add flags
	prListCmd.Flags().IntVarP(&prListLimit, "limit", "l", 10, "Maximum number of PRs to list")
	prMergeCmd.Flags().StringVarP(&prMergeMethod, "method", "m", "", "Merge method: merge, squash, or rebase (default: squash)")
	prMergeCmd.Flags().StringVarP(&prMergeCommit, "commit-message", "c", "", "Custom commit message for merge (only applies to merge and squash)")
}

func runPRList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	julesClient := jules.NewClient(cfg.Jules.APIKey, cfg.Jules.BaseURL, cfg.Jules.Timeout, cfg.Jules.RetryAttempts)

	ghClient := ghclient.NewClient(cfg.GitHub.Token, julesClient)
	if ghClient == nil {
		return fmt.Errorf("GitHub client not configured - please set GITHUB_TOKEN")
	}

	ctx := context.Background()

	// Get recent sessions
	sessions, err := julesClient.ListSessions(ctx, prListLimit*2) // Get more to account for sessions without PRs
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	fmt.Println("🔍 Jules Session Pull Requests")
	fmt.Println("================================")

	prCount := 0
	for _, session := range sessions {
		if prCount >= prListLimit {
			break
		}

		// Try to get PR for this session
		pr, err := ghClient.PullRequests.GetSessionPullRequest(ctx, session.ID)
		if err != nil {
			// Skip sessions without PRs or with errors
			continue
		}

		displayPR(session, pr)
		prCount++
	}

	if prCount == 0 {
		fmt.Println("No pull requests found from recent Jules sessions.")
		fmt.Println("Try running more sessions or check that sessions have created PRs.")
	}

	return nil
}

func runPRGet(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	julesClient := jules.NewClient(cfg.Jules.APIKey, cfg.Jules.BaseURL, cfg.Jules.Timeout, cfg.Jules.RetryAttempts)

	ghClient := ghclient.NewClient(cfg.GitHub.Token, julesClient)
	if ghClient == nil {
		return fmt.Errorf("GitHub client not configured - please set GITHUB_TOKEN")
	}

	ctx := context.Background()

	// Get session details
	_, err = julesClient.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session %s: %w", sessionID, err)
	}

	// Get PR details
	pr, err := ghClient.PullRequests.GetSessionPullRequest(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get PR for session %s: %w", sessionID, err)
	}

	fmt.Printf("📝 Pull Request #%d\n", pr.GetNumber())
	fmt.Printf("Title: %s\n", pr.GetTitle())
	fmt.Printf("Repository: %s\n", pr.GetBase().GetRepo().GetFullName())
	fmt.Printf("Branch: %s → %s\n", pr.GetHead().GetRef(), pr.GetBase().GetRef())
	fmt.Printf("Author: %s\n", pr.GetUser().GetLogin())
	fmt.Printf("Status: %s\n", getPRStatus(pr))
	fmt.Printf("URL: %s\n", pr.GetHTMLURL())

	if pr.GetBody() != "" {
		fmt.Printf("\nDescription:\n%s\n", pr.GetBody())
	}

	return nil
}

func runPRMerge(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	julesClient := jules.NewClient(cfg.Jules.APIKey, cfg.Jules.BaseURL, cfg.Jules.Timeout, cfg.Jules.RetryAttempts)

	ghClient := ghclient.NewClient(cfg.GitHub.Token, julesClient)
	if ghClient == nil {
		return fmt.Errorf("GitHub client not configured - please set GITHUB_TOKEN")
	}

	ctx := context.Background()

	// Get PR details first
	pr, err := ghClient.PullRequests.GetSessionPullRequest(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get PR for session %s: %w", sessionID, err)
	}

	// Check if PR can be merged
	if !pr.GetMergeable() {
		return fmt.Errorf("PR #%d cannot be merged - it may have conflicts or failing checks", pr.GetNumber())
	}

	// Determine merge method
	mergeMethod := prMergeMethod
	if mergeMethod == "" {
		mergeMethod = cfg.GitHub.PR.DefaultMergeMethod
		if mergeMethod == "" {
			mergeMethod = "squash"
		}
	}

	// Confirm merge
	fmt.Printf("🔄 Merging PR #%d: %s\n", pr.GetNumber(), pr.GetTitle())
	fmt.Printf("Repository: %s\n", pr.GetBase().GetRepo().GetFullName())
	fmt.Printf("Method: %s\n", mergeMethod)

	if !confirmAction("Are you sure you want to merge this PR?") {
		fmt.Println("Merge cancelled.")
		return nil
	}

	// Perform merge
	err = ghClient.PullRequests.MergePullRequest(ctx, pr.GetHTMLURL(), mergeMethod)
	if err != nil {
		return fmt.Errorf("failed to merge PR: %w", err)
	}

	fmt.Printf("✅ Successfully merged PR #%d\n", pr.GetNumber())
	return nil
}

func runPRDiff(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	julesClient := jules.NewClient(cfg.Jules.APIKey, cfg.Jules.BaseURL, cfg.Jules.Timeout, cfg.Jules.RetryAttempts)

	ghClient := ghclient.NewClient(cfg.GitHub.Token, julesClient)
	if ghClient == nil {
		return fmt.Errorf("GitHub client not configured - please set GITHUB_TOKEN")
	}

	ctx := context.Background()

	// Get PR details
	pr, err := ghClient.PullRequests.GetSessionPullRequest(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get PR for session %s: %w", sessionID, err)
	}

	fmt.Printf("📋 Diff for PR #%d: %s\n", pr.GetNumber(), pr.GetTitle())
	fmt.Printf("Repository: %s\n", pr.GetBase().GetRepo().GetFullName())
	fmt.Printf("Branch: %s → %s\n", pr.GetHead().GetRef(), pr.GetBase().GetRef())
	fmt.Println("================================================================")

	// Get the actual diff
	diff, err := ghClient.PullRequests.GetPullRequestDiff(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get PR diff: %w", err)
	}

	if diff == "" {
		fmt.Println("No diff available for this PR.")
	} else {
		fmt.Println(diff)
	}

	return nil
}

// Helper functions

func displayPR(session jules.Session, pr *github.PullRequest) {
	fmt.Printf("\n⚡ Session: %s\n", session.ID)
	fmt.Printf("📝 PR #%d: %s\n", pr.GetNumber(), pr.GetTitle())
	fmt.Printf("📁 Repository: %s\n", pr.GetBase().GetRepo().GetFullName())
	fmt.Printf("🌿 Branch: %s → %s\n", pr.GetHead().GetRef(), pr.GetBase().GetRef())
	fmt.Printf("📊 Status: %s\n", getPRStatus(pr))
	fmt.Printf("🔗 URL: %s\n", pr.GetHTMLURL())
}

func getPRStatus(pr *github.PullRequest) string {
	if pr.GetMerged() {
		return "✅ Merged"
	}
	if pr.GetState() == "closed" {
		return "❌ Closed"
	}
	if !pr.GetMergeable() {
		return "⚠️  Cannot merge (conflicts or failing checks)"
	}
	return "🟢 Ready to merge"
}

// NewPRCommand creates the pr command
func NewPRCommand(cfg *config.Config) *cobra.Command {
	return prCmd
}

func confirmAction(prompt string) bool {
	fmt.Printf("%s (y/N): ", prompt)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
