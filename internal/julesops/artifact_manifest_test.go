package julesops

import (
	"testing"
	"time"

	"github.com/SamyRai/juleson/pkg/jules"
	"github.com/stretchr/testify/assert"
)

func TestBuildArtifactManifest(t *testing.T) {
	exitCode := 7
	created := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	activity := jules.Activity{
		ID:         "activity-1",
		Name:       "sessions/session-1/activities/activity-1",
		CreateTime: created,
	}

	patchArtifact := jules.Artifact{
		ChangeSet: &jules.ChangeSet{GitPatch: &jules.GitPatch{
			UnidiffPatch: `diff --git a/a.txt b/a.txt
--- a/a.txt
+++ b/a.txt
@@ -1 +1,2 @@
 line
+new
`,
			BaseCommitID:           "abc123",
			SuggestedCommitMessage: "Update a.txt",
		}},
	}
	patchManifest := BuildArtifactManifest(activity, 0, patchArtifact)
	assert.Equal(t, "change_set", patchManifest.Type)
	assert.Equal(t, 1, patchManifest.FileCount)
	assert.Equal(t, "abc123", patchManifest.BaseCommitID)
	assert.Equal(t, "Update a.txt", patchManifest.SuggestedCommitMessage)
	assert.Equal(t, created, patchManifest.ActivityCreateTime)

	bashManifest := BuildArtifactManifest(activity, 1, jules.Artifact{
		BashOutput: &jules.BashOutput{Command: "go test ./...", ExitCode: exitCode},
	})
	assert.Equal(t, "bash_output", bashManifest.Type)
	assert.Equal(t, "go test ./...", bashManifest.BashCommand)
	assert.Equal(t, &exitCode, bashManifest.BashExitCode)

	mediaManifest := BuildArtifactManifest(activity, 2, jules.Artifact{
		Media: &jules.Media{MimeType: "image/png"},
	})
	assert.Equal(t, "media", mediaManifest.Type)
	assert.Equal(t, "image/png", mediaManifest.MediaMIMEType)
}
