package julesops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/SamyRai/go-jules"
)

// PatchApplicationOptions represents options for applying patches
type PatchApplicationOptions struct {
	WorkingDir        string // Working directory where patches should be applied (default: current directory)
	DryRun            bool   // Whether to perform a dry-run without actually applying changes
	CreateBackup      bool   // Whether to create backup files before applying patches
	Force             bool   // Whether to force application even if some hunks fail
	StripComponents   int    // Number of leading path components to strip (default: 1 for git patches)
	ActivityID        string // Optional activity ID/resource name to scope patch handling
	ArtifactIndex     int    // Optional artifact index to scope patch handling
	HasArtifactIndex  bool   // Whether ArtifactIndex should be applied as a filter
	AllowBaseMismatch bool   // Whether to allow actual apply when gitPatch.baseCommitId differs from HEAD
}

// PatchApplicationResult represents the result of applying patches
type PatchApplicationResult struct {
	ActivityID              string   // Activity from which patches were applied
	PatchesApplied          int      // Number of patches successfully applied
	PatchesFailed           int      // Number of patches that failed to apply
	FilesModified           []string // List of files that were modified
	SuggestedCommitMessages []string // Suggested commit messages from patch artifacts
	Warnings                []string // Non-fatal warnings encountered during preview/apply
	BaseCommitMismatches    []string // Base commit mismatch warnings
	Errors                  []string // List of errors encountered
	DryRun                  bool     // Whether this was a dry-run
}

