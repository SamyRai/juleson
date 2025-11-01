package analyzer

import (
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
	structure := make(map[string]int)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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

	return structure, err
}

func shouldSkipDir(name string) bool {
	skipDirs := []string{
		".git", "node_modules", "vendor", ".idea", ".vscode",
		"dist", "build", "bin", ".cache", "tmp",
	}

	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}
	return false
}
