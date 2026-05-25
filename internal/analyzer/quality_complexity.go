package analyzer

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// analyzeCodeComplexity analyzes code complexity.
func (c *complexityAnalyzer) analyzeCodeComplexity(projectPath string, languages []string) (float64, error) {
	totalComplexity := 0.0
	fileCount := 0

	for _, lang := range languages {
		switch lang {
		case "go":
			if complexity, count, err := c.analyzeGoComplexity(projectPath); err == nil {
				totalComplexity += complexity
				fileCount += count
			}
		case "python":
			if complexity, count, err := c.analyzePythonComplexity(projectPath); err == nil {
				totalComplexity += complexity
				fileCount += count
			}
		case "javascript", "typescript":
			if complexity, count, err := c.analyzeJavaScriptComplexity(projectPath); err == nil {
				totalComplexity += complexity
				fileCount += count
			}
		}
	}

	if fileCount > 0 {
		return totalComplexity / float64(fileCount), nil
	}
	return 0.0, fmt.Errorf("no complexity analysis available")
}

func (c *complexityAnalyzer) analyzeGoComplexity(projectPath string) (float64, int, error) {
	if _, err := exec.LookPath("gocyclo"); err != nil {
		return 0.0, 0, fmt.Errorf("gocyclo not found")
	}

	cmd := exec.Command("gocyclo", "-over", "10", ".")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return 0.0, 0, err
	}

	lines := strings.Split(string(output), "\n")
	totalComplexity := 0
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			if complexity, err := strconv.Atoi(parts[0]); err == nil {
				totalComplexity += complexity
				count++
			}
		}
	}

	if count > 0 {
		return float64(totalComplexity) / float64(count), count, nil
	}
	return 0.0, 0, fmt.Errorf("no complexity data")
}

func (c *complexityAnalyzer) analyzePythonComplexity(projectPath string) (float64, int, error) {
	if _, err := exec.LookPath("radon"); err != nil {
		return 0.0, 0, fmt.Errorf("radon not found")
	}

	cmd := exec.Command("radon", "cc", "-a", ".")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return 0.0, 0, err
	}
	return c.parseRadonComplexity(string(output))
}

func (c *complexityAnalyzer) analyzeJavaScriptComplexity(projectPath string) (float64, int, error) {
	if _, err := exec.LookPath("npx"); err != nil {
		return 0.0, 0, fmt.Errorf("npx not found")
	}

	cmd := exec.Command("npx", "eslint", "--ext", ".js,.ts,.jsx,.tsx", ".", "--format", "json")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return 0.0, 0, err
	}
	return c.parseESLintComplexity(string(output))
}

func (c *complexityAnalyzer) parseRadonComplexity(output string) (float64, int, error) {
	return 0.0, 0, fmt.Errorf("not implemented")
}

func (c *complexityAnalyzer) parseESLintComplexity(output string) (float64, int, error) {
	return 0.0, 0, fmt.Errorf("not implemented")
}
