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

// CodeQualityAnalyzer analyzes code quality metrics
type CodeQualityAnalyzer struct{}

// NewCodeQualityAnalyzer creates a new code quality analyzer
func NewCodeQualityAnalyzer() *CodeQualityAnalyzer {
	return &CodeQualityAnalyzer{}
}

// CodeQualityMetrics represents various code quality metrics
type CodeQualityMetrics struct {
	TestCoverage      float64            `json:"test_coverage"`
	CodeComplexity    float64            `json:"code_complexity"`
	Maintainability   float64            `json:"maintainability"`
	DuplicationRate   float64            `json:"duplication_rate"`
	SecurityIssues    []SecurityIssue    `json:"security_issues"`
	CodeSmells        []CodeSmell        `json:"code_smells"`
	PerformanceIssues []PerformanceIssue `json:"performance_issues"`
}

// SecurityIssue represents a security vulnerability
type SecurityIssue struct {
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	CVE         string `json:"cve,omitempty"`
}

// CodeSmell represents a code quality issue
type CodeSmell struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Suggestion  string `json:"suggestion"`
}

// PerformanceIssue represents a performance problem
type PerformanceIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Suggestion  string `json:"suggestion"`
}

// Analyze performs comprehensive code quality analysis
func (c *CodeQualityAnalyzer) Analyze(projectPath string, languages []string) (*CodeQualityMetrics, error) {
	metrics := &CodeQualityMetrics{
		TestCoverage:      0.0,
		CodeComplexity:    0.0,
		Maintainability:   0.0,
		DuplicationRate:   0.0,
		SecurityIssues:    make([]SecurityIssue, 0),
		CodeSmells:        make([]CodeSmell, 0),
		PerformanceIssues: make([]PerformanceIssue, 0),
	}

	// Analyze test coverage
	if coverage, err := c.analyzeTestCoverage(projectPath, languages); err == nil {
		metrics.TestCoverage = coverage
	}

	// Analyze code complexity
	if complexity, err := c.analyzeCodeComplexity(projectPath, languages); err == nil {
		metrics.CodeComplexity = complexity
	}

	// Analyze security issues
	if issues, err := c.analyzeSecurityIssues(projectPath, languages); err == nil {
		metrics.SecurityIssues = issues
	}

	// Analyze code smells
	if smells, err := c.analyzeCodeSmells(projectPath, languages); err == nil {
		metrics.CodeSmells = smells
	}

	// Calculate maintainability index
	metrics.Maintainability = c.calculateMaintainabilityIndex(metrics)

	return metrics, nil
}

// analyzeTestCoverage analyzes test coverage for supported languages
func (c *CodeQualityAnalyzer) analyzeTestCoverage(projectPath string, languages []string) (float64, error) {
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

// analyzeGoTestCoverage analyzes Go test coverage
func (c *CodeQualityAnalyzer) analyzeGoTestCoverage(projectPath string) (float64, error) {
	// Run go test with coverage
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	cmd.Dir = projectPath

	if err := cmd.Run(); err != nil {
		// If tests fail, try to get coverage anyway
		cmd = exec.Command("go", "test", "-cover", "./...")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err != nil {
			return 0.0, err
		}

		// Parse coverage from output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "coverage:") {
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
		}
		return 0.0, fmt.Errorf("could not parse coverage")
	}

	// Read coverage profile
	coverageFile := filepath.Join(projectPath, "coverage.out")
	defer func() {
		if removeErr := os.Remove(coverageFile); removeErr != nil {
			// Log error but don't override the main error
			// In a real implementation, you'd use a proper logger
		}
	}() // Clean up

	file, err := os.Open(coverageFile)
	if err != nil {
		return 0.0, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log error but don't override the main error
			// In a real implementation, you'd use a proper logger
		}
	}()

	scanner := bufio.NewScanner(file)
	totalCoverage := 0.0
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "total:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				coverageStr := strings.TrimSuffix(parts[2], "%")
				if coverage, err := strconv.ParseFloat(coverageStr, 64); err == nil {
					return coverage, nil
				}
			}
		}
	}

	if count > 0 {
		return totalCoverage / float64(count), nil
	}
	return 0.0, fmt.Errorf("no coverage data found")
}

