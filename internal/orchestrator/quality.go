package orchestrator

import (
	"context"
	"fmt"
)

// Lint runs linters on the codebase
func (s *Service) Lint(ctx context.Context) error {
	result := s.LintWithResult(ctx, DefaultLintConfig())
	return result.Error
}

// Format formats the codebase
func (s *Service) Format(ctx context.Context) error {
	return s.FormatCode(ctx, false, "./...")
}

// RunAllChecks runs all quality checks (lint, test, build)
func (s *Service) RunAllChecks(ctx context.Context) error {
	if _, err := s.RunQualityChecks(ctx, QualityOptions{
		Format: true,
		Lint:   true,
		Test:   true,
		Build:  true,
	}); err != nil {
		return fmt.Errorf("quality checks failed: %w", err)
	}
	return nil
}
