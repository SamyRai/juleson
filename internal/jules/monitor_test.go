package jules

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MonitorTestSuite defines the test suite for session monitoring
type MonitorTestSuite struct {
	suite.Suite
	client  *Client
	monitor *SessionMonitor
}

// SetupTest is called before each test
func (suite *MonitorTestSuite) SetupTest() {
	httpmock.Activate()
	suite.client = NewClient("test-api-key", "https://api.jules.ai", 30*time.Second, 3)
	suite.monitor = NewSessionMonitor(suite.client, "test-session-1")
}

// TearDownTest is called after each test
func (suite *MonitorTestSuite) TearDownTest() {
	httpmock.DeactivateAndReset()
}

// TestNewSessionMonitor tests monitor creation
func (suite *MonitorTestSuite) TestNewSessionMonitor() {
	monitor := NewSessionMonitor(suite.client, "session-1")

	assert.Equal(suite.T(), suite.client, monitor.client)
	assert.Equal(suite.T(), "session-1", monitor.sessionID)
	assert.Equal(suite.T(), 5*time.Second, monitor.interval)
	assert.Equal(suite.T(), 30*time.Minute, monitor.maxWait)
	assert.Nil(suite.T(), monitor.onProgress)
	assert.Nil(suite.T(), monitor.onComplete)
}

// TestWithInterval tests setting custom interval
func (suite *MonitorTestSuite) TestWithInterval() {
	monitor := suite.monitor.WithInterval(10 * time.Second)

	assert.Equal(suite.T(), 10*time.Second, monitor.interval)
}

// TestWithMaxWait tests setting custom max wait
func (suite *MonitorTestSuite) TestWithMaxWait() {
	monitor := suite.monitor.WithMaxWait(1 * time.Hour)

	assert.Equal(suite.T(), 1*time.Hour, monitor.maxWait)
}

// TestOnProgress tests setting progress callback
func (suite *MonitorTestSuite) TestOnProgress() {
	called := false
	var capturedStatus *SessionStatus

	monitor := suite.monitor.OnProgress(func(status *SessionStatus) {
		called = true
		capturedStatus = status
	})

	assert.NotNil(suite.T(), monitor.onProgress)

	// Test callback
	mockStatus := &SessionStatus{State: SessionStatePlanning}
	monitor.onProgress(mockStatus)

	assert.True(suite.T(), called)
	assert.Equal(suite.T(), SessionStatePlanning, capturedStatus.State)
}

// TestOnComplete tests setting completion callback
func (suite *MonitorTestSuite) TestOnComplete() {
	called := false
	var capturedStatus *SessionStatus

	monitor := suite.monitor.OnComplete(func(status *SessionStatus) {
		called = true
		capturedStatus = status
	})

	assert.NotNil(suite.T(), monitor.onComplete)

	// Test callback
	mockStatus := &SessionStatus{State: SessionStateCompleted}
	monitor.onComplete(mockStatus)

	assert.True(suite.T(), called)
	assert.Equal(suite.T(), SessionStateCompleted, capturedStatus.State)
}

// TestGetSessionStatus tests getting session status
func (suite *MonitorTestSuite) TestGetSessionStatus() {
	mockSession := Session{
		ID:    "test-session-1",
		State: "COMPLETED",
	}

	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/test-session-1",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockSession)
			return resp, nil
		})

	status, err := suite.monitor.getSessionStatus(context.Background())

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), SessionStateCompleted, status.State)
	assert.True(suite.T(), status.IsDone)
	assert.True(suite.T(), status.IsSuccess)
	assert.Equal(suite.T(), "test-session-1", status.Session.ID)
}

// TestGetSessionStatusNotFound tests handling of not found sessions
func (suite *MonitorTestSuite) TestGetSessionStatusNotFound() {
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/test-session-1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(404, `{"error": "Session not found"}`), nil
		})

	status, err := suite.monitor.getSessionStatus(context.Background())

	require.NoError(suite.T(), err) // Should not return error for 404
	assert.Equal(suite.T(), SessionStateFailed, status.State)
	assert.True(suite.T(), status.IsDone)
	assert.False(suite.T(), status.IsSuccess)
	assert.Equal(suite.T(), "session not found", status.Error)
}

// TestWaitForCompletion tests waiting for session completion
func (suite *MonitorTestSuite) TestWaitForCompletion() {
	callCount := 0
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/test-session-1",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			var state string
			if callCount >= 3 {
				state = "COMPLETED"
			} else {
				state = "IN_PROGRESS"
			}

			mockSession := Session{
				ID:    "test-session-1",
				State: state,
			}
			resp, _ := httpmock.NewJsonResponse(200, mockSession)
			return resp, nil
		})

	// Use a short interval for testing
	suite.monitor.WithInterval(10 * time.Millisecond).WithMaxWait(1 * time.Second)

	status, err := suite.monitor.WaitForCompletion(context.Background())

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), SessionStateCompleted, status.State)
	assert.True(suite.T(), status.IsDone)
	assert.True(suite.T(), status.IsSuccess)
	assert.GreaterOrEqual(suite.T(), callCount, 3)
}

// TestWaitForPlan tests waiting for plan generation
func (suite *MonitorTestSuite) TestWaitForPlan() {
	callCount := 0
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/test-session-1",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			mockSession := Session{
				ID:    "test-session-1",
				State: "PLANNING",
			}
			resp, _ := httpmock.NewJsonResponse(200, mockSession)
			return resp, nil
		})

	// Mock activities endpoint to return activities with plan
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/test-session-1/activities?pageSize=10",
		func(req *http.Request) (*http.Response, error) {
			var activities []Activity
			if callCount >= 2 {
				activities = []Activity{
					{
						ID: "activity-1",
						PlanGenerated: &PlanGenerated{
							Plan: Plan{ID: "plan-1"},
						},
					},
				}
			}

			resp, _ := httpmock.NewJsonResponse(200, ActivitiesResponse{Activities: activities})
			return resp, nil
		})

	// Use a short interval for testing
	suite.monitor.WithInterval(10 * time.Millisecond).WithMaxWait(1 * time.Second)

	status, err := suite.monitor.WaitForPlan(context.Background())

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), SessionStatePlanning, status.State)
	assert.GreaterOrEqual(suite.T(), callCount, 2)
}

// TestPollUntilCompleteTimeout tests timeout behavior
func (suite *MonitorTestSuite) TestPollUntilCompleteTimeout() {
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/test-session-1",
		func(req *http.Request) (*http.Response, error) {
			mockSession := Session{
				ID:    "test-session-1",
				State: "IN_PROGRESS", // Never completes
			}
			resp, _ := httpmock.NewJsonResponse(200, mockSession)
			return resp, nil
		})

	// Use very short timeout for testing
	suite.monitor.WithInterval(100 * time.Millisecond).WithMaxWait(10 * time.Millisecond)

	_, err := suite.monitor.WaitForCompletion(context.Background())

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "timeout")
}

// TestContextCancellation tests context cancellation during polling
func (suite *MonitorTestSuite) TestContextCancellation() {
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/test-session-1",
		func(req *http.Request) (*http.Response, error) {
			time.Sleep(100 * time.Millisecond) // Slow response
			mockSession := Session{
				ID:    "test-session-1",
				State: "IN_PROGRESS",
			}
			resp, _ := httpmock.NewJsonResponse(200, mockSession)
			return resp, nil
		})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := suite.monitor.WaitForCompletion(ctx)

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "context")
}

// TestRunMonitorTestSuite runs the test suite
func TestMonitorTestSuite(t *testing.T) {
	suite.Run(t, new(MonitorTestSuite))
}
