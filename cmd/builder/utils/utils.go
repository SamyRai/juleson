package utils

import (
	"os"
	"os/exec"
	"strings"
)

// RunCommand executes a command with the given arguments
func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GetGitCommit returns the current git commit hash
func GetGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// GetCurrentDir returns the current working directory
func GetCurrentDir() string {
	cmd := exec.Command("pwd")
	output, err := cmd.Output()
	if err != nil {
		return "."
	}
	return strings.TrimSpace(string(output))
}
