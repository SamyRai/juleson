package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/SamyRai/juleson/pkg/jules"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSessionAllowsRepolessInput(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("POST", "https://jules.googleapis.com/v1alpha/sessions",
		func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			assert.False(t, strings.Contains(string(body), "sourceContext"))

			var received jules.CreateSessionRequest
			require.NoError(t, json.Unmarshal(body, &received))
			assert.Equal(t, "Draft a migration plan", received.Prompt)

			return httpmock.NewJsonResponse(201, jules.Session{
				ID:    "session-1",
				Title: "Repoless",
				State: "PLANNING",
			})
		})

	_, output, err := createSession(context.Background(), nil, CreateSessionInput{
		Prompt: "Draft a migration plan",
	}, client)

	require.NoError(t, err)
	assert.Equal(t, "session-1", output.SessionID)
}

func TestDeleteSessionRequiresConfirmation(t *testing.T) {
	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))

	result, output, err := deleteSession(context.Background(), nil, DeleteSessionInput{
		SessionID: "session-1",
		Confirm:   false,
	}, client)

	require.Error(t, err)
	assert.True(t, result.IsError)
	assert.Empty(t, output.SessionID)
}

func TestDeleteSessionConfirmed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("DELETE", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, ""), nil
		})

	_, output, err := deleteSession(context.Background(), nil, DeleteSessionInput{
		SessionID: "session-1",
		Confirm:   true,
	}, client)

	require.NoError(t, err)
	assert.Equal(t, "session-1", output.SessionID)
	assert.Equal(t, "deleted", output.Status)
}
