package jules

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func ExampleClient_CreateSession() {
	client, cleanup := exampleClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name":"sessions/1","id":"1","state":"QUEUED","createTime":"2026-01-26T09:00:00Z"}`)
	}))
	defer cleanup()

	session, _ := client.CreateSession(context.Background(), &CreateSessionRequest{
		Prompt: "Add tests for authentication",
		SourceContext: &SourceContext{
			Source: "github/owner/repo",
			GithubRepoContext: &GithubRepoContext{
				StartingBranch: "main",
			},
		},
		AutomationMode: AutomationModeAutoCreatePR,
	})

	fmt.Println(session.ID, session.State)
	// Output: 1 QUEUED
}

func ExampleClient_ListActivitiesSince() {
	client, cleanup := exampleClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"activities":[{"name":"sessions/1/activities/a1","id":"a1","originator":"agent","createTime":"2026-01-26T09:01:00Z"}]}`)
	}))
	defer cleanup()

	cursor := time.Date(2026, 1, 26, 9, 0, 0, 0, time.UTC)
	activities, _ := client.ListActivitiesSince(context.Background(), "sessions/1", cursor, 50)

	fmt.Println(len(activities), ActivityCursor(activities).Format(time.RFC3339))
	// Output: 1 2026-01-26T09:01:00Z
}

func ExampleSessionMonitor_WaitForCompletion() {
	calls := 0
	client, cleanup := exampleClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		state := SessionStateInProgress
		if calls > 1 {
			state = SessionStateCompleted
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"name":"sessions/1","id":"1","state":%q}`, state)
	}))
	defer cleanup()

	status, _ := NewSessionMonitor(client, "1").
		WithInterval(time.Millisecond).
		WithMaxWait(time.Second).
		WaitForCompletion(context.Background())

	fmt.Println(status.State, status.IsSuccess)
	// Output: COMPLETED true
}

func ExampleArtifactContent() {
	content, _ := ArtifactContent(Artifact{
		BashOutput: &BashOutput{
			Command:  "go test ./...",
			Output:   "ok ./...",
			ExitCode: 0,
		},
	})

	fmt.Print(string(content))
	// Output:
	// $ go test ./...
	// ok ./...
	// exit code: 0
}

func exampleClient(handler http.Handler) (*Client, func()) {
	server := httptest.NewServer(handler)
	client := NewClient("test-api-key", WithBaseURL(server.URL), WithRetryAttempts(0))
	return client, server.Close
}