// analyzePythonTestCoverage analyzes Python test coverage
func (c *CodeQualityAnalyzer) analyzePythonTestCoverage(projectPath string) (float64, error) {
	// Check if coverage is available
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

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "TOTAL") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				coverageStr := strings.TrimSuffix(parts[3], "%")
				if coverage, err := strconv.ParseFloat(coverageStr, 64); err == nil {
					return coverage, nil
				}
			}
		}
	}

	return 0.0, fmt.Errorf("could not parse coverage")
}

// analyzeJavaScriptTestCoverage analyzes JavaScript/TypeScript test coverage
func (c *CodeQualityAnalyzer) analyzeJavaScriptTestCoverage(projectPath string) (float64, error) {
	// Check for various test runners
	if c.hasPackageScript(projectPath, "test:coverage") {
		cmd := exec.Command("npm", "run", "test:coverage")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			return c.parseJestCoverageOutput(string(output))
		}
	}

	// Try nyc/istanbul
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

// analyzeJavaTestCoverage analyzes Java test coverage
func (c *CodeQualityAnalyzer) analyzeJavaTestCoverage(projectPath string) (float64, error) {
	// Check for JaCoCo or similar
	if _, err := exec.LookPath("mvn"); err == nil {
		cmd := exec.Command("mvn", "test", "jacoco:report")
		cmd.Dir = projectPath
		if err := cmd.Run(); err == nil {
			// Parse JaCoCo report
			return c.parseJaCoCoReport(projectPath)
		}
	}

	return 0.0, fmt.Errorf("no Java coverage tool found")
}

// analyzeCSharpTestCoverage analyzes C# test coverage
func (c *CodeQualityAnalyzer) analyzeCSharpTestCoverage(projectPath string) (float64, error) {
	if _, err := exec.LookPath("dotnet"); err == nil {
		cmd := exec.Command("dotnet", "test", "--collect:\"XPlat Code Coverage\"")
		cmd.Dir = projectPath
		if err := cmd.Run(); err == nil {
			// Parse coverage report
			return c.parseDotNetCoverage(projectPath)
		}
	}

	return 0.0, fmt.Errorf("no .NET coverage tool found")
}

// analyzeCodeComplexity analyzes code complexity
func (c *CodeQualityAnalyzer) analyzeCodeComplexity(projectPath string, languages []string) (float64, error) {
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

// analyzeGoComplexity analyzes Go code complexity
func (c *CodeQualityAnalyzer) analyzeGoComplexity(projectPath string) (float64, int, error) {
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

// analyzePythonComplexity analyzes Python code complexity
func (c *CodeQualityAnalyzer) analyzePythonComplexity(projectPath string) (float64, int, error) {
	// Use radon for complexity analysis
	if _, err := exec.LookPath("radon"); err != nil {
		return 0.0, 0, fmt.Errorf("radon not found")
	}

	cmd := exec.Command("radon", "cc", "-a", ".")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return 0.0, 0, err
	}

	// Parse radon output
	return c.parseRadonComplexity(string(output))
}

// analyzeJavaScriptComplexity analyzes JavaScript/TypeScript complexity
func (c *CodeQualityAnalyzer) analyzeJavaScriptComplexity(projectPath string) (float64, int, error) {
	// Use eslint complexity plugin or similar
	if _, err := exec.LookPath("npx"); err != nil {
		return 0.0, 0, fmt.Errorf("npx not found")
	}

	cmd := exec.Command("npx", "eslint", "--ext", ".js,.ts,.jsx,.tsx", ".", "--format", "json")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return 0.0, 0, err
	}

	// Parse ESLint complexity data
	return c.parseESLintComplexity(string(output))
}

