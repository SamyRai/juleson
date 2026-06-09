package builder

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/pkg/build"
)

// Test runs tests with the given options.
func (s *Service) Test(ctx context.Context, options TestOptions) error {
	config := build.DefaultTestConfig()
	config.Verbose = options.Verbose
	config.Race = options.Race
	config.Cover = options.Cover
	config.Short = options.Short
	config.Packages = options.Packages
	if options.Cover {
		config.CoverProfile = s.config.CoverageFile
	}
	result := s.RunTestsWithResult(ctx, config)
	return result.Error
}

// Coverage generates test coverage report.
func (s *Service) Coverage(ctx context.Context) error {
	// Run tests with coverage
	if err := s.Test(ctx, TestOptions{
		Verbose: true,
		Race:    true,
		Cover:   true,
	}); err != nil {
		return err
	}

	if err := s.GenerateCoverageHTML(ctx, build.TestConfig{CoverProfile: s.config.CoverageFile}, s.config.CoverageHTML); err != nil {
		return fmt.Errorf("failed to generate coverage report: %w", err)
	}

	return nil
}
