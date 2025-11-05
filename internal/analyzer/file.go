package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileStructureAnalyzer analyzes project file structure
type FileStructureAnalyzer struct{}

// NewFileStructureAnalyzer creates a new file structure analyzer
func NewFileStructureAnalyzer() *FileStructureAnalyzer {
	return &FileStructureAnalyzer{}
}

// Analyze analyzes the file structure and returns file counts by extension
func (f *FileStructureAnalyzer) Analyze(projectPath string) (map[string]int, error) {
	if projectPath == "" {
		return nil, fmt.Errorf("project path cannot be empty")
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project path does not exist: %s", projectPath)
	}

	structure := make(map[string]int)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log the error but continue walking
			return nil
		}

		// Skip directories and common ignore patterns
		if info.IsDir() {
			name := info.Name()
			if shouldSkipDir(name) {
				return filepath.SkipDir
			}
			return nil
		}

		// Count files by extension
		ext := strings.ToLower(filepath.Ext(path))
		if ext != "" {
			structure[ext]++
		} else {
			structure["no-extension"]++
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk project directory: %w", err)
	}

	return structure, nil
}

func shouldSkipDir(name string) bool {
	skipDirs := []string{
		".git", "node_modules", "vendor", ".idea", ".vscode",
		"dist", "build", "bin", ".cache", "tmp", "__pycache__",
		".next", ".nuxt", ".vuepress", "target", ".gradle",
		"cmake-build-debug", "cmake-build-release", ".cargo",
		".bundle", "vendor/bundle", ".meteor", ".expo",
		".expo-shared", "coverage", ".nyc_output",
	}

	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}
	return false
}
