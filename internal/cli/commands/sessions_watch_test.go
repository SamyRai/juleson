package commands

import (
	"bytes"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/jarcoal/httpmock"
)

func captureOutput(f func()) string {
	var buf bytes.Buffer
	old := os.Stdout
	defer func() {
		os.Stdout = old
	}()

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer r.Close()
	os.Stdout = w

	f()

	_ = w.Close()
	if _, err := buf.ReadFrom(r); err != nil {
		panic(err)
	}
	return buf.String()
}

func TestWatchSession_WakeOnStatusChange(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &config.Config{
		Jules: config.JulesConfig{
			APIKey: "test-api-key",
		},
	}

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		httpmock.NewJsonResponderOrPanic(200, jules.Session{
			ID:    "session-1",
			State: jules.SessionStateInProgress,
		}))

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=25",
		httpmock.NewJsonResponderOrPanic(200, jules.ActivitiesResponse{}))

	out := captureOutput(func() {
		err := watchSession(cfg, "session-1", "100ms", "1s", false, "", "", string(jules.SessionStateQueued), true, false, "")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	if !strings.Contains(out, "Wake reason: session state changed from QUEUED to IN_PROGRESS.") {
		t.Errorf("expected output to contain wake reason for status change, got: %s", out)
	}
}

func TestWatchSession_WakeOnAgentMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &config.Config{
		Jules: config.JulesConfig{
			APIKey: "test-api-key",
		},
	}

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		httpmock.NewJsonResponderOrPanic(200, jules.Session{
			ID:    "session-1",
			State: jules.SessionStateInProgress,
		}))

	called := false
	httpmock.RegisterResponder("GET", "=~^https://jules.googleapis.com/v1alpha/sessions/session-1/activities",
		func(req *http.Request) (*http.Response, error) {
			if !called {
				called = true
				return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
					Activities: []jules.Activity{
						{
							ID:            "activity-1",
							CreateTime:    time.Now(),
							AgentMessaged: &jules.AgentMessaged{AgentMessage: "hello"},
						},
					},
				})
			}
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
				Activities: []jules.Activity{
					{
						ID:            "activity-2",
						CreateTime:    time.Now().Add(time.Second),
						AgentMessaged: &jules.AgentMessaged{AgentMessage: "hello again"},
					},
				},
			})
		})

	out := captureOutput(func() {
		err := watchSession(cfg, "session-1", "100ms", "1s", false, "", "", "", false, true, "")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	if !strings.Contains(out, "Wake reason: Jules sent a new message.") {
		t.Errorf("expected output to contain wake reason for agent message, got: %s", out)
	}
}

func TestWatchSession_TerminalState(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &config.Config{
		Jules: config.JulesConfig{
			APIKey: "test-api-key",
		},
	}

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		httpmock.NewJsonResponderOrPanic(200, jules.Session{
			ID:    "session-1",
			State: jules.SessionStateFailed,
		}))

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=25",
		httpmock.NewJsonResponderOrPanic(200, jules.ActivitiesResponse{}))

	out := captureOutput(func() {
		err := watchSession(cfg, "session-1", "100ms", "1s", false, "", "", "", false, false, "")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	if !strings.Contains(out, "Next action: inspect failure details with 'juleson sessions get session-1'.") {
		t.Errorf("expected output to contain failure next action, got: %s", out)
	}
}

func TestWatchSession_Timeout(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &config.Config{
		Jules: config.JulesConfig{
			APIKey: "test-api-key",
		},
	}

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		httpmock.NewJsonResponderOrPanic(200, jules.Session{
			ID:    "session-1",
			State: jules.SessionStateInProgress,
		}))

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=25",
		httpmock.NewJsonResponderOrPanic(200, jules.ActivitiesResponse{}))

	err := watchSession(cfg, "session-1", "100ms", "200ms", false, "", "", "", false, false, "")
	if err == nil {
		t.Errorf("expected timeout error")
	} else if !strings.Contains(err.Error(), "timeout watching session") {
		t.Errorf("expected timeout error message, got %v", err)
	}
}
