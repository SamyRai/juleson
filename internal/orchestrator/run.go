package orchestrator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Install installs binaries to the target path (defaults to GOPATH/bin)
func (s *Service) Install(ctx context.Context, targetPath string) error {
	// Build first
	if err := s.BuildAll(ctx); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Determine target path
	if targetPath == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			return fmt.Errorf("GOPATH not set and no target path provided")
		}
		targetPath = filepath.Join(gopath, "bin")
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Install CLI binary
	cliSrc := filepath.Join(s.config.BinDir, s.config.BinaryCLI)
	cliDst := filepath.Join(targetPath, s.config.BinaryCLI)
	if err := s.copyFile(cliSrc, cliDst); err != nil {
		return fmt.Errorf("failed to install CLI binary: %w", err)
	}

	// Install MCP binary
	mcpSrc := filepath.Join(s.config.BinDir, s.config.BinaryMCP)
	mcpDst := filepath.Join(targetPath, s.config.BinaryMCP)
	if err := s.copyFile(mcpSrc, mcpDst); err != nil {
		return fmt.Errorf("failed to install MCP binary: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst and sets executable permissions
func (s *Service) copyFile(src, dst string) error {
	srcFile, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err := os.WriteFile(dst, srcFile, 0755); err != nil {
		return err
	}

	return nil
}

// RunCLI runs the CLI binary with the given arguments
func (s *Service) RunCLI(ctx context.Context, args []string) error {
	// Build first
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

// RunMCP runs the MCP server binary
func (s *Service) RunMCP(ctx context.Context) error {
	// Build first
	if err := s.BuildMCP(ctx); err != nil {
		return fmt.Errorf("failed to build MCP: %w", err)
	}

	binaryPath := filepath.Join(s.config.BinDir, s.config.BinaryMCP)
	if err := s.runCommand(ctx, binaryPath); err != nil {
		return fmt.Errorf("failed to run MCP: %w", err)
	}

	return nil
}

// StartDev starts development mode with live reload (requires air)
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
