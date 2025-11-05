package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Config holds orchestrator configuration
type Config struct {
	BinaryCLI    string
	BinaryMCP    string
	BinDir       string
	CmdCLIDir    string
	CmdMCPDir    string
	CoverageFile string
	CoverageHTML string
	DockerImage  string
	Version      string
	BuildDate    string
	GitCommit    string
}

// DefaultConfig returns default configuration for Juleson project
func DefaultConfig(version, buildDate, gitCommit string) *Config {
	return &Config{
		BinaryCLI:    "juleson",
		BinaryMCP:    "juleson-mcp",
		BinDir:       "bin",
		CmdCLIDir:    "cmd/juleson",
		CmdMCPDir:    "cmd/jules-mcp",
		CoverageFile: "coverage.out",
		CoverageHTML: "coverage.html",
		DockerImage:  "juleson:latest",
		Version:      version,
		BuildDate:    buildDate,
		GitCommit:    gitCommit,
	}
}

// Orchestrator defines the interface for project orchestration operations
type Orchestrator interface {
	// Build operations
	BuildAll(ctx context.Context) error
	BuildCLI(ctx context.Context) error
	BuildMCP(ctx context.Context) error

	// Clean operations
	Clean(ctx context.Context) error

	// Test operations
	Test(ctx context.Context, options TestOptions) error
	Coverage(ctx context.Context) error

	// Code quality operations
	Lint(ctx context.Context) error
	Format(ctx context.Context) error

	// Dependency operations
	DownloadDeps(ctx context.Context) error
	TidyDeps(ctx context.Context) error

	// Install operations
	Install(ctx context.Context, targetPath string) error

	// Run operations
	RunCLI(ctx context.Context, args []string) error
	RunMCP(ctx context.Context) error

	// Development operations
	StartDev(ctx context.Context) error

	// Check operations
	RunAllChecks(ctx context.Context) error

	// Docker operations
	DockerBuild(ctx context.Context) error
	DockerRun(ctx context.Context, args []string) error
	DockerRunCLI(ctx context.Context, args []string) error
	DockerRunMCP(ctx context.Context) error
	DockerPush(ctx context.Context) error
	DockerComposeUp(ctx context.Context) error
	DockerComposeDown(ctx context.Context) error
	DockerClean(ctx context.Context) error

	// Info operations
	GetVersion() VersionInfo
}

// TestOptions holds options for running tests
type TestOptions struct {
	Verbose  bool
	Race     bool
	Cover    bool
	Short    bool
	Packages []string
}

// VersionInfo holds version information
type VersionInfo struct {
	Version   string
	BuildDate string
	GitCommit string
}

// Service implements the Orchestrator interface
type Service struct {
	config *Config
	stdout io.Writer
	stderr io.Writer
}

// NewService creates a new orchestrator service
func NewService(config *Config) *Service {
	return &Service{
		config: config,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// WithOutput sets custom output writers for stdout and stderr
func (s *Service) WithOutput(stdout, stderr io.Writer) *Service {
	s.stdout = stdout
	s.stderr = stderr
	return s
}

// runCommand executes a command with the given arguments
func (s *Service) runCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = s.stdout
	cmd.Stderr = s.stderr
	return cmd.Run()
}

// ensureBinDir creates the bin directory if it doesn't exist
func (s *Service) ensureBinDir() error {
	return os.MkdirAll(s.config.BinDir, 0755)
}

// buildBinary builds a Go binary with ldflags
func (s *Service) buildBinary(ctx context.Context, outputPath, sourceDir string) error {
	ldflags := fmt.Sprintf("-s -w -X 'github.com/SamyRai/juleson/internal/cli/commands.Version=%s' "+
		"-X 'github.com/SamyRai/juleson/internal/cli/commands.BuildDate=%s' "+
		"-X 'github.com/SamyRai/juleson/internal/cli/commands.GitCommit=%s'",
		s.config.Version, s.config.BuildDate, s.config.GitCommit)

	return s.runCommand(ctx, "go", "build", "-trimpath", "-ldflags", ldflags, "-o", outputPath, "./"+sourceDir)
}

// GetVersion returns version information
func (s *Service) GetVersion() VersionInfo {
	return VersionInfo{
		Version:   s.config.Version,
		BuildDate: s.config.BuildDate,
		GitCommit: s.config.GitCommit,
	}
}

// BuildAll builds all binaries
func (s *Service) BuildAll(ctx context.Context) error {
	if err := s.ensureBinDir(); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	if err := s.BuildCLI(ctx); err != nil {
		return fmt.Errorf("failed to build CLI: %w", err)
	}

	if err := s.BuildMCP(ctx); err != nil {
		return fmt.Errorf("failed to build MCP: %w", err)
	}

	return nil
}

// BuildCLI builds the CLI binary
func (s *Service) BuildCLI(ctx context.Context) error {
	if err := s.ensureBinDir(); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	outputPath := filepath.Join(s.config.BinDir, s.config.BinaryCLI)
	return s.buildBinary(ctx, outputPath, s.config.CmdCLIDir)
}

// BuildMCP builds the MCP server binary
func (s *Service) BuildMCP(ctx context.Context) error {
	if err := s.ensureBinDir(); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	outputPath := filepath.Join(s.config.BinDir, s.config.BinaryMCP)
	return s.buildBinary(ctx, outputPath, s.config.CmdMCPDir)
}

// Clean removes build artifacts
func (s *Service) Clean(ctx context.Context) error {
	if err := s.runCommand(ctx, "go", "clean"); err != nil {
		return fmt.Errorf("go clean failed: %w", err)
	}

	if err := os.RemoveAll(s.config.BinDir); err != nil {
		return fmt.Errorf("failed to remove bin directory: %w", err)
	}

	// Remove coverage files
	os.Remove(s.config.CoverageFile)
	os.Remove(s.config.CoverageHTML)

	return nil
}
