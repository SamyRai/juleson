package jules

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ArtifactDownloadOptions represents options for downloading artifacts
type ArtifactDownloadOptions struct {
	DestinationDir string // Directory to save artifacts (default: current directory)
	Overwrite      bool   // Whether to overwrite existing files
	CreateDir      bool   // Whether to create destination directory if it doesn't exist
}

// DownloadArtifactFromActivity downloads artifacts from a specific activity
func (c *Client) DownloadArtifactFromActivity(ctx context.Context, sessionID, activityID string, options *ArtifactDownloadOptions) ([]string, error) {
	// Get the activity to access its artifacts
	activity, err := c.GetActivity(ctx, sessionID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	var downloadedFiles []string

	for i, artifact := range activity.Artifacts {
		filename, err := c.downloadSingleArtifact(ctx, sessionID, activityID, i, artifact, options)
		if err != nil {
			return downloadedFiles, fmt.Errorf("failed to download artifact %d: %w", i, err)
		}
		downloadedFiles = append(downloadedFiles, filename)
	}

	return downloadedFiles, nil
}

// downloadSingleArtifact downloads a single artifact from an activity
func (c *Client) downloadSingleArtifact(ctx context.Context, sessionID, activityID string, artifactIndex int, artifact Artifact, options *ArtifactDownloadOptions) (string, error) {
	// Set default options
	if options == nil {
		options = &ArtifactDownloadOptions{}
	}
	if options.DestinationDir == "" {
		options.DestinationDir = "."
	}

	// Create destination directory if needed
	if options.CreateDir {
		if err := os.MkdirAll(options.DestinationDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create destination directory: %w", err)
		}
	}

	// Build download URL for activity artifact
	downloadURL := fmt.Sprintf("%s/sessions/%s/activities/%s/artifacts/%d/download", c.BaseURL, sessionID, activityID, artifactIndex)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Accept", "*/*")

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download artifact: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Determine filename
	filename := c.generateArtifactFilename(artifact, artifactIndex)

	// Build full file path
	filePath := filepath.Join(options.DestinationDir, filename)

	// Check if file exists and handle overwrite
	if !options.Overwrite {
		if _, err := os.Stat(filePath); err == nil {
			return "", fmt.Errorf("file already exists: %s", filePath)
		}
	}

	// Create destination file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filename, nil
}

// generateArtifactFilename generates a filename for an artifact based on its type
func (c *Client) generateArtifactFilename(artifact Artifact, index int) string {
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
		ext := c.getExtensionFromMimeType(artifact.Media.MimeType)
		return fmt.Sprintf("media_%d%s", index, ext)
	}
	return fmt.Sprintf("artifact_%d.bin", index)
}

// getExtensionFromMimeType returns file extension based on MIME type
func (c *Client) getExtensionFromMimeType(mimeType string) string {
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

// DownloadAllSessionArtifacts downloads all artifacts from all activities in a session
func (c *Client) DownloadAllSessionArtifacts(ctx context.Context, sessionID string, options *ArtifactDownloadOptions) ([]string, error) {
	// Get all activities for the session
	activities, err := c.ListActivities(ctx, sessionID, 100) // Get up to 100 activities
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	var allDownloadedFiles []string

	for _, activity := range activities {
		if len(activity.Artifacts) > 0 {
			files, err := c.DownloadArtifactFromActivity(ctx, sessionID, activity.ID, options)
			if err != nil {
				return allDownloadedFiles, fmt.Errorf("failed to download artifacts for activity %s: %w", activity.ID, err)
			}
			allDownloadedFiles = append(allDownloadedFiles, files...)
		}
	}

	return allDownloadedFiles, nil
}

// GetArtifactsFromActivity retrieves all artifacts from a specific activity
func (c *Client) GetArtifactsFromActivity(ctx context.Context, sessionID, activityID string) ([]Artifact, error) {
	activity, err := c.GetActivity(ctx, sessionID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return activity.Artifacts, nil
}

// GetAllSessionArtifacts retrieves all artifacts from all activities in a session
func (c *Client) GetAllSessionArtifacts(ctx context.Context, sessionID string) ([]ActivityArtifact, error) {
	activities, err := c.ListActivities(ctx, sessionID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	var allArtifacts []ActivityArtifact

	for _, activity := range activities {
		for i, artifact := range activity.Artifacts {
			allArtifacts = append(allArtifacts, ActivityArtifact{
				ActivityID: activity.ID,
				Index:      i,
				Artifact:   artifact,
			})
		}
	}

	return allArtifacts, nil
}

// ActivityArtifact represents an artifact with its activity context
type ActivityArtifact struct {
	ActivityID string
	Index      int
	Artifact   Artifact
}

// AnalyzeArtifact analyzes an artifact's content and structure
func (c *Client) AnalyzeArtifact(ctx context.Context, sessionID, activityID string, artifactIndex int) (*ArtifactAnalysis, error) {
	url := fmt.Sprintf("%s/sessions/%s/activities/%s/artifacts/%d/analyze", c.BaseURL, sessionID, activityID, artifactIndex)

	var analysis ArtifactAnalysis
	if err := c.doRequestWithJSON(ctx, "GET", url, nil, &analysis); err != nil {
		return nil, fmt.Errorf("failed to analyze artifact: %w", err)
	}

	return &analysis, nil
}

// ArtifactAnalysis represents the analysis result of an artifact
type ArtifactAnalysis struct {
	ActivityID    string                 `json:"activityId"`
	ArtifactIndex int                    `json:"artifactIndex"`
	ContentType   string                 `json:"contentType"`
	Size          int64                  `json:"size"`
	LineCount     int                    `json:"lineCount,omitempty"`
	Language      string                 `json:"language,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Summary       string                 `json:"summary,omitempty"`
	KeyInsights   []string               `json:"keyInsights,omitempty"`
	Issues        []ArtifactIssue        `json:"issues,omitempty"`
}

// ArtifactIssue represents an issue found in an artifact
type ArtifactIssue struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Line       int    `json:"line,omitempty"`
	Column     int    `json:"column,omitempty"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// GetArtifactContent retrieves the raw content of an artifact from an activity
func (c *Client) GetArtifactContent(ctx context.Context, sessionID, activityID string, artifactIndex int) ([]byte, error) {
	url := fmt.Sprintf("%s/sessions/%s/activities/%s/artifacts/%d/content", c.BaseURL, sessionID, activityID, artifactIndex)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create content request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Accept", "*/*")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get artifact content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("content request failed with status %d: %s", resp.StatusCode, string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact content: %w", err)
	}

	return content, nil
}

// DownloadAllArtifacts downloads all artifacts from a session
func (c *Client) DownloadAllArtifacts(ctx context.Context, sessionID string, options *ArtifactDownloadOptions) ([]string, error) {
	return c.DownloadAllSessionArtifacts(ctx, sessionID, options)
}

// GetArtifact retrieves artifact metadata by ID (deprecated - use GetArtifactsFromActivity)
func (c *Client) GetArtifact(ctx context.Context, sessionID, artifactID string) (*Artifact, error) {
	return nil, fmt.Errorf("deprecated: artifacts are accessed through activities, use GetArtifactsFromActivity instead")
}

// ListArtifacts lists all artifacts for a session (deprecated - use GetAllSessionArtifacts)
func (c *Client) ListArtifacts(ctx context.Context, sessionID string) ([]*Artifact, error) {
	artifacts, err := c.GetAllSessionArtifacts(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var result []*Artifact
	for _, aa := range artifacts {
		artifact := aa.Artifact // Copy the artifact
		result = append(result, &artifact)
	}

	return result, nil
}
