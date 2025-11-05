package orchestrator

import (
	"context"
	"fmt"
	"os/exec"
)

// Lint runs linters on the codebase
func (s *Service) Lint(ctx context.Context) error {
	// Run go vet
	if err := s.runCommand(ctx, "go", "vet", "./..."); err != nil {
		return fmt.Errorf("go vet failed: %w", err)
	}

	// Check if golangci-lint is installed
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return fmt.Errorf("golangci-lint not installed: install from https://golangci-lint.run/usage/install/")
	}

	// Run golangci-lint
	if err := s.runCommand(ctx, "golangci-lint", "run", "./..."); err != nil {
		return fmt.Errorf("golangci-lint failed: %w", err)
	}

	return nil
}

// Format formats the codebase
func (s *Service) Format(ctx context.Context) error {
	if err := s.runCommand(ctx, "go", "fmt", "./..."); err != nil {
		return fmt.Errorf("go fmt failed: %w", err)
	}

	return nil
}

// RunAllChecks runs all quality checks (lint, test, build)
func (s *Service) RunAllChecks(ctx context.Context) error {
	// Run lint
	if err := s.Lint(ctx); err != nil {
		return fmt.Errorf("lint check failed: %w", err)
	}

	// Run tests
	if err := s.Test(ctx, TestOptions{
		Verbose: true,
		Race:    true,
	}); err != nil {
		return fmt.Errorf("test check failed: %w", err)
	}

	// Build
	if err := s.BuildAll(ctx); err != nil {
		return fmt.Errorf("build check failed: %w", err)
	}

	return nil
}
