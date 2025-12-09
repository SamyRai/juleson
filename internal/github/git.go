package github

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitRemoteParser handles parsing of Git remote URLs and repository detection
type GitRemoteParser struct{}

// NewGitRemoteParser creates a new git remote parser
func NewGitRemoteParser() *GitRemoteParser {
	return &GitRemoteParser{}
}

// GetRepoFromGitRemote detects the GitHub repository from the current directory's git remote
func (p *GitRemoteParser) GetRepoFromGitRemote() (*Repository, error) {
	// Run git remote -v to get remote URLs
	cmd := exec.Command("git", "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git remote: %w", err)
	}

	// Parse the output to find origin remote
	lines := strings.Split(string(output), "\n")
	var remoteURL string

	for _, line := range lines {
		// Look for origin remote (fetch)
		if strings.Contains(line, "origin") && strings.Contains(line, "(fetch)") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				remoteURL = parts[1]
				break
			}
		}
	}

	if remoteURL == "" {
		return nil, fmt.Errorf("no origin remote found in git repository")
	}

	// Parse GitHub URL
	return p.ParseGitHubURL(remoteURL)
}

// ParseGitHubURL parses a GitHub URL and extracts owner and repository name
// Supports both HTTPS and SSH URL formats:
// - https://github.com/owner/repo.git
// - git@github.com:owner/repo.git
func (p *GitRemoteParser) ParseGitHubURL(remoteURL string) (*Repository, error) {
	// Remove .git suffix if present
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	var owner, repo string

	if strings.HasPrefix(remoteURL, "https://github.com/") {
		// HTTPS URL: https://github.com/owner/repo
		path := strings.TrimPrefix(remoteURL, "https://github.com/")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			owner = parts[0]
			repo = parts[1]
		}
	} else if strings.HasPrefix(remoteURL, "git@github.com:") {
		// SSH URL: git@github.com:owner/repo
		path := strings.TrimPrefix(remoteURL, "git@github.com:")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			owner = parts[0]
			repo = parts[1]
		}
	} else {
		return nil, fmt.Errorf("unsupported GitHub URL format: %s", remoteURL)
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("failed to parse owner/repo from URL: %s", remoteURL)
	}

	return &Repository{
		Owner:    owner,
		Name:     repo,
		FullName: fmt.Sprintf("%s/%s", owner, repo),
	}, nil
}
