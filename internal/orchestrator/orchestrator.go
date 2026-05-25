package orchestrator

import (
	"context"
	"io"
	"os"
	"os/exec"
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
		BinaryMCP:    "jules-mcp",
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

// Service coordinates project-level orchestration workflows.
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

// GetVersion returns version information
func (s *Service) GetVersion() VersionInfo {
	return VersionInfo{
		Version:   s.config.Version,
		BuildDate: s.config.BuildDate,
		GitCommit: s.config.GitCommit,
	}
}
