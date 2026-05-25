package julesops

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SamyRai/juleson/pkg/jules"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ArtifactsTestSuite defines the test suite for artifact operations
type ArtifactsTestSuite struct {
	suite.Suite
	client *jules.Client
}

// SetupTest is called before each test
func (suite *ArtifactsTestSuite) SetupTest() {
	httpmock.Activate()
	suite.client = jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(3))
}

// TearDownTest is called after each test
func (suite *ArtifactsTestSuite) TearDownTest() {
	httpmock.DeactivateAndReset()
}

// TestDownloadArtifactFromActivity tests downloading artifacts from an activity
func (suite *ArtifactsTestSuite) TestDownloadArtifactFromActivity() {
	// Create temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "jules_test_*")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(tempDir)

	// Mock activity endpoint
	mockActivity := Activity{
		ID: "activity-1",
		Artifacts: []Artifact{
			{BashOutput: &BashOutput{Command: "echo hello", Output: "hello"}},
		},
	}
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities/activity-1",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockActivity)
			return resp, nil
		})

	options := &ArtifactDownloadOptions{
		DestinationDir: tempDir,
		Overwrite:      true,
		CreateDir:      true,
	}
	files, err := DownloadArtifactFromActivity(context.Background(), suite.client, "session-1", "activity-1", options)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), files, 1)
	assert.Contains(suite.T(), files[0], "bash_output_0.txt")

	// Verify file was created
	filePath := filepath.Join(tempDir, files[0])
	assert.FileExists(suite.T(), filePath)
	content, err := os.ReadFile(filePath)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(content), "echo hello")
	assert.Contains(suite.T(), string(content), "hello")
}

// TestDownloadAllSessionArtifacts tests downloading all artifacts from a session
func (suite *ArtifactsTestSuite) TestDownloadAllSessionArtifacts() {
	// Create temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "jules_test_*")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(tempDir)

	// Mock activities list
	mockActivities := []Activity{
		{
			ID: "activity-1",
			Artifacts: []Artifact{
				{BashOutput: &BashOutput{Command: "echo hello", Output: "hello"}},
			},
		},
	}
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, ActivitiesResponse{Activities: mockActivities})
			return resp, nil
		})

	// Mock activity details
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities/activity-1",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockActivities[0])
			return resp, nil
		})

	options := &ArtifactDownloadOptions{
		DestinationDir: tempDir,
		Overwrite:      true,
		CreateDir:      true,
	}
	files, err := DownloadAllSessionArtifacts(context.Background(), suite.client, "session-1", options)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), files, 1)

	// Verify file was created
	filePath := filepath.Join(tempDir, files[0])
	assert.FileExists(suite.T(), filePath)
}

func (suite *ArtifactsTestSuite) TestDownloadMediaFromEmbeddedBase64() {
	tempDir, err := os.MkdirTemp("", "jules_test_*")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(tempDir)

	mockActivity := Activity{
		ID: "activity-1",
		Artifacts: []Artifact{
			{Media: &Media{MimeType: "image/png", Data: "aGVsbG8="}},
		},
	}
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities/activity-1",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockActivity)
			return resp, nil
		})

	files, err := DownloadArtifactFromActivity(context.Background(), suite.client, "session-1", "activity-1", &ArtifactDownloadOptions{
		DestinationDir: tempDir,
		Overwrite:      true,
		CreateDir:      true,
	})

	require.NoError(suite.T(), err)
	require.Len(suite.T(), files, 1)
	content, err := os.ReadFile(filepath.Join(tempDir, files[0]))
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), []byte("hello"), content)

	for request := range httpmock.GetCallCountInfo() {
		assert.NotContains(suite.T(), request, "/artifacts/")
	}
}

// TestGetArtifactsFromActivity tests getting artifacts from an activity
func (suite *ArtifactsTestSuite) TestGetArtifactsFromActivity() {
	mockActivity := Activity{
		ID: "activity-1",
		Artifacts: []Artifact{
			{BashOutput: &BashOutput{Command: "echo hello", Output: "hello"}},
			{ChangeSet: &ChangeSet{GitPatch: &GitPatch{UnidiffPatch: "diff content"}}},
		},
	}
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities/activity-1",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockActivity)
			return resp, nil
		})

	artifacts, err := suite.client.GetArtifactsFromActivity(context.Background(), "session-1", "activity-1")

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), artifacts, 2)
	assert.NotNil(suite.T(), artifacts[0].BashOutput)
	assert.NotNil(suite.T(), artifacts[1].ChangeSet)
}

// TestGetAllSessionArtifacts tests getting all artifacts from a session
func (suite *ArtifactsTestSuite) TestGetAllSessionArtifacts() {
	mockActivities := []Activity{
		{
			ID: "activity-1",
			Artifacts: []Artifact{
				{BashOutput: &BashOutput{Command: "echo hello"}},
			},
		},
		{
			ID: "activity-2",
			Artifacts: []Artifact{
				{Media: &Media{Data: "image data", MimeType: "image/png"}},
			},
		},
	}
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, ActivitiesResponse{Activities: mockActivities})
			return resp, nil
		})

	artifacts, err := suite.client.GetAllSessionArtifacts(context.Background(), "session-1")

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), artifacts, 2)
	assert.Equal(suite.T(), "activity-1", artifacts[0].ActivityID)
	assert.Equal(suite.T(), 0, artifacts[0].Index)
	assert.Equal(suite.T(), "activity-2", artifacts[1].ActivityID)
	assert.Equal(suite.T(), 0, artifacts[1].Index)
}

