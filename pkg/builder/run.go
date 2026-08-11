package builder

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

// Install installs binaries to the target path (defaults to GOPATH/bin).
func (s *Service) Install(ctx context.Context, targetPath string) error {
	_, err := s.InstallWithResult(ctx, InstallOptions{Path: targetPath})
	return err
}

// RunCLI runs the CLI binary with the given arguments.
func (s *Service) RunCLI(ctx context.Context, args []string) error {
	if err := s.BuildCLI(ctx); err != nil {
		return fmt.Errorf("failed to build CLI: %w", err)
	}

	binaryPath := filepath.Join(s.config.BinDir, s.config.BinaryCLI)

	cmdArgs := append([]string{}, args...)
	if err := s.runCommand(ctx, binaryPath, cmdArgs...); err != nil {
		return fmt.Errorf("failed to run CLI: %w", err)
	}

	return nil
}

// StartDev starts development mode with live reload (requires air).
func (s *Service) StartDev(ctx context.Context) error {
	// Check if air is installed
	if _, err := exec.LookPath("air"); err != nil {
		return fmt.Errorf("air not installed: install with 'go install github.com/cosmtrek/air@latest'")
	}

	if err := s.runCommand(ctx, "air"); err != nil {
		return fmt.Errorf("air failed: %w", err)
	}

	return nil
}
