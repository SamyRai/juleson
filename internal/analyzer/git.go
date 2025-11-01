package analyzer

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// GitAnalyzer analyzes git repository status
type GitAnalyzer struct{}

// NewGitAnalyzer creates a new git analyzer
func NewGitAnalyzer() *GitAnalyzer {
	return &GitAnalyzer{}
}

// GetStatus returns the git status of the project
func (g *GitAnalyzer) GetStatus(projectPath string) (string, error) {
	// Check if .git directory exists
	gitDir := filepath.Join(projectPath, ".git")
	cmd := exec.Command("test", "-d", gitDir)
	if err := cmd.Run(); err != nil {
		return "not-a-git-repo", nil
	}

	// Get git status
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return "unknown", err
	}

	status := strings.TrimSpace(string(output))
	if status == "" {
		return "clean", nil
	}

	// Count changes
	lines := strings.Split(status, "\n")
	if len(lines) > 10 {
		return "many-changes", nil
	}

	return "has-changes", nil
}

// GetBranch returns the current git branch
func (g *GitAnalyzer) GetBranch(projectPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
