package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCodeQualityAnalyzer_Analyze(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "quality_analyzer_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	createQualityTestProject(t, tempDir)

	// Test analyzer
	analyzer := NewCodeQualityAnalyzer()
	languages := []string{"go", "python", "javascript"}
	metrics, err := analyzer.Analyze(tempDir, languages)
	if err != nil {
		t.Fatalf("Quality analysis failed: %v", err)
	}

	// Validate results
	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// Check that we have some basic metrics
	if metrics.TestCoverage < 0 || metrics.TestCoverage > 100 {
		t.Errorf("Test coverage should be between 0 and 100, got: %f", metrics.TestCoverage)
	}

	if metrics.CodeComplexity < 0 {
		t.Errorf("Code complexity should be non-negative, got: %f", metrics.CodeComplexity)
	}

	if metrics.Maintainability < 0 || metrics.Maintainability > 100 {
		t.Errorf("Maintainability should be between 0 and 100, got: %f", metrics.Maintainability)
	}
}

func TestCodeQualityAnalyzer_AnalyzeGoProject(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "go_quality_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create Go test files
	goFiles := map[string]string{
		"main.go": `package main

import "fmt"

// CalculateSum calculates the sum of two numbers
func CalculateSum(a, b int) int {
	result := a + b
	return result
}

func main() {
	fmt.Println(CalculateSum(1, 2))
}`,
		"main_test.go": `package main

import "testing"

func TestCalculateSum(t *testing.T) {
	result := CalculateSum(2, 3)
	expected := 5
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}`,
	}

	for filename, content := range goFiles {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test analyzer
	analyzer := NewCodeQualityAnalyzer()
	languages := []string{"go"}
	metrics, err := analyzer.Analyze(tempDir, languages)
	if err != nil {
		t.Fatalf("Go quality analysis failed: %v", err)
	}

	// Validate results
	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// Should have some security issues or code smells detected
	// (exact validation depends on available tools)
}

func TestCodeQualityAnalyzer_AnalyzePythonProject(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "python_quality_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create Python test files
	pythonFiles := map[string]string{
		"app.py": `def calculate_sum(a, b):
    """Calculate the sum of two numbers."""
    result = a + b
    return result

if __name__ == "__main__":
    print(calculate_sum(1, 2))`,
		"test_app.py": `import unittest
from app import calculate_sum

class TestApp(unittest.TestCase):
    def test_calculate_sum(self):
        result = calculate_sum(2, 3)
        self.assertEqual(result, 5)

if __name__ == "__main__":
    unittest.main()`,
		"requirements.txt": "pytest==6.2.0\ncoverage==5.5",
	}

	for filename, content := range pythonFiles {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test analyzer
	analyzer := NewCodeQualityAnalyzer()
	languages := []string{"python"}
	metrics, err := analyzer.Analyze(tempDir, languages)
	if err != nil {
		t.Fatalf("Python quality analysis failed: %v", err)
	}

	// Validate results
	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// Should have some metrics calculated
	if metrics.TestCoverage < 0 || metrics.TestCoverage > 100 {
		t.Errorf("Test coverage should be between 0 and 100, got: %f", metrics.TestCoverage)
	}
}

func TestCodeQualityAnalyzer_AnalyzeJavaScriptProject(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "js_quality_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create JavaScript test files
	jsFiles := map[string]string{
		"app.js": `function calculateSum(a, b) {
  // Calculate the sum of two numbers
  const result = a + b;
  return result;
}

console.log(calculateSum(1, 2));`,
		"app.test.js": `const { calculateSum } = require('./app');

test('calculates sum correctly', () => {
  expect(calculateSum(2, 3)).toBe(5);
});`,
		"package.json": `{
  "name": "test-app",
  "version": "1.0.0",
  "scripts": {
    "test": "jest",
    "lint": "eslint ."
  },
  "devDependencies": {
    "jest": "^27.0.0",
    "eslint": "^8.0.0"
  }
}`,
	}

	for filename, content := range jsFiles {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test analyzer
	analyzer := NewCodeQualityAnalyzer()
	languages := []string{"javascript"}
	metrics, err := analyzer.Analyze(tempDir, languages)
	if err != nil {
		t.Fatalf("JavaScript quality analysis failed: %v", err)
	}

	// Validate results
	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// Should have some metrics calculated
	if metrics.TestCoverage < 0 || metrics.TestCoverage > 100 {
		t.Errorf("Test coverage should be between 0 and 100, got: %f", metrics.TestCoverage)
	}
}

func TestCodeQualityAnalyzer_SecurityAnalysis(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "security_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create Go file with potential security issues
	goFile := `package main

import (
	"os"
	"os/exec"
)

func main() {
	// Potential security issue: command injection
	cmd := exec.Command("ls", os.Args[1])
	cmd.Run()

	// Another potential issue: hardcoded password
	password := "admin123"
	_ = password
}`
	path := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(path, []byte(goFile), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test analyzer
	analyzer := NewCodeQualityAnalyzer()
	languages := []string{"go"}
	metrics, err := analyzer.Analyze(tempDir, languages)
	if err != nil {
		t.Fatalf("Security analysis failed: %v", err)
	}

	// Validate results
	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// Security issues should be detected (if gosec is available)
	// Maintainability should be calculated
	if metrics.Maintainability < 0 || metrics.Maintainability > 100 {
		t.Errorf("Maintainability should be between 0 and 100, got: %f", metrics.Maintainability)
	}
}

// Helper functions

func createQualityTestProject(t *testing.T, dir string) {
	// Create a comprehensive test project
	files := map[string]string{
		"go.mod": `module test-quality

go 1.21

require github.com/stretchr/testify v1.8.0`,
		"main.go": `package main

import "fmt"

// CalculateSum calculates the sum of two numbers
func CalculateSum(a, b int) int {
	result := a + b
	return result
}

func main() {
	fmt.Println(CalculateSum(1, 2))
}`,
		"main_test.go": `package main

import "testing"

func TestCalculateSum(t *testing.T) {
	result := CalculateSum(2, 3)
	expected := 5
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}`,
		"utils.go": `package main

// ComplexFunction demonstrates cyclomatic complexity
func ComplexFunction(x, y int) int {
	if x > 0 {
		if y > 0 {
			return x + y
		} else if y < 0 {
			return x - y
		} else {
			return x
		}
	} else if x < 0 {
		if y > 0 {
			return x * y
		} else {
			return x / y
		}
	}
	return 0
}`,
		"README.md": `# Test Quality Project

This project demonstrates code quality analysis.`,
		"package.json": `{
  "name": "test-quality",
  "version": "1.0.0"
}`,
		"requirements.txt": "pytest==6.2.0",
	}

	for filename, content := range files {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
