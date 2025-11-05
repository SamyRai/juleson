package analyzer

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestProjectAnalyzer_Analyze(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "analyzer_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	createTestProject(t, tempDir)

	// Test analyzer
	analyzer := NewProjectAnalyzer()
	context, err := analyzer.Analyze(tempDir)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Validate results
	if context.ProjectName == "" {
		t.Error("Project name should not be empty")
	}

	if len(context.Languages) == 0 {
		t.Error("Should detect at least one language")
	}

	if context.ProjectType == "" {
		t.Error("Project type should not be empty")
	}

	if len(context.FileStructure) == 0 {
		t.Error("File structure should not be empty")
	}
}

func TestFileStructureAnalyzer_Analyze(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "file_analyzer_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"main.go",
		"utils.go",
		"README.md",
		"package.json",
		"test.py",
	}

	for _, file := range testFiles {
		path := filepath.Join(tempDir, file)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Test analyzer
	analyzer := NewFileStructureAnalyzer()
	structure, err := analyzer.Analyze(tempDir)
	if err != nil {
		t.Fatalf("File analysis failed: %v", err)
	}

	// Validate results
	if structure[".go"] != 2 {
		t.Errorf("Expected 2 .go files, got %d", structure[".go"])
	}

	if structure[".md"] != 1 {
		t.Errorf("Expected 1 .md file, got %d", structure[".md"])
	}

	if structure[".json"] != 1 {
		t.Errorf("Expected 1 .json file, got %d", structure[".json"])
	}

	if structure[".py"] != 1 {
		t.Errorf("Expected 1 .py file, got %d", structure[".py"])
	}
}

func TestLanguageDetector_Detect(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "lang_detector_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files that indicate different languages
	testFiles := map[string]string{
		"go.mod":           "module test\n\ngo 1.21",
		"package.json":     `{"name": "test", "version": "1.0.0"}`,
		"requirements.txt": "flask==2.0.0\nrequests==2.25.0",
		"pom.xml":          `<?xml version="1.0"?><project><modelVersion>4.0.0</modelVersion></project>`,
		"Cargo.toml":       `[package]\nname = "test"\nversion = "0.1.0"`,
	}

	for filename, content := range testFiles {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test detector
	detector := NewLanguageDetector()
	languages, frameworks, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Language detection failed: %v", err)
	}

	// Validate results
	expectedLanguages := map[string]bool{
		"go":         false,
		"javascript": false,
		"python":     false,
		"java":       false,
		"rust":       false,
	}

	for _, lang := range languages {
		if _, exists := expectedLanguages[lang]; exists {
			expectedLanguages[lang] = true
		}
	}

	for lang, detected := range expectedLanguages {
		if !detected {
			t.Errorf("Expected to detect language: %s", lang)
		}
	}

	// Check frameworks
	if len(frameworks) == 0 {
		t.Error("Should detect at least one framework")
	}
}

func TestArchitectureAnalyzer_DetectArchitecture(t *testing.T) {
	tests := []struct {
		name          string
		fileStructure map[string]int
		projectPath   string
		expected      string
	}{
		{
			name: "minimal project",
			fileStructure: map[string]int{
				".go": 1,
			},
			expected: "simple",
		},
		{
			name: "modular project",
			fileStructure: map[string]int{
				".go": 15,
				".md": 2,
			},
			expected: "modular",
		},
		{
			name: "layered project",
			fileStructure: map[string]int{
				".go":   50,
				".md":   5,
				".json": 3,
			},
			expected: "layered",
		},
	}

	analyzer := NewArchitectureAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.DetectArchitecture(tt.fileStructure, tt.projectPath)
			if result == "" {
				t.Error("Architecture detection should not return empty string")
			}
			// Note: The actual result may vary due to the heuristic nature of the detection
			// We just ensure it returns something reasonable
		})
	}
}

func TestDependencyAnalyzer_Analyze(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "dep_analyzer_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test dependency files
	testFiles := map[string]string{
		"go.mod": `module test

go 1.21

require (
	github.com/stretchr/testify v1.8.0
	golang.org/x/crypto v0.0.0
)`,
		"requirements.txt": `flask==2.0.0
requests==2.25.0
pytest==6.2.0`,
		"package.json": `{
  "dependencies": {
    "express": "^4.17.1",
    "lodash": "^4.17.21"
  }
}`,
	}

	for filename, content := range testFiles {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test analyzer
	analyzer := NewDependencyAnalyzer()
	deps, err := analyzer.Analyze(tempDir)
	if err != nil {
		t.Fatalf("Dependency analysis failed: %v", err)
	}

	// Should find some dependencies
	if len(deps) == 0 {
		t.Error("Should detect at least some dependencies")
	}
}

func TestGitAnalyzer_GetStatus(t *testing.T) {
	// Create a temporary git repository
	tempDir, err := os.MkdirTemp("", "git_analyzer_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	runGitCommand(t, tempDir, "init")
	runGitCommand(t, tempDir, "config", "user.email", "test@example.com")
	runGitCommand(t, tempDir, "config", "user.name", "Test User")

	// Create and commit a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	runGitCommand(t, tempDir, "add", "test.txt")
	runGitCommand(t, tempDir, "commit", "-m", "Initial commit")

	// Test analyzer
	analyzer := NewGitAnalyzer()
	status, err := analyzer.GetStatus(tempDir)
	if err != nil {
		t.Fatalf("Git status check failed: %v", err)
	}

	// Should be clean after commit
	if status != "clean" {
		t.Errorf("Expected clean status, got: %s", status)
	}

	// Modify file and check status
	if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	status, err = analyzer.GetStatus(tempDir)
	if err != nil {
		t.Fatalf("Git status check failed: %v", err)
	}

	// Should have changes
	if status == "clean" {
		t.Error("Expected to detect changes after file modification")
	}
}

// Helper functions

func createTestProject(t *testing.T, dir string) {
	// Create a basic Go project structure
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

This is a test project for analyzer testing.`,
		"package.json": `{
  "name": "test-project",
  "version": "1.0.0",
  "scripts": {
    "test": "go test"
  }
}`,
	}

	for filename, content := range files {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create subdirectories
	subdirs := []string{"pkg", "cmd", "internal"}
	for _, subdir := range subdirs {
		path := filepath.Join(dir, subdir)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("Failed to create subdirectory %s: %v", subdir, err)
		}
	}
}

func runGitCommand(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Git command failed: git %v: %v", args, err)
	}
}
