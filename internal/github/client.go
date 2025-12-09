package github

import (
	"context"

	"github.com/SamyRai/juleson/internal/jules"
	"github.com/google/go-github/v76/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client with Jules integration
// It acts as a facade that delegates to specialized services following SOLID principles:
// - Single Responsibility: Each service handles one domain area
// - Open/Closed: New services can be added without modifying existing code
// - Liskov Substitution: Services can be mocked for testing
// - Interface Segregation: Each service exposes only relevant methods
// - Dependency Inversion: Services depend on abstractions (interfaces)
type Client struct {
	*github.Client
	token string

	// Specialized services - each responsible for a specific domain
	Repositories *RepositoryService
	Actions      *ActionsService
	PullRequests *PullRequestService
	Sessions     *SessionService
	Issues       *IssuesService
	Milestones   *MilestonesService
	Projects     *ProjectsService
}

// NewClient creates a new GitHub client with authentication and initializes all services
// This is the main entry point for GitHub operations
func NewClient(token string, julesClient *jules.Client) *Client {
	if token == "" {
		return nil
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	client := &Client{
		Client: github.NewClient(tc),
		token:  token,
	}

	// Initialize specialized services with proper dependency injection
	client.Repositories = NewRepositoryService(client, julesClient)
	client.Actions = NewActionsService(client)
	client.PullRequests = NewPullRequestService(client, julesClient)
	client.Sessions = NewSessionService(client, julesClient, client.Repositories)
	client.Issues = NewIssuesService(client)
	client.Milestones = NewMilestonesService(client)
	client.Projects = NewProjectsService(client)

	return client
}
