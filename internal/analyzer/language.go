package analyzer

import (
	"os"
	"path/filepath"
)

// LanguageDetector detects programming languages used in a project
type LanguageDetector struct{}

// NewLanguageDetector creates a new language detector
func NewLanguageDetector() *LanguageDetector {
	return &LanguageDetector{}
}

// Detect detects languages based on file extensions and special files
func (l *LanguageDetector) Detect(projectPath string) ([]string, []string, error) {
	languageMarkers := make(map[string]bool)
	frameworkMarkers := make(map[string]bool)

	// Check for language-specific files
	checkFile := func(filename, language, framework string) {
		path := filepath.Join(projectPath, filename)
		if _, err := os.Stat(path); err == nil {
			if language != "" {
				languageMarkers[language] = true
			}
			if framework != "" {
				frameworkMarkers[framework] = true
			}
		}
	}

	// Go
	checkFile("go.mod", "go", "")
	checkFile("go.sum", "go", "")

	// JavaScript/TypeScript
	checkFile("package.json", "javascript", "")
	checkFile("tsconfig.json", "typescript", "")
	checkFile("yarn.lock", "javascript", "")
	checkFile("pnpm-lock.yaml", "javascript", "")

	// Python
	checkFile("requirements.txt", "python", "")
	checkFile("Pipfile", "python", "")
	checkFile("pyproject.toml", "python", "")
	checkFile("setup.py", "python", "")

	// Java
	checkFile("pom.xml", "java", "maven")
	checkFile("build.gradle", "java", "gradle")
	checkFile("build.gradle.kts", "java", "gradle")

	// Rust
	checkFile("Cargo.toml", "rust", "")

	// Ruby
	checkFile("Gemfile", "ruby", "")

	// PHP
	checkFile("composer.json", "php", "")

	// .NET
	checkFile("*.csproj", "csharp", "")
	checkFile("*.sln", "csharp", "")

	// Frameworks - Next.js
	checkFile("next.config.js", "", "next.js")
	checkFile("next.config.ts", "", "next.js")

	// Convert maps to slices
	languages := make([]string, 0, len(languageMarkers))
	for lang := range languageMarkers {
		languages = append(languages, lang)
	}

	frameworks := make([]string, 0, len(frameworkMarkers))
	for fw := range frameworkMarkers {
		frameworks = append(frameworks, fw)
	}

	// Default if nothing detected
	if len(languages) == 0 {
		languages = []string{"unknown"}
	}

	return languages, frameworks, nil
}
