package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/internal/sessionops"
)

func applySessionChanges(cfg *config.Config, sessionID, projectPath string, options ApplySessionOptions) error {
	julesClient := newJulesClient(cfg)
	ctx := context.Background()

	preparation, err := sessionops.PreparePatchApplication(ctx, sessionops.PatchRequest{
		WorkingDir:        projectPath,
		Confirm:           options.Confirm,
		AllowDirty:        options.AllowDirty,
		ActivityID:        options.ActivityID,
		ArtifactIndex:     options.ArtifactIndex,
		HasArtifactIndex:  options.HasArtifactIndex,
		AllowBaseMismatch: options.AllowBaseMismatch,
	})
	if err != nil {
		if preparation != nil && preparation.Blocker != "" {
			return fmt.Errorf("target worktree has local changes; commit/stash them or pass --allow-dirty\n%s", preparation.WorktreeStatus)
		}
		return err
	}
	patchOptions := preparation.Options
	changes, previewErr := julesops.PreviewSessionPatchesWithOptions(ctx, julesClient, sessionID, patchOptions)
	if changes != nil {
		printSessionChangesSummary(changes)
	}
	if preparation.DryRun {
		if previewErr != nil {
			return previewErr
		}
		fmt.Printf("\nDry-run only. Re-run with --confirm to apply patches.\n")
		return nil
	}
	if previewErr != nil {
		return fmt.Errorf("refusing to apply because preview failed: %w", previewErr)
	}

	result, err := julesops.ApplySessionPatches(ctx, julesClient, sessionID, patchOptions)
	if err != nil {
		return fmt.Errorf("failed to apply session patches: %w", err)
	}
	for _, warning := range result.Warnings {
		fmt.Printf("⚠️  %s\n", warning)
	}
	if len(result.Errors) > 0 {
		return fmt.Errorf("some patches failed: %s", strings.Join(result.Errors, "; "))
	}

	fmt.Printf("\n✅ Applied %d patch(es) touching %d file(s).\n", result.PatchesApplied, len(result.FilesModified))
	return nil
}

func printSessionChangesSummary(changes *julesops.SessionChanges) {
	totalAdded := 0
	totalRemoved := 0
	for _, file := range changes.Files {
		totalAdded += file.LinesAdded
		totalRemoved += file.LinesRemoved
	}
	fmt.Printf("Patch summary: %d patch(es), %d file(s), +%d -%d\n", changes.TotalPatches, len(changes.Files), totalAdded, totalRemoved)
	for _, file := range changes.Files {
		fmt.Printf("  %s (+%d -%d)\n", file.Path, file.LinesAdded, file.LinesRemoved)
	}
	for _, message := range changes.SuggestedCommitMessages {
		fmt.Printf("Suggested commit message: %s\n", message)
	}
	for _, warning := range changes.Warnings {
		fmt.Printf("Warning: %s\n", warning)
	}
}
