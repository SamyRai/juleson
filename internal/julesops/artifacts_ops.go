package julesops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SamyRai/juleson/pkg/jules"
)

// ArtifactDownloadOptions represents options for downloading artifacts.
type ArtifactDownloadOptions struct {
	DestinationDir string // Directory to save artifacts (default: current directory)
	Overwrite      bool   // Whether to overwrite existing files
	CreateDir      bool   // Whether to create destination directory if it doesn't exist
}

// DownloadArtifactFromActivity downloads artifacts from a specific activity.
func DownloadArtifactFromActivity(ctx context.Context, client *jules.Client, sessionID, activityID string, options *ArtifactDownloadOptions) ([]string, error) {
	activity, err := client.GetActivity(ctx, sessionID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	var downloadedFiles []string
	for i, artifact := range activity.Artifacts {
		filename, err := downloadSingleArtifact(i, artifact, options)
		if err != nil {
			return downloadedFiles, fmt.Errorf("failed to download artifact %d: %w", i, err)
		}
		downloadedFiles = append(downloadedFiles, filename)
	}

	return downloadedFiles, nil
}

func downloadSingleArtifact(artifactIndex int, artifact jules.Artifact, options *ArtifactDownloadOptions) (string, error) {
	if options == nil {
		options = &ArtifactDownloadOptions{}
	}
	if options.DestinationDir == "" {
		options.DestinationDir = "."
	}

	if options.CreateDir {
		if err := os.MkdirAll(options.DestinationDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create destination directory: %w", err)
		}
	}

	filename := GenerateArtifactFilename(artifact, artifactIndex)
	filePath := filepath.Join(options.DestinationDir, filename)
	if !options.Overwrite {
		if _, err := os.Stat(filePath); err == nil {
			return "", fmt.Errorf("file already exists: %s", filePath)
		}
	}

	content, err := jules.ArtifactContent(artifact)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded artifact content: %w", err)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filename, nil
}

// GenerateArtifactFilename generates a filename for an artifact based on its type.
func GenerateArtifactFilename(artifact jules.Artifact, index int) string {
	if artifact.BashOutput != nil {
		return fmt.Sprintf("bash_output_%d.txt", index)
	}
	if artifact.ChangeSet != nil {
		if artifact.ChangeSet.GitPatch != nil {
			return fmt.Sprintf("changeset_%d.patch", index)
		}
		return fmt.Sprintf("changeset_%d.txt", index)
	}
	if artifact.Media != nil {
		ext := extensionFromMimeType(artifact.Media.MimeType)
		return fmt.Sprintf("media_%d%s", index, ext)
	}
	return fmt.Sprintf("artifact_%d.bin", index)
}

func extensionFromMimeType(mimeType string) string {
	switch {
	case strings.Contains(mimeType, "png"):
		return ".png"
	case strings.Contains(mimeType, "jpeg"), strings.Contains(mimeType, "jpg"):
		return ".jpg"
	case strings.Contains(mimeType, "gif"):
		return ".gif"
	case strings.Contains(mimeType, "json"):
		return ".json"
	case strings.Contains(mimeType, "text"):
		return ".txt"
	default:
		return ".bin"
	}
}

// DownloadAllSessionArtifacts downloads all artifacts from all activities in a session.
func DownloadAllSessionArtifacts(ctx context.Context, client *jules.Client, sessionID string, options *ArtifactDownloadOptions) ([]string, error) {
	activities, err := client.ListActivities(ctx, sessionID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	var allDownloadedFiles []string
	for _, activity := range activities {
		if len(activity.Artifacts) == 0 {
			continue
		}
		files, err := DownloadArtifactFromActivity(ctx, client, sessionID, activity.ID, options)
		if err != nil {
			return allDownloadedFiles, fmt.Errorf("failed to download artifacts for activity %s: %w", activity.ID, err)
		}
		allDownloadedFiles = append(allDownloadedFiles, files...)
	}

	return allDownloadedFiles, nil
}

// DownloadAllArtifacts downloads all artifacts from a session.
func DownloadAllArtifacts(ctx context.Context, client *jules.Client, sessionID string, options *ArtifactDownloadOptions) ([]string, error) {
	return DownloadAllSessionArtifacts(ctx, client, sessionID, options)
}

// SessionHasDeliverables reports whether a completed session exposes anything
// an operator can retrieve or act on through documented Jules API fields.
func SessionHasDeliverables(ctx context.Context, client *jules.Client, session *jules.Session) (bool, error) {
	if session == nil {
		return false, nil
	}
	for _, output := range session.Outputs {
		if output.PullRequest != nil && strings.TrimSpace(output.PullRequest.URL) != "" {
			return true, nil
		}
	}

	activities, err := client.ListAllActivities(ctx, session.ID, 100)
	if err != nil {
		return false, fmt.Errorf("failed to list activities for deliverable check: %w", err)
	}
	for _, activity := range activities {
		for _, artifact := range activity.Artifacts {
			if artifact.ChangeSet == nil || artifact.ChangeSet.GitPatch == nil {
				continue
			}
			if strings.TrimSpace(artifact.ChangeSet.GitPatch.UnidiffPatch) != "" {
				return true, nil
			}
		}
	}
	return false, nil
}
