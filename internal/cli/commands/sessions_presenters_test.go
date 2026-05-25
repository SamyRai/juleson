package commands

import (
	"testing"

	"github.com/SamyRai/juleson/pkg/jules"
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
			if got := getSessionStatusText(state); got != want {
				t.Fatalf("getSessionStatusText(%q) = %q, want %q", state, got, want)
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
			if got := getSessionStatusIcon(state); got != want {
				t.Fatalf("getSessionStatusIcon(%q) = %q, want %q", state, got, want)
			}
		})
	}
}
