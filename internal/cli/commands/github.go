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

// githubCmd represents the github command
var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "Manage GitHub integration",
	Long: `Manage GitHub integration settings including authentication,
repository discovery, and connection management.`,
}

// githubLoginCmd represents the github login command
var githubLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Set up GitHub authentication",
	Long: `Interactively set up GitHub authentication by providing a Personal Access Token.
This token needs 'repo', 'workflow', and 'read:org' scopes for full functionality.`,
	RunE: runGitHubLogin,
}

// githubStatusCmd represents the github status command
var githubStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check GitHub integration status",
	Long: `Check the current status of GitHub integration including authentication
and accessible repositories.`,
	RunE: runGitHubStatus,
}

// githubReposCmd represents the github repos command
var githubReposCmd = &cobra.Command{
	Use:   "repos",
	Short: "List accessible GitHub repositories",
	Long: `List all GitHub repositories that the authenticated user can access.
Shows repository details including stars, forks, and open issues.`,
	RunE: runGitHubRepos,
}

// githubCurrentCmd represents the github current command
var githubCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current repository information",
	Long: `Detect and display information about the current GitHub repository
based on the git remote URL in the current directory.`,
	RunE: runGitHubCurrent,
}

// githubSearchCmd represents the github search command
var githubSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for GitHub repositories",
	Long: `Search for GitHub repositories using the GitHub Search API.
Supports advanced search qualifiers like language, stars, forks, etc.

Examples:
  juleson github search "machine learning"           # Search by topic
  juleson github search "language:go stars:>100"     # Go repos with >100 stars
  juleson github search "user:octocat"               # Repos by specific user
  juleson github search "org:github forks:>50"       # GitHub org repos with >50 forks

Search qualifiers:
  • language:LANGUAGE     - Filter by programming language
  • stars:>N or stars:<N  - Filter by star count
  • forks:>N or forks:<N  - Filter by fork count
  • size:>N or size:<N    - Filter by repository size (KB)
  • user:USERNAME         - Repositories from specific user
  • org:ORGNAME           - Repositories from specific organization
  • in:name or in:description or in:topics - Search in specific fields
  • created:>YYYY-MM-DD   - Created after date
  • pushed:>YYYY-MM-DD    - Pushed to after date
  • license:LICENSE       - Filter by license
  • is:public or is:private - Repository visibility
  • archived:true/false   - Include/exclude archived repos`,
	RunE: runGitHubSearch,
}

var (
	githubReposLimit  int
	githubSearchLimit int
	githubSearchSort  string
	githubSearchOrder string
)

func init() {
	// Add subcommands to github
	githubCmd.AddCommand(githubLoginCmd)
	githubCmd.AddCommand(githubStatusCmd)
	githubCmd.AddCommand(githubReposCmd)
	githubCmd.AddCommand(githubCurrentCmd)
	githubCmd.AddCommand(githubSearchCmd)

	// Add flags
	githubReposCmd.Flags().IntVarP(&githubReposLimit, "limit", "l", 20, "Maximum number of repositories to list")
	githubSearchCmd.Flags().IntVarP(&githubSearchLimit, "limit", "l", 30, "Maximum number of search results to return")
	githubSearchCmd.Flags().StringVarP(&githubSearchSort, "sort", "s", "", "Sort results by: stars, forks, updated (default: best match)")
	githubSearchCmd.Flags().StringVarP(&githubSearchOrder, "order", "o", "desc", "Sort order: asc or desc")
}

