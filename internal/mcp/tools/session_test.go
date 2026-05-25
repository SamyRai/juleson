package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
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

func TestCreateSessionReadsPromptFile(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	promptFile := t.TempDir() + "/prompt.txt"
	require.NoError(t, os.WriteFile(promptFile, []byte("Implement cursor watch"), 0644))

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sources/github/acme/widgets",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.Source{
				Name: "sources/github/acme/widgets",
				GithubRepo: &jules.GithubRepo{
					DefaultBranch: &jules.Branch{DisplayName: "main"},
				},
			})
		})
	httpmock.RegisterResponder("POST", "https://jules.googleapis.com/v1alpha/sessions",
		func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			var received jules.CreateSessionRequest
			require.NoError(t, json.Unmarshal(body, &received))
			assert.Equal(t, "Implement cursor watch", received.Prompt)
			assert.Equal(t, "sources/github/acme/widgets", received.SourceContext.Source)
			require.NotNil(t, received.SourceContext.GithubRepoContext)
			assert.Equal(t, "main", received.SourceContext.GithubRepoContext.StartingBranch)
			return httpmock.NewJsonResponse(201, jules.Session{ID: "session-1", State: jules.SessionStatePlanning})
		})

	_, output, err := createSession(context.Background(), nil, CreateSessionInput{
		Source:     "github/acme/widgets",
		PromptFile: promptFile,
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

func TestListSessionsSuccessOutput(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions?pageSize=2",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.SessionsResponse{
				Sessions: []jules.Session{
					{ID: "session-1", State: jules.SessionStatePlanning},
					{ID: "session-2", State: jules.SessionStateCompleted},
				},
				NextPageToken: "next",
			})
		})

	result, output, err := listSessions(context.Background(), nil, ListSessionsInput{Limit: 2}, client)

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 2, output.TotalCount)
	assert.Equal(t, "next", output.NextCursor)
	assert.Len(t, output.Sessions, 2)
}

func TestGetSessionStatusSummarizesSessions(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.SessionsResponse{
				Sessions: []jules.Session{
					{ID: "session-1", State: jules.SessionStatePlanning},
					{ID: "session-2", State: jules.SessionStateInProgress},
					{ID: "session-3", State: jules.SessionStateCompleted},
					{ID: "session-4", State: jules.SessionStateQueued},
					{ID: "session-5", State: jules.SessionStateAwaitingPlanApproval},
				},
			})
		})

	result, output, err := getSessionStatus(context.Background(), nil, GetSessionStatusInput{}, client)

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 5, output.TotalSessions)
	assert.Equal(t, 3, output.ActiveSessions)
	assert.Equal(t, 1, output.UserActionSessions)
	assert.Equal(t, 1, output.StateBreakdown[string(jules.SessionStateCompleted)])
	assert.Len(t, output.RecentSessions, 5)
	assert.Equal(t, "Found 5 total sessions with 3 currently active and 1 needing user action", output.Summary)
}

func TestWatchSessionReturnsOnStatusChange(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.Session{
				ID:    "session-1",
				State: jules.SessionStateInProgress,
			})
		})
	httpmock.RegisterRegexpResponder("GET", regexp.MustCompile(`^https://jules\.googleapis\.com/v1alpha/sessions/session-1/activities\?.*`),
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{})
		})

	result, output, err := watchSession(context.Background(), nil, WatchSessionInput{
		SessionID:            "session-1",
		InitialState:         string(jules.SessionStatePlanning),
		ReturnOnStatusChange: true,
	}, client)

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "status_change", output.WakeReason)
	assert.Equal(t, string(jules.SessionStateInProgress), output.State)
}

