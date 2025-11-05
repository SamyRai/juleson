package analyzer

import (
	"os"
	"testing"
)

func TestCoverageAnalyzer(t *testing.T) {
	// Create a temporary directory for our mock project
	tmpDir, err := os.MkdirTemp("", "test-project-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock Go file
	mainGoContent := `
package main

func main() {}

func Add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(tmpDir+"/main.go", []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	// Create a mock Go test file
	mainTestGoContent := `
package main

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Error("Add(1, 2) should be 3")
	}
}
`
	if err := os.WriteFile(tmpDir+"/main_test.go", []byte(mainTestGoContent), 0644); err != nil {
		t.Fatalf("Failed to write main_test.go: %v", err)
	}

	// Create a go.mod file
	goModContent := "module test-project"
	if err := os.WriteFile(tmpDir+"/go.mod", []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	analyzer := NewCoverageAnalyzer()
	coverage, err := analyzer.Analyze(tmpDir)
	if err != nil {
		t.Fatalf("Coverage analysis failed: %v", err)
	}

	// In this controlled environment, we expect 100% coverage
	if coverage != 100.0 {
		t.Errorf("Expected 100.0%% coverage, but got %.1f%%", coverage)
	}
}
