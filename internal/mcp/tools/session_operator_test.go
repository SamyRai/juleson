package tools

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/jarcoal/httpmock"
)

func TestGetSessionPlansIncludesActivitiesAndSummaries(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=50",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
				Activities: []jules.Activity{
					{ID: "activity-plan", PlanGenerated: &jules.PlanGenerated{Plan: jules.Plan{
						ID: "plan-1",
						Steps: []jules.Step{
							{Index: 1, Title: "Inspect", Description: "Read files"},
						},
					}}},
				},
			})
		})

	result, output, err := getActivitiesWithPlans(context.Background(), nil, GetActivitiesWithPlansInput{SessionID: "session-1"}, client)
	if err != nil {
		t.Fatalf("getActivitiesWithPlans returned error: %v", err)
	}
	if result != nil {
		t.Fatalf("result = %+v, want nil", result)
	}
	if len(output.Activities) != 1 || output.Activities[0].ID != "activity-plan" {
		t.Fatalf("activities missing: %+v", output.Activities)
	}
	if len(output.Plans) != 1 || output.Plans[0].PlanID != "plan-1" || output.Plans[0].Steps[0].Description != "Read files" {
		t.Fatalf("plans missing: %+v", output.Plans)
	}
}

func TestReviewSessionStructuredOutputIsReadOnly(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	tmpDir := mcpCleanGitRepo(t)
	base := mcpGitHead(t, tmpDir)
	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	activity := jules.Activity{
		ID: "activity-patch",
		Artifacts: []jules.Artifact{
			{ChangeSet: &jules.ChangeSet{GitPatch: &jules.GitPatch{
				BaseCommitID:           base,
				UnidiffPatch:           mcpFilePatch(),
				SuggestedCommitMessage: "Update file",
			}}},
		},
	}
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.Session{ID: "session-1", State: jules.SessionStateCompleted})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{Activities: []jules.Activity{activity}})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities/activity-patch",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, activity)
		})

	result, output, err := reviewSession(context.Background(), nil, ReviewSessionInput{
		SessionID:  "session-1",
		WorkingDir: tmpDir,
	}, client)
	if err != nil {
		t.Fatalf("reviewSession returned error: %v", err)
	}
	if result != nil {
		t.Fatalf("result = %+v, want nil", result)
	}
	if output.Review.PatchPreview.TotalPatches != 1 || !output.Review.PatchPreview.CanApply {
		t.Fatalf("patch preview unexpected: %+v", output.Review.PatchPreview)
	}
	if !strings.Contains(output.Review.NextActions[0].Command, "sessions apply") {
		t.Fatalf("next actions missing apply suggestion: %+v", output.Review.NextActions)
	}
	data, err := os.ReadFile(tmpDir + "/file.txt")
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "one\n" {
		t.Fatalf("review mutated file: %q", string(data))
	}
}

func mcpCleanGitRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	mcpRunGit(t, tmpDir, "init")
	if err := os.WriteFile(tmpDir+"/file.txt", []byte("one\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	mcpRunGit(t, tmpDir, "add", "file.txt")
	mcpRunGit(t, tmpDir, "-c", "user.email=test@example.com", "-c", "user.name=Test User", "commit", "-m", "initial")
	return tmpDir
}

func mcpGitHead(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse: %v\n%s", err, string(output))
	}
	return strings.TrimSpace(string(output))
}

func mcpRunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, string(output))
	}
}

func mcpFilePatch() string {
	return "diff --git a/file.txt b/file.txt\n" +
		"index 5626abf..814f4a4 100644\n" +
		"--- a/file.txt\n" +
		"+++ b/file.txt\n" +
		"@@ -1 +1,2 @@\n" +
		" one\n" +
		"+two\n"
}
