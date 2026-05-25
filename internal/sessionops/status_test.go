package sessionops

import (
	"testing"

	"github.com/SamyRai/go-jules"
)

func TestSummarizeSessions(t *testing.T) {
	summary := SummarizeSessions([]jules.Session{
		{ID: "session-1", State: jules.SessionStatePlanning},
		{ID: "session-2", State: jules.SessionStateInProgress},
		{ID: "session-3", State: jules.SessionStateCompleted},
		{ID: "session-4", State: jules.SessionStateQueued},
		{ID: "session-5", State: jules.SessionStateAwaitingPlanApproval},
	}, 3)

	if summary.TotalSessions != 5 {
		t.Fatalf("TotalSessions = %d", summary.TotalSessions)
	}
	if summary.ActiveSessions != 3 {
		t.Fatalf("ActiveSessions = %d", summary.ActiveSessions)
	}
	if summary.UserActionSessions != 1 {
		t.Fatalf("UserActionSessions = %d", summary.UserActionSessions)
	}
	if summary.StateBreakdown[string(jules.SessionStateCompleted)] != 1 {
		t.Fatalf("completed count = %d", summary.StateBreakdown[string(jules.SessionStateCompleted)])
	}
	if len(summary.RecentSessions) != 3 {
		t.Fatalf("RecentSessions length = %d", len(summary.RecentSessions))
	}
	if summary.Summary != "Found 5 total sessions with 3 currently active and 1 needing user action" {
		t.Fatalf("Summary = %q", summary.Summary)
	}
}

func TestSummarizeSessionsHandlesNilAndNegativeRecentLimit(t *testing.T) {
	summary := SummarizeSessions(nil, -1)
	if summary.TotalSessions != 0 || len(summary.RecentSessions) != 0 {
		t.Fatalf("unexpected empty summary: %+v", summary)
	}
	if summary.StateBreakdown == nil {
		t.Fatal("StateBreakdown should be initialized")
	}
}

func TestDocumentedOutputsFiltersPullRequests(t *testing.T) {
	outputs := DocumentedOutputs(&jules.Session{
		Outputs: []jules.Output{
			{},
			{PullRequest: &jules.PullRequest{URL: "https://github.com/acme/widgets/pull/1"}},
		},
	})
	if len(outputs) != 1 {
		t.Fatalf("len(outputs) = %d", len(outputs))
	}
	if outputs[0].PullRequest.URL != "https://github.com/acme/widgets/pull/1" {
		t.Fatalf("unexpected output: %+v", outputs[0])
	}
}