func TestWatchSessionReturnsOnJulesAgentMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cursor := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.Session{
				ID:    "session-1",
				State: jules.SessionStateInProgress,
			})
		})
	httpmock.RegisterRegexpResponder("GET", regexp.MustCompile(`^https://jules\.googleapis\.com/v1alpha/sessions/session-1/activities\?.*`),
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
				Activities: []jules.Activity{
					{
						ID:            "activity-1",
						CreateTime:    cursor.Add(time.Minute),
						Originator:    jules.ActivityOriginatorAgent,
						AgentMessaged: &jules.AgentMessaged{AgentMessage: "I need feedback on the plan."},
					},
				},
			})
		})

	result, output, err := watchSession(context.Background(), nil, WatchSessionInput{
		SessionID:                 "session-1",
		Since:                     cursor.Format(time.RFC3339Nano),
		ReturnOnJulesAgentMessage: true,
	}, client)

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "jules_agent_message", output.WakeReason)
	assert.Len(t, output.RecentActivities, 1)
}

func TestCurrentWatchSessionOutputCompletedWithoutDeliverables(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.Session{
				ID:    "session-1",
				State: jules.SessionStateCompleted,
			})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=25",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
				Activities: []jules.Activity{
					{
						ID: "activity-1",
						Artifacts: []jules.Artifact{
							{ChangeSet: &jules.ChangeSet{GitPatch: &jules.GitPatch{UnidiffPatch: " \n"}}},
						},
					},
				},
			})
		})

	output, err := currentWatchSessionOutput(context.Background(), "session-1", client, time.Time{})

	require.NoError(t, err)
	assert.Equal(t, string(jules.SessionStateCompleted), output.State)
	assert.Contains(t, output.NextAction, "no retrievable deliverable")
}

func TestCurrentWatchSessionOutputCompletedWithDeliverables(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.Session{
				ID:    "session-1",
				State: jules.SessionStateCompleted,
			})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=25",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
				Activities: []jules.Activity{
					{
						ID: "activity-1",
						Artifacts: []jules.Artifact{
							{ChangeSet: &jules.ChangeSet{GitPatch: &jules.GitPatch{UnidiffPatch: "diff --git a/file b/file\n"}}},
						},
					},
				},
			})
		})

	output, err := currentWatchSessionOutput(context.Background(), "session-1", client, time.Time{})

	require.NoError(t, err)
	assert.Equal(t, string(jules.SessionStateCompleted), output.State)
	assert.Contains(t, output.NextAction, "preview_session_changes")
}

func TestApplySessionPatchesRequiresCleanWorktreeForMutation(t *testing.T) {
	tmpDir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())
	require.NoError(t, os.WriteFile(tmpDir+"/dirty.txt", []byte("dirty"), 0644))

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	result, output, err := applySessionPatches(context.Background(), nil, ApplySessionPatchesInput{
		SessionID:    "session-1",
		WorkingDir:   tmpDir,
		ConfirmApply: true,
	}, client)

	require.Error(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
	assert.True(t, output.DryRun)
	assert.NotEmpty(t, output.Blockers)
}

func TestListSessionArtifactsMCP(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
				Activities: []jules.Activity{
					{
						ID: "activity-1",
						Artifacts: []jules.Artifact{
							{BashOutput: &jules.BashOutput{Command: "go test ./...", ExitCode: 0}},
						},
					},
				},
			})
		})

	result, output, err := listSessionArtifacts(context.Background(), nil, ListSessionArtifactsInput{SessionID: "session-1"}, client)

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, output.TotalCount)
	assert.Equal(t, "bash_output", output.Artifacts[0].Type)
}

func TestGetSessionOutputsMCP(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.Session{
				ID: "session-1",
				Outputs: []jules.Output{
					{PullRequest: &jules.PullRequest{URL: "https://github.com/acme/widgets/pull/1", Title: "Update widgets"}},
				},
			})
		})

	result, output, err := getSessionOutputs(context.Background(), nil, GetSessionOutputsInput{SessionID: "session-1"}, client)

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, output.TotalCount)
	assert.Equal(t, "https://github.com/acme/widgets/pull/1", output.Outputs[0].PullRequest.URL)
}

func TestGetSessionOutputsFiltersEmptyPayloads(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.Session{
				ID:      "session-1",
				Outputs: []jules.Output{{}},
			})
		})

	result, output, err := getSessionOutputs(context.Background(), nil, GetSessionOutputsInput{SessionID: "session-1"}, client)

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, output.TotalCount)
	assert.Empty(t, output.Outputs)
}
