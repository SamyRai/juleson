package analyzer

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// CoverageAnalyzer is responsible for analyzing test coverage
type CoverageAnalyzer struct{}

// NewCoverageAnalyzer creates a new CoverageAnalyzer
func NewCoverageAnalyzer() *CoverageAnalyzer {
	return &CoverageAnalyzer{}
}

// Analyze calculates the test coverage for a given project path
func (c *CoverageAnalyzer) Analyze(projectPath string) (float64, error) {
	// For Go projects, we can use 'go test -cover'
	// This is a simplified example; a real implementation would need to
	// handle different languages and testing frameworks.

	cmd := exec.Command("go", "test", "-cover", "./...")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If 'no test files' is the error, it's not a failure, but 0% coverage
		if strings.Contains(string(output), "no test files") {
			return 0.0, nil
		}
		return 0.0, fmt.Errorf("failed to run coverage analysis: %w\nOutput: %s", err, string(output))
	}

	// Parse the output to find the coverage percentage
	// Example output: "coverage: 83.3% of statements"
	re := regexp.MustCompile(`coverage: (\d+\.\d+)% of statements`)
	matches := re.FindStringSubmatch(string(output))

	if len(matches) < 2 {
		// If no coverage percentage is found, assume 0%
		return 0.0, nil
	}

	coverage, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0.0, fmt.Errorf("failed to parse coverage percentage: %w", err)
	}

	return coverage, nil
}
