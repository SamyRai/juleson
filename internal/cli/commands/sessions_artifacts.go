package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/internal/sessionops"
)

func listSessionArtifacts(cfg *config.Config, sessionID string) error {
	julesClient := newJulesClient(cfg)
	manifests, err := julesops.ListSessionArtifactManifests(context.Background(), julesClient, sessionID)
	if err != nil {
		return fmt.Errorf("failed to list session artifacts: %w", err)
	}
	if len(manifests) == 0 {
		fmt.Println("No artifacts found.")
		return nil
	}
	fmt.Printf("Artifacts for session %s\n", sessionID)
	fmt.Println(strings.Repeat("=", 60))
	for _, manifest := range manifests {
		fmt.Printf("Activity: %s  Index: %d  Type: %s\n", manifest.ActivityID, manifest.Index, manifest.Type)
		if !manifest.ActivityCreateTime.IsZero() {
			fmt.Printf("  Created: %s\n", manifest.ActivityCreateTime.Format(time.RFC3339))
		}
		if manifest.FileCount > 0 {
			fmt.Printf("  Files: %d\n", manifest.FileCount)
			for _, file := range manifest.Files {
				fmt.Printf("    %s (+%d -%d)\n", file.Path, file.LinesAdded, file.LinesRemoved)
			}
		} else if manifest.Empty {
			fmt.Printf("  Empty changeset: no diff content\n")
		}
		if manifest.BaseCommitID != "" {
			fmt.Printf("  Base commit: %s\n", manifest.BaseCommitID)
		}
		if manifest.SuggestedCommitMessage != "" {
			fmt.Printf("  Suggested commit: %s\n", manifest.SuggestedCommitMessage)
		}
		if manifest.MediaMIMEType != "" {
			fmt.Printf("  Media MIME: %s\n", manifest.MediaMIMEType)
		}
		if manifest.BashCommand != "" {
			fmt.Printf("  Bash command: %s\n", manifest.BashCommand)
		}
		if manifest.BashExitCode != nil {
			fmt.Printf("  Bash exit code: %d\n", *manifest.BashExitCode)
		}
		fmt.Println()
	}
	return nil
}

