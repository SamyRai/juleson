package sessions

import (
	"testing"

	"github.com/SamyRai/go-jules"
	julessessions "github.com/SamyRai/juleson/internal/jules/sessions"
	"github.com/SamyRai/juleson/internal/presentation/views"
)

func TestGetSessionStatusText(t *testing.T) {
	tests := map[jules.SessionState]string{
		jules.SessionStateQueued:               "ACTIVE",
		jules.SessionStatePlanning:             "ACTIVE",
		jules.SessionStateInProgress:           "ACTIVE",
		jules.SessionStateAwaitingPlanApproval: "NEEDS_USER_ACTION",
		jules.SessionStateAwaitingUserFeedback: "NEEDS_USER_ACTION",
		jules.SessionStateCompleted:            "COMPLETED",
		jules.SessionStateFailed:               "FAILED",
		jules.SessionStateUnspecified:          string(jules.SessionStateUnspecified),
	}

	for state, want := range tests {
		t.Run(string(state), func(t *testing.T) {
			if got := views.SessionStatusText(string(state)); got != want {
				t.Fatalf("SessionStatusText(%q) = %q, want %q", state, got, want)
			}
		})
	}
}

func TestGetSessionStatusIcon(t *testing.T) {
	tests := map[jules.SessionState]string{
		jules.SessionStateQueued:               "⚡",
		jules.SessionStateAwaitingPlanApproval: "⏸",
		jules.SessionStateCompleted:            "✅",
		jules.SessionStateFailed:               "❌",
		jules.SessionStateUnspecified:          "📋",
	}

	for state, want := range tests {
		t.Run(string(state), func(t *testing.T) {
			if got := views.SessionStatusIcon(string(state)); got != want {
				t.Fatalf("SessionStatusIcon(%q) = %q, want %q", state, got, want)
			}
		})
	}
}

func TestCLIWakePolicyCompatibilityFlag(t *testing.T) {
	got, err := cliWakePolicy(true, string(julessessions.WakePolicyActionable))
	if err != nil {
		t.Fatalf("cliWakePolicy returned error: %v", err)
	}
	if got != julessessions.WakePolicyAnyStatus {
		t.Fatalf("cliWakePolicy = %q, want %q", got, julessessions.WakePolicyAnyStatus)
	}
}

func TestCLIWakePolicyUsesConfiguredDefault(t *testing.T) {
	got, err := cliWakePolicy(false, "")
	if err != nil {
		t.Fatalf("cliWakePolicy returned error: %v", err)
	}
	if got != julessessions.WakePolicyActionable {
		t.Fatalf("cliWakePolicy = %q, want %q", got, julessessions.WakePolicyActionable)
	}
}
