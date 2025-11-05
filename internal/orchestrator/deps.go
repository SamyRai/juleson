package orchestrator

import (
	"context"
	"fmt"
)

// DownloadDeps downloads Go module dependencies
func (s *Service) DownloadDeps(ctx context.Context) error {
	if err := s.runCommand(ctx, "go", "mod", "download"); err != nil {
		return fmt.Errorf("go mod download failed: %w", err)
	}

	if err := s.runCommand(ctx, "go", "mod", "verify"); err != nil {
		return fmt.Errorf("go mod verify failed: %w", err)
	}

	return nil
}

// TidyDeps tidies Go module dependencies
func (s *Service) TidyDeps(ctx context.Context) error {
	if err := s.runCommand(ctx, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	return nil
}
