package sessions

import (
	"context"
	"fmt"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
	"strings"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/internal/presentation/tui/conflict"
	"github.com/SamyRai/juleson/internal/sessionops"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func applySessionChanges(cfg *config.Config, sessionID, projectPath string, options ApplySessionOptions) error {
	julesClient := core.NewJulesClient(cfg)
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
		fmt.Printf("\n⚠️  Some patches failed to apply: %s\n", strings.Join(result.Errors, "; "))

		// Prompt the user to resolve conflict agentically
		var resolve bool
		err := huh.NewConfirm().
			Title("Would you like to resolve these conflicts with Jules?").
			Value(&resolve).
			Run()

		if err == nil && resolve {
			return resolveConflictAgentically(ctx, julesClient, sessionID, projectPath, patchOptions)
		}

		return fmt.Errorf("some patches failed")
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

func resolveConflictAgentically(ctx context.Context, client *jules.Client, sessionID, projectPath string, patchOptions *julesops.PatchApplicationOptions) error {
	// For simplicity, we get the last patch details to send
	changes, err := julesops.GetSessionChangesWithOptions(ctx, client, sessionID, patchOptions)
	if err != nil || changes == nil || len(changes.Files) == 0 {
		return fmt.Errorf("could not gather patch details for resolution: %v", err)
	}

	// Launch TUI
	filename := changes.Files[0].Path
	opts, err := conflict.RunWizard(filename)
	if err != nil {
		return fmt.Errorf("conflict resolution cancelled: %w", err)
	}

	fmt.Println("Gathering context...")

	// We need the raw patch content which is tricky to get easily here, so we re-fetch the activity/activities
	var rawPatch string
	if patchOptions.ActivityID != "" {
		activity, err := client.Activities().Get(ctx, sessionID, patchOptions.ActivityID)
		if err == nil {
			for _, artifact := range activity.Artifacts {
				if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
					rawPatch = artifact.ChangeSet.GitPatch.UnidiffPatch
					break
				}
			}
		}
	} else {
		// Fetch all activities and try to find the last patch
		response, err := client.Activities().List(ctx, sessionID, &jules.ListActivitiesOptions{PageSize: 100})
		if err == nil {
			for i := len(response.Activities) - 1; i >= 0; i-- {
				for _, artifact := range response.Activities[i].Artifacts {
					if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
						rawPatch = artifact.ChangeSet.GitPatch.UnidiffPatch
						break
					}
				}
				if rawPatch != "" {
					break
				}
			}
		}
	}

	payload, err := conflict.BuildContextPayload(ctx, projectPath, filename, rawPatch, opts)
	if err != nil {
		return fmt.Errorf("failed to build context payload: %w", err)
	}

	// Dispatch context to Jules via SendMessage
	req := &jules.SendMessageRequest{
		Prompt: payload,
	}

	err = client.Sessions().SendMessage(ctx, sessionID, req)
	if err != nil {
		return fmt.Errorf("failed to send resolution request to agent: %w", err)
	}

	// Launch async waiting UI
	p := tea.NewProgram(conflict.InitialSpinnerModel("Waiting for Jules to resolve conflict..."))

	// Run spinner in a goroutine while we wait
	go func() {
		// Wait for the next relevant activity
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				p.Send(conflict.QuitMsg{})
				return
			case <-ticker.C:
				snapshot, watchErr := sessionops.CurrentWatchSnapshot(ctx, client, sessionID, time.Time{}, sessionops.CurrentWatchOptions{
					FetchActivities: true,
				})
				if watchErr != nil {
					continue // Ignore transient errors and keep polling
				}

				// Very simple wait logic: stop if we see Jules agent replied or state completes
				if snapshot.HasJulesAgentMessage || snapshot.Decision.Kind == sessionops.WatchDecisionCompletedWithDeliverables || snapshot.Decision.Kind == sessionops.WatchDecisionCompletedNoDeliverables || snapshot.Decision.Kind == sessionops.WatchDecisionFailed {
					p.Send(conflict.QuitMsg{})
					return
				}
			}
		}
	}()

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("spinner error: %w", err)
	}

	return handleResolutionResponse(ctx, client, sessionID)
}

func handleResolutionResponse(ctx context.Context, client *jules.Client, sessionID string) error {
	// Parse the final patch and resolution_report.md
	fmt.Println("\n✅ Agent finished processing.")

	// Fetch the latest activities to grab the new patch and report
	response, err := client.Activities().List(ctx, sessionID, &jules.ListActivitiesOptions{PageSize: 10})
	if err != nil || len(response.Activities) == 0 {
		return fmt.Errorf("could not fetch new activities: %v", err)
	}

	latestActivity := response.Activities[0]
	var reportContent string
	var hasNewPatch bool

	for _, artifact := range latestActivity.Artifacts {
		if artifact.Media != nil && strings.HasSuffix(artifact.Media.MimeType, "markdown") {
			reportContent = artifact.Media.Data
		} else if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
			hasNewPatch = true
		}
	}

	if reportContent != "" {
		fmt.Printf("\n--- Resolution Report ---\n%s\n-------------------------\n", reportContent)
	} else {
		fmt.Println("No resolution report artifact found.")
	}

	if hasNewPatch {
		fmt.Println("A new patch was created! Use 'juleson sessions apply <session_id> <project_path>' to preview it.")
	} else {
		fmt.Println("No new patch was found in the latest activity.")
	}

	return nil
}
