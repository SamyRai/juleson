package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SamyRai/go-jules"
)

// PatchApplicationOptions represents options for applying patches.
type PatchApplicationOptions struct {
	WorkingDir        string
	ActivityID        string
	StripComponents   int
	ArtifactIndex     int
	DryRun            bool
	CreateBackup      bool
	Force             bool
	HasArtifactIndex  bool
	AllowBaseMismatch bool
}

// PatchApplicationResult represents the result of applying patches.
type PatchApplicationResult struct {
	ActivityID              string
	FilesModified           []string
	SuggestedCommitMessages []string
	Warnings                []string
	BaseCommitMismatches    []string
	Errors                  []string
	PatchesApplied          int
	PatchesFailed           int
	DryRun                  bool
}

// PatchService orchestrates fetching and applying patches from Jules.
type PatchService struct {
	client *jules.Client
}

// NewPatchService creates a new PatchService.
func NewPatchService(client *jules.Client) *PatchService {
	return &PatchService{client: client}
}

// ApplySessionPatches applies all git patches from a session's activities.
func ApplySessionPatches(ctx context.Context, client *jules.Client, sessionID string, options *PatchApplicationOptions) (*PatchApplicationResult, error) {
	svc := NewPatchService(client)
	return svc.ApplySessionPatches(ctx, sessionID, options)
}

func (s *PatchService) ApplySessionPatches(ctx context.Context, sessionID string, options *PatchApplicationOptions) (*PatchApplicationResult, error) {
	if options == nil {
		options = &PatchApplicationOptions{}
	}
	if options.WorkingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		options.WorkingDir = wd
	}
	if options.StripComponents == 0 {
		options.StripComponents = 1
	}

	gitClient := NewGitClient(options.WorkingDir)

	if options.ActivityID != "" {
		return s.applyActivityPatches(ctx, sessionID, options.ActivityID, options, gitClient)
	}

	response, err := s.client.Activities().List(ctx, sessionID, &jules.ListActivitiesOptions{PageSize: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	result := &PatchApplicationResult{
		DryRun: options.DryRun,
	}

	for _, activity := range response.Activities {
		activityResult, err := s.applyActivityPatches(ctx, sessionID, activity.ID, options, gitClient)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Activity %s: %v", activity.ID, err))
			continue
		}

		if activityResult != nil {
			result.PatchesApplied += activityResult.PatchesApplied
			result.PatchesFailed += activityResult.PatchesFailed
			result.FilesModified = append(result.FilesModified, activityResult.FilesModified...)
			result.SuggestedCommitMessages = appendUniqueStrings(result.SuggestedCommitMessages, activityResult.SuggestedCommitMessages...)
			result.Warnings = append(result.Warnings, activityResult.Warnings...)
			result.BaseCommitMismatches = append(result.BaseCommitMismatches, activityResult.BaseCommitMismatches...)
			result.Errors = append(result.Errors, activityResult.Errors...)
		}
	}

	return result, nil
}

// ApplyActivityPatches applies git patches from a specific activity.
func ApplyActivityPatches(ctx context.Context, client *jules.Client, sessionID, activityID string, options *PatchApplicationOptions) (*PatchApplicationResult, error) {
	svc := NewPatchService(client)
	if options == nil {
		options = &PatchApplicationOptions{}
	}
	if options.WorkingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		options.WorkingDir = wd
	}
	if options.StripComponents == 0 {
		options.StripComponents = 1
	}

	return svc.applyActivityPatches(ctx, sessionID, activityID, options, NewGitClient(options.WorkingDir))
}

func (s *PatchService) applyActivityPatches(ctx context.Context, sessionID, activityID string, options *PatchApplicationOptions, gitClient GitClient) (*PatchApplicationResult, error) {
	activity, err := s.client.Activities().Get(ctx, sessionID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	result := &PatchApplicationResult{
		ActivityID: activityID,
		DryRun:     options.DryRun,
	}

	for i, artifact := range activity.Artifacts {
		if options.HasArtifactIndex && i != options.ArtifactIndex {
			continue
		}
		if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
			gitPatch := artifact.ChangeSet.GitPatch
			patchContent := gitPatch.UnidiffPatch
			result.SuggestedCommitMessages = appendUniqueStrings(result.SuggestedCommitMessages, gitPatch.SuggestedCommitMessage)
			if patchContent == "" {
				result.Errors = append(result.Errors, fmt.Sprintf("Artifact %d: empty patch content", i))
				result.PatchesFailed++
				continue
			}

			if gitPatch.BaseCommitID != "" {
				mismatch, warning, err := s.checkBaseCommitMismatch(ctx, gitClient, gitPatch.BaseCommitID, i)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Artifact %d: %v", i, err))
					result.PatchesFailed++
					continue
				}
				if mismatch {
					result.Warnings = append(result.Warnings, warning)
					result.BaseCommitMismatches = append(result.BaseCommitMismatches, warning)
					if !options.DryRun && !options.AllowBaseMismatch {
						result.Errors = append(result.Errors, warning+"; pass --allow-base-mismatch to apply anyway")
						result.PatchesFailed++
						continue
					}
				}
			}

			files, err := s.applyGitPatch(ctx, patchContent, options, gitClient)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Artifact %d: %v", i, err))
				result.PatchesFailed++
				continue
			}

			result.PatchesApplied++
			result.FilesModified = append(result.FilesModified, files...)
		}
	}

	return result, nil
}

