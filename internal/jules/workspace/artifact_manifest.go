package workspace

import (
	"context"
	"strings"
	"time"

	"github.com/SamyRai/go-jules"
)

// ArtifactManifest summarizes a documented Jules activity artifact without
// embedding large payload content.
type ArtifactManifest struct {
	ActivityCreateTime     time.Time    `json:"activity_create_time,omitempty"`
	BashExitCode           *int         `json:"bash_exit_code,omitempty"`
	MediaMIMEType          string       `json:"media_mime_type,omitempty"`
	ActivityName           string       `json:"activity_name,omitempty"`
	ActivityID             string       `json:"activity_id"`
	Type                   string       `json:"type"`
	BashCommand            string       `json:"bash_command,omitempty"`
	BaseCommitID           string       `json:"base_commit_id,omitempty"`
	SuggestedCommitMessage string       `json:"suggested_commit_message,omitempty"`
	Files                  []FileChange `json:"files,omitempty"`
	Index                  int          `json:"index"`
	FileCount              int          `json:"file_count,omitempty"`
	Empty                  bool         `json:"empty,omitempty"`
}

// ListSessionArtifactManifests returns manifests for all artifacts in a session.
func ListSessionArtifactManifests(ctx context.Context, client *jules.Client, sessionID string) ([]ArtifactManifest, error) {
	activities, err := client.Activities().ListAll(ctx, sessionID, 100)
	if err != nil {
		return nil, err
	}

	var manifests []ArtifactManifest
	for _, activity := range activities {
		for i, artifact := range activity.Artifacts {
			manifests = append(manifests, BuildArtifactManifest(activity, i, artifact))
		}
	}
	return manifests, nil
}

// BuildArtifactManifest summarizes one activity artifact.
func BuildArtifactManifest(activity jules.Activity, index int, artifact jules.Artifact) ArtifactManifest {
	manifest := ArtifactManifest{
		ActivityID:         activity.ID,
		ActivityName:       activity.Name,
		ActivityCreateTime: activity.CreateTime,
		Index:              index,
		Type:               "unknown",
	}

	switch {
	case artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil:
		manifest.Type = "change_set"
		manifest.BaseCommitID = artifact.ChangeSet.GitPatch.BaseCommitID
		manifest.SuggestedCommitMessage = artifact.ChangeSet.GitPatch.SuggestedCommitMessage
		manifest.Empty = strings.TrimSpace(artifact.ChangeSet.GitPatch.UnidiffPatch) == ""
		manifest.Files = parsePatchFiles(artifact.ChangeSet.GitPatch.UnidiffPatch)
		manifest.FileCount = len(manifest.Files)
	case artifact.BashOutput != nil:
		manifest.Type = "bash_output"
		manifest.BashCommand = artifact.BashOutput.Command
		exitCode := artifact.BashOutput.ExitCode
		manifest.BashExitCode = &exitCode
	case artifact.Media != nil:
		manifest.Type = "media"
		manifest.MediaMIMEType = artifact.Media.MimeType
	}

	return manifest
}
