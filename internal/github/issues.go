package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v76/github"
)

// IssuesService handles GitHub Issues operations
type IssuesService struct {
	client *Client
}

// NewIssuesService creates a new issues service
func NewIssuesService(client *Client) *IssuesService {
	return &IssuesService{
		client: client,
	}
}

// Issue represents a GitHub issue with relevant metadata.
type Issue struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	URL       string     `json:"url"`
	HTMLURL   string     `json:"html_url"`
	Milestone string     `json:"milestone,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
	Assignees []string   `json:"assignees"`
	Labels    []string   `json:"labels"`
}

// IssueCreateRequest represents parameters for creating a new issue
type IssueCreateRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Milestone int      `json:"milestone,omitempty"`
}

// IssueUpdateRequest represents parameters for updating an issue
type IssueUpdateRequest struct {
	Title     *string   `json:"title,omitempty"`
	Body      *string   `json:"body,omitempty"`
	State     *string   `json:"state,omitempty"`
	Assignees *[]string `json:"assignees,omitempty"`
	Labels    *[]string `json:"labels,omitempty"`
	Milestone *int      `json:"milestone,omitempty"`
}

// CreateIssue creates a new issue in a repository.
func (s *IssuesService) CreateIssue(ctx context.Context, owner, repo string, req *IssueCreateRequest) (*Issue, error) {
	issueReq := &github.IssueRequest{
		Title: &req.Title,
	}

	if req.Body != "" {
		issueReq.Body = &req.Body
	}

	if len(req.Assignees) > 0 {
		issueReq.Assignees = &req.Assignees
	}

	if len(req.Labels) > 0 {
		issueReq.Labels = &req.Labels
	}

	if req.Milestone > 0 {
		issueReq.Milestone = &req.Milestone
	}

	issue, _, err := s.client.Client.Issues.Create(ctx, owner, repo, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return s.mapGitHubIssue(issue), nil
}

// ListIssues lists issues for a repository
func (s *IssuesService) ListIssues(ctx context.Context, owner, repo string, state string, labels []string) ([]*Issue, error) {
	opts := &github.IssueListByRepoOptions{
		State: state,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	if len(labels) > 0 {
		opts.Labels = labels
	}

	issues, _, err := s.client.Client.Issues.ListByRepo(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	result := make([]*Issue, 0, len(issues))
	for _, issue := range issues {
		// Skip pull requests (they're also returned by Issues API)
		if issue.PullRequestLinks != nil {
			continue
		}
		result = append(result, s.mapGitHubIssue(issue))
	}

	return result, nil
}

// GetIssue retrieves a specific issue
func (s *IssuesService) GetIssue(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	issue, _, err := s.client.Client.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue #%d: %w", number, err)
	}

	return s.mapGitHubIssue(issue), nil
}

// UpdateIssue updates an existing issue
func (s *IssuesService) UpdateIssue(ctx context.Context, owner, repo string, number int, req IssueUpdateRequest) (*Issue, error) {
	issueReq := &github.IssueRequest{}

	if req.Title != nil {
		issueReq.Title = req.Title
	}

	if req.Body != nil {
		issueReq.Body = req.Body
	}

	if req.State != nil {
		issueReq.State = req.State
	}

	if req.Assignees != nil {
		issueReq.Assignees = req.Assignees
	}

	if req.Labels != nil {
		issueReq.Labels = req.Labels
	}

	if req.Milestone != nil {
		issueReq.Milestone = req.Milestone
	}

	issue, _, err := s.client.Client.Issues.Edit(ctx, owner, repo, number, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue #%d: %w", number, err)
	}

	return s.mapGitHubIssue(issue), nil
}

// CloseIssue closes an issue
func (s *IssuesService) CloseIssue(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	state := "closed"
	return s.UpdateIssue(ctx, owner, repo, number, IssueUpdateRequest{
		State: &state,
	})
}

// AddLabels adds labels to an issue
func (s *IssuesService) AddLabels(ctx context.Context, owner, repo string, number int, labels []string) error {
	_, _, err := s.client.Client.Issues.AddLabelsToIssue(ctx, owner, repo, number, labels)
	if err != nil {
		return fmt.Errorf("failed to add labels to issue #%d: %w", number, err)
	}
	return nil
}

// RemoveLabel removes a label from an issue
func (s *IssuesService) RemoveLabel(ctx context.Context, owner, repo string, number int, label string) error {
	_, err := s.client.Client.Issues.RemoveLabelForIssue(ctx, owner, repo, number, label)
	if err != nil {
		return fmt.Errorf("failed to remove label from issue #%d: %w", number, err)
	}
	return nil
}

// AddComment adds a comment to an issue
func (s *IssuesService) AddComment(ctx context.Context, owner, repo string, number int, body string) error {
	comment := &github.IssueComment{
		Body: &body,
	}

	_, _, err := s.client.Client.Issues.CreateComment(ctx, owner, repo, number, comment)
	if err != nil {
		return fmt.Errorf("failed to add comment to issue #%d: %w", number, err)
	}
	return nil
}

// AssignIssue assigns users to an issue
func (s *IssuesService) AssignIssue(ctx context.Context, owner, repo string, number int, assignees []string) error {
	_, _, err := s.client.Client.Issues.AddAssignees(ctx, owner, repo, number, assignees)
	if err != nil {
		return fmt.Errorf("failed to assign issue #%d: %w", number, err)
	}
	return nil
}

// mapGitHubIssue converts a GitHub API issue to our Issue type
func (s *IssuesService) mapGitHubIssue(ghIssue *github.Issue) *Issue {
	issue := &Issue{
		Number:    ghIssue.GetNumber(),
		Title:     ghIssue.GetTitle(),
		Body:      ghIssue.GetBody(),
		State:     ghIssue.GetState(),
		URL:       ghIssue.GetURL(),
		HTMLURL:   ghIssue.GetHTMLURL(),
		CreatedAt: ghIssue.GetCreatedAt().Time,
		UpdatedAt: ghIssue.GetUpdatedAt().Time,
	}

	if ghIssue.ClosedAt != nil {
		closedAt := ghIssue.GetClosedAt().Time
		issue.ClosedAt = &closedAt
	}

	// Map assignees
	if len(ghIssue.Assignees) > 0 {
		issue.Assignees = make([]string, len(ghIssue.Assignees))
		for i, assignee := range ghIssue.Assignees {
			issue.Assignees[i] = assignee.GetLogin()
		}
	}

	// Map labels
	if len(ghIssue.Labels) > 0 {
		issue.Labels = make([]string, len(ghIssue.Labels))
		for i, label := range ghIssue.Labels {
			issue.Labels[i] = label.GetName()
		}
	}

	// Map milestone
	if ghIssue.Milestone != nil {
		issue.Milestone = ghIssue.Milestone.GetTitle()
	}

	return issue
}
