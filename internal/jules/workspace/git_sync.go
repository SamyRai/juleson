package workspace

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type GitSyncOptions struct {
	Stdout      io.Writer
	Stderr      io.Writer
	ProjectPath string
	Remote      string
	Branch      string
	Pull        bool
	Push        bool
}

func SyncGitRepository(ctx context.Context, options GitSyncOptions) error {
	absPath, err := filepath.Abs(options.ProjectPath)
	if err != nil {
		return fmt.Errorf("invalid project path: %w", err)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", absPath)
	}
	if _, err := os.Stat(filepath.Join(absPath, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository: %s", absPath)
	}

	branch := options.Branch
	if branch == "" {
		branch = "main"
	}

	if options.Pull {
		if err := runGit(ctx, absPath, options.Stdout, options.Stderr, "pull", options.Remote, branch); err != nil {
			return fmt.Errorf("failed to pull changes: %w", err)
		}
	}
	if options.Push {
		if err := runGit(ctx, absPath, options.Stdout, options.Stderr, "push", options.Remote, branch); err != nil {
			return fmt.Errorf("failed to push changes: %w", err)
		}
	}
	if !options.Pull && !options.Push {
		if err := runGit(ctx, absPath, options.Stdout, options.Stderr, "fetch", options.Remote); err != nil {
			return fmt.Errorf("failed to fetch changes: %w", err)
		}
	}

	return nil
}

func runGit(ctx context.Context, dir string, stdout, stderr io.Writer, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
