package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/jules"
	"github.com/google/go-github/v76/github"
)

// PullRequestService handles pull request operations
type PullRequestService struct {
	client      *Client
	julesClient *jules.Client
}

// NewPullRequestService creates a new pull request service
func NewPullRequestService(client *Client, julesClient *jules.Client) *PullRequestService {
	return &PullRequestService{
		client:      client,
		julesClient: julesClient,
	}
}

// GetSessionPullRequest retrieves the PR created by a Jules session
func (s *PullRequestService) GetSessionPullRequest(ctx context.Context, sessionID string) (*github.PullRequest, error) {
	if s.julesClient == nil {
		return nil, fmt.Errorf("Jules client not available")
	}

	session, err := s.julesClient.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Parse PR URL from session metadata
	if session.URL == "" {
		return nil, fmt.Errorf("session has no URL - PR may not be created yet")
	}

	// Extract owner/repo/PR# from URL
	owner, repo, prNumber, err := s.parsePRURL(session.URL)
	if err != nil {
		return nil, err
	}

	pr, _, err := s.client.Client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	return pr, nil
}

// MergePullRequest merges a PR created by Jules
func (s *PullRequestService) MergePullRequest(ctx context.Context, prURL string, mergeMethod string) error {
	// Parse PR URL to extract owner, repo, and PR number
	owner, repo, prNumber, err := s.parsePRURL(prURL)
	if err != nil {
		return err
	}

	// Default merge method
	if mergeMethod == "" {
		mergeMethod = "squash"
	}

	_, _, err = s.client.Client.PullRequests.Merge(ctx, owner, repo, prNumber, "", &github.PullRequestOptions{
		MergeMethod: mergeMethod,
	})

	if err != nil {
		return fmt.Errorf("failed to merge PR: %w", err)
	}

	return nil
}

// GetPullRequestDiff retrieves the diff for a PR created by a Jules session
func (s *PullRequestService) GetPullRequestDiff(ctx context.Context, sessionID string) (string, error) {
	if s.julesClient == nil {
		return "", fmt.Errorf("Jules client not available")
	}

	session, err := s.julesClient.GetSession(ctx, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	// Parse PR URL from session metadata
	if session.URL == "" {
		return "", fmt.Errorf("session has no URL - PR may not be created yet")
	}

	// Extract owner/repo/PR# from URL
	owner, repo, prNumber, err := s.parsePRURL(session.URL)
	if err != nil {
		return "", err
	}

	// Get the diff using GitHub API
	diff, _, err := s.client.Client.PullRequests.GetRaw(ctx, owner, repo, prNumber, github.RawOptions{Type: github.Diff})
	if err != nil {
		return "", fmt.Errorf("failed to get PR diff: %w", err)
	}

	return diff, nil
}

// parsePRURL parses a GitHub PR URL and extracts owner, repo, and PR number
// URL format: https://github.com/owner/repo/pull/123
func (s *PullRequestService) parsePRURL(prURL string) (owner, repo string, prNumber int, err error) {
	parts := strings.Split(prURL, "/")
	if len(parts) < 7 || parts[5] != "pull" {
		return "", "", 0, fmt.Errorf("invalid PR URL format: %s", prURL)
	}

	owner = parts[3]
	repo = parts[4]
	prNumber = parseInt(parts[6])

	return owner, repo, prNumber, nil
}
