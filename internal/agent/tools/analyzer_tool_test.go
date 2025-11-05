package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/SamyRai/juleson/internal/agent"
)

func TestAnalyzerTool_CanHandle(t *testing.T) {
	tool := NewAnalyzerTool()

	tests := []struct {
		task     agent.Task
		expected bool
	}{
		{
			task: agent.Task{
				Description: "analyze the codebase",
				Prompt:      "Please analyze the project structure",
			},
			expected: true,
		},
		{
			task: agent.Task{
				Description: "review code quality",
				Prompt:      "Check the code quality metrics",
			},
			expected: true,
		},
		{
			task: agent.Task{
				Description: "build the project",
				Prompt:      "Compile the application",
			},
			expected: false,
		},
		{
			task: agent.Task{
				Description: "write documentation",
				Prompt:      "Create README file",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.task.Description, func(t *testing.T) {
			result := tool.CanHandle(tt.task)
			if result != tt.expected {
				t.Errorf("CanHandle(%q) = %v, expected %v", tt.task.Description, result, tt.expected)
			}
		})
	}
}

func TestAnalyzerTool_Execute(t *testing.T) {
	// Create a temporary test directory
	tempDir := t.TempDir()

	// Create test project files
	createTestProjectForTool(t, tempDir)

	// Create analyzer tool
	tool := NewAnalyzerTool()

	// Create parameters for analysis
	params := map[string]interface{}{
		"action":       "analyze_project",
		"project_path": tempDir,
	}

	// Execute analysis
	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	// Validate result
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !result.Success {
		t.Errorf("Analysis should succeed, got error: %v", result.Error)
	}

	// Check output contains expected fields
	output, ok := result.Output.(map[string]interface{})
	if !ok {
		t.Fatal("Output should be a map")
	}

	if _, exists := output["project_name"]; !exists {
		t.Error("Output should contain project_name")
	}

	if _, exists := output["languages"]; !exists {
		t.Error("Output should contain languages")
	}

	if _, exists := output["project_type"]; !exists {
		t.Error("Output should contain project_type")
	}
}

func TestAnalyzerTool_ExecuteWithInvalidPath(t *testing.T) {
	tool := NewAnalyzerTool()

	params := map[string]interface{}{
		"action":       "analyze_project",
		"project_path": "/nonexistent/path",
	}

	_, err := tool.Execute(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

// Helper functions

func createTestProjectForTool(t *testing.T, dir string) {
	files := map[string]string{
		"go.mod": `module test-project

go 1.21

require github.com/stretchr/testify v1.8.0`,
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`,
		"utils.go": `package main

func Add(a, b int) int {
	return a + b
}`,
		"README.md": `# Test Project

This is a test project for analyzer tool testing.`,
		"package.json": `{
  "name": "test-project",
  "version": "1.0.0"
}`,
	}

	for filename, content := range files {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}
}

func createQualityTestProjectForTool(t *testing.T, dir string) {
	files := map[string]string{
		"go.mod": `module quality-test

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
		"complex.go": `package main

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
		"README.md": `# Quality Test Project

This project demonstrates code quality analysis.`,
	}

	for filename, content := range files {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}
}
