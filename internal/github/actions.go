package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v76/github"
)

// ActionsService handles GitHub Actions operations
type ActionsService struct {
	client *Client
}

// NewActionsService creates a new actions service
func NewActionsService(client *Client) *ActionsService {
	return &ActionsService{
		client: client,
	}
}

// Workflow operations

// ListWorkflows lists all workflows in a repository
func (s *ActionsService) ListWorkflows(ctx context.Context, owner, repo string) ([]*Workflow, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	workflows, _, err := s.client.Client.Actions.ListWorkflows(ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}

	var result []*Workflow
	for _, wf := range workflows.Workflows {
		result = append(result, s.mapWorkflow(wf))
	}

	return result, nil
}

// GetWorkflow gets a specific workflow by ID or filename
func (s *ActionsService) GetWorkflow(ctx context.Context, owner, repo, workflowIDOrFile string) (*Workflow, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	// Try as filename first
	wf, _, err := s.client.Client.Actions.GetWorkflowByFileName(ctx, owner, repo, workflowIDOrFile)
	if err != nil {
		// Try as ID
		workflowID := int64(parseInt(workflowIDOrFile))
		wf, _, err = s.client.Client.Actions.GetWorkflowByID(ctx, owner, repo, workflowID)
		if err != nil {
			return nil, fmt.Errorf("failed to get workflow: %w", err)
		}
	}

	return s.mapWorkflow(wf), nil
}

// TriggerWorkflow manually triggers a workflow dispatch event
func (s *ActionsService) TriggerWorkflow(ctx context.Context, owner, repo, workflowIDOrFile, ref string, inputs map[string]interface{}) error {
	if s.client == nil {
		return fmt.Errorf("GitHub client not configured")
	}

	event := github.CreateWorkflowDispatchEventRequest{
		Ref:    ref,
		Inputs: inputs,
	}

	// Try as filename first
	_, err := s.client.Client.Actions.CreateWorkflowDispatchEventByFileName(ctx, owner, repo, workflowIDOrFile, event)
	if err != nil {
		// Try as ID
		workflowID := int64(parseInt(workflowIDOrFile))
		_, err = s.client.Client.Actions.CreateWorkflowDispatchEventByID(ctx, owner, repo, workflowID, event)
		if err != nil {
			return fmt.Errorf("failed to trigger workflow: %w", err)
		}
	}

	return nil
}

// Workflow run operations

// ListWorkflowRuns lists workflow runs for a repository or specific workflow
func (s *ActionsService) ListWorkflowRuns(ctx context.Context, owner, repo string, workflowIDOrFile string, opts *github.ListWorkflowRunsOptions) ([]*WorkflowRun, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	var runs *github.WorkflowRuns
	var err error

	if workflowIDOrFile == "" {
		// List all workflow runs for repository
		runs, _, err = s.client.Client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, opts)
	} else {
		// Try as filename first
		runs, _, err = s.client.Client.Actions.ListWorkflowRunsByFileName(ctx, owner, repo, workflowIDOrFile, opts)
		if err != nil {
			// Try as ID
			workflowID := int64(parseInt(workflowIDOrFile))
			runs, _, err = s.client.Client.Actions.ListWorkflowRunsByID(ctx, owner, repo, workflowID, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list workflow runs: %w", err)
			}
		}
	}

	var result []*WorkflowRun
	for _, run := range runs.WorkflowRuns {
		result = append(result, s.mapWorkflowRun(run))
	}

	return result, nil
}

// GetWorkflowRun gets a specific workflow run by ID
func (s *ActionsService) GetWorkflowRun(ctx context.Context, owner, repo string, runID int64) (*WorkflowRun, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	run, _, err := s.client.Client.Actions.GetWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow run: %w", err)
	}

	return s.mapWorkflowRun(run), nil
}

// RerunWorkflow re-runs a workflow run
func (s *ActionsService) RerunWorkflow(ctx context.Context, owner, repo string, runID int64) error {
	if s.client == nil {
		return fmt.Errorf("GitHub client not configured")
	}

	_, err := s.client.Client.Actions.RerunWorkflowByID(ctx, owner, repo, runID)
	if err != nil {
		return fmt.Errorf("failed to rerun workflow: %w", err)
	}

	return nil
}

