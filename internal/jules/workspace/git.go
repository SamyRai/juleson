package workspace

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitClient abstracts git interactions for patch application.
type GitClient interface {
	ApplyPatch(ctx context.Context, patchPath string, dryRun bool, stripComponents int, force bool) ([]string, error)
	GetHeadCommit(ctx context.Context) (string, error)
	IsClean(ctx context.Context) (bool, string, error)
}

type execGitClient struct {
	workingDir string
}

// NewGitClient returns a GitClient that uses os/exec.
func NewGitClient(workingDir string) GitClient {
	return &execGitClient{workingDir: workingDir}
}

func (c *execGitClient) ApplyPatch(ctx context.Context, patchPath string, dryRun bool, stripComponents int, force bool) ([]string, error) {
	args := []string{"apply"}
	if dryRun {
		args = append(args, "--check")
	}
	args = append(args, fmt.Sprintf("-p%d", stripComponents))
	if force {
		args = append(args, "--3way")
	}
	args = append(args, "--verbose", patchPath)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = c.workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git apply failed: %w\nOutput: %s", err, string(output))
	}

	return parseGitApplyOutput(string(output)), nil
}

func (c *execGitClient) GetHeadCommit(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = c.workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to resolve target HEAD: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func (c *execGitClient) IsClean(ctx context.Context) (bool, string, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = c.workingDir
	output, err := cmd.CombinedOutput()
	status := strings.TrimSpace(string(output))
	if err != nil {
		return false, status, fmt.Errorf("git status failed: %w\nOutput: %s", err, status)
	}
	return status == "", status, nil
}
