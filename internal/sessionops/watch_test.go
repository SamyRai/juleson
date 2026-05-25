package sessionops

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/jarcoal/httpmock"
)

func TestEvaluateWatchDecision(t *testing.T) {
	hasDeliverables := true
	noDeliverables := false
	tests := []struct {
		name            string
		session         *jules.Session
		hasDeliverables *bool
		deliverablesErr error
		want            WatchDecisionKind
		stop            bool
	}{
		{
			name:    "needs user action",
			session: &jules.Session{State: jules.SessionStateAwaitingPlanApproval},
			want:    WatchDecisionNeedsUserAction,
			stop:    true,
		},
		{
			name:    "failed",
			session: &jules.Session{State: jules.SessionStateFailed},
			want:    WatchDecisionFailed,
			stop:    true,
		},
		{
			name:            "completed with deliverables",
			session:         &jules.Session{State: jules.SessionStateCompleted},
			hasDeliverables: &hasDeliverables,
			want:            WatchDecisionCompletedWithDeliverables,
			stop:            true,
		},
		{
			name:            "completed without deliverables",
			session:         &jules.Session{State: jules.SessionStateCompleted},
			hasDeliverables: &noDeliverables,
			want:            WatchDecisionCompletedNoDeliverables,
			stop:            true,
		},
		{
			name:    "outputs",
			session: &jules.Session{State: jules.SessionStateInProgress, Outputs: []jules.Output{{PullRequest: &jules.PullRequest{URL: "https://example.test/pr"}}}},
			want:    WatchDecisionOutputs,
			stop:    true,
		},
		{
			name:    "continue",
			session: &jules.Session{State: jules.SessionStateInProgress},
			want:    WatchDecisionContinue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateWatchDecision(tt.session, tt.hasDeliverables, tt.deliverablesErr)
			if got.Kind != tt.want || got.Stop != tt.stop {
				t.Fatalf("decision = %+v, want kind %s stop %t", got, tt.want, tt.stop)
			}
		})
	}
}

func TestHasJulesAgentMessageAfter(t *testing.T) {
	cursor := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	activities := []jules.Activity{
		{
			CreateTime:    cursor.Add(-time.Minute),
			AgentMessaged: &jules.AgentMessaged{AgentMessage: "old"},
		},
		{
			CreateTime:    cursor.Add(time.Minute),
			AgentMessaged: &jules.AgentMessaged{AgentMessage: "new"},
		},
	}
	if !HasJulesAgentMessageAfter(activities, cursor) {
		t.Fatal("expected agent message after cursor")
	}
	if HasJulesAgentMessageAfter(activities[:1], cursor) {
		t.Fatal("did not expect old agent message to match")
	}
}

func TestCurrentWatchSnapshotCompletedWithDeliverables(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.Session{ID: "session-1", State: jules.SessionStateCompleted})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=25",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
				Activities: []jules.Activity{{
					ID: "activity-1",
					Artifacts: []jules.Artifact{{
						ChangeSet: &jules.ChangeSet{GitPatch: &jules.GitPatch{UnidiffPatch: "diff --git a/file b/file\n"}},
					}},
				}},
			})
		})

	snapshot, err := CurrentWatchSnapshot(context.Background(), client, "session-1", time.Time{}, CurrentWatchOptions{FetchActivities: true})
	if err != nil {
		t.Fatalf("CurrentWatchSnapshot returned error: %v", err)
	}
	if snapshot.Decision.Kind != WatchDecisionCompletedWithDeliverables {
		t.Fatalf("decision = %+v", snapshot.Decision)
	}
	if got := MCPNextAction(snapshot); got != "call preview_session_changes, then apply_session_patches with confirm_apply=true if acceptable" {
		t.Fatalf("MCPNextAction = %q", got)
	}
	if got := DefaultWatchWakeReason(snapshot.Decision); got != "terminal_state" {
		t.Fatalf("DefaultWatchWakeReason = %q", got)
	}
}