// ApplySessionPatches applies all git patches from a session's activities
func ApplySessionPatches(ctx context.Context, client *jules.Client, sessionID string, options *PatchApplicationOptions) (*PatchApplicationResult, error) {
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
		options.StripComponents = 1 // Default for git patches
	}

	if options.ActivityID != "" {
		return ApplyActivityPatches(ctx, client, sessionID, options.ActivityID, options)
	}

	// Get all activities for the session
	response, err := client.Activities().List(ctx, sessionID, &jules.ListActivitiesOptions{PageSize: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	result := &PatchApplicationResult{
		DryRun: options.DryRun,
	}

	// Process each activity looking for patches
	for _, activity := range response.Activities {
		activityResult, err := applyActivityPatches(ctx, client, sessionID, activity.ID, options)
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

// ApplyActivityPatches applies git patches from a specific activity
func ApplyActivityPatches(ctx context.Context, client *jules.Client, sessionID, activityID string, options *PatchApplicationOptions) (*PatchApplicationResult, error) {
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

	return applyActivityPatches(ctx, client, sessionID, activityID, options)
}

// applyActivityPatches is the internal implementation for applying patches from an activity
func applyActivityPatches(ctx context.Context, client *jules.Client, sessionID, activityID string, options *PatchApplicationOptions) (*PatchApplicationResult, error) {
	// Get the activity to access its artifacts
	activity, err := client.Activities().Get(ctx, sessionID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	result := &PatchApplicationResult{
		ActivityID: activityID,
		DryRun:     options.DryRun,
	}

	// Look for changeset artifacts with git patches
	for i, artifact := range activity.Artifacts {
		if options.HasArtifactIndex && i != options.ArtifactIndex {
			continue
		}
		if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
			// Get the patch content
			gitPatch := artifact.ChangeSet.GitPatch
			patchContent := gitPatch.UnidiffPatch
			result.SuggestedCommitMessages = appendUniqueStrings(result.SuggestedCommitMessages, gitPatch.SuggestedCommitMessage)
			if patchContent == "" {
				result.Errors = append(result.Errors, fmt.Sprintf("Artifact %d: empty patch content", i))
				result.PatchesFailed++
				continue
			}

			if gitPatch.BaseCommitID != "" {
				mismatch, warning, err := checkBaseCommitMismatch(ctx, options.WorkingDir, gitPatch.BaseCommitID, i)
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

			// Apply the patch
			files, err := applyGitPatch(ctx, patchContent, options)
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

// applyGitPatch applies a single git patch using the git apply command
func applyGitPatch(ctx context.Context, patchContent string, options *PatchApplicationOptions) ([]string, error) {
	// Create a temporary file for the patch
	tmpFile, err := os.CreateTemp("", "jules-patch-*.patch")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write patch content to temp file
	if _, err := tmpFile.WriteString(patchContent); err != nil {
		return nil, fmt.Errorf("failed to write patch: %w", err)
	}
	tmpFile.Close()

	// Build git apply command
	args := []string{"apply"}

	// Add dry-run flag if requested
	if options.DryRun {
		args = append(args, "--check")
	}

	// Add strip components
	args = append(args, fmt.Sprintf("-p%d", options.StripComponents))

	// Add 3-way merge if force is enabled
	if options.Force {
		args = append(args, "--3way")
	}

	// Add verbose output
	args = append(args, "--verbose")

	// Add the patch file
	args = append(args, tmpFile.Name())

	// Execute git apply
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = options.WorkingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git apply failed: %w\nOutput: %s", err, string(output))
	}

	// Parse the output to extract modified files
	files := parseGitApplyOutput(string(output))

	// Create backups if requested and not a dry-run
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

// parseGitApplyOutput parses the output of git apply to extract modified files
func parseGitApplyOutput(output string) []string {
	var files []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines like "Checking patch file.txt..." or "Applying patch to file.txt..."
		if strings.HasPrefix(line, "Checking patch ") {
			file := strings.TrimPrefix(line, "Checking patch ")
			file = strings.TrimSuffix(file, "...")
			files = append(files, file)
		} else if strings.HasPrefix(line, "Applying patch to ") {
			file := strings.TrimPrefix(line, "Applying patch to ")
			file = strings.TrimSuffix(file, "...")
			files = append(files, file)
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, file := range files {
		if !seen[file] {
			seen[file] = true
			unique = append(unique, file)
		}
	}

	return unique
}

func checkBaseCommitMismatch(ctx context.Context, workingDir, baseCommitID string, artifactIndex int) (bool, string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", fmt.Errorf("failed to resolve target HEAD for base commit check: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}
	head := strings.TrimSpace(string(output))
	base := strings.TrimSpace(baseCommitID)
	if head == "" || base == "" || strings.EqualFold(head, base) || strings.HasPrefix(head, base) || strings.HasPrefix(base, head) {
		return false, "", nil
	}
	return true, fmt.Sprintf("Artifact %d base commit %s does not match target HEAD %s", artifactIndex, base, head), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// IsGitWorkingTreeClean reports whether a git working tree has no tracked or
// untracked changes. It returns an error when dir is not a git repository.
func IsGitWorkingTreeClean(ctx context.Context, dir string) (bool, string, error) {
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return false, "", fmt.Errorf("failed to get working directory: %w", err)
		}
		dir = wd
	}

	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	status := strings.TrimSpace(string(output))
	if err != nil {
		return false, status, fmt.Errorf("git status failed: %w\nOutput: %s", err, status)
	}
	return status == "", status, nil
}

func appendUniqueStrings(values []string, candidates ...string) []string {
	seen := make(map[string]bool, len(values)+len(candidates))
	for _, value := range values {
		seen[value] = true
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || seen[candidate] {
			continue
		}
		values = append(values, candidate)
		seen[candidate] = true
	}
	return values
}

// GetSessionChanges retrieves a summary of all changes in a session
func GetSessionChanges(ctx context.Context, client *jules.Client, sessionID string) (*SessionChanges, error) {
	return GetSessionChangesWithOptions(ctx, client, sessionID, &PatchApplicationOptions{})
}

// GetSessionChangesWithOptions retrieves a summary of changes in a session,
// optionally scoped to one activity or artifact index.
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

				// Parse the patch to extract file changes
				patch := artifact.ChangeSet.GitPatch.UnidiffPatch
				fileChanges := parsePatchFiles(patch)

				for _, fc := range fileChanges {
					// Check if we've seen this file before
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

// SessionChanges represents a summary of changes in a session
type SessionChanges struct {
	SessionID               string       `json:"sessionId"`
	TotalPatches            int          `json:"totalPatches"`
	Files                   []FileChange `json:"files"`
	SuggestedCommitMessages []string     `json:"suggestedCommitMessages,omitempty"`
	Warnings                []string     `json:"warnings,omitempty"`
	BaseCommitMismatches    []string     `json:"baseCommitMismatches,omitempty"`
}

// FileChange represents changes to a single file
type FileChange struct {
	Path         string `json:"path"`
	LinesAdded   int    `json:"linesAdded"`
	LinesRemoved int    `json:"linesRemoved"`
}

// parsePatchFiles extracts file changes from a git patch
func parsePatchFiles(patch string) []FileChange {
	var changes []FileChange
	lines := strings.Split(patch, "\n")

	var currentFile *FileChange

	for _, line := range lines {
		// Look for file headers like "diff --git a/file.txt b/file.txt"
		if strings.HasPrefix(line, "diff --git ") {
			if currentFile != nil {
				changes = append(changes, *currentFile)
			}

			currentFile = &FileChange{Path: extractDiffGitPath(line)}
		} else if currentFile != nil {
			if renamedPath, ok := strings.CutPrefix(line, "rename to "); ok {
				currentFile.Path = strings.TrimSpace(renamedPath)
				continue
			}
			// Count additions and deletions
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				currentFile.LinesAdded++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				currentFile.LinesRemoved++
			}
		}
	}

	// Add the last file
	if currentFile != nil {
		changes = append(changes, *currentFile)
	}

	return changes
}

func extractDiffGitPath(line string) string {
	remainder := strings.TrimSpace(strings.TrimPrefix(line, "diff --git "))
	first, rest := nextPatchToken(remainder)
	second, _ := nextPatchToken(rest)
	if second != "" && second != "/dev/null" {
		return stripPatchPrefix(second)
	}
	return stripPatchPrefix(first)
}

func nextPatchToken(value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}
	if value[0] == '"' {
		for i := 1; i < len(value); i++ {
			if value[i] == '"' && value[i-1] != '\\' {
				token := value[:i+1]
				unquoted, err := strconv.Unquote(token)
				if err != nil {
					unquoted = strings.Trim(token, `"`)
				}
				return unquoted, value[i+1:]
			}
		}
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func stripPatchPrefix(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "a/")
	path = strings.TrimPrefix(path, "b/")
	return path
}

// PreviewSessionPatches shows what would be changed if patches were applied
func PreviewSessionPatches(ctx context.Context, client *jules.Client, sessionID string, workingDir string) (*SessionChanges, error) {
	return PreviewSessionPatchesWithOptions(ctx, client, sessionID, &PatchApplicationOptions{WorkingDir: workingDir})
}

// PreviewSessionPatchesWithOptions shows what would be changed if patches were applied.
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

	// Get the changes summary
	changes, err := GetSessionChangesWithOptions(ctx, client, sessionID, options)
	if err != nil {
		return nil, err
	}

	// Test if patches can be applied (dry-run)
	previewOptions := *options
	previewOptions.DryRun = true
	result, err := ApplySessionPatches(ctx, client, sessionID, &previewOptions)

	if err != nil {
		return changes, fmt.Errorf("preview failed: %w", err)
	}
	changes.Warnings = append(changes.Warnings, result.Warnings...)
	changes.BaseCommitMismatches = append(changes.BaseCommitMismatches, result.BaseCommitMismatches...)

	// Add error information to the changes
	if len(result.Errors) > 0 {
		return changes, fmt.Errorf("some patches would fail to apply: %v", result.Errors)
	}

	return changes, nil
}