func (s *PatchService) checkBaseCommitMismatch(ctx context.Context, gitClient GitClient, baseCommitID string, artifactIndex int) (bool, string, error) {
	head, err := gitClient.GetHeadCommit(ctx)
	if err != nil {
		return false, "", err
	}
	base := strings.TrimSpace(baseCommitID)
	if head == "" || base == "" || strings.EqualFold(head, base) || strings.HasPrefix(head, base) || strings.HasPrefix(base, head) {
		return false, "", nil
	}
	return true, fmt.Sprintf("Artifact %d base commit %s does not match target HEAD %s", artifactIndex, base, head), nil
}

func (s *PatchService) applyGitPatch(ctx context.Context, patchContent string, options *PatchApplicationOptions, gitClient GitClient) ([]string, error) {
	tmpFile, err := os.CreateTemp("", "jules-patch-*.patch")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(patchContent); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write patch: %w", err)
	}
	tmpFile.Close()

	files, err := gitClient.ApplyPatch(ctx, tmpFile.Name(), options.DryRun, options.StripComponents, options.Force)
	if err != nil {
		return nil, err
	}

	if options.CreateBackup && !options.DryRun {
		for _, file := range files {
			filePath := filepath.Join(options.WorkingDir, file)
			backupPath := filePath + ".backup"
			if err := copyFile(filePath, backupPath); err != nil {
				return files, fmt.Errorf("failed to create backup for %s: %w", file, err)
			}
		}
	}

	return files, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}

// IsGitWorkingTreeClean reports whether a git working tree has no tracked or untracked changes.
func IsGitWorkingTreeClean(ctx context.Context, dir string) (bool, string, error) {
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return false, "", fmt.Errorf("failed to get working directory: %w", err)
		}
		dir = wd
	}
	gitClient := NewGitClient(dir)
	return gitClient.IsClean(ctx)
}

func GetSessionChanges(ctx context.Context, client *jules.Client, sessionID string) (*SessionChanges, error) {
	return GetSessionChangesWithOptions(ctx, client, sessionID, &PatchApplicationOptions{})
}

func GetSessionChangesWithOptions(ctx context.Context, client *jules.Client, sessionID string, options *PatchApplicationOptions) (*SessionChanges, error) {
	if options == nil {
		options = &PatchApplicationOptions{}
	}
	if options.ActivityID != "" {
		activity, err := client.Activities().Get(ctx, sessionID, options.ActivityID)
		if err != nil {
			return nil, fmt.Errorf("failed to get activity: %w", err)
		}
		return changesFromActivities(sessionID, []jules.Activity{*activity}, options), nil
	}

	response, err := client.Activities().List(ctx, sessionID, &jules.ListActivitiesOptions{PageSize: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	return changesFromActivities(sessionID, response.Activities, options), nil
}

func changesFromActivities(sessionID string, activities []jules.Activity, options *PatchApplicationOptions) *SessionChanges {
	changes := &SessionChanges{
		SessionID: sessionID,
	}

	for _, activity := range activities {
		for i, artifact := range activity.Artifacts {
			if options.HasArtifactIndex && i != options.ArtifactIndex {
				continue
			}
			if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
				changes.TotalPatches++
				changes.SuggestedCommitMessages = appendUniqueStrings(changes.SuggestedCommitMessages, artifact.ChangeSet.GitPatch.SuggestedCommitMessage)

				patch := artifact.ChangeSet.GitPatch.UnidiffPatch
				fileChanges := parsePatchFiles(patch)

				for _, fc := range fileChanges {
					found := false
					for i, existing := range changes.Files {
						if existing.Path == fc.Path {
							changes.Files[i].LinesAdded += fc.LinesAdded
							changes.Files[i].LinesRemoved += fc.LinesRemoved
							found = true
							break
						}
					}
					if !found {
						changes.Files = append(changes.Files, fc)
					}
				}
			}
		}
	}

	return changes
}

func PreviewSessionPatches(ctx context.Context, client *jules.Client, sessionID string, workingDir string) (*SessionChanges, error) {
	return PreviewSessionPatchesWithOptions(ctx, client, sessionID, &PatchApplicationOptions{WorkingDir: workingDir})
}

func PreviewSessionPatchesWithOptions(ctx context.Context, client *jules.Client, sessionID string, options *PatchApplicationOptions) (*SessionChanges, error) {
	if options == nil {
		options = &PatchApplicationOptions{}
	}
	if options.WorkingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		options.WorkingDir = wd
	}

	changes, err := GetSessionChangesWithOptions(ctx, client, sessionID, options)
	if err != nil {
		return nil, err
	}

	previewOptions := *options
	previewOptions.DryRun = true
	result, err := ApplySessionPatches(ctx, client, sessionID, &previewOptions)

	if err != nil {
		return changes, fmt.Errorf("preview failed: %w", err)
	}
	changes.Warnings = append(changes.Warnings, result.Warnings...)
	changes.BaseCommitMismatches = append(changes.BaseCommitMismatches, result.BaseCommitMismatches...)

	if len(result.Errors) > 0 {
		return changes, fmt.Errorf("some patches would fail to apply: %v", result.Errors)
	}

	return changes, nil
}