func showSessionOutputs(cfg *config.Config, sessionID string) error {
	julesClient := newJulesClient(cfg)
	session, err := julesClient.GetSession(context.Background(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if len(session.Outputs) == 0 {
		fmt.Println("No outputs found.")
		return nil
	}
	outputs := sessionops.DocumentedOutputs(session)
	fmt.Printf("Outputs for session %s\n", sessionID)
	fmt.Println(strings.Repeat("=", 60))
	if len(outputs) == 0 {
		fmt.Println("No supported documented output payloads found.")
		return nil
	}
	for i, output := range outputs {
		fmt.Printf("%d. ", i+1)
		fmt.Println("Pull Request")
		fmt.Printf("   URL: %s\n", output.PullRequest.URL)
		fmt.Printf("   Title: %s\n", output.PullRequest.Title)
		if output.PullRequest.Description != "" {
			fmt.Printf("   Description: %s\n", output.PullRequest.Description)
		}
	}
	return nil
}

// downloadSessionArtifacts downloads all artifacts from all activities in a session
func downloadSessionArtifacts(cfg *config.Config, sessionID string, outputDir string) error {
	julesClient := newJulesClient(cfg)
	ctx := context.Background()

	fmt.Printf("📥 Downloading artifacts from session: %s\n", sessionID)
	fmt.Printf("📁 Output directory: %s\n", outputDir)
	fmt.Println(strings.Repeat("=", 60))

	options := &julesops.ArtifactDownloadOptions{
		DestinationDir: outputDir,
		CreateDir:      true,
		Overwrite:      false,
	}

	downloadedFiles, err := julesops.DownloadAllSessionArtifacts(ctx, julesClient, sessionID, options)
	if err != nil {
		return fmt.Errorf("failed to download session artifacts: %w", err)
	}

	if len(downloadedFiles) == 0 {
		fmt.Println("📭 No artifacts found in this session.")
		return nil
	}

	fmt.Printf("✅ Successfully downloaded %d artifact(s):\n", len(downloadedFiles))
	for i, filename := range downloadedFiles {
		fmt.Printf("  %d. %s\n", i+1, filename)
	}

	fmt.Printf("\n💡 Artifacts saved to: %s\n", outputDir)
	return nil
}

// downloadActivityArtifacts downloads all artifacts from a specific activity
func downloadActivityArtifacts(cfg *config.Config, sessionID string, activityID string, outputDir string) error {
	julesClient := newJulesClient(cfg)
	ctx := context.Background()

	fmt.Printf("📥 Downloading artifacts from activity: %s\n", activityID)
	fmt.Printf("📁 Session: %s\n", sessionID)
	fmt.Printf("📁 Output directory: %s\n", outputDir)
	fmt.Println(strings.Repeat("=", 60))

	options := &julesops.ArtifactDownloadOptions{
		DestinationDir: outputDir,
		CreateDir:      true,
		Overwrite:      false,
	}

	downloadedFiles, err := julesops.DownloadArtifactFromActivity(ctx, julesClient, sessionID, activityID, options)
	if err != nil {
		return fmt.Errorf("failed to download activity artifacts: %w", err)
	}

	if len(downloadedFiles) == 0 {
		fmt.Println("📭 No artifacts found in this activity.")
		return nil
	}

	fmt.Printf("✅ Successfully downloaded %d artifact(s):\n", len(downloadedFiles))
	for i, filename := range downloadedFiles {
		fmt.Printf("  %d. %s\n", i+1, filename)
	}

	fmt.Printf("\n💡 Artifacts saved to: %s\n", outputDir)
	return nil
}

// previewSessionArtifacts previews all artifacts from all activities in a session
func previewSessionArtifacts(cfg *config.Config, sessionID string) error {
	julesClient := newJulesClient(cfg)
	ctx := context.Background()

	fmt.Printf("👁️  Previewing artifacts from session: %s\n", sessionID)
	fmt.Println(strings.Repeat("=", 60))

	activities, err := julesClient.ListActivities(ctx, sessionID, 100)
	if err != nil {
		return fmt.Errorf("failed to list activities: %w", err)
	}

	if len(activities) == 0 {
		fmt.Println("📭 No activities found in this session.")
		return nil
	}

	totalArtifacts := 0
	for i, activity := range activities {
		if len(activity.Artifacts) > 0 {
			fmt.Printf("\n📋 Activity %d: %s\n", i+1, activity.ID)
			err := previewActivityArtifactsContent(activity.Artifacts)
			if err != nil {
				fmt.Printf("⚠️  Failed to preview activity %s: %v\n", activity.ID, err)
			} else {
				totalArtifacts += len(activity.Artifacts)
			}
		}
	}

	if totalArtifacts == 0 {
		fmt.Println("📭 No artifacts found in this session.")
	} else {
		fmt.Printf("\n✅ Previewed %d artifact(s) total\n", totalArtifacts)
	}

	return nil
}

// previewActivityArtifacts previews all artifacts from a specific activity
func previewActivityArtifacts(cfg *config.Config, sessionID string, activityID string) error {
	julesClient := newJulesClient(cfg)
	ctx := context.Background()

	fmt.Printf("👁️  Previewing artifacts from activity: %s\n", activityID)
	fmt.Printf("📁 Session: %s\n", sessionID)
	fmt.Println(strings.Repeat("=", 60))

	activity, err := julesClient.GetActivity(ctx, sessionID, activityID)
	if err != nil {
		return fmt.Errorf("failed to get activity: %w", err)
	}

	if len(activity.Artifacts) == 0 {
		fmt.Println("📭 No artifacts found in this activity.")
		return nil
	}

	err = previewActivityArtifactsContent(activity.Artifacts)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Previewed %d artifact(s)\n", len(activity.Artifacts))
	return nil
}
