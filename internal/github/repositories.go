package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/jules"
	"github.com/google/go-github/v76/github"
)

// RepositoryService handles repository-related operations
type RepositoryService struct {
	client      *Client
	julesClient *jules.Client
	gitParser   *GitRemoteParser
}

// NewRepositoryService creates a new repository service
func NewRepositoryService(client *Client, julesClient *jules.Client) *RepositoryService {
	return &RepositoryService{
		client:      client,
		julesClient: julesClient,
		gitParser:   NewGitRemoteParser(),
	}
}

// DiscoverCurrentRepo detects the GitHub repository from the current git remote
func (s *RepositoryService) DiscoverCurrentRepo(ctx context.Context) (*Repository, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	// Get current directory's git remote
	repo, err := s.gitParser.GetRepoFromGitRemote()
	if err != nil {
		return nil, fmt.Errorf("failed to detect repository from git remote: %w", err)
	}

	// Verify repository exists and get metadata
	ghRepo, _, err := s.client.Client.Repositories.Get(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository info: %w", err)
	}

	return s.mapGitHubRepo(ghRepo), nil
}

// ListConnectedRepos fetches repositories connected to Jules
func (s *RepositoryService) ListConnectedRepos(ctx context.Context) ([]*Repository, error) {
	if s.julesClient == nil {
		return nil, fmt.Errorf("Jules client not available")
	}

	sources, err := s.julesClient.ListSources(ctx, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list Jules sources: %w", err)
	}

	var repos []*Repository
	for _, source := range sources {
		// Parse GitHub sources (format: sources/github/owner/repo)
		if strings.HasPrefix(source.Name, "sources/github/") {
			parts := strings.Split(source.Name, "/")
			if len(parts) >= 4 {
				owner := parts[2]
				name := parts[3]

				// Get repository metadata from GitHub
				ghRepo, _, err := s.client.Client.Repositories.Get(ctx, owner, name)
				if err != nil {
					// Skip repositories we can't access
					continue
				}

				repos = append(repos, s.mapGitHubRepo(ghRepo))
			}
		}
	}

	return repos, nil
}

// ListAccessibleRepos lists repositories the user can access
func (s *RepositoryService) ListAccessibleRepos(ctx context.Context) ([]*Repository, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	opts := &github.RepositoryListOptions{
		Sort:        "updated",
		Affiliation: "owner,collaborator,organization_member",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []*Repository
	for {
		repos, resp, err := s.client.Client.Repositories.List(ctx, "", opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		for _, repo := range repos {
			allRepos = append(allRepos, s.mapGitHubRepo(repo))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// SyncRepoWithJules ensures a repository is connected to Jules
func (s *RepositoryService) SyncRepoWithJules(ctx context.Context, owner, repo string) error {
	if s.julesClient == nil {
		return fmt.Errorf("Jules client not available")
	}

	sourceID := fmt.Sprintf("sources/github/%s/%s", owner, repo)

	// Check if source already exists
	sources, err := s.julesClient.ListSources(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to list sources: %w", err)
	}

	for _, source := range sources {
		if source.Name == sourceID {
			// Source already exists
			return nil
		}
	}

	// Note: Jules sources are typically connected via the web UI
	// This method currently only checks if a source exists
	return fmt.Errorf("source %s not found - please connect this repository via the Jules web UI first", sourceID)
}

// SearchRepositories searches for GitHub repositories using the GitHub Search API
func (s *RepositoryService) SearchRepositories(ctx context.Context, query string, opts *github.SearchOptions) ([]*Repository, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	if opts == nil {
		opts = &github.SearchOptions{}
	}

	// Set default pagination if not specified
	if opts.PerPage == 0 {
		opts.PerPage = 30 // GitHub's default for search
	}
	if opts.PerPage > 100 {
		opts.PerPage = 100 // GitHub's maximum
	}

	result, _, err := s.client.Client.Search.Repositories(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search repositories: %w", err)
	}

	var repos []*Repository
	for _, ghRepo := range result.Repositories {
		repos = append(repos, s.mapGitHubRepo(ghRepo))
	}

	return repos, nil
}

// mapGitHubRepo converts a github.Repository to our Repository type
func (s *RepositoryService) mapGitHubRepo(ghRepo *github.Repository) *Repository {
	return &Repository{
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
}
