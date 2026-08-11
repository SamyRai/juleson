package views

import (
	"strings"
	"testing"
	"time"
)

func TestSessionStatusPresentation(t *testing.T) {
	tests := map[string]struct {
		icon string
		text string
	}{
		"QUEUED":                 {icon: "⚡", text: "ACTIVE"},
		"PLANNING":               {icon: "⚡", text: "ACTIVE"},
		"IN_PROGRESS":            {icon: "⚡", text: "ACTIVE"},
		"AWAITING_PLAN_APPROVAL": {icon: "⏸", text: "NEEDS_USER_ACTION"},
		"AWAITING_USER_FEEDBACK": {icon: "⏸", text: "NEEDS_USER_ACTION"},
		"COMPLETED":              {icon: "✅", text: "COMPLETED"},
		"FAILED":                 {icon: "❌", text: "FAILED"},
		"SESSION_STATE_UNKNOWN":  {icon: "📋", text: "SESSION_STATE_UNKNOWN"},
	}

	for state, want := range tests {
		t.Run(state, func(t *testing.T) {
			if got := SessionStatusIcon(state); got != want.icon {
				t.Fatalf("SessionStatusIcon(%q) = %q, want %q", state, got, want.icon)
			}
			if got := SessionStatusText(state); got != want.text {
				t.Fatalf("SessionStatusText(%q) = %q, want %q", state, got, want.text)
			}
		})
	}
}

func TestSessionFormatterUsesDTO(t *testing.T) {
	created := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	output := NewSessionFormatter().FormatList([]SessionView{{
		ID:                  "session-1234567890",
		Title:               "Refactor",
		State:               "AWAITING_PLAN_APPROVAL",
		CreateTime:          created,
		Source:              "sources/github/acme/widgets",
		RequirePlanApproval: true,
		AutomationMode:      "AUTO_CREATE_PR",
		OutputCount:         2,
	}})

	for _, want := range []string{
		"Session: session-1234567890",
		"Title: Refactor",
		"State: 📋 AWAITING_PLAN_APPROVAL",
		"Source: sources/github/acme/widgets",
		"Plan Approval Required: Yes",
		"Automation Mode: AUTO_CREATE_PR",
		"Outputs: 2",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("formatted output missing %q:\n%s", want, output)
		}
	}
}
