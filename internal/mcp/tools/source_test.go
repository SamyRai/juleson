package tools

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSourcesMCP(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sources?pageSize=2",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.SourcesResponse{
				Sources: []jules.Source{
					{Name: "sources/github/owner/repo", ID: "github/owner/repo"},
				},
				NextPageToken: "next",
			})
		})

	result, output, err := listSourcesMCP(context.Background(), ListSourcesInput{PageSize: 2}, client)

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, output.TotalCount)
	assert.Equal(t, "next", output.NextPageToken)
	assert.Equal(t, "sources/github/owner/repo", output.Sources[0].Name)
}

func TestGetSourceMCP(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sources/github/owner/repo",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.Source{
				Name: "sources/github/owner/repo",
				ID:   "github/owner/repo",
			})
		})

	result, output, err := getSourceMCP(context.Background(), GetSourceInput{SourceID: "github/owner/repo"}, client)

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "sources/github/owner/repo", output.Source.Name)
}
