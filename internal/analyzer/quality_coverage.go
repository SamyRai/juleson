package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// analyzeTestCoverage analyzes test coverage for supported languages.
func (c *coverageAnalyzer) analyzeTestCoverage(projectPath string, languages []string) (float64, error) {
	for _, lang := range languages {
		switch lang {
		case "go":
			return c.analyzeGoTestCoverage(projectPath)
		case "python":
			return c.analyzePythonTestCoverage(projectPath)
		case "javascript", "typescript":
			return c.analyzeJavaScriptTestCoverage(projectPath)
		case "java":
			return c.analyzeJavaTestCoverage(projectPath)
		case "csharp":
			return c.analyzeCSharpTestCoverage(projectPath)
		}
	}
	return 0.0, fmt.Errorf("no supported language for test coverage analysis")
}

func (c *coverageAnalyzer) analyzeGoTestCoverage(projectPath string) (float64, error) {
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	cmd.Dir = projectPath

	if err := cmd.Run(); err != nil {
		cmd = exec.Command("go", "test", "-cover", "./...")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err != nil {
			return 0.0, err
		}
		return parseCoveragePercent(string(output))
	}

	coverageFile := filepath.Join(projectPath, "coverage.out")
	defer func() { _ = os.Remove(coverageFile) }()

	file, err := os.Open(coverageFile)
	if err != nil {
		return 0.0, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "total:") {
			return parseCoveragePercent(line)
		}
	}
	if err := scanner.Err(); err != nil {
		return 0.0, err
	}
	return 0.0, fmt.Errorf("no coverage data found")
}

func (c *coverageAnalyzer) analyzePythonTestCoverage(projectPath string) (float64, error) {
	if _, err := exec.LookPath("coverage"); err != nil {
		return 0.0, fmt.Errorf("coverage tool not found")
	}

	cmd := exec.Command("coverage", "run", "--source=.", "-m", "pytest")
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		return 0.0, err
	}

	cmd = exec.Command("coverage", "report")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return 0.0, err
	}
	return parseCoveragePercent(string(output))
}

func (c *coverageAnalyzer) analyzeJavaScriptTestCoverage(projectPath string) (float64, error) {
	if c.hasPackageScript(projectPath, "test:coverage") {
		cmd := exec.Command("npm", "run", "test:coverage")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			return c.parseJestCoverageOutput(string(output))
		}
	}

	if _, err := exec.LookPath("nyc"); err == nil {
		cmd := exec.Command("nyc", "npm", "test")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			return c.parseNYCCoverageOutput(string(output))
		}
	}

	return 0.0, fmt.Errorf("no coverage tool configured")
}

func (c *coverageAnalyzer) analyzeJavaTestCoverage(projectPath string) (float64, error) {
	if _, err := exec.LookPath("mvn"); err == nil {
		cmd := exec.Command("mvn", "test", "jacoco:report")
		cmd.Dir = projectPath
		if err := cmd.Run(); err == nil {
			return c.parseJaCoCoReport(projectPath)
		}
	}
	return 0.0, fmt.Errorf("no Java coverage tool found")
}

func (c *coverageAnalyzer) analyzeCSharpTestCoverage(projectPath string) (float64, error) {
	if _, err := exec.LookPath("dotnet"); err == nil {
		cmd := exec.Command("dotnet", "test", "--collect:\"XPlat Code Coverage\"")
		cmd.Dir = projectPath
		if err := cmd.Run(); err == nil {
			return c.parseDotNetCoverage(projectPath)
		}
	}
	return 0.0, fmt.Errorf("no .NET coverage tool found")
}

func (c *coverageAnalyzer) hasPackageScript(projectPath, script string) bool {
	return true
}

func (c *coverageAnalyzer) parseJestCoverageOutput(output string) (float64, error) {
	return 0.0, fmt.Errorf("not implemented")
}

func (c *coverageAnalyzer) parseNYCCoverageOutput(output string) (float64, error) {
	return 0.0, fmt.Errorf("not implemented")
}

func (c *coverageAnalyzer) parseJaCoCoReport(projectPath string) (float64, error) {
	return 0.0, fmt.Errorf("not implemented")
}

func (c *coverageAnalyzer) parseDotNetCoverage(projectPath string) (float64, error) {
	return 0.0, fmt.Errorf("not implemented")
}

func parseCoveragePercent(output string) (float64, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if !strings.Contains(line, "%") {
			continue
		}
		parts := strings.Fields(line)
		for _, part := range parts {
			if strings.HasSuffix(part, "%") {
				coverageStr := strings.TrimSuffix(part, "%")
				if coverage, err := strconv.ParseFloat(coverageStr, 64); err == nil {
					return coverage, nil
				}
			}
		}
	}
	return 0.0, fmt.Errorf("could not parse coverage")
}
