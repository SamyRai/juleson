package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/github"
	"github.com/SamyRai/juleson/internal/jules"
	ghapi "github.com/google/go-github/v76/github"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterGitHubTools registers all GitHub-related MCP tools
func RegisterGitHubTools(server *mcp.Server, cfg *config.Config, julesClient *jules.Client) {
	// Only register GitHub tools if GitHub token is configured
	if cfg.GitHub.Token == "" {
		return
	}

	ghClient := github.NewClient(cfg.GitHub.Token, julesClient)
	if ghClient == nil {
		return
	}

	// Create Session from GitHub Repo Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "github_create_session_from_repo",
		Description: "Create Jules session auto-detecting GitHub repo from current directory or explicit repo path",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GitHubCreateSessionInput) (*mcp.CallToolResult, GitHubCreateSessionOutput, error) {
		return createSessionFromRepo(ctx, req, input, cfg, ghClient)
	})

	// Merge Session PR Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "github_merge_session_pr",
		Description: "Merge PR created by Jules session",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GitHubMergePRInput) (*mcp.CallToolResult, GitHubMergePROutput, error) {
		return mergeSessionPR(ctx, req, input, ghClient)
	})

	// List Accessible Repos Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "github_list_repos",
		Description: "List GitHub repositories accessible to the authenticated user",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GitHubListReposInput) (*mcp.CallToolResult, GitHubListReposOutput, error) {
		return listAccessibleRepos(ctx, req, input, ghClient)
	})

	// Get Current Repo Info Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "github_current_repo",
		Description: "Get information about the current GitHub repository (detected from git remote)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GitHubCurrentRepoInput) (*mcp.CallToolResult, GitHubCurrentRepoOutput, error) {
		return getCurrentRepo(ctx, req, input, ghClient)
	})

	// List Connected Repos Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "github_list_connected_repos",
		Description: "List GitHub repositories connected to Jules",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GitHubListConnectedInput) (*mcp.CallToolResult, GitHubListConnectedOutput, error) {
		return listConnectedRepos(ctx, req, input, ghClient)
	})

	// Search Repositories Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "github_search_repos",
		Description: "Search for GitHub repositories using advanced search qualifiers",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GitHubSearchReposInput) (*mcp.CallToolResult, GitHubSearchReposOutput, error) {
		return searchRepositories(ctx, req, input, ghClient)
	})
}

// GitHubCreateSessionInput represents input for github_create_session_from_repo tool
type GitHubCreateSessionInput struct {
	Prompt string `json:"prompt" jsonschema:"Prompt for the Jules session"`
	Repo   string `json:"repo,omitempty" jsonschema:"GitHub repo in format 'owner/repo' (optional - auto-detects from current directory)"`
	Branch string `json:"branch,omitempty" jsonschema:"Branch to create session on (optional - uses default branch)"`
}

// GitHubCreateSessionOutput represents output for github_create_session_from_repo tool
type GitHubCreateSessionOutput struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
	Repo      string `json:"repo"`
	Branch    string `json:"branch"`
	Message   string `json:"message"`
}

