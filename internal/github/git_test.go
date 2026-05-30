package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGitHubURL(t *testing.T) {
	parser := NewGitRemoteParser()

	cases := []struct {
		name        string
		url         string
		expectError bool
		owner       string
		repo        string
	}{
		{
			name:        "https valid",
			url:         "https://github.com/SamyRai/juleson.git",
			expectError: false,
			owner:       "SamyRai",
			repo:        "juleson",
		},
		{
			name:        "https without git",
			url:         "https://github.com/SamyRai/juleson",
			expectError: false,
			owner:       "SamyRai",
			repo:        "juleson",
		},
		{
			name:        "ssh valid",
			url:         "git@github.com:SamyRai/juleson.git",
			expectError: false,
			owner:       "SamyRai",
			repo:        "juleson",
		},
		{
			name:        "unsupported url",
			url:         "https://gitlab.com/SamyRai/juleson.git",
			expectError: true,
		},
		{
			name:        "invalid https format",
			url:         "https://github.com/SamyRai", // Missing repo
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, err := parser.ParseGitHubURL(tc.url)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, repo)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, repo)
				assert.Equal(t, tc.owner, repo.Owner)
				assert.Equal(t, tc.repo, repo.Name)
				assert.Equal(t, tc.owner+"/"+tc.repo, repo.FullName)
			}
		})
	}
}