// analyzeSecurityIssues analyzes security vulnerabilities
func (c *CodeQualityAnalyzer) analyzeSecurityIssues(projectPath string, languages []string) ([]SecurityIssue, error) {
	issues := make([]SecurityIssue, 0)

	for _, lang := range languages {
		switch lang {
		case "javascript", "typescript":
			if jsIssues, err := c.analyzeJavaScriptSecurity(projectPath); err == nil {
				issues = append(issues, jsIssues...)
			}
		case "python":
			if pyIssues, err := c.analyzePythonSecurity(projectPath); err == nil {
				issues = append(issues, pyIssues...)
			}
		case "go":
			if goIssues, err := c.analyzeGoSecurity(projectPath); err == nil {
				issues = append(issues, goIssues...)
			}
		}
	}

	return issues, nil
}

// analyzeJavaScriptSecurity analyzes JavaScript security issues
func (c *CodeQualityAnalyzer) analyzeJavaScriptSecurity(projectPath string) ([]SecurityIssue, error) {
	issues := make([]SecurityIssue, 0)

	// Check for npm audit
	if c.hasPackageJSON(projectPath) {
		cmd := exec.Command("npm", "audit", "--json")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			if auditIssues, err := c.parseNPMAudit(string(output)); err == nil {
				issues = append(issues, auditIssues...)
			}
		}
	}

	return issues, nil
}

// analyzePythonSecurity analyzes Python security issues
func (c *CodeQualityAnalyzer) analyzePythonSecurity(projectPath string) ([]SecurityIssue, error) {
	issues := make([]SecurityIssue, 0)

	// Check for safety
	if _, err := exec.LookPath("safety"); err == nil {
		cmd := exec.Command("safety", "check", "--json")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			if safetyIssues, err := c.parseSafetyOutput(string(output)); err == nil {
				issues = append(issues, safetyIssues...)
			}
		}
	}

	return issues, nil
}

// analyzeGoSecurity analyzes Go security issues
func (c *CodeQualityAnalyzer) analyzeGoSecurity(projectPath string) ([]SecurityIssue, error) {
	issues := make([]SecurityIssue, 0)

	// Use gosec if available
	if _, err := exec.LookPath("gosec"); err == nil {
		cmd := exec.Command("gosec", "-fmt=json", "./...")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			if gosecIssues, err := c.parseGosecOutput(string(output)); err == nil {
				issues = append(issues, gosecIssues...)
			}
		}
	}

	return issues, nil
}

// analyzeCodeSmells analyzes code quality issues
func (c *CodeQualityAnalyzer) analyzeCodeSmells(projectPath string, languages []string) ([]CodeSmell, error) {
	smells := make([]CodeSmell, 0)

	for _, lang := range languages {
		switch lang {
		case "go":
			if goSmells, err := c.analyzeGoCodeSmells(projectPath); err == nil {
				smells = append(smells, goSmells...)
			}
		case "python":
			if pySmells, err := c.analyzePythonCodeSmells(projectPath); err == nil {
				smells = append(smells, pySmells...)
			}
		case "javascript", "typescript":
			if jsSmells, err := c.analyzeJavaScriptCodeSmells(projectPath); err == nil {
				smells = append(smells, jsSmells...)
			}
		}
	}

	return smells, nil
}

// analyzeGoCodeSmells analyzes Go code smells
func (c *CodeQualityAnalyzer) analyzeGoCodeSmells(projectPath string) ([]CodeSmell, error) {
	smells := make([]CodeSmell, 0)

	// Use golint or revive
	if _, err := exec.LookPath("revive"); err == nil {
		cmd := exec.Command("revive", "-formatter", "json", "./...")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			if reviveSmells, err := c.parseReviveOutput(string(output)); err == nil {
				smells = append(smells, reviveSmells...)
			}
		}
	}

	return smells, nil
}

// analyzePythonCodeSmells analyzes Python code smells
func (c *CodeQualityAnalyzer) analyzePythonCodeSmells(projectPath string) ([]CodeSmell, error) {
	smells := make([]CodeSmell, 0)

	// Use pylint
	if _, err := exec.LookPath("pylint"); err == nil {
		cmd := exec.Command("pylint", "--output-format=json", ".")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			if pylintSmells, err := c.parsePylintOutput(string(output)); err == nil {
				smells = append(smells, pylintSmells...)
			}
		}
	}

	return smells, nil
}

