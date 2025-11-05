package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v76/github"
)

// MilestonesService handles GitHub Milestones operations
type MilestonesService struct {
	client *Client
}

// NewMilestonesService creates a new milestones service
func NewMilestonesService(client *Client) *MilestonesService {
	return &MilestonesService{
		client: client,
	}
}

// Milestone represents a GitHub milestone.
type Milestone struct {
	Number       int        `json:"number"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	State        string     `json:"state"`
	URL          string     `json:"url"`
	HTMLURL      string     `json:"html_url"`
	OpenIssues   int        `json:"open_issues"`
	ClosedIssues int        `json:"closed_issues"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DueOn        *time.Time `json:"due_on,omitempty"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
}

// MilestoneCreateRequest represents parameters for creating a milestone
type MilestoneCreateRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	DueOn       *time.Time `json:"due_on,omitempty"`
	State       string     `json:"state,omitempty"`
}

// CreateMilestone creates a new milestone
func (s *MilestonesService) CreateMilestone(ctx context.Context, owner, repo string, req MilestoneCreateRequest) (*Milestone, error) {
	milestoneReq := &github.Milestone{
		Title: &req.Title,
	}

	if req.Description != "" {
		milestoneReq.Description = &req.Description
	}

	if req.DueOn != nil {
		dueOn := github.Timestamp{Time: *req.DueOn}
		milestoneReq.DueOn = &dueOn
	}

	if req.State != "" {
		milestoneReq.State = &req.State
	}

	milestone, _, err := s.client.Client.Issues.CreateMilestone(ctx, owner, repo, milestoneReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create milestone: %w", err)
	}

	return s.mapGitHubMilestone(milestone), nil
}

// ListMilestones lists milestones for a repository
func (s *MilestonesService) ListMilestones(ctx context.Context, owner, repo string, state string) ([]*Milestone, error) {
	opts := &github.MilestoneListOptions{
		State: state,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	milestones, _, err := s.client.Client.Issues.ListMilestones(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list milestones: %w", err)
	}

	result := make([]*Milestone, len(milestones))
	for i, milestone := range milestones {
		result[i] = s.mapGitHubMilestone(milestone)
	}

	return result, nil
}

// GetMilestone retrieves a specific milestone
func (s *MilestonesService) GetMilestone(ctx context.Context, owner, repo string, number int) (*Milestone, error) {
	milestone, _, err := s.client.Client.Issues.GetMilestone(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone #%d: %w", number, err)
	}

	return s.mapGitHubMilestone(milestone), nil
}

// UpdateMilestone updates an existing milestone
func (s *MilestonesService) UpdateMilestone(ctx context.Context, owner, repo string, number int, req MilestoneCreateRequest) (*Milestone, error) {
	milestoneReq := &github.Milestone{}

	if req.Title != "" {
		milestoneReq.Title = &req.Title
	}

	if req.Description != "" {
		milestoneReq.Description = &req.Description
	}

	if req.DueOn != nil {
		dueOn := github.Timestamp{Time: *req.DueOn}
		milestoneReq.DueOn = &dueOn
	}

	if req.State != "" {
		milestoneReq.State = &req.State
	}

	milestone, _, err := s.client.Client.Issues.EditMilestone(ctx, owner, repo, number, milestoneReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update milestone #%d: %w", number, err)
	}

	return s.mapGitHubMilestone(milestone), nil
}

// CloseMilestone closes a milestone
func (s *MilestonesService) CloseMilestone(ctx context.Context, owner, repo string, number int) (*Milestone, error) {
	state := "closed"
	return s.UpdateMilestone(ctx, owner, repo, number, MilestoneCreateRequest{
		State: state,
	})
}

// DeleteMilestone deletes a milestone
func (s *MilestonesService) DeleteMilestone(ctx context.Context, owner, repo string, number int) error {
	_, err := s.client.Client.Issues.DeleteMilestone(ctx, owner, repo, number)
	if err != nil {
		return fmt.Errorf("failed to delete milestone #%d: %w", number, err)
	}
	return nil
}

// mapGitHubMilestone converts a GitHub API milestone to our Milestone type
func (s *MilestonesService) mapGitHubMilestone(ghMilestone *github.Milestone) *Milestone {
	milestone := &Milestone{
		Number:       ghMilestone.GetNumber(),
		Title:        ghMilestone.GetTitle(),
		Description:  ghMilestone.GetDescription(),
		State:        ghMilestone.GetState(),
		URL:          ghMilestone.GetURL(),
		HTMLURL:      ghMilestone.GetHTMLURL(),
		OpenIssues:   ghMilestone.GetOpenIssues(),
		ClosedIssues: ghMilestone.GetClosedIssues(),
		CreatedAt:    ghMilestone.GetCreatedAt().Time,
		UpdatedAt:    ghMilestone.GetUpdatedAt().Time,
	}

	if ghMilestone.DueOn != nil {
		dueOn := ghMilestone.GetDueOn().Time
		milestone.DueOn = &dueOn
	}

	if ghMilestone.ClosedAt != nil {
		closedAt := ghMilestone.GetClosedAt().Time
		milestone.ClosedAt = &closedAt
	}

	return milestone
}
