package sessions

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
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

func TestPreviewGitPatch_Fallback(t *testing.T) {
	cfg := &config.Config{
		Diff: config.DiffConfig{
			ForceNative: true, // Bypass external pagers
		},
	}
	patch := &jules.GitPatch{
		UnidiffPatch: "--- a/test.txt\n+++ b/test.txt\n@@ -1 +1 @@\n-old\n+new\n",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := previewGitPatch(cfg, patch)

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("previewGitPatch failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Git Patch") {
		t.Errorf("Expected 'Git Patch' in output, got: %s", output)
	}
	if !strings.Contains(output, "+new") {
		t.Errorf("Expected diff content '+new' in output")
	}
}
