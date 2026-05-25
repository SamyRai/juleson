package automation

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/SamyRai/juleson/pkg/jules"
)

// matchGitRepoToSource matches the current git repository to a Jules source.
func (e *Engine) matchGitRepoToSource(sources []jules.Source) (*jules.Source, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = e.projectPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git remote: %w (ensure project is a git repository)", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	var owner, repo string
	if strings.Contains(remoteURL, "github.com") {
		remoteURL = strings.TrimSuffix(remoteURL, ".git")

		if strings.HasPrefix(remoteURL, "https://") {
			parts := strings.Split(remoteURL, "/")
			if len(parts) >= 2 {
				repo = parts[len(parts)-1]
				owner = parts[len(parts)-2]
			}
		} else if strings.HasPrefix(remoteURL, "git@") {
			parts := strings.Split(remoteURL, ":")
			if len(parts) == 2 {
				pathParts := strings.Split(parts[1], "/")
				if len(pathParts) == 2 {
					owner = pathParts[0]
					repo = pathParts[1]
				}
			}
		}
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("failed to parse GitHub owner/repo from remote URL: %s", remoteURL)
	}

	expectedSourceName := fmt.Sprintf("sources/github/%s/%s", owner, repo)
	for i := range sources {
		if sources[i].Name == expectedSourceName {
			return &sources[i], nil
		}
	}

	return nil, fmt.Errorf("repository %s/%s not found in connected sources - connect it via Jules web UI at https://jules.google.com", owner, repo)
}