// createSessionFromRepo creates a Jules session from a GitHub repository
func createSessionFromRepo(ctx context.Context, req *mcp.CallToolRequest, input GitHubCreateSessionInput, cfg *config.Config, ghClient *github.Client) (
	*mcp.CallToolResult,
	GitHubCreateSessionOutput,
	error,
) {
	var session *jules.Session
	var repo *github.Repository
	var err error

	if input.Repo != "" {
		// Explicit repo provided - parse it
		parts := strings.Split(input.Repo, "/")
		if len(parts) != 2 {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid repo format. Use 'owner/repo'"},
				},
			}, GitHubCreateSessionOutput{}, fmt.Errorf("invalid repo format: %s", input.Repo)
		}

		owner, name := parts[0], parts[1]

		// Verify repo exists and get metadata
		ghRepo, _, err := ghClient.Client.Repositories.Get(ctx, owner, name)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to access repository %s/%s: %v", owner, name, err)},
				},
			}, GitHubCreateSessionOutput{}, err
		}

		repo = &github.Repository{
			Owner:         ghRepo.GetOwner().GetLogin(),
			Name:          ghRepo.GetName(),
			FullName:      ghRepo.GetFullName(),
			Description:   ghRepo.GetDescription(),
			Stars:         ghRepo.GetStargazersCount(),
			Forks:         ghRepo.GetForksCount(),
			OpenIssues:    ghRepo.GetOpenIssuesCount(),
			DefaultBranch: ghRepo.GetDefaultBranch(),
			Private:       ghRepo.GetPrivate(),
			URL:           ghRepo.GetHTMLURL(),
			UpdatedAt:     ghRepo.GetUpdatedAt().Format("2006-01-02T15:04:05Z"),
		}

		// Create session with explicit repo
		session, err = ghClient.Sessions.CreateSessionFromRepo(ctx, input.Prompt, owner, name, input.Branch)
	} else {
		// Auto-detect from current directory
		session, err = ghClient.Sessions.CreateSessionFromCurrentRepo(ctx, input.Prompt, input.Branch)
		if err == nil {
			// Get repo info for output
			repo, _ = ghClient.Repositories.DiscoverCurrentRepo(ctx)
		}
	}

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create session: %v", err)},
			},
		}, GitHubCreateSessionOutput{}, err
	}

	branch := input.Branch
	if branch == "" && repo != nil {
		branch = repo.DefaultBranch
	}

	repoName := ""
	if repo != nil {
		repoName = repo.FullName
	}

	output := GitHubCreateSessionOutput{
		SessionID: session.ID,
		URL:       session.URL,
		Repo:      repoName,
		Branch:    branch,
		Message:   fmt.Sprintf("Session created successfully: %s", session.URL),
	}

	return nil, output, nil
}

// GitHubMergePRInput represents input for github_merge_session_pr tool
type GitHubMergePRInput struct {
	SessionID   string `json:"session_id" jsonschema:"ID of the Jules session"`
	MergeMethod string `json:"merge_method,omitempty" jsonschema:"Merge method: 'merge', 'squash', or 'rebase' (default: 'squash')"`
}

// GitHubMergePROutput represents output for github_merge_session_pr tool
type GitHubMergePROutput struct {
	PRURL    string `json:"pr_url"`
	Merged   bool   `json:"merged"`
	MergeSHA string `json:"merge_sha,omitempty"`
	Message  string `json:"message"`
}

// mergeSessionPR merges the PR created by a Jules session
func mergeSessionPR(ctx context.Context, req *mcp.CallToolRequest, input GitHubMergePRInput, ghClient *github.Client) (
	*mcp.CallToolResult,
	GitHubMergePROutput,
	error,
) {
	// Get session to find PR URL
	session, err := ghClient.PullRequests.GetSessionPullRequest(ctx, input.SessionID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get session PR: %v", err)},
			},
		}, GitHubMergePROutput{}, err
	}

	prURL := session.GetHTMLURL()
	err = ghClient.PullRequests.MergePullRequest(ctx, prURL, input.MergeMethod)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to merge PR: %v", err)},
			},
		}, GitHubMergePROutput{}, err
	}

	output := GitHubMergePROutput{
		PRURL:   prURL,
		Merged:  true,
		Message: fmt.Sprintf("PR merged successfully: %s", prURL),
	}

	return nil, output, nil
}

