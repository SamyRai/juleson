package jules

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ArtifactsTestSuite defines the test suite for artifact operations
type ArtifactsTestSuite struct {
	suite.Suite
	client *Client
}

// SetupTest is called before each test
func (suite *ArtifactsTestSuite) SetupTest() {
	httpmock.Activate()
	suite.client = NewClient("test-api-key", "https://api.jules.ai", 30*time.Second, 3)
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
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1/activities/activity-1",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockActivity)
			return resp, nil
		})

	// Mock download endpoint
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1/activities/activity-1/artifacts/0/download",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "artifact content"), nil
		})

	options := &ArtifactDownloadOptions{
		DestinationDir: tempDir,
		Overwrite:      true,
		CreateDir:      true,
	}
	files, err := suite.client.DownloadArtifactFromActivity(context.Background(), "session-1", "activity-1", options)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), files, 1)
	assert.Contains(suite.T(), files[0], "bash_output_0.txt")

	// Verify file was created
	filePath := filepath.Join(tempDir, files[0])
	assert.FileExists(suite.T(), filePath)
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
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1/activities?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, ActivitiesResponse{Activities: mockActivities})
			return resp, nil
		})

	// Mock activity details
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1/activities/activity-1",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockActivities[0])
			return resp, nil
		})

	// Mock download endpoint
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1/activities/activity-1/artifacts/0/download",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "artifact content"), nil
		})

	options := &ArtifactDownloadOptions{
		DestinationDir: tempDir,
		Overwrite:      true,
		CreateDir:      true,
	}
	files, err := suite.client.DownloadAllSessionArtifacts(context.Background(), "session-1", options)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), files, 1)

	// Verify file was created
	filePath := filepath.Join(tempDir, files[0])
	assert.FileExists(suite.T(), filePath)
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
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1/activities/activity-1",
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
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1/activities?pageSize=100",
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

// TestAnalyzeArtifact tests analyzing an artifact
func (suite *ArtifactsTestSuite) TestAnalyzeArtifact() {
	mockAnalysis := ArtifactAnalysis{
		ActivityID:    "activity-1",
		ArtifactIndex: 0,
		ContentType:   "text/plain",
		Size:          1024,
		Language:      "bash",
		Summary:       "Bash script output",
		KeyInsights:   []string{"Command executed successfully"},
		Issues:        []ArtifactIssue{},
	}

	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1/activities/activity-1/artifacts/0/analyze",
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(200, mockAnalysis)
			return resp, nil
		})

	analysis, err := suite.client.AnalyzeArtifact(context.Background(), "session-1", "activity-1", 0)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "activity-1", analysis.ActivityID)
	assert.Equal(suite.T(), 0, analysis.ArtifactIndex)
	assert.Equal(suite.T(), "text/plain", analysis.ContentType)
	assert.Equal(suite.T(), "bash", analysis.Language)
}

// TestGetArtifactContent tests getting raw artifact content
func (suite *ArtifactsTestSuite) TestGetArtifactContent() {
	content := "This is the artifact content"
	httpmock.RegisterResponder("GET", "https://api.jules.ai/sessions/session-1/activities/activity-1/artifacts/0/content",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, content), nil
		})

	result, err := suite.client.GetArtifactContent(context.Background(), "session-1", "activity-1", 0)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), content, string(result))
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
			result := suite.client.generateArtifactFilename(tc.artifact, tc.index)
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
			result := suite.client.getExtensionFromMimeType(tc.mimeType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestRunArtifactsTestSuite runs the test suite
func TestArtifactsTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactsTestSuite))
}