// RerunFailedJobs re-runs only failed jobs in a workflow run
func (s *ActionsService) RerunFailedJobs(ctx context.Context, owner, repo string, runID int64) error {
	if s.client == nil {
		return fmt.Errorf("GitHub client not configured")
	}

	_, err := s.client.Client.Actions.RerunFailedJobsByID(ctx, owner, repo, runID)
	if err != nil {
		return fmt.Errorf("failed to rerun failed jobs: %w", err)
	}

	return nil
}

// CancelWorkflow cancels a workflow run
func (s *ActionsService) CancelWorkflow(ctx context.Context, owner, repo string, runID int64) error {
	if s.client == nil {
		return fmt.Errorf("GitHub client not configured")
	}

	_, err := s.client.Client.Actions.CancelWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	return nil
}

// DownloadWorkflowLogs downloads logs for a workflow run
func (s *ActionsService) DownloadWorkflowLogs(ctx context.Context, owner, repo string, runID int64) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("GitHub client not configured")
	}

	// Get redirect URL for logs
	url, _, err := s.client.Client.Actions.GetWorkflowRunLogs(ctx, owner, repo, runID, 1)
	if err != nil {
		return "", fmt.Errorf("failed to get workflow logs URL: %w", err)
	}

	return url.String(), nil
}

// Job operations

// ListWorkflowJobs lists jobs for a workflow run
func (s *ActionsService) ListWorkflowJobs(ctx context.Context, owner, repo string, runID int64, filter string) ([]*WorkflowJob, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	opts := &github.ListWorkflowJobsOptions{
		Filter: filter,
	}

	jobs, _, err := s.client.Client.Actions.ListWorkflowJobs(ctx, owner, repo, runID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow jobs: %w", err)
	}

	var result []*WorkflowJob
	for _, job := range jobs.Jobs {
		result = append(result, s.mapWorkflowJob(job))
	}

	return result, nil
}

// GetWorkflowJob gets a specific job by ID
func (s *ActionsService) GetWorkflowJob(ctx context.Context, owner, repo string, jobID int64) (*WorkflowJob, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	job, _, err := s.client.Client.Actions.GetWorkflowJobByID(ctx, owner, repo, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow job: %w", err)
	}

	return s.mapWorkflowJob(job), nil
}

// RerunJob re-runs a specific job
func (s *ActionsService) RerunJob(ctx context.Context, owner, repo string, jobID int64) error {
	if s.client == nil {
		return fmt.Errorf("GitHub client not configured")
	}

	_, err := s.client.Client.Actions.RerunJobByID(ctx, owner, repo, jobID)
	if err != nil {
		return fmt.Errorf("failed to rerun job: %w", err)
	}

	return nil
}

// DownloadJobLogs downloads logs for a specific job
func (s *ActionsService) DownloadJobLogs(ctx context.Context, owner, repo string, jobID int64) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("GitHub client not configured")
	}

	// Get redirect URL for job logs
	url, _, err := s.client.Client.Actions.GetWorkflowJobLogs(ctx, owner, repo, jobID, 1)
	if err != nil {
		return "", fmt.Errorf("failed to get job logs URL: %w", err)
	}

	return url.String(), nil
}

// Artifact operations

// ListArtifacts lists artifacts for a repository or workflow run
func (s *ActionsService) ListArtifacts(ctx context.Context, owner, repo string, runID int64) ([]*github.Artifact, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	var artifacts *github.ArtifactList
	var err error

	if runID == 0 {
		// List all artifacts for repository
		artifacts, _, err = s.client.Client.Actions.ListArtifacts(ctx, owner, repo, nil)
	} else {
		// List artifacts for specific workflow run
		artifacts, _, err = s.client.Client.Actions.ListWorkflowRunArtifacts(ctx, owner, repo, runID, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}

	return artifacts.Artifacts, nil
}

// GetArtifact gets a specific artifact by ID
func (s *ActionsService) GetArtifact(ctx context.Context, owner, repo string, artifactID int64) (*github.Artifact, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	artifact, _, err := s.client.Client.Actions.GetArtifact(ctx, owner, repo, artifactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get artifact: %w", err)
	}

	return artifact, nil
}

// DownloadArtifact downloads an artifact and returns the download URL
func (s *ActionsService) DownloadArtifact(ctx context.Context, owner, repo string, artifactID int64) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("GitHub client not configured")
	}

	// Get redirect URL for artifact download
	url, _, err := s.client.Client.Actions.DownloadArtifact(ctx, owner, repo, artifactID, 1)
	if err != nil {
		return "", fmt.Errorf("failed to get artifact download URL: %w", err)
	}

	return url.String(), nil
}