func runGitHubLogin(cmd *cobra.Command, args []string) error {
	fmt.Println("🔐 GitHub Authentication Setup")
	fmt.Println("==============================")
	fmt.Println()
	fmt.Println("To integrate Juleson with GitHub, you need a Personal Access Token.")
	fmt.Println("Create one at: https://github.com/settings/tokens")
	fmt.Println()
	fmt.Println("Required scopes:")
	fmt.Println("  ✅ repo          - Full control of private repositories")
	fmt.Println("  ✅ workflow      - Update GitHub Action workflows")
	fmt.Println("  ✅ read:org      - Read org and team membership (optional)")
	fmt.Println()
	fmt.Println("The token will be stored securely in your Juleson configuration.")
	fmt.Println()

	// Check if token already exists
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.GitHub.Token != "" {
		fmt.Println("⚠️  GitHub token is already configured.")
		fmt.Print("Do you want to update it? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Login cancelled.")
			return nil
		}
	}

	// Prompt for token
	fmt.Print("Enter your GitHub Personal Access Token: ")
	var token string
	fmt.Scanln(&token)
	token = strings.TrimSpace(token)

	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Validate token format (basic check)
	if !strings.HasPrefix(token, "ghp_") && !strings.HasPrefix(token, "github_pat_") {
		fmt.Println("⚠️  Warning: Token doesn't start with 'ghp_' or 'github_pat_'")
		fmt.Println("   This might not be a valid GitHub Personal Access Token.")
		fmt.Print("Continue anyway? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Login cancelled.")
			return nil
		}
	}

	// Test the token
	fmt.Println("🔍 Testing GitHub authentication...")
	julesClient := jules.NewClient(cfg.Jules.APIKey, cfg.Jules.BaseURL, cfg.Jules.Timeout, cfg.Jules.RetryAttempts)
	ghClient := ghclient.NewClient(token, julesClient)

	if ghClient == nil {
		return fmt.Errorf("failed to create GitHub client with provided token")
	}

	ctx := context.Background()
	user, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to authenticate with GitHub: %w", err)
	}

	// Save the token to config
	cfg.GitHub.Token = token

	// Save the configuration
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Successfully authenticated as: %s\n", user.GetLogin())
	fmt.Println("✅ GitHub token saved to configuration!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • Run 'juleson github status' to verify setup")
	fmt.Println("  • Run 'juleson github repos' to see accessible repositories")
	fmt.Println("  • Run 'juleson sessions create \"prompt\"' to create sessions")

	return nil
}

func runGitHubStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("📊 GitHub Integration Status")
	fmt.Println("============================")

	// Check if token is configured
	if cfg.GitHub.Token == "" {
		fmt.Println("❌ GitHub token not configured")
		fmt.Println()
		fmt.Println("To set up GitHub integration:")
		fmt.Println("  1. Run: juleson github login")
		fmt.Println("  2. Or set GITHUB_TOKEN environment variable")
		fmt.Println("  3. Or add 'token: your_token' to ~/.juleson.yaml")
		return nil
	}

	fmt.Println("✅ GitHub token configured")

	// Test authentication
	julesClient := jules.NewClient(cfg.Jules.APIKey, cfg.Jules.BaseURL, cfg.Jules.Timeout, cfg.Jules.RetryAttempts)
	ghClient := ghclient.NewClient(cfg.GitHub.Token, julesClient)

	if ghClient == nil {
		fmt.Println("❌ Failed to create GitHub client")
		return nil
	}

	ctx := context.Background()
	user, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		fmt.Println("❌ GitHub authentication failed")
		fmt.Printf("Error: %v\n", err)
		return nil
	}

	fmt.Printf("👤 Authenticated as: %s\n", user.GetLogin())
	fmt.Printf("📧 Email: %s\n", user.GetEmail())
	fmt.Printf("🏢 Company: %s\n", user.GetCompany())

	// Get rate limit info
	rate, _, err := ghClient.RateLimits(ctx)
	if err == nil && rate != nil {
		fmt.Printf("📊 API Rate Limit: %d/%d remaining\n", rate.Core.Remaining, rate.Core.Limit)
	}

	// Check connected repos
	repos, err := ghClient.Repositories.ListConnectedRepos(ctx)
	if err != nil {
		fmt.Printf("⚠️  Could not check connected repositories: %v\n", err)
	} else {
		fmt.Printf("🔗 Connected repositories: %d\n", len(repos))
	}

	fmt.Println()
	fmt.Println("🎉 GitHub integration is working correctly!")

	return nil
}

func runGitHubRepos(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.GitHub.Token == "" {
		return fmt.Errorf("GitHub token not configured. Run 'juleson github login' first")
	}

	julesClient := jules.NewClient(cfg.Jules.APIKey, cfg.Jules.BaseURL, cfg.Jules.Timeout, cfg.Jules.RetryAttempts)
	ghClient := ghclient.NewClient(cfg.GitHub.Token, julesClient)

	if ghClient == nil {
		return fmt.Errorf("failed to create GitHub client")
	}

	ctx := context.Background()

	fmt.Printf("🔍 Fetching accessible repositories (limit: %d)...\n", githubReposLimit)

	repos, err := ghClient.Repositories.ListAccessibleRepos(ctx)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %v", err)
	}

	if len(repos) == 0 {
		fmt.Println("📭 No accessible repositories found.")
		return nil
	}

	// Limit results
	if len(repos) > githubReposLimit {
		repos = repos[:githubReposLimit]
	}

	fmt.Printf("📊 Found %d accessible repositories:\n\n", len(repos))

	// Display in a nice table format
	fmt.Printf("%-40s %-8s %-8s %-12s %-10s\n", "Repository", "Stars", "Forks", "Issues", "Private")
	fmt.Println(strings.Repeat("-", 80))

	for _, repo := range repos {
		private := "Public"
		if repo.Private {
			private = "Private"
		}

		fmt.Printf("%-40s %-8d %-8d %-12d %-10s\n",
			fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
			repo.Stars,
			repo.Forks,
			repo.OpenIssues,
			private)
	}

	fmt.Println()
	fmt.Println("💡 Use 'juleson sources connect owner/repo' to connect repositories to Jules")

	return nil
}

