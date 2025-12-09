# GitHub Integration Package

This package provides a well-structured, maintainable GitHub API integration following SOLID principles.

## Architecture Overview

The package is organized following **Single Responsibility Principle (SRP)** with each file handling a specific domain:

### Core Files

#### `client.go` - Main Client & Facade

- **Purpose**: Entry point and orchestrator for all GitHub operations
- **Pattern**: Facade pattern that delegates to specialized services
- **SOLID Principles**:
  - **Single Responsibility**: Only handles client initialization and service coordination
  - **Open/Closed**: New services can be added without modifying existing code
  - **Dependency Inversion**: Depends on service abstractions

```go
client := github.NewClient(token, julesClient)
workflows, err := client.Actions.ListWorkflows(ctx, owner, repo)
repos, err := client.Repositories.ListAccessibleRepos(ctx)
```

#### `types.go` - Domain Types

- **Purpose**: Centralized domain models
- **Contains**: `Repository`, `Workflow`, `WorkflowRun`, `WorkflowJob`
- **Benefit**: Single source of truth for data structures

### Services

Each service encapsulates a specific domain with clear responsibilities:

#### `actions.go` - ActionsService

- **Responsibility**: GitHub Actions operations (workflows, runs, jobs, artifacts, caches)
- **Methods**:
  - Workflow operations: `ListWorkflows`, `GetWorkflow`, `TriggerWorkflow`
  - Run management: `ListWorkflowRuns`, `GetWorkflowRun`, `RerunWorkflow`, `CancelWorkflow`
  - Job control: `ListWorkflowJobs`, `GetWorkflowJob`, `RerunJob`
  - Artifacts: `ListArtifacts`, `DownloadArtifact`, `DeleteArtifact`
  - Caches: `ListCaches`, `DeleteCachesByKey`, `DeleteCacheByID`

#### `repositories.go` - RepositoryService

- **Responsibility**: Repository discovery and management
- **Methods**:
  - `DiscoverCurrentRepo`: Detect repository from git remote
  - `ListConnectedRepos`: Fetch Jules-connected repositories
  - `ListAccessibleRepos`: List all accessible repositories
  - `SyncRepoWithJules`: Ensure repository is connected to Jules

#### `pullrequests.go` - PullRequestService

- **Responsibility**: Pull request operations
- **Methods**:
  - `GetSessionPullRequest`: Get PR for Jules session
  - `MergePullRequest`: Merge PR with specified method
  - `GetPullRequestDiff`: Retrieve PR diff

#### `sessions.go` - SessionService

- **Responsibility**: Jules session management with GitHub context
- **Methods**:
  - `CreateSessionFromRepo`: Create session for specific repository
  - `CreateSessionFromCurrentRepo`: Create session using git context

### Utilities

#### `git.go` - GitRemoteParser

- **Responsibility**: Git remote URL parsing
- **Methods**:
  - `GetRepoFromGitRemote`: Detect repository from git remote
  - `ParseGitHubURL`: Parse owner/repo from URLs (HTTPS & SSH)

#### `utils.go` - Helper Functions

- **Responsibility**: Shared utility functions
- **Contains**: `parseInt` and other common helpers

## Design Patterns

### 1. Facade Pattern

The `Client` struct acts as a facade, providing a simple interface to the complex GitHub API subsystem.

### 2. Service Layer Pattern

Each service (`ActionsService`, `RepositoryService`, etc.) encapsulates business logic for its domain.

### 3. Dependency Injection

Services receive dependencies through constructors, making them testable and loosely coupled.

```go
client.Repositories = NewRepositoryService(client, julesClient)
client.Actions = NewActionsService(client)
```

## SOLID Principles Applied

### Single Responsibility Principle (SRP)

- ✅ Each service handles one domain area
- ✅ `client.go` only orchestrates, doesn't implement business logic
- ✅ `git.go` only handles git operations
- ✅ `types.go` only defines data structures

### Open/Closed Principle

- ✅ New services can be added without modifying existing services
- ✅ New methods can be added to services without breaking clients

### Liskov Substitution Principle

- ✅ Services can be replaced with mocks for testing
- ✅ All services follow consistent patterns

### Interface Segregation Principle

- ✅ Each service exposes only relevant methods
- ✅ CLI commands use only the services they need

### Dependency Inversion Principle

- ✅ High-level Client depends on service abstractions
- ✅ Services depend on interfaces (jules.Client), not concrete implementations

## Usage Examples

### Actions

```go
// List workflows
workflows, err := client.Actions.ListWorkflows(ctx, "owner", "repo")

// Trigger workflow
err := client.Actions.TriggerWorkflow(ctx, "owner", "repo", "workflow.yml", "main", inputs)

// List workflow runs
runs, err := client.Actions.ListWorkflowRuns(ctx, "owner", "repo", "", nil)
```

### Repositories

```go
// Discover current repository
repo, err := client.Repositories.DiscoverCurrentRepo(ctx)

// List accessible repositories
repos, err := client.Repositories.ListAccessibleRepos(ctx)
```

### Pull Requests

```go
// Get PR for session
pr, err := client.PullRequests.GetSessionPullRequest(ctx, sessionID)

// Merge PR
err := client.PullRequests.MergePullRequest(ctx, prURL, "squash")
```

### Sessions

```go
// Create session from current repo
session, err := client.Sessions.CreateSessionFromCurrentRepo(ctx, prompt, branch)
```

## Benefits of This Architecture

1. **Maintainability**: Clear separation of concerns makes code easier to understand and modify
2. **Testability**: Services can be mocked and tested independently
3. **Scalability**: New features can be added without touching existing code
4. **Readability**: Intent is clear from the service name (e.g., `client.Actions.ListWorkflows`)
5. **Reusability**: Services can be reused across different parts of the application
6. **Type Safety**: Strong typing with domain models prevents errors

## Migration from Old Structure

Old (monolithic):

```go
workflows, err := client.ListWorkflows(ctx, owner, repo)
```

New (service-based):

```go
workflows, err := client.Actions.ListWorkflows(ctx, owner, repo)
```

All CLI commands have been updated to use the new structure.

## Testing

Each service can be tested independently:

```go
func TestActionsService_ListWorkflows(t *testing.T) {
    mockClient := &github.Client{...}
    service := NewActionsService(mockClient)

    workflows, err := service.ListWorkflows(ctx, "owner", "repo")
    // assertions...
}
```

## Future Enhancements

- [ ] Add interfaces for each service to enable better mocking
- [ ] Add comprehensive unit tests for each service
- [ ] Add integration tests
- [ ] Add caching layer for frequently accessed data
- [ ] Add rate limiting and retry logic
- [ ] Add metrics and logging
