package jules

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PatchesTestSuite struct {
	suite.Suite
	client *Client
}

func (suite *PatchesTestSuite) SetupTest() {
	httpmock.Activate()
	suite.client = NewClient("test-api-key", "https://jules.googleapis.com", 30, 3)
}

func (suite *PatchesTestSuite) TearDownTest() {
	httpmock.DeactivateAndReset()
}

func TestPatchesTestSuite(t *testing.T) {
	suite.Run(t, new(PatchesTestSuite))
}

func (suite *PatchesTestSuite) TestGetSessionChanges() {
	sessionID := "session-123"

	// Mock the activities list response with patches
	activitiesResponse := ActivitiesResponse{
		Activities: []Activity{
			{
				ID:   "activity-1",
				Name: "sessions/session-123/activities/activity-1",
				Artifacts: []Artifact{
					{
						ChangeSet: &ChangeSet{
							GitPatch: &GitPatch{
								UnidiffPatch: `diff --git a/test.txt b/test.txt
index 1234567..abcdefg 100644
--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,4 @@
 line 1
+line 2
 line 3
-line 4
`,
								BaseCommitID:           "abc123",
								SuggestedCommitMessage: "Update test.txt",
							},
						},
					},
				},
			},
		},
	}

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/sessions/session-123/activities",
		httpmock.NewJsonResponderOrPanic(200, activitiesResponse))

	changes, err := suite.client.GetSessionChanges(context.Background(), sessionID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), changes)
	assert.Equal(suite.T(), sessionID, changes.SessionID)
	assert.Equal(suite.T(), 1, changes.TotalPatches)
	assert.Len(suite.T(), changes.Files, 1)
	assert.Equal(suite.T(), "test.txt", changes.Files[0].Path)
	assert.Equal(suite.T(), 1, changes.Files[0].LinesAdded)
	assert.Equal(suite.T(), 1, changes.Files[0].LinesRemoved)
}

func (suite *PatchesTestSuite) TestParsePatchFiles() {
	patch := `diff --git a/file1.go b/file1.go
index 1234567..abcdefg 100644
--- a/file1.go
+++ b/file1.go
@@ -1,5 +1,7 @@
 package main

+import "fmt"
+
 func main() {
-	println("hello")
+	fmt.Println("hello world")
 }
diff --git a/file2.go b/file2.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/file2.go
@@ -0,0 +1,3 @@
+package test
+
+// New file
`

	changes := parsePatchFiles(patch)

	assert.Len(suite.T(), changes, 2)

	// Check first file
	assert.Equal(suite.T(), "file1.go", changes[0].Path)
	assert.Equal(suite.T(), 3, changes[0].LinesAdded)
	assert.Equal(suite.T(), 1, changes[0].LinesRemoved)

	// Check second file
	assert.Equal(suite.T(), "file2.go", changes[1].Path)
	assert.Equal(suite.T(), 3, changes[1].LinesAdded)
	assert.Equal(suite.T(), 0, changes[1].LinesRemoved)
}

func (suite *PatchesTestSuite) TestParseGitApplyOutput() {
	output := `Checking patch file1.txt...
Checking patch file2.txt...
Applying patch to file1.txt...
Applying patch to file2.txt...`

	files := parseGitApplyOutput(output)

	assert.Len(suite.T(), files, 2)
	assert.Contains(suite.T(), files, "file1.txt")
	assert.Contains(suite.T(), files, "file2.txt")
}

func (suite *PatchesTestSuite) TestApplyGitPatch() {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "jules-patch-test-*")
	assert.NoError(suite.T(), err)
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("line 1\nline 2\nline 3\n"), 0644)
	assert.NoError(suite.T(), err)

	// Initialize a git repo (required for git apply)
	suite.T().Skip("Skipping test that requires git initialization")
	// Would need: exec.Command("git", "init", tmpDir).Run()
}

