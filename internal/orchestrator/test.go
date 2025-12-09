package orchestrator

import (
	"context"
	"fmt"
)

// Test runs tests with the given options
func (s *Service) Test(ctx context.Context, options TestOptions) error {
	args := []string{"test"}

	if options.Verbose {
		args = append(args, "-v")
	}

	if options.Race {
		args = append(args, "-race")
	}

	if options.Cover {
		args = append(args, "-coverprofile="+s.config.CoverageFile, "-covermode=atomic")
	}

	if options.Short {
		args = append(args, "-short")
	}

	// Add packages or default to all
	if len(options.Packages) > 0 {
		args = append(args, options.Packages...)
	} else {
		args = append(args, "./...")
	}

	if err := s.runCommand(ctx, "go", args...); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	return nil
}

// Coverage generates test coverage report
func (s *Service) Coverage(ctx context.Context) error {
	// Run tests with coverage
	if err := s.Test(ctx, TestOptions{
		Verbose: true,
		Race:    true,
		Cover:   true,
	}); err != nil {
		return err
	}

	// Generate HTML report
	if err := s.runCommand(ctx, "go", "tool", "cover",
		"-html="+s.config.CoverageFile,
		"-o", s.config.CoverageHTML); err != nil {
		return fmt.Errorf("failed to generate coverage report: %w", err)
	}

	return nil
}
