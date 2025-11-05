package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v76/github"
)

// ProjectsService handles GitHub Projects (v2) operations
// Note: Projects v2 uses GraphQL API for most operations
// This service provides helper methods for common use cases
type ProjectsService struct {
	client *Client
}

// NewProjectsService creates a new projects service
func NewProjectsService(client *Client) *ProjectsService {
	return &ProjectsService{
		client: client,
	}
}

// Project represents a GitHub Project (v2).
type Project struct {
	ID          int64      `json:"id"`
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	State       string     `json:"state"`
	URL         string     `json:"url"`
	HTMLURL     string     `json:"html_url"`
	Public      bool       `json:"public"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
}

// ProjectField represents a field in a GitHub Projects v2 project.
type ProjectField struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	DataType string   `json:"data_type"`
	Options  []string `json:"options,omitempty"`
}

// ListOrganizationProjects lists Projects v2 for an organization
func (s *ProjectsService) ListOrganizationProjects(ctx context.Context, org string) ([]*Project, error) {
	opts := &github.ListProjectsOptions{}

	projects, _, err := s.client.Client.Projects.ListProjectsForOrg(ctx, org, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization projects: %w", err)
	}

	result := make([]*Project, len(projects))
	for i, proj := range projects {
		result[i] = s.mapGitHubProjectV2(proj)
	}

	return result, nil
}

// GetOrganizationProject retrieves a specific organization project
func (s *ProjectsService) GetOrganizationProject(ctx context.Context, org string, projectNumber int) (*Project, error) {
	project, _, err := s.client.Client.Projects.GetProjectForOrg(ctx, org, projectNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization project #%d: %w", projectNumber, err)
	}

	return s.mapGitHubProjectV2(project), nil
}

// ListUserProjects lists Projects v2 for a user
func (s *ProjectsService) ListUserProjects(ctx context.Context, username string) ([]*Project, error) {
	opts := &github.ListProjectsOptions{}

	projects, _, err := s.client.Client.Projects.ListProjectsForUser(ctx, username, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list user projects: %w", err)
	}

	result := make([]*Project, len(projects))
	for i, proj := range projects {
		result[i] = s.mapGitHubProjectV2(proj)
	}

	return result, nil
}

// GetUserProject retrieves a specific user project
func (s *ProjectsService) GetUserProject(ctx context.Context, username string, projectNumber int) (*Project, error) {
	project, _, err := s.client.Client.Projects.GetProjectForUser(ctx, username, projectNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get user project #%d: %w", projectNumber, err)
	}

	return s.mapGitHubProjectV2(project), nil
}

// ListProjectFields lists fields for a project
func (s *ProjectsService) ListProjectFields(ctx context.Context, org string, projectNumber int) ([]*ProjectField, error) {
	opts := &github.ListProjectsOptions{}

	fields, _, err := s.client.Client.Projects.ListProjectFieldsForOrg(ctx, org, projectNumber, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list project fields: %w", err)
	}

	result := make([]*ProjectField, len(fields))
	for i, field := range fields {
		result[i] = s.mapGitHubProjectField(field)
	}

	return result, nil
}

// mapGitHubProjectV2 converts a GitHub API ProjectV2 to our Project type
func (s *ProjectsService) mapGitHubProjectV2(ghProject *github.ProjectV2) *Project {
	project := &Project{
		ID:          ghProject.GetID(),
		Number:      ghProject.GetNumber(),
		Title:       ghProject.GetTitle(),
		Description: ghProject.GetDescription(),
		URL:         ghProject.GetURL(),
		HTMLURL:     ghProject.GetHTMLURL(),
		Public:      ghProject.GetPublic(),
		CreatedAt:   ghProject.GetCreatedAt().Time,
		UpdatedAt:   ghProject.GetUpdatedAt().Time,
	}

	if ghProject.ClosedAt != nil {
		closedAt := ghProject.GetClosedAt().Time
		project.ClosedAt = &closedAt
	}

	// Map state
	if ghProject.GetPublic() {
		project.State = "open"
	} else {
		project.State = "private"
	}

	return project
}

// mapGitHubProjectField converts a GitHub API ProjectV2Field to our ProjectField type
func (s *ProjectsService) mapGitHubProjectField(ghField *github.ProjectV2Field) *ProjectField {
	field := &ProjectField{
		ID:       ghField.GetID(),
		Name:     ghField.Name,
		DataType: ghField.DataType,
	}

	return field
}
