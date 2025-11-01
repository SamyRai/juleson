package jules

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PatchApplicationOptions represents options for applying patches
type PatchApplicationOptions struct {
	WorkingDir      string // Working directory where patches should be applied (default: current directory)
	DryRun          bool   // Whether to perform a dry-run without actually applying changes
	CreateBackup    bool   // Whether to create backup files before applying patches
	Force           bool   // Whether to force application even if some hunks fail
	StripComponents int    // Number of leading path components to strip (default: 1 for git patches)
}

// PatchApplicationResult represents the result of applying patches
type PatchApplicationResult struct {
	ActivityID     string   // Activity from which patches were applied
	PatchesApplied int      // Number of patches successfully applied
	PatchesFailed  int      // Number of patches that failed to apply
	FilesModified  []string // List of files that were modified
	Errors         []string // List of errors encountered
	DryRun         bool     // Whether this was a dry-run
}

// ApplySessionPatches applies all git patches from a session's activities
func (c *Client) ApplySessionPatches(ctx context.Context, sessionID string, options *PatchApplicationOptions) (*PatchApplicationResult, error) {
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

	// Get all activities for the session
	activities, err := c.ListActivities(ctx, sessionID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	result := &PatchApplicationResult{
		DryRun: options.DryRun,
	}

	// Process each activity looking for patches
	for _, activity := range activities {
		activityResult, err := c.applyActivityPatches(ctx, sessionID, activity.ID, options)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Activity %s: %v", activity.ID, err))
			continue
		}

		if activityResult != nil {
			result.PatchesApplied += activityResult.PatchesApplied
			result.PatchesFailed += activityResult.PatchesFailed
			result.FilesModified = append(result.FilesModified, activityResult.FilesModified...)
		}
	}

	return result, nil
}

// ApplyActivityPatches applies git patches from a specific activity
func (c *Client) ApplyActivityPatches(ctx context.Context, sessionID, activityID string, options *PatchApplicationOptions) (*PatchApplicationResult, error) {
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

	return c.applyActivityPatches(ctx, sessionID, activityID, options)
}

// applyActivityPatches is the internal implementation for applying patches from an activity
func (c *Client) applyActivityPatches(ctx context.Context, sessionID, activityID string, options *PatchApplicationOptions) (*PatchApplicationResult, error) {
	// Get the activity to access its artifacts
	activity, err := c.GetActivity(ctx, sessionID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	result := &PatchApplicationResult{
		ActivityID: activityID,
		DryRun:     options.DryRun,
	}

	// Look for changeset artifacts with git patches
	for i, artifact := range activity.Artifacts {
		if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
			// Get the patch content
			patchContent := artifact.ChangeSet.GitPatch.UnidiffPatch
			if patchContent == "" {
				result.Errors = append(result.Errors, fmt.Sprintf("Artifact %d: empty patch content", i))
				result.PatchesFailed++
				continue
			}

			// Apply the patch
			files, err := c.applyGitPatch(ctx, patchContent, options)
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
func (c *Client) applyGitPatch(ctx context.Context, patchContent string, options *PatchApplicationOptions) ([]string, error) {
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

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// GetSessionChanges retrieves a summary of all changes in a session
func (c *Client) GetSessionChanges(ctx context.Context, sessionID string) (*SessionChanges, error) {
	activities, err := c.ListActivities(ctx, sessionID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	changes := &SessionChanges{
		SessionID: sessionID,
	}

	for _, activity := range activities {
		for _, artifact := range activity.Artifacts {
			if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
				changes.TotalPatches++

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

	return changes, nil
}

// SessionChanges represents a summary of changes in a session
type SessionChanges struct {
	SessionID    string       `json:"sessionId"`
	TotalPatches int          `json:"totalPatches"`
	Files        []FileChange `json:"files"`
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

			// Extract file path (use the b/ version for new path)
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				path := strings.TrimPrefix(parts[3], "b/")
				currentFile = &FileChange{Path: path}
			}
		} else if currentFile != nil {
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

// PreviewSessionPatches shows what would be changed if patches were applied
func (c *Client) PreviewSessionPatches(ctx context.Context, sessionID string, workingDir string) (*SessionChanges, error) {
	if workingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		workingDir = wd
	}

	// Get the changes summary
	changes, err := c.GetSessionChanges(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Test if patches can be applied (dry-run)
	result, err := c.ApplySessionPatches(ctx, sessionID, &PatchApplicationOptions{
		WorkingDir: workingDir,
		DryRun:     true,
	})

	if err != nil {
		return changes, fmt.Errorf("preview failed: %w", err)
	}

	// Add error information to the changes
	if len(result.Errors) > 0 {
		return changes, fmt.Errorf("some patches would fail to apply: %v", result.Errors)
	}

	return changes, nil
}
