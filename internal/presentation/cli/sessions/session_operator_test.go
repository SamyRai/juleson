package sessions

import (
	"bytes"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	julessessions "github.com/SamyRai/juleson/internal/jules/sessions"
	"github.com/jarcoal/httpmock"
)

func TestSessionsPlansCommandRegistered(t *testing.T) {
	cmd := NewSessionsCommand(operatorTestConfig())
	plans, _, err := cmd.Find([]string{"plans", "session-1"})
	if err != nil {
		t.Fatalf("find plans command: %v", err)
	}
	if plans == nil || plans.Name() != "plans" {
		t.Fatalf("plans command not found: %+v", plans)
	}
	if plans.Flags().Lookup("latest") == nil || plans.Flags().Lookup("json") == nil {
		t.Fatalf("plans flags not registered")
	}
}

func TestPrintPlanSummariesIncludesFullStepsAndIDs(t *testing.T) {
	output := captureStdout(t, func() {
		printPlanSummaries("session-1", []julessessions.PlanSummary{
			{
				ActivityID:   "activity-plan",
				ActivityName: "sessions/session-1/activities/activity-plan",
				PlanID:       "plan-1",
				Approved:     true,
				Steps: []julessessions.PlanStepSummary{
					{Index: 1, Title: "Inspect", Description: "Read all relevant files"},
					{Index: 2, Title: "Patch", Description: "Apply the scoped fix"},
				},
			},
		})
	})

	for _, want := range []string{
		"Activity ID: activity-plan",
		"Activity Name: sessions/session-1/activities/activity-plan",
		"Plan ID: plan-1",
		"Approved: true",
		"Read all relevant files",
		"juleson sessions review session-1 <project-path>",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestShowSessionPlansLatestJSON(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	oldTime := time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(time.Hour)
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
				Activities: []jules.Activity{
					{ID: "old", CreateTime: oldTime, PlanGenerated: &jules.PlanGenerated{Plan: jules.Plan{ID: "old-plan", CreateTime: oldTime}}},
					{ID: "new", CreateTime: newTime, PlanGenerated: &jules.PlanGenerated{Plan: jules.Plan{ID: "new-plan", CreateTime: newTime}}},
				},
			})
		})

	output := captureStdout(t, func() {
		if err := showSessionPlans(operatorTestConfig(), "session-1", true, true); err != nil {
			t.Fatalf("showSessionPlans returned error: %v", err)
		}
	})
	if !strings.Contains(output, `"plan_id": "new-plan"`) || strings.Contains(output, `"plan_id": "old-plan"`) {
		t.Fatalf("latest json output unexpected:\n%s", output)
	}
}

func TestActivitiesListShowsIDAndName(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
				Activities: []jules.Activity{
					{ID: "activity-1", Name: "sessions/session-1/activities/activity-1", Originator: jules.ActivityOriginatorAgent},
				},
			})
		})

	output := captureStdout(t, func() {
		if err := core.ListActivities(operatorTestConfig(), "session-1", "", ""); err != nil {
			t.Fatalf("listActivities returned error: %v", err)
		}
	})
	if !strings.Contains(output, "ID: activity-1") || !strings.Contains(output, "Name: sessions/session-1/activities/activity-1") {
		t.Fatalf("activity ID/name missing:\n%s", output)
	}
}

func TestPrintSessionReviewNextActions(t *testing.T) {
	review := &julessessions.SessionReview{
		SessionID: "session-1",
		Session:   jules.Session{ID: "session-1", State: jules.SessionStateCompleted, Title: "Done"},
		PatchPreview: julessessions.PatchPreviewSummary{
			TotalPatches: 1,
			CanApply:     true,
			Summary:      "1 patches affecting 1 files (+1 -0 lines)",
		},
		Worktree: julessessions.WorktreeReview{WorkingDir: "/tmp/project", Clean: true},
		NextActions: []julessessions.ReviewNextAction{
			{Label: "apply patches", Command: "juleson sessions apply session-1 /tmp/project --confirm", Reason: "dry-run passed"},
		},
	}
	output := captureStdout(t, func() {
		printSessionReview(review)
	})
	if !strings.Contains(output, "Patch preview: 1 patches affecting 1 files") ||
		!strings.Contains(output, "apply patches: juleson sessions apply session-1 /tmp/project --confirm") {
		t.Fatalf("review output missing next action:\n%s", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = writer
	fn()
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	os.Stdout = original
	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, reader); err != nil {
		t.Fatalf("copy stdout: %v", err)
	}
	return buffer.String()
}

func operatorTestConfig() *config.Config {
	return &config.Config{
		Jules: config.JulesConfig{
			APIKey:        "test-api-key",
			BaseURL:       "https://jules.googleapis.com/v1alpha",
			Timeout:       30 * time.Second,
			RetryAttempts: 0,
		},
	}
}