// GitHubListReposInput represents input for github_list_repos tool
type GitHubListReposInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"Maximum number of repositories to return (default: 50)"`
}

// GitHubListReposOutput represents output for github_list_repos tool
type GitHubListReposOutput struct {
	Repos   []*github.Repository `json:"repos"`
	Count   int                  `json:"count"`
	Message string               `json:"message"`
}

// listAccessibleRepos lists repositories accessible to the user
func listAccessibleRepos(ctx context.Context, req *mcp.CallToolRequest, input GitHubListReposInput, ghClient *github.Client) (
	*mcp.CallToolResult,
	GitHubListReposOutput,
	error,
) {
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}

	repos, err := ghClient.Repositories.ListAccessibleRepos(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list repositories: %v", err)},
			},
		}, GitHubListReposOutput{}, err
	}

	// Limit results
	if len(repos) > limit {
		repos = repos[:limit]
	}

	output := GitHubListReposOutput{
		Repos:   repos,
		Count:   len(repos),
		Message: fmt.Sprintf("Found %d accessible repositories", len(repos)),
	}

	return nil, output, nil
}

// GitHubCurrentRepoInput represents input for github_current_repo tool
type GitHubCurrentRepoInput struct{}

// GitHubCurrentRepoOutput represents output for github_current_repo tool
type GitHubCurrentRepoOutput struct {
	Repo    *github.Repository `json:"repo,omitempty"`
	Message string             `json:"message"`
}

// getCurrentRepo gets information about the current repository
func getCurrentRepo(ctx context.Context, req *mcp.CallToolRequest, input GitHubCurrentRepoInput, ghClient *github.Client) (
	*mcp.CallToolResult,
	GitHubCurrentRepoOutput,
	error,
) {
	repo, err := ghClient.Repositories.DiscoverCurrentRepo(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to detect current repository: %v", err)},
			},
		}, GitHubCurrentRepoOutput{}, err
	}

	output := GitHubCurrentRepoOutput{
		Repo:    repo,
		Message: fmt.Sprintf("Current repository: %s", repo.FullName),
	}

	return nil, output, nil
}

// GitHubListConnectedInput represents input for github_list_connected_repos tool
type GitHubListConnectedInput struct{}

// GitHubListConnectedOutput represents output for github_list_connected_repos tool
type GitHubListConnectedOutput struct {
	Repos   []*github.Repository `json:"repos"`
	Count   int                  `json:"count"`
	Message string               `json:"message"`
}

// listConnectedRepos lists repositories connected to Jules
func listConnectedRepos(ctx context.Context, req *mcp.CallToolRequest, input GitHubListConnectedInput, ghClient *github.Client) (
	*mcp.CallToolResult,
	GitHubListConnectedOutput,
	error,
) {
	repos, err := ghClient.Repositories.ListConnectedRepos(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list connected repositories: %v", err)},
			},
		}, GitHubListConnectedOutput{}, err
	}

	output := GitHubListConnectedOutput{
		Repos:   repos,
		Count:   len(repos),
		Message: fmt.Sprintf("Found %d repositories connected to Jules", len(repos)),
	}

	return nil, output, nil
}

// GitHubSearchReposInput represents input for github_search_repos tool
type GitHubSearchReposInput struct {
	Query string `json:"query" jsonschema:"Search query with optional GitHub search qualifiers"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum number of results to return (default: 30, max: 100)"`
	Sort  string `json:"sort,omitempty" jsonschema:"Sort results by: stars, forks, updated (default: best match)"`
	Order string `json:"order,omitempty" jsonschema:"Sort order: asc or desc (default: desc)"`
}

// GitHubSearchReposOutput represents output for github_search_repos tool
type GitHubSearchReposOutput struct {
	Repos   []*github.Repository `json:"repos"`
	Count   int                  `json:"count"`
	Query   string               `json:"query"`
	Message string               `json:"message"`
}

// searchRepositories searches for GitHub repositories
func searchRepositories(ctx context.Context, req *mcp.CallToolRequest, input GitHubSearchReposInput, ghClient *github.Client) (
	*mcp.CallToolResult,
	GitHubSearchReposOutput,
	error,
) {
	if input.Query == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Search query cannot be empty"},
			},
		}, GitHubSearchReposOutput{}, fmt.Errorf("search query is required")
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	// Prepare search options
	opts := &ghapi.SearchOptions{
		Sort:  input.Sort,
		Order: input.Order,
		ListOptions: ghapi.ListOptions{
			PerPage: limit,
		},
	}

	repos, err := ghClient.Repositories.SearchRepositories(ctx, input.Query, opts)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to search repositories: %v", err)},
			},
		}, GitHubSearchReposOutput{}, err
	}

	output := GitHubSearchReposOutput{
		Repos:   repos,
		Count:   len(repos),
		Query:   input.Query,
		Message: fmt.Sprintf("Found %d repositories matching '%s'", len(repos), input.Query),
	}

	return nil, output, nil
}
