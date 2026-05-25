package tools

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/internal/sessionops"
	"github.com/SamyRai/juleson/pkg/jules"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// applySessionPatches applies git patches from a session to the working directory
func applySessionPatches(ctx context.Context, req *mcp.CallToolRequest, input ApplySessionPatchesInput, client *jules.Client) (
	*mcp.CallToolResult,
	ApplySessionPatchesOutput,
	error,
) {
	artifactIndex := 0
	hasArtifactIndex := false
	if input.ArtifactIndex != nil {
		artifactIndex = *input.ArtifactIndex
		hasArtifactIndex = true
	}
	preparation, err := sessionops.PreparePatchApplication(ctx, sessionops.PatchRequest{
		WorkingDir:        input.WorkingDir,
		DryRun:            input.DryRun,
		Confirm:           input.ConfirmApply,
		AllowDirty:        input.AllowDirty,
		Force:             input.Force,
		CreateBackup:      input.CreateBackup,
		ActivityID:        input.ActivityID,
		ArtifactIndex:     artifactIndex,
		HasArtifactIndex:  hasArtifactIndex,
		AllowBaseMismatch: input.AllowBaseMismatch,
	})
	if err != nil {
		if preparation != nil && preparation.Blocker != "" {
			return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: preparation.Blocker},
					},
				}, ApplySessionPatchesOutput{
					SessionID: input.SessionID,
					DryRun:    true,
					Blockers:  []string{preparation.Blocker},
					Message:   "Refusing to apply patches to a dirty working tree",
				}, err
		}
		return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to inspect working tree: %v", err)},
				},
			}, ApplySessionPatchesOutput{
				SessionID: input.SessionID,
				DryRun:    true,
				Blockers:  []string{err.Error()},
				Message:   "Refusing to apply patches because working tree status could not be checked",
			}, err
	}

	result, err := julesops.ApplySessionPatches(ctx, client, input.SessionID, preparation.Options)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to apply session patches: %v", err)},
			},
		}, ApplySessionPatchesOutput{}, err
	}

	message := fmt.Sprintf("Successfully applied %d patches", result.PatchesApplied)
	if result.DryRun {
		message = fmt.Sprintf("Dry-run: %d patches can be applied", result.PatchesApplied)
	}
	if result.PatchesFailed > 0 {
		message += fmt.Sprintf(", %d patches failed", result.PatchesFailed)
	}

	output := ApplySessionPatchesOutput{
		SessionID:               input.SessionID,
		PatchesApplied:          result.PatchesApplied,
		PatchesFailed:           result.PatchesFailed,
		FilesModified:           result.FilesModified,
		SuggestedCommitMessages: result.SuggestedCommitMessages,
		Warnings:                result.Warnings,
		BaseCommitMismatches:    result.BaseCommitMismatches,
		Errors:                  result.Errors,
		DryRun:                  result.DryRun,
		Message:                 message,
	}

	return nil, output, nil
}

// previewSessionChanges previews what changes would be made if patches were applied
func previewSessionChanges(ctx context.Context, req *mcp.CallToolRequest, input PreviewSessionChangesInput, client *jules.Client) (
	*mcp.CallToolResult,
	PreviewSessionChangesOutput,
	error,
) {
	options := &julesops.PatchApplicationOptions{
		WorkingDir: input.WorkingDir,
		ActivityID: input.ActivityID,
	}
	if input.ArtifactIndex != nil {
		options.ArtifactIndex = *input.ArtifactIndex
		options.HasArtifactIndex = true
	}
	changes, err := julesops.PreviewSessionPatchesWithOptions(ctx, client, input.SessionID, options)

	canApply := true
	var errors []string
	if err != nil {
		canApply = false
		errors = append(errors, err.Error())
	}

	if changes == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to preview session changes: %v", err)},
			},
		}, PreviewSessionChangesOutput{}, err
	}

	summary, _, _ := sessionops.SessionChangesSummary(changes)

	if !canApply {
		summary += " - WARNING: Some patches may fail to apply"
	}

	output := PreviewSessionChangesOutput{
		SessionID:               input.SessionID,
		TotalPatches:            changes.TotalPatches,
		Files:                   changes.Files,
		SuggestedCommitMessages: changes.SuggestedCommitMessages,
		Warnings:                changes.Warnings,
		BaseCommitMismatches:    changes.BaseCommitMismatches,
		CanApply:                canApply,
		Errors:                  errors,
		Summary:                 summary,
	}

	return nil, output, nil
}