// DeleteArtifact deletes an artifact
func (s *ActionsService) DeleteArtifact(ctx context.Context, owner, repo string, artifactID int64) error {
	if s.client == nil {
		return fmt.Errorf("GitHub client not configured")
	}

	_, err := s.client.Client.Actions.DeleteArtifact(ctx, owner, repo, artifactID)
	if err != nil {
		return fmt.Errorf("failed to delete artifact: %w", err)
	}

	return nil
}

// Cache operations

// ListCaches lists GitHub Actions caches for a repository
func (s *ActionsService) ListCaches(ctx context.Context, owner, repo string, ref *string) ([]*github.ActionsCache, error) {
	if s.client == nil {
		return nil, fmt.Errorf("GitHub client not configured")
	}

	opts := &github.ActionsCacheListOptions{}
	if ref != nil {
		opts.Ref = ref
	}

	caches, _, err := s.client.Client.Actions.ListCaches(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list caches: %w", err)
	}

	return caches.ActionsCaches, nil
}

// DeleteCachesByKey deletes caches by key
func (s *ActionsService) DeleteCachesByKey(ctx context.Context, owner, repo, key string, ref *string) error {
	if s.client == nil {
		return fmt.Errorf("GitHub client not configured")
	}

	_, err := s.client.Client.Actions.DeleteCachesByKey(ctx, owner, repo, key, ref)
	if err != nil {
		return fmt.Errorf("failed to delete caches: %w", err)
	}

	return nil
}

// DeleteCacheByID deletes a cache by ID
func (s *ActionsService) DeleteCacheByID(ctx context.Context, owner, repo string, cacheID int64) error {
	if s.client == nil {
		return fmt.Errorf("GitHub client not configured")
	}

	_, err := s.client.Client.Actions.DeleteCachesByID(ctx, owner, repo, cacheID)
	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

// Mapper functions

func (s *ActionsService) mapWorkflow(wf *github.Workflow) *Workflow {
	return &Workflow{
		ID:        wf.GetID(),
		Name:      wf.GetName(),
		Path:      wf.GetPath(),
		State:     wf.GetState(),
		CreatedAt: wf.GetCreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt: wf.GetUpdatedAt().Format("2006-01-02T15:04:05Z"),
		URL:       wf.GetHTMLURL(),
		BadgeURL:  wf.GetBadgeURL(),
	}
}

func (s *ActionsService) mapWorkflowRun(run *github.WorkflowRun) *WorkflowRun {
	return &WorkflowRun{
		ID:           run.GetID(),
		Name:         run.GetName(),
		HeadBranch:   run.GetHeadBranch(),
		Status:       run.GetStatus(),
		Conclusion:   run.GetConclusion(),
		WorkflowID:   run.GetWorkflowID(),
		URL:          run.GetHTMLURL(),
		CreatedAt:    run.GetCreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    run.GetUpdatedAt().Format("2006-01-02T15:04:05Z"),
		RunNumber:    run.GetRunNumber(),
		Event:        run.GetEvent(),
		Actor:        run.GetActor().GetLogin(),
		RunAttempt:   run.GetRunAttempt(),
		RunStartedAt: run.GetRunStartedAt().Format("2006-01-02T15:04:05Z"),
	}
}

func (s *ActionsService) mapWorkflowJob(job *github.WorkflowJob) *WorkflowJob {
	return &WorkflowJob{
		ID:          job.GetID(),
		RunID:       job.GetRunID(),
		Name:        job.GetName(),
		Status:      job.GetStatus(),
		Conclusion:  job.GetConclusion(),
		StartedAt:   job.GetStartedAt().Format("2006-01-02T15:04:05Z"),
		CompletedAt: job.GetCompletedAt().Format("2006-01-02T15:04:05Z"),
		URL:         job.GetHTMLURL(),
		RunnerName:  job.GetRunnerName(),
	}
}
