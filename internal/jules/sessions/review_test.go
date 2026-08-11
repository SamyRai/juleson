package sessions

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/jarcoal/httpmock"
)

func TestBuildSessionReviewActiveSession(t *testing.T) {
	tmpDir := cleanGitRepo(t)
	client := mockedReviewClient(t, &jules.Session{
		ID:    "session-1",
		State: jules.SessionStateInProgress,
	}, []jules.Activity{})

	review, err := BuildSessionReview(context.Background(), client, ReviewRequest{
		SessionID:  "session-1",
		WorkingDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("BuildSessionReview returned error: %v", err)
	}
	if review.Session.State != jules.SessionStateInProgress || len(review.Plans) != 0 {
		t.Fatalf("unexpected active review: %+v", review)
	}
	if !review.Worktree.Clean || len(review.Blockers) != 0 {
		t.Fatalf("expected clean non-blocked review: %+v", review)
	}
}

func TestBuildSessionReviewAwaitingPlanNextAction(t *testing.T) {
	tmpDir := cleanGitRepo(t)
	client := mockedReviewClient(t, &jules.Session{
		ID:    "session-1",
		State: jules.SessionStateAwaitingPlanApproval,
	}, []jules.Activity{
		{ID: "activity-plan", PlanGenerated: &jules.PlanGenerated{Plan: jules.Plan{ID: "plan-1", Steps: []jules.Step{{Index: 1, Title: "Inspect"}}}}},
	})

	review, err := BuildSessionReview(context.Background(), client, ReviewRequest{SessionID: "session-1", WorkingDir: tmpDir})
	if err != nil {
		t.Fatalf("BuildSessionReview returned error: %v", err)
	}
	if review.LatestPlan == nil || review.LatestPlan.PlanID != "plan-1" {
		t.Fatalf("latest plan missing: %+v", review.LatestPlan)
	}
	if !hasAction(review.NextActions, "approve plan") {
		t.Fatalf("approve action missing: %+v", review.NextActions)
	}
}

func TestBuildSessionReviewCompletedWithArtifacts(t *testing.T) {
	tmpDir := cleanGitRepo(t)
	base := gitHead(t, tmpDir)
	patch := filePatch(base)
	client := mockedReviewClient(t, &jules.Session{
		ID:    "session-1",
		State: jules.SessionStateCompleted,
		Outputs: []jules.Output{
			{PullRequest: &jules.PullRequest{URL: "https://github.com/acme/widgets/pull/1", Title: "Update widget"}},
		},
	}, []jules.Activity{
		{
			ID: "activity-patch",
			Artifacts: []jules.Artifact{
				{ChangeSet: &jules.ChangeSet{GitPatch: &jules.GitPatch{
					BaseCommitID:           base,
					UnidiffPatch:           patch,
					SuggestedCommitMessage: "Update widget file",
				}}},
			},
		},
	})

	review, err := BuildSessionReview(context.Background(), client, ReviewRequest{SessionID: "session-1", WorkingDir: tmpDir})
	if err != nil {
		t.Fatalf("BuildSessionReview returned error: %v", err)
	}
	if len(review.Outputs) != 1 || len(review.ArtifactManifests) != 1 {
		t.Fatalf("outputs/artifacts missing: %+v", review)
	}
	if review.PatchPreview.TotalPatches != 1 || !review.PatchPreview.CanApply {
		t.Fatalf("patch preview unexpected: %+v", review.PatchPreview)
	}
	if !hasAction(review.NextActions, "apply patches") {
		t.Fatalf("apply action missing: %+v", review.NextActions)
	}
}

func TestBuildSessionReviewCompletedWithoutArtifacts(t *testing.T) {
	tmpDir := cleanGitRepo(t)
	client := mockedReviewClient(t, &jules.Session{
		ID:    "session-1",
		State: jules.SessionStateCompleted,
	}, []jules.Activity{{ID: "activity-complete", SessionCompleted: &jules.SessionCompleted{}}})

	review, err := BuildSessionReview(context.Background(), client, ReviewRequest{SessionID: "session-1", WorkingDir: tmpDir})
	if err != nil {
		t.Fatalf("BuildSessionReview returned error: %v", err)
	}
	if review.PatchPreview.TotalPatches != 0 || len(review.ArtifactManifests) != 0 {
		t.Fatalf("expected no patches/artifacts: %+v", review)
	}
	if hasAction(review.NextActions, "apply patches") {
		t.Fatalf("unexpected apply action: %+v", review.NextActions)
	}
}

func TestBuildSessionReviewBaseMismatchBlocksApply(t *testing.T) {
	tmpDir := cleanGitRepo(t)
	patch := filePatch("0000000000000000000000000000000000000000")
	client := mockedReviewClient(t, &jules.Session{ID: "session-1", State: jules.SessionStateCompleted}, []jules.Activity{
		{
			ID: "activity-patch",
			Artifacts: []jules.Artifact{
				{ChangeSet: &jules.ChangeSet{GitPatch: &jules.GitPatch{
					BaseCommitID: "0000000000000000000000000000000000000000",
					UnidiffPatch: patch,
				}}},
			},
		},
	})

	review, err := BuildSessionReview(context.Background(), client, ReviewRequest{SessionID: "session-1", WorkingDir: tmpDir})
	if err != nil {
		t.Fatalf("BuildSessionReview returned error: %v", err)
	}
	if len(review.PatchPreview.BaseCommitMismatches) == 0 || !hasBlocker(review.Blockers, "base commit") {
		t.Fatalf("base mismatch not reported: preview=%+v blockers=%+v", review.PatchPreview, review.Blockers)
	}
	if hasAction(review.NextActions, "apply patches") {
		t.Fatalf("unexpected apply action: %+v", review.NextActions)
	}
}

func TestBuildSessionReviewDirtyWorktreeBlocksApply(t *testing.T) {
	tmpDir := cleanGitRepo(t)
	base := gitHead(t, tmpDir)
	client := mockedReviewClient(t, &jules.Session{ID: "session-1", State: jules.SessionStateCompleted}, []jules.Activity{
		{
			ID: "activity-patch",
			Artifacts: []jules.Artifact{
				{ChangeSet: &jules.ChangeSet{GitPatch: &jules.GitPatch{
					BaseCommitID: base,
					UnidiffPatch: filePatch(base),
				}}},
			},
		},
	})
	if err := os.WriteFile(tmpDir+"/dirty.txt", []byte("dirty"), 0o600); err != nil {
		t.Fatalf("write dirty file: %v", err)
	}

	review, err := BuildSessionReview(context.Background(), client, ReviewRequest{SessionID: "session-1", WorkingDir: tmpDir})
	if err != nil {
		t.Fatalf("BuildSessionReview returned error: %v", err)
	}
	if review.Worktree.Clean || !hasBlocker(review.Blockers, "local changes") {
		t.Fatalf("dirty worktree not blocked: worktree=%+v blockers=%+v", review.Worktree, review.Blockers)
	}
	if hasAction(review.NextActions, "apply patches") {
		t.Fatalf("unexpected apply action: %+v", review.NextActions)
	}
}

func mockedReviewClient(t *testing.T, session *jules.Session, activities []jules.Activity) *jules.Client {
	t.Helper()
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, session)
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{Activities: activities})
		})
	for i := range activities {
		activity := activities[i]
		httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities/"+activity.ID,
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewJsonResponse(200, activity)
			})
	}
	httpmock.RegisterRegexpResponder("GET", regexp.MustCompile(`^https://jules\.googleapis\.com/v1alpha/sessions/session-1/activities\?pageSize=100&pageToken=`),
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{})
		})
	return client
}

func cleanGitRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	runGit(t, tmpDir, "init")
	if err := os.WriteFile(tmpDir+"/file.txt", []byte("one\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, tmpDir, "add", "file.txt")
	runGit(t, tmpDir, "-c", "user.email=test@example.com", "-c", "user.name=Test User", "commit", "-m", "initial")
	return tmpDir
}

func gitHead(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse: %v\n%s", err, string(output))
	}
	return string(output[:len(output)-1])
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, string(output))
	}
}

func filePatch(base string) string {
	return "diff --git a/file.txt b/file.txt\n" +
		"index 5626abf..814f4a4 100644\n" +
		"--- a/file.txt\n" +
		"+++ b/file.txt\n" +
		"@@ -1 +1,2 @@\n" +
		" one\n" +
		"+two\n"
}

func hasAction(actions []ReviewNextAction, label string) bool {
	for _, action := range actions {
		if action.Label == label {
			return true
		}
	}
	return false
}

func hasBlocker(blockers []string, fragment string) bool {
	for _, blocker := range blockers {
		if regexp.MustCompile(regexp.QuoteMeta(fragment)).MatchString(blocker) {
			return true
		}
	}
	return false
}