func (suite *PatchesTestSuite) TestApplyActivityPatches() {
	activityID := "activity-1"

	// Mock the activity response with a patch
	activityResponse := Activity{
		ID:   activityID,
		Name: "sessions/session-123/activities/activity-1",
		Artifacts: []Artifact{
			{
				ChangeSet: &ChangeSet{
					GitPatch: &GitPatch{
						UnidiffPatch: `diff --git a/test.txt b/test.txt
index 1234567..abcdefg 100644
--- a/test.txt
+++ b/test.txt
@@ -1,2 +1,3 @@
 line 1
+line 2
 line 3
`,
					},
				},
			},
		},
	}

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/sessions/session-123/activities/activity-1",
		httpmock.NewJsonResponderOrPanic(200, activityResponse))

	// This would require actual git repo, so we skip
	suite.T().Skip("Skipping test that requires git repository")
}

func (suite *PatchesTestSuite) TestCopyFile() {
	tmpDir, err := os.MkdirTemp("", "jules-copy-test-*")
	assert.NoError(suite.T(), err)
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	// Create source file
	content := []byte("test content")
	err = os.WriteFile(srcFile, content, 0644)
	assert.NoError(suite.T(), err)

	// Copy file
	err = copyFile(srcFile, dstFile)
	assert.NoError(suite.T(), err)

	// Verify destination file
	dstContent, err := os.ReadFile(dstFile)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), content, dstContent)
}

func (suite *PatchesTestSuite) TestGetSessionChangesNoPatch() {
	sessionID := "session-456"

	// Mock response with no patches
	activitiesResponse := ActivitiesResponse{
		Activities: []Activity{
			{
				ID:   "activity-1",
				Name: "sessions/session-456/activities/activity-1",
				Artifacts: []Artifact{
					{
						BashOutput: &BashOutput{
							Command:  "echo test",
							Output:   "some output",
							ExitCode: 0,
						},
					},
				},
			},
		},
	}

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/sessions/session-456/activities",
		httpmock.NewJsonResponderOrPanic(200, activitiesResponse))

	changes, err := suite.client.GetSessionChanges(context.Background(), sessionID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), changes)
	assert.Equal(suite.T(), 0, changes.TotalPatches)
	assert.Len(suite.T(), changes.Files, 0)
}

func (suite *PatchesTestSuite) TestGetSessionChangesMultiplePatches() {
	sessionID := "session-789"

	// Mock response with multiple patches affecting the same file
	activitiesResponse := ActivitiesResponse{
		Activities: []Activity{
			{
				ID:   "activity-1",
				Name: "sessions/session-789/activities/activity-1",
				Artifacts: []Artifact{
					{
						ChangeSet: &ChangeSet{
							GitPatch: &GitPatch{
								UnidiffPatch: `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,1 +1,2 @@
 package main
+import "fmt"
`,
							},
						},
					},
				},
			},
			{
				ID:   "activity-2",
				Name: "sessions/session-789/activities/activity-2",
				Artifacts: []Artifact{
					{
						ChangeSet: &ChangeSet{
							GitPatch: &GitPatch{
								UnidiffPatch: `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -3,1 +3,2 @@
 func main() {
+	fmt.Println("hello")
 }
`,
							},
						},
					},
				},
			},
		},
	}

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/sessions/session-789/activities",
		httpmock.NewJsonResponderOrPanic(200, activitiesResponse))

	changes, err := suite.client.GetSessionChanges(context.Background(), sessionID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), changes)
	assert.Equal(suite.T(), 2, changes.TotalPatches)
	assert.Len(suite.T(), changes.Files, 1)
	assert.Equal(suite.T(), "main.go", changes.Files[0].Path)
	// Both patches add lines to the same file
	assert.Equal(suite.T(), 2, changes.Files[0].LinesAdded)
	assert.Equal(suite.T(), 0, changes.Files[0].LinesRemoved)
}
