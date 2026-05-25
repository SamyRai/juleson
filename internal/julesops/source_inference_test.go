package julesops

import (
	"context"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGitHubRemoteURL(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		owner     string
		repo      string
	}{
		{name: "https", remoteURL: "https://github.com/acme/widgets.git", owner: "acme", repo: "widgets"},
		{name: "ssh scp", remoteURL: "git@github.com:acme/widgets.git", owner: "acme", repo: "widgets"},
		{name: "ssh url", remoteURL: "ssh://git@github.com/acme/widgets.git", owner: "acme", repo: "widgets"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseGitHubRemoteURL(tt.remoteURL)
			require.NoError(t, err)
			assert.Equal(t, tt.owner, owner)
			assert.Equal(t, tt.repo, repo)
		})
	}
}

func TestInferSourceFromGitRemote(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	tmpDir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "remote", "add", "origin", "git@github.com:acme/widgets.git")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(0))
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sources?pageSize=100",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.SourcesResponse{
				Sources: []jules.Source{
					{Name: "sources/github/acme/widgets", ID: "github/acme/widgets"},
				},
			})
		})

	source, err := InferSourceFromGitRemote(context.Background(), client, tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "sources/github/acme/widgets", source.Name)
}
