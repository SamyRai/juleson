package julesops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VerificationOptions controls repository verification.
type VerificationOptions struct {
	WorkingDir string
	Command    string
	Packages   []string
	Short      bool
}

// VerificationResult captures a verification command result.
type VerificationResult struct {
	WorkingDir string
	Success    bool
	Command    string
	Output     string
	Summary    string
}

// VerifyProjectChanges chooses a conservative repo-native verification command.
func VerifyProjectChanges(ctx context.Context, options VerificationOptions) (*VerificationResult, error) {
	workingDir := options.WorkingDir
	if workingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		workingDir = wd
	}

	args, display, err := verificationCommand(workingDir, options)
	if err != nil {
		return &VerificationResult{
			WorkingDir: workingDir,
			Success:    false,
			Summary:    err.Error(),
		}, nil
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = workingDir
	output, runErr := cmd.CombinedOutput()
	result := &VerificationResult{
		WorkingDir: workingDir,
		Command:    display,
		Output:     string(output),
	}
	if runErr != nil {
		result.Success = false
		result.Summary = fmt.Sprintf("verification failed: %v", runErr)
		return result, nil
	}
	result.Success = true
	result.Summary = "verification passed"
	return result, nil
}

func verificationCommand(workingDir string, options VerificationOptions) ([]string, string, error) {
	if strings.TrimSpace(options.Command) != "" {
		fields := strings.Fields(options.Command)
		if len(fields) == 0 {
			return nil, "", fmt.Errorf("verification command cannot be empty")
		}
		return fields, strings.Join(fields, " "), nil
	}

	switch {
	case fileExists(workingDir, "go.mod"):
		args := []string{"go", "test"}
		if options.Short {
			args = append(args, "-short")
		}
		packages := options.Packages
		if len(packages) == 0 {
			packages = []string{"./..."}
		}
		args = append(args, packages...)
		return args, strings.Join(args, " "), nil
	case fileExists(workingDir, "yarn.lock"):
		return []string{"yarn", "test"}, "yarn test", nil
	case fileExists(workingDir, "package.json"):
		return []string{"yarn", "test"}, "yarn test", nil
	case fileExists(workingDir, "pyproject.toml") || fileExists(workingDir, "uv.lock"):
		return []string{"uv", "run", "pytest"}, "uv run pytest", nil
	case fileExists(workingDir, "Cargo.toml"):
		return []string{"cargo", "test"}, "cargo test", nil
	default:
		return nil, "", fmt.Errorf("no supported verification target found; pass an explicit command")
	}
}

func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}
