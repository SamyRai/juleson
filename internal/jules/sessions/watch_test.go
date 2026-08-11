package sessions

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

func TestWatchUpdateTypeForDecision(t *testing.T) {
	tests := []struct {
		name     string
		want     WatchUpdateType
		decision WatchDecision
	}{
		{name: "progress", decision: WatchDecision{Kind: WatchDecisionContinue}, want: WatchUpdateProgress},
		{name: "needs user action", decision: WatchDecision{Kind: WatchDecisionNeedsUserAction}, want: WatchUpdateNeedsUserAction},
		{name: "failed", decision: WatchDecision{Kind: WatchDecisionFailed}, want: WatchUpdateTerminalFailure},
		{name: "completed", decision: WatchDecision{Kind: WatchDecisionCompletedWithDeliverables}, want: WatchUpdateTerminalSuccess},
		{name: "outputs", decision: WatchDecision{Kind: WatchDecisionOutputs}, want: WatchUpdateOutputsAvailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WatchUpdateTypeForDecision(tt.decision); got != tt.want {
				t.Fatalf("WatchUpdateTypeForDecision = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEvaluateWatchWake(t *testing.T) {
	tests := []struct {
		name         string
		policy       WakePolicy
		updateType   WatchUpdateType
		wantReason   string
		stateChanged bool
		wantWake     bool
	}{
		{name: "actionable progress does not wake", policy: WakePolicyActionable, updateType: WatchUpdateProgress},
		{name: "actionable user action wakes", policy: WakePolicyActionable, updateType: WatchUpdateNeedsUserAction, wantWake: true, wantReason: "user_action"},
		{name: "actionable completed wakes", policy: WakePolicyActionable, updateType: WatchUpdateTerminalSuccess, wantWake: true, wantReason: "terminal_state"},
		{name: "actionable failed wakes", policy: WakePolicyActionable, updateType: WatchUpdateTerminalFailure, wantWake: true, wantReason: "terminal_state"},
		{name: "actionable outputs wake", policy: WakePolicyActionable, updateType: WatchUpdateOutputsAvailable, wantWake: true, wantReason: "session_outputs"},
		{name: "any status wakes on state change", policy: WakePolicyAnyStatus, updateType: WatchUpdateProgress, stateChanged: true, wantWake: true, wantReason: "status_change"},
		{name: "terminal ignores user action", policy: WakePolicyTerminal, updateType: WatchUpdateNeedsUserAction},
		{name: "terminal wakes on completed", policy: WakePolicyTerminal, updateType: WatchUpdateTerminalSuccess, wantWake: true, wantReason: "terminal_state"},
		{name: "none ignores outputs", policy: WakePolicyNone, updateType: WatchUpdateOutputsAvailable},
		{name: "agent message wakes regardless policy", policy: WakePolicyNone, updateType: WatchUpdateAgentMessage, wantWake: true, wantReason: "agent_message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateWatchWake(tt.policy, tt.updateType, tt.stateChanged)
			if got.ShouldWake != tt.wantWake || got.WakeReason != tt.wantReason || got.UpdateType != tt.updateType {
				t.Fatalf("EvaluateWatchWake = %+v, want wake %t reason %q type %q", got, tt.wantWake, tt.wantReason, tt.updateType)
			}
		})
	}
}

func TestParseWakePolicy(t *testing.T) {
	for _, value := range []string{"", "actionable", "any-status", "terminal", "none"} {
		t.Run(value, func(t *testing.T) {
			if _, err := ParseWakePolicy(value); err != nil {
				t.Fatalf("ParseWakePolicy(%q) returned error: %v", value, err)
			}
		})
	}
	if _, err := ParseWakePolicy("later"); err == nil {
		t.Fatal("expected invalid wake policy to fail")
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

func TestCurrentWatchSnapshotFilteringAndCursors(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.Session{ID: "session-1", State: jules.SessionStateInProgress})
		})

	cursorTime := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	httpmock.RegisterResponder("GET", "=~^https://jules.googleapis.com/v1alpha/sessions/session-1/activities",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
				Activities: []jules.Activity{
					{
						ID:            "activity-1",
						Name:          "activity-1",
						CreateTime:    cursorTime.Add(time.Minute),
						AgentMessaged: &jules.AgentMessaged{AgentMessage: "hello"},
					},
					{
						ID:         "activity-2",
						CreateTime: cursorTime.Add(-time.Minute),
					},
				},
			})
		})

	snapshot, err := CurrentWatchSnapshot(context.Background(), client, "session-1", cursorTime, CurrentWatchOptions{FetchActivities: true})
	if err != nil {
		t.Fatalf("CurrentWatchSnapshot returned error: %v", err)
	}

	if !snapshot.HasJulesAgentMessage {
		t.Fatalf("Expected HasJulesAgentMessage to be true")
	}
}

func TestEvaluateWatchDecisionUserAction(t *testing.T) {
	decision := EvaluateWatchDecision(&jules.Session{State: jules.SessionStateAwaitingUserFeedback}, nil, nil)
	if decision.Kind != WatchDecisionNeedsUserAction {
		t.Fatalf("Expected WatchDecisionNeedsUserAction, got %v", decision.Kind)
	}

	if got := DefaultWatchWakeReason(decision); got != "user_action" {
		t.Fatalf("Expected wake reason user_action, got %v", got)
	}
}