// TestGenerateArtifactFilename tests filename generation for different artifact types
func (suite *ArtifactsTestSuite) TestGenerateArtifactFilename() {
	testCases := []struct {
		name     string
		artifact Artifact
		index    int
		expected string
	}{
		{
			name: "bash output",
			artifact: Artifact{
				BashOutput: &BashOutput{Command: "echo hello"},
			},
			index:    0,
			expected: "bash_output_0.txt",
		},
		{
			name: "change set with patch",
			artifact: Artifact{
				ChangeSet: &ChangeSet{
					GitPatch: &GitPatch{UnidiffPatch: "diff content"},
				},
			},
			index:    1,
			expected: "changeset_1.patch",
		},
		{
			name: "change set without patch",
			artifact: Artifact{
				ChangeSet: &ChangeSet{Source: "file.txt"},
			},
			index:    2,
			expected: "changeset_2.txt",
		},
		{
			name: "media PNG",
			artifact: Artifact{
				Media: &Media{MimeType: "image/png"},
			},
			index:    3,
			expected: "media_3.png",
		},
		{
			name: "media JPEG",
			artifact: Artifact{
				Media: &Media{MimeType: "image/jpeg"},
			},
			index:    4,
			expected: "media_4.jpg",
		},
		{
			name:     "unknown type",
			artifact: Artifact{},
			index:    5,
			expected: "artifact_5.bin",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result := GenerateArtifactFilename(tc.artifact, tc.index)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGetExtensionFromMimeType tests MIME type to extension conversion
func (suite *ArtifactsTestSuite) TestGetExtensionFromMimeType() {
	testCases := []struct {
		mimeType string
		expected string
	}{
		{"image/png", ".png"},
		{"image/jpeg", ".jpg"},
		{"image/jpg", ".jpg"},
		{"image/gif", ".gif"},
		{"application/json", ".json"},
		{"text/plain", ".txt"},
		{"application/octet-stream", ".bin"},
		{"unknown/type", ".bin"},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.mimeType, func(t *testing.T) {
			result := extensionFromMimeType(tc.mimeType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func (suite *ArtifactsTestSuite) TestSessionHasDeliverablesNilSession() {
	hasDeliverables, err := SessionHasDeliverables(context.Background(), suite.client, nil)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), hasDeliverables)
}

func (suite *ArtifactsTestSuite) TestSessionHasDeliverablesPullRequestOutput() {
	hasDeliverables, err := SessionHasDeliverables(context.Background(), suite.client, &jules.Session{
		ID: "session-1",
		Outputs: []jules.Output{
			{PullRequest: &jules.PullRequest{URL: "https://github.com/SamyRai/juleson/pull/1"}},
		},
	})

	require.NoError(suite.T(), err)
	assert.True(suite.T(), hasDeliverables)
}

func (suite *ArtifactsTestSuite) TestSessionHasDeliverablesNonEmptyPatch() {
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, ActivitiesResponse{
				Activities: []Activity{
					{
						ID: "activity-1",
						Artifacts: []Artifact{
							{ChangeSet: &ChangeSet{GitPatch: &GitPatch{UnidiffPatch: "diff --git a/file b/file\n"}}},
						},
					},
				},
			})
		})

	hasDeliverables, err := SessionHasDeliverables(context.Background(), suite.client, &jules.Session{ID: "session-1"})

	require.NoError(suite.T(), err)
	assert.True(suite.T(), hasDeliverables)
}

func (suite *ArtifactsTestSuite) TestSessionHasDeliverablesEmptyPatch() {
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, ActivitiesResponse{
				Activities: []Activity{
					{
						ID: "activity-1",
						Artifacts: []Artifact{
							{ChangeSet: &ChangeSet{GitPatch: &GitPatch{UnidiffPatch: "   \n"}}},
						},
					},
				},
			})
		})

	hasDeliverables, err := SessionHasDeliverables(context.Background(), suite.client, &jules.Session{ID: "session-1"})

	require.NoError(suite.T(), err)
	assert.False(suite.T(), hasDeliverables)
}

func (suite *ArtifactsTestSuite) TestSessionHasDeliverablesFindsPatchOnLaterPage() {
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, ActivitiesResponse{
				NextPageToken: "next",
				Activities: []Activity{
					{
						ID: "activity-1",
						Artifacts: []Artifact{
							{ChangeSet: &ChangeSet{GitPatch: &GitPatch{UnidiffPatch: ""}}},
						},
					},
				},
			})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=100&pageToken=next",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, ActivitiesResponse{
				Activities: []Activity{
					{
						ID: "activity-2",
						Artifacts: []Artifact{
							{ChangeSet: &ChangeSet{GitPatch: &GitPatch{UnidiffPatch: "diff --git a/file b/file\n"}}},
						},
					},
				},
			})
		})

	hasDeliverables, err := SessionHasDeliverables(context.Background(), suite.client, &jules.Session{ID: "session-1"})

	require.NoError(suite.T(), err)
	assert.True(suite.T(), hasDeliverables)
}

// TestRunArtifactsTestSuite runs the test suite
func TestArtifactsTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactsTestSuite))
}