func runGitHubCurrent(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.GitHub.Token == "" {
		return fmt.Errorf("GitHub token not configured. Run 'juleson github login' first")
	}

	julesClient := jules.NewClient(cfg.Jules.APIKey, cfg.Jules.BaseURL, cfg.Jules.Timeout, cfg.Jules.RetryAttempts)
	ghClient := ghclient.NewClient(cfg.GitHub.Token, julesClient)

	if ghClient == nil {
		return fmt.Errorf("failed to create GitHub client")
	}

	ctx := context.Background()

	fmt.Println("🔍 Detecting current repository...")

	repo, err := ghClient.Repositories.DiscoverCurrentRepo(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect current repository: %v", err)
	}

	fmt.Printf("📁 Current Repository: %s\n", repo.FullName)
	fmt.Printf("👤 Owner: %s\n", repo.Owner)
	fmt.Printf("📦 Name: %s\n", repo.Name)
	fmt.Printf("⭐ Stars: %d\n", repo.Stars)
	fmt.Printf("🍴 Forks: %d\n", repo.Forks)
	fmt.Printf("📋 Open Issues: %d\n", repo.OpenIssues)
	fmt.Printf("🌿 Default Branch: %s\n", repo.DefaultBranch)
	fmt.Printf("🔒 Private: %t\n", repo.Private)
	fmt.Printf("🔗 URL: %s\n", repo.URL)

	if repo.Description != "" {
		fmt.Printf("📝 Description: %s\n", repo.Description)
	}

	if repo.UpdatedAt != "" {
		fmt.Printf("🕒 Last Updated: %s\n", repo.UpdatedAt)
	}

	fmt.Println()
	fmt.Println("💡 You can now create Jules sessions for this repository:")
	fmt.Printf("   juleson sessions create \"Your prompt here\"\n")

	return nil
}

func runGitHubSearch(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.GitHub.Token == "" {
		return fmt.Errorf("GitHub token not configured. Run 'juleson github login' first")
	}

	if len(args) == 0 {
		return fmt.Errorf("search query is required. Use 'juleson github search --help' for examples")
	}

	query := strings.Join(args, " ")

	julesClient := jules.NewClient(cfg.Jules.APIKey, cfg.Jules.BaseURL, cfg.Jules.Timeout, cfg.Jules.RetryAttempts)
	ghClient := ghclient.NewClient(cfg.GitHub.Token, julesClient)

	if ghClient == nil {
		return fmt.Errorf("failed to create GitHub client")
	}

	ctx := context.Background()

	fmt.Printf("🔍 Searching GitHub repositories for: %s\n", query)

	// Prepare search options
	opts := &github.SearchOptions{
		Sort:  githubSearchSort,
		Order: githubSearchOrder,
		ListOptions: github.ListOptions{
			PerPage: githubSearchLimit,
		},
	}

	repos, err := ghClient.Repositories.SearchRepositories(ctx, query, opts)
	if err != nil {
		return fmt.Errorf("failed to search repositories: %v", err)
	}

	if len(repos) == 0 {
		fmt.Println("📭 No repositories found matching your search.")
		return nil
	}

	fmt.Printf("📊 Found %d repositories:\n\n", len(repos))

	// Display in a nice table format
	fmt.Printf("%-40s %-8s %-8s %-12s %-10s\n", "Repository", "Stars", "Forks", "Issues", "Private")
	fmt.Println(strings.Repeat("-", 80))

	for _, repo := range repos {
		private := "Public"
		if repo.Private {
			private = "Private"
		}

		fmt.Printf("%-40s %-8d %-8d %-12d %-10s\n",
			fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
			repo.Stars,
			repo.Forks,
			repo.OpenIssues,
			private)
	}

	fmt.Println()
	fmt.Println("💡 Use 'juleson sources connect owner/repo' to connect repositories to Jules")

	return nil
}

// NewGitHubCommand creates the github command
func NewGitHubCommand(cfg *config.Config) *cobra.Command {
	return githubCmd
}
