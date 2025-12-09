package github

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/jules"
)

// SessionService handles Jules session operations with GitHub integration
type SessionService struct {
	client      *Client
	julesClient *jules.Client
	repoService *RepositoryService
}

// NewSessionService creates a new session service
func NewSessionService(client *Client, julesClient *jules.Client, repoService *RepositoryService) *SessionService {
	return &SessionService{
		client:      client,
		julesClient: julesClient,
		repoService: repoService,
	}
}

// CreateSessionFromRepo creates a Jules session for a specific GitHub repository
func (s *SessionService) CreateSessionFromRepo(ctx context.Context, prompt, owner, repo, branch string) (*jules.Session, error) {
	if s.julesClient == nil {
		return nil, fmt.Errorf("Jules client not available")
	}

	// Ensure repository is connected to Jules
	err := s.repoService.SyncRepoWithJules(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to sync repo with Jules: %w", err)
	}

	sourceID := fmt.Sprintf("sources/github/%s/%s", owner, repo)

	// Use specified branch or get default branch
	if branch == "" {
		ghRepo, _, err := s.client.Client.Repositories.Get(ctx, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get repository info: %w", err)
		}
		branch = ghRepo.GetDefaultBranch()
	}

	session, err := s.julesClient.CreateSession(ctx, &jules.CreateSessionRequest{
		Prompt: prompt,
		SourceContext: &jules.SourceContext{
			Source: sourceID,
			GithubRepoContext: &jules.GithubRepoContext{
				StartingBranch: branch,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// CreateSessionFromCurrentRepo creates a Jules session using git context
func (s *SessionService) CreateSessionFromCurrentRepo(ctx context.Context, prompt string, branch string) (*jules.Session, error) {
	if s.julesClient == nil {
		return nil, fmt.Errorf("Jules client not available")
	}

	repo, err := s.repoService.DiscoverCurrentRepo(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure repository is connected to Jules
	err = s.repoService.SyncRepoWithJules(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to sync repo with Jules: %w", err)
	}

	sourceID := fmt.Sprintf("sources/github/%s/%s", repo.Owner, repo.Name)

	// Use specified branch or default
	if branch == "" {
		branch = repo.DefaultBranch
	}

	session, err := s.julesClient.CreateSession(ctx, &jules.CreateSessionRequest{
		Prompt: prompt,
		SourceContext: &jules.SourceContext{
			Source: sourceID,
			GithubRepoContext: &jules.GithubRepoContext{
				StartingBranch: branch,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}
