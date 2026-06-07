package github

import (
	"testing"
	"time"

	"github.com/google/go-github/v76/github"
	"github.com/stretchr/testify/assert"
)

func TestMapGitHubRepo(t *testing.T) {
	// Setup service (client is not needed for mapping)
	svc := NewRepositoryService(nil, nil)

	login := "SamyRai"
	name := "juleson"
	fullName := "SamyRai/juleson"
	desc := "A great project"
	stars := 10
	forks := 2
	branch := "main"
	private := false
	url := "https://github.com/SamyRai/juleson"
	now := time.Now()
	githubTime := &github.Timestamp{Time: now}

	ghRepo := &github.Repository{
		Owner: &github.User{
			Login: &login,
		},
		Name:            &name,
		FullName:        &fullName,
		Description:     &desc,
		StargazersCount: &stars,
		ForksCount:      &forks,
		DefaultBranch:   &branch,
		Private:         &private,
		HTMLURL:         &url,
		UpdatedAt:       githubTime,
	}

	repo := svc.mapGitHubRepo(ghRepo)

	assert.NotNil(t, repo)
	assert.Equal(t, login, repo.Owner)
	assert.Equal(t, name, repo.Name)
	assert.Equal(t, fullName, repo.FullName)
	assert.Equal(t, desc, repo.Description)
	assert.Equal(t, stars, repo.Stars)
	assert.Equal(t, forks, repo.Forks)
	assert.Equal(t, branch, repo.DefaultBranch)
	assert.Equal(t, private, repo.Private)
	assert.Equal(t, url, repo.URL)
	assert.Equal(t, now.Format("2006-01-02T15:04:05Z"), repo.UpdatedAt)
}
