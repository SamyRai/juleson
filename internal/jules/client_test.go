package jules

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ClientTestSuite defines the test suite for Jules client
type ClientTestSuite struct {
	suite.Suite
	client *Client
}

// SetupTest is called before each test
func (suite *ClientTestSuite) SetupTest() {
	httpmock.Activate()
	suite.client = NewClient("test-api-key", "https://api.jules.ai", 30*time.Second, 3)
}

// TearDownTest is called after each test
func (suite *ClientTestSuite) TearDownTest() {
	httpmock.DeactivateAndReset()
}

// TestNewClient tests client creation
func (suite *ClientTestSuite) TestNewClient() {
	client := NewClient("api-key", "https://api.example.com", 10*time.Second, 2)

	assert.Equal(suite.T(), "api-key", client.APIKey)
	assert.Equal(suite.T(), "https://api.example.com", client.BaseURL)
	assert.Equal(suite.T(), 2, client.RetryAttempts)
	assert.NotNil(suite.T(), client.HTTPClient)
}

// TestListSessions tests listing sessions
func (suite *ClientTestSuite) TestListSessions() {
	mockResponse := SessionsResponse{
		Sessions: []Session{
			{
				ID:    "session-1",
				Title: "Test Session",
				State: "COMPLETED",
			},
		},
		NextPageToken: "",
	}

	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions?pageSize=10",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockResponse)
			return resp, nil
		})

	sessions, err := suite.client.ListSessions(context.Background(), 10)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 1)
	assert.Equal(suite.T(), "session-1", sessions[0].ID)
	assert.Equal(suite.T(), "Test Session", sessions[0].Title)
}

// TestListSessionsWithPagination tests listing sessions with pagination
func (suite *ClientTestSuite) TestListSessionsWithPagination() {
	mockResponse := SessionsResponse{
		Sessions: []Session{
			{
				ID:    "session-1",
				Title: "Test Session",
				State: "COMPLETED",
			},
		},
		NextPageToken: "next-token",
	}

	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions?pageSize=5&pageToken=test-token",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockResponse)
			return resp, nil
		})

	sessions, err := suite.client.ListSessionsWithPagination(context.Background(), 5, "test-token")

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions.Sessions, 1)
	assert.Equal(suite.T(), "next-token", sessions.NextPageToken)
}

// TestGetSession tests getting a specific session
func (suite *ClientTestSuite) TestGetSession() {
	mockSession := Session{
		ID:         "session-1",
		Title:      "Test Session",
		State:      "COMPLETED",
		CreateTime: "2024-01-01T00:00:00Z",
		UpdateTime: "2024-01-01T01:00:00Z",
		Prompt:     "Test prompt",
		URL:        "https://app.jules.ai/sessions/session-1",
	}

	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockSession)
			return resp, nil
		})

	session, err := suite.client.GetSession(context.Background(), "session-1")

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "session-1", session.ID)
	assert.Equal(suite.T(), "Test Session", session.Title)
	assert.Equal(suite.T(), "COMPLETED", session.State)
}

// TestCreateSession tests creating a new session
func (suite *ClientTestSuite) TestCreateSession() {
	request := CreateSessionRequest{
		Prompt:              "Create a new feature",
		RequirePlanApproval: true,
		AutomationMode:      "AUTO_CREATE_PR",
	}

	expectedResponse := Session{
		ID:    "new-session-1",
		Title: "New Session",
		State: "PLANNING",
	}

	httpmock.RegisterResponder("POST", "https://api.jules.ai/sessions",
		func(req *http.Request) (*http.Response, error) {
			// Verify request body
			var receivedRequest CreateSessionRequest
			body, _ := io.ReadAll(req.Body)
			json.Unmarshal(body, &receivedRequest)

			assert.Equal(suite.T(), request.Prompt, receivedRequest.Prompt)
			assert.Equal(suite.T(), request.RequirePlanApproval, receivedRequest.RequirePlanApproval)
			assert.Equal(suite.T(), request.AutomationMode, receivedRequest.AutomationMode)

			resp, _ := httpmock.NewJsonResponse(201, expectedResponse)
			return resp, nil
		})

	session, err := suite.client.CreateSession(context.Background(), &request)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "new-session-1", session.ID)
	assert.Equal(suite.T(), "New Session", session.Title)
}

// TestSendMessage tests sending a message to a session
func (suite *ClientTestSuite) TestSendMessage() {
	request := SendMessageRequest{
		Prompt: "Please implement this feature",
	}

	httpmock.RegisterResponder("POST", "https://api.jules.ai/sessions/session-1:sendMessage",
		func(req *http.Request) (*http.Response, error) {
			var receivedRequest SendMessageRequest
			body, _ := io.ReadAll(req.Body)
			json.Unmarshal(body, &receivedRequest)

			assert.Equal(suite.T(), request.Prompt, receivedRequest.Prompt)

			return httpmock.NewStringResponse(200, ""), nil
		})

	err := suite.client.SendMessage(context.Background(), "session-1", &request)

	require.NoError(suite.T(), err)
}