// analyzeJavaScriptCodeSmells analyzes JavaScript code smells
func (c *CodeQualityAnalyzer) analyzeJavaScriptCodeSmells(projectPath string) ([]CodeSmell, error) {
	smells := make([]CodeSmell, 0)

	// Use eslint
	if c.hasPackageJSON(projectPath) {
		cmd := exec.Command("npx", "eslint", ".", "--format", "json")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			if eslintSmells, err := c.parseESLintOutput(string(output)); err == nil {
				smells = append(smells, eslintSmells...)
			}
		}
	}

	return smells, nil
}

// calculateMaintainabilityIndex calculates the maintainability index
func (c *CodeQualityAnalyzer) calculateMaintainabilityIndex(metrics *CodeQualityMetrics) float64 {
	// Simplified maintainability index calculation
	// Higher is better (0-100)
	baseScore := 100.0

	// Penalize for complexity
	if metrics.CodeComplexity > 10 {
		baseScore -= (metrics.CodeComplexity - 10) * 2
	}

	// Penalize for security issues
	baseScore -= float64(len(metrics.SecurityIssues)) * 5

	// Penalize for code smells
	baseScore -= float64(len(metrics.CodeSmells)) * 2

	// Penalize for low test coverage
	if metrics.TestCoverage < 80 {
		baseScore -= (80 - metrics.TestCoverage) * 0.5
	}

	// Ensure bounds
	if baseScore < 0 {
		baseScore = 0
	}
	if baseScore > 100 {
		baseScore = 100
	}

	return baseScore
}

// Helper methods for parsing outputs
func (c *CodeQualityAnalyzer) hasPackageJSON(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, "package.json"))
	return err == nil
}

func (c *CodeQualityAnalyzer) hasPackageScript(projectPath, script string) bool {
	// Simplified check - in production, parse package.json
	return true // Assume it exists for now
}

func (c *CodeQualityAnalyzer) parseJestCoverageOutput(output string) (float64, error) {
	// Parse Jest coverage output
	return 0.0, fmt.Errorf("not implemented")
}

func (c *CodeQualityAnalyzer) parseNYCCoverageOutput(output string) (float64, error) {
	// Parse NYC coverage output
	return 0.0, fmt.Errorf("not implemented")
}

func (c *CodeQualityAnalyzer) parseJaCoCoReport(projectPath string) (float64, error) {
	// Parse JaCoCo XML report
	return 0.0, fmt.Errorf("not implemented")
}

func (c *CodeQualityAnalyzer) parseDotNetCoverage(projectPath string) (float64, error) {
	// Parse .NET coverage report
	return 0.0, fmt.Errorf("not implemented")
}

func (c *CodeQualityAnalyzer) parseRadonComplexity(output string) (float64, int, error) {
	// Parse radon complexity output
	return 0.0, 0, fmt.Errorf("not implemented")
}

func (c *CodeQualityAnalyzer) parseESLintComplexity(output string) (float64, int, error) {
	// Parse ESLint complexity output
	return 0.0, 0, fmt.Errorf("not implemented")
}

func (c *CodeQualityAnalyzer) parseNPMAudit(output string) ([]SecurityIssue, error) {
	// Parse npm audit JSON output
	return nil, fmt.Errorf("not implemented")
}

func (c *CodeQualityAnalyzer) parseSafetyOutput(output string) ([]SecurityIssue, error) {
	// Parse safety JSON output
	return nil, fmt.Errorf("not implemented")
}

func (c *CodeQualityAnalyzer) parseGosecOutput(output string) ([]SecurityIssue, error) {
	// Parse gosec JSON output
	return nil, fmt.Errorf("not implemented")
}

func (c *CodeQualityAnalyzer) parseReviveOutput(output string) ([]CodeSmell, error) {
	// Parse revive JSON output
	return nil, fmt.Errorf("not implemented")
}

func (c *CodeQualityAnalyzer) parsePylintOutput(output string) ([]CodeSmell, error) {
	// Parse pylint JSON output
	return nil, fmt.Errorf("not implemented")
}

func (c *CodeQualityAnalyzer) parseESLintOutput(output string) ([]CodeSmell, error) {
	// Parse ESLint JSON output
	return nil, fmt.Errorf("not implemented")
}
