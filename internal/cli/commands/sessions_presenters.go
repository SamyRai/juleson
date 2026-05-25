package commands

import (
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/pkg/jules"
)

// getSessionStatusIcon returns the appropriate icon for a session state
func getSessionStatusIcon(state jules.SessionState) string {
	switch state {
	case jules.SessionStateInProgress, jules.SessionStatePlanning, jules.SessionStateQueued:
		return "⚡"
	case jules.SessionStateAwaitingPlanApproval, jules.SessionStateAwaitingUserFeedback:
		return "⏸"
	case jules.SessionStateCompleted:
		return "✅"
	case jules.SessionStateFailed:
		return "❌"
	default:
		return "📋"
	}
}

// getSessionStatusText returns the status text for a session state
func getSessionStatusText(state jules.SessionState) string {
	switch state {
	case jules.SessionStateInProgress, jules.SessionStatePlanning, jules.SessionStateQueued:
		return "ACTIVE"
	case jules.SessionStateAwaitingPlanApproval, jules.SessionStateAwaitingUserFeedback:
		return "NEEDS_USER_ACTION"
	case jules.SessionStateCompleted:
		return "COMPLETED"
	case jules.SessionStateFailed:
		return "FAILED"
	default:
		return string(state)
	}
}

// previewActivityArtifactsContent displays artifact content based on type
func previewActivityArtifactsContent(artifacts []jules.Artifact) error {
	for i, artifact := range artifacts {
		fmt.Printf("\n  📄 Artifact %d:\n", i+1)

		// Handle different artifact types
		if artifact.BashOutput != nil {
			previewBashOutput(artifact.BashOutput)
		} else if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
			err := previewGitPatch(artifact.ChangeSet.GitPatch)
			if err != nil {
				fmt.Printf("    ⚠️  Failed to preview git patch: %v\n", err)
			}
		} else if artifact.Media != nil {
			previewMedia(artifact.Media)
		} else {
			fmt.Printf("    📄 Unknown artifact type\n")
		}
	}
	return nil
}

// previewBashOutput displays bash command output
func previewBashOutput(output *jules.BashOutput) error {
	fmt.Printf("    🖥️  Bash Output:\n")
	fmt.Printf("    Command: %s\n", output.Command)
	fmt.Printf("    Exit Code: %d\n", output.ExitCode)

	// Truncate output if too long
	content := output.Output
	if len(content) > 1000 {
		content = content[:1000] + "\n... (truncated)"
	}

	fmt.Printf("    Output:\n")
	fmt.Printf("    ```\n")
	for _, line := range strings.Split(content, "\n") {
		fmt.Printf("    %s\n", line)
	}
	fmt.Printf("    ```\n")
	return nil
}

// previewGitPatch displays git diff content
func previewGitPatch(patch *jules.GitPatch) error {
	fmt.Printf("    🔀 Git Patch:\n")

	if patch.SuggestedCommitMessage != "" {
		fmt.Printf("    Commit Message: %s\n", patch.SuggestedCommitMessage)
	}

	if patch.BaseCommitID != "" {
		fmt.Printf("    Base Commit: %s\n", patch.BaseCommitID)
	}

	// If we have unidiff content, display it
	if patch.UnidiffPatch != "" {
		fmt.Printf("    Diff:\n")
		fmt.Printf("    ```diff\n")

		// Split into lines and add proper indentation
		lines := strings.Split(patch.UnidiffPatch, "\n")
		for _, line := range lines {
			if len(line) > 120 { // Truncate very long lines
				line = line[:120] + "..."
			}
			fmt.Printf("    %s\n", line)
		}
		fmt.Printf("    ```\n")
	} else {
		fmt.Printf("    No diff content.\n")
	}

	return nil
}

// previewMedia displays media artifact information
func previewMedia(media *jules.Media) error {
	fmt.Printf("    🖼️  Media:\n")
	fmt.Printf("    Type: %s\n", media.MimeType)
	fmt.Printf("    Size: %d bytes\n", len(media.Data))

	// Don't display binary data, just metadata
	if strings.Contains(media.MimeType, "image/") {
		fmt.Printf("    📷 Image data (base64 encoded)\n")
	} else {
		fmt.Printf("    📄 Binary data\n")
	}

	return nil
}