// TestApprovePlan tests approving a plan
func (suite *ClientTestSuite) TestApprovePlan() {
	httpmock.RegisterResponder("POST", "https://api.jules.ai/sessions/session-1:approvePlan",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, ""), nil
		})

	err := suite.client.ApprovePlan(context.Background(), "session-1")

	require.NoError(suite.T(), err)
}

// TestCancelSession tests canceling a session
func (suite *ClientTestSuite) TestCancelSession() {
	httpmock.RegisterResponder("POST", "https://api.jules.ai/sessions/session-1:cancel",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, ""), nil
		})

	err := suite.client.CancelSession(context.Background(), "session-1")

	require.NoError(suite.T(), err)
}

// TestDeleteSession tests deleting a session
func (suite *ClientTestSuite) TestDeleteSession() {
	httpmock.RegisterResponder("DELETE", "https://api.jules.ai/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(204, ""), nil
		})

	err := suite.client.DeleteSession(context.Background(), "session-1")

	require.NoError(suite.T(), err)
}

// TestListActivities tests listing activities
func (suite *ClientTestSuite) TestListActivities() {
	mockResponse := ActivitiesResponse{
		Activities: []Activity{
			{
				ID:         "activity-1",
				Name:       "Plan Generated",
				Originator: "agent",
				PlanGenerated: &PlanGenerated{
					Plan: Plan{
						ID: "plan-1",
						Steps: []Step{
							{ID: "step-1", Title: "Step 1", Index: 1},
						},
					},
				},
			},
		},
		NextPageToken: "",
	}

	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1/activities?pageSize=10",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockResponse)
			return resp, nil
		})

	activities, err := suite.client.ListActivities(context.Background(), "session-1", 10)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), activities, 1)
	assert.Equal(suite.T(), "activity-1", activities[0].ID)
	assert.Equal(suite.T(), "Plan Generated", activities[0].Name)
}

// TestGetActivity tests getting a specific activity
func (suite *ClientTestSuite) TestGetActivity() {
	mockActivity := Activity{
		ID:         "activity-1",
		Name:       "Plan Generated",
		Originator: "agent",
		CreateTime: "2024-01-01T00:00:00Z",
	}

	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1/activities/activity-1",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockActivity)
			return resp, nil
		})

	activity, err := suite.client.GetActivity(context.Background(), "session-1", "activity-1")

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "activity-1", activity.ID)
	assert.Equal(suite.T(), "Plan Generated", activity.Name)
}

// TestListSources tests listing sources
func (suite *ClientTestSuite) TestListSources() {
	mockResponse := SourcesResponse{
		Sources: []Source{
			{
				ID:   "source-1",
				Name: "test-repo",
				GithubRepo: &GithubRepo{
					Owner: "testuser",
					Repo:  "test-repo",
				},
			},
		},
		NextPageToken: "",
	}

	httpmock.RegisterResponder("GET", "https://api.jules.ai/sources?pageSize=10",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockResponse)
			return resp, nil
		})

	sources, err := suite.client.ListSources(context.Background(), 10)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), sources, 1)
	assert.Equal(suite.T(), "source-1", sources[0].ID)
	assert.Equal(suite.T(), "test-repo", sources[0].Name)
}

// TestGetSource tests getting a specific source
func (suite *ClientTestSuite) TestGetSource() {
	mockSource := Source{
		ID:   "source-1",
		Name: "test-repo",
		GithubRepo: &GithubRepo{
			Owner: "testuser",
			Repo:  "test-repo",
		},
	}

	httpmock.RegisterResponder("GET", "https://api.jules.ai/sources/source-1",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockSource)
			return resp, nil
		})

	source, err := suite.client.GetSource(context.Background(), "source-1")

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "source-1", source.ID)
	assert.Equal(suite.T(), "test-repo", source.Name)
}

// TestErrorHandling tests error handling for API errors
func (suite *ClientTestSuite) TestErrorHandling() {
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(404, `{"error": "Session not found"}`), nil
		})

	_, err := suite.client.GetSession(context.Background(), "session-1")

	require.Error(suite.T(), err)
	var apiErr *APIError
	assert.ErrorAs(suite.T(), err, &apiErr)
	assert.True(suite.T(), apiErr.IsNotFound())
}

// TestRetryLogic tests retry logic for failed requests
func (suite *ClientTestSuite) TestRetryLogic() {
	callCount := 0
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount < 3 {
				return httpmock.NewStringResponse(500, "Internal Server Error"), nil
			}
			mockSession := Session{ID: "session-1", Title: "Test Session"}
			resp, _ := httpmock.NewJsonResponse(200, mockSession)
			return resp, nil
		})

	session, err := suite.client.GetSession(context.Background(), "session-1")

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "session-1", session.ID)
	assert.Equal(suite.T(), 3, callCount) // Should have retried 3 times
}

// TestContextCancellation tests context cancellation
func (suite *ClientTestSuite) TestContextCancellation() {
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1",
		func(req *http.Request) (*http.Response, error) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			mockSession := Session{ID: "session-1", Title: "Test Session"}
			resp, _ := httpmock.NewJsonResponse(200, mockSession)
			return resp, nil
		})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := suite.client.GetSession(ctx, "session-1")

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "context")
}

// TestRunClientTestSuite runs the test suite
func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
