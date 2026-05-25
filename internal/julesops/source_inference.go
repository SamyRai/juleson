package julesops

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/SamyRai/juleson/pkg/jules"
)

// InferSourceFromGitRemote matches a local git repository's origin remote to a
// connected Jules source.
func InferSourceFromGitRemote(ctx context.Context, client *jules.Client, projectPath string) (*jules.Source, error) {
	owner, repo, err := GitRemoteOwnerRepo(ctx, projectPath)
	if err != nil {
		return nil, err
	}

	sources, err := client.ListAllSources(ctx, 100, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list Jules sources: %w", err)
	}

	var matches []jules.Source
	expectedName := fmt.Sprintf("sources/github/%s/%s", owner, repo)
	for _, source := range sources {
		if source.Name == expectedName || source.ID == fmt.Sprintf("github/%s/%s", owner, repo) {
			matches = append(matches, source)
			continue
		}
		if source.GithubRepo != nil && strings.EqualFold(source.GithubRepo.Owner, owner) && strings.EqualFold(source.GithubRepo.Repo, repo) {
			matches = append(matches, source)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no connected Jules source matches git remote %s/%s", owner, repo)
	case 1:
		return &matches[0], nil
	default:
		names := make([]string, 0, len(matches))
		for _, match := range matches {
			names = append(names, match.Name)
		}
		return nil, fmt.Errorf("multiple Jules sources match git remote %s/%s: %s", owner, repo, strings.Join(names, ", "))
	}
}

// GitRemoteOwnerRepo returns the GitHub owner/repo for origin in a local git repository.
func GitRemoteOwnerRepo(ctx context.Context, projectPath string) (string, string, error) {
	if projectPath == "" {
		projectPath = "."
	}
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to read git origin remote in %s: %w", projectPath, err)
	}
	owner, repo, err := ParseGitHubRemoteURL(strings.TrimSpace(string(output)))
	if err != nil {
		return "", "", err
	}
	return owner, repo, nil
}

// ParseGitHubRemoteURL parses common GitHub HTTPS and SSH remote URL forms.
func ParseGitHubRemoteURL(remoteURL string) (string, string, error) {
	remoteURL = strings.TrimSpace(remoteURL)
	if remoteURL == "" {
		return "", "", fmt.Errorf("git remote URL is empty")
	}

	if strings.HasPrefix(remoteURL, "git@github.com:") {
		return splitOwnerRepo(strings.TrimPrefix(remoteURL, "git@github.com:"))
	}

	sshRE := regexp.MustCompile(`^ssh://git@github\.com[:/](.+)$`)
	if match := sshRE.FindStringSubmatch(remoteURL); len(match) == 2 {
		return splitOwnerRepo(match[1])
	}

	parsed, err := url.Parse(remoteURL)
	if err == nil && strings.EqualFold(parsed.Host, "github.com") {
		return splitOwnerRepo(strings.TrimPrefix(parsed.Path, "/"))
	}

	return "", "", fmt.Errorf("unsupported GitHub remote URL: %s", remoteURL)
}

func splitOwnerRepo(ownerRepo string) (string, string, error) {
	ownerRepo = strings.TrimSuffix(ownerRepo, ".git")
	ownerRepo = strings.Trim(ownerRepo, "/")
	owner := path.Dir(ownerRepo)
	repo := path.Base(ownerRepo)
	if owner == "." || owner == "" || repo == "." || repo == "" {
		return "", "", fmt.Errorf("invalid GitHub owner/repo in remote path: %s", ownerRepo)
	}
	if strings.Contains(owner, "/") {
		return "", "", fmt.Errorf("unsupported nested GitHub owner path: %s", ownerRepo)
	}
	return owner, repo, nil
}
