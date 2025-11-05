package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LanguageDetector detects programming languages used in a project
type LanguageDetector struct{}

// NewLanguageDetector creates a new language detector
func NewLanguageDetector() *LanguageDetector {
	return &LanguageDetector{}
}

// Detect detects languages based on file extensions and special files
func (l *LanguageDetector) Detect(projectPath string) ([]string, []string, error) {
	if projectPath == "" {
		return nil, nil, fmt.Errorf("project path cannot be empty")
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("project path does not exist: %s", projectPath)
	}

	languageMarkers := make(map[string]bool)
	frameworkMarkers := make(map[string]bool)

	// Check for language-specific files and patterns
	err := l.detectByFiles(projectPath, languageMarkers, frameworkMarkers)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to detect languages by files: %w", err)
	}

	// Check for file extensions in the project
	err = l.detectByExtensions(projectPath, languageMarkers)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to detect languages by extensions: %w", err)
	}

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

// detectByFiles detects languages and frameworks by checking for specific files
func (l *LanguageDetector) detectByFiles(projectPath string, languageMarkers, frameworkMarkers map[string]bool) error {
	// Go
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

	// Go ecosystem
	checkFile("go.mod", "go", "")
	checkFile("go.sum", "go", "")
	checkFile("main.go", "go", "")
	checkFile("cmd/main.go", "go", "")

	// JavaScript/TypeScript ecosystem
	checkFile("package.json", "javascript", "")
	checkFile("tsconfig.json", "typescript", "")
	checkFile("yarn.lock", "javascript", "")
	checkFile("pnpm-lock.yaml", "javascript", "")
	checkFile("bun.lockb", "javascript", "")
	checkFile("node_modules", "javascript", "")

	// Python ecosystem
	checkFile("requirements.txt", "python", "")
	checkFile("Pipfile", "python", "")
	checkFile("pyproject.toml", "python", "")
	checkFile("setup.py", "python", "")
	checkFile("setup.cfg", "python", "")
	checkFile("poetry.lock", "python", "")
	checkFile("Pipfile.lock", "python", "")

	// Java ecosystem
	checkFile("pom.xml", "java", "maven")
	checkFile("build.gradle", "java", "gradle")
	checkFile("build.gradle.kts", "java", "gradle")
	checkFile("settings.gradle", "java", "gradle")

	// Kotlin
	checkFile("build.gradle.kts", "kotlin", "")

	// C#/.NET ecosystem
	checkFile("*.csproj", "csharp", "")
	checkFile("*.sln", "csharp", "")
	checkFile("*.fsproj", "fsharp", "")
	checkFile("*.vbproj", "vbnet", "")
	checkFile("global.json", "csharp", "")
	checkFile("Directory.Build.props", "csharp", "")

	// Rust ecosystem
	checkFile("Cargo.toml", "rust", "")
	checkFile("Cargo.lock", "rust", "")

	// C/C++ ecosystem
	checkFile("CMakeLists.txt", "cpp", "")
	checkFile("Makefile", "cpp", "")
	checkFile("configure.ac", "cpp", "")
	checkFile("meson.build", "cpp", "")

	// Ruby ecosystem
	checkFile("Gemfile", "ruby", "")
	checkFile("Gemfile.lock", "ruby", "")
	checkFile("Rakefile", "ruby", "")

	// PHP ecosystem
	checkFile("composer.json", "php", "")
	checkFile("composer.lock", "php", "")
	checkFile("artisan", "php", "laravel")

	// Swift ecosystem
	checkFile("Package.swift", "swift", "")

	// Dart/Flutter ecosystem
	checkFile("pubspec.yaml", "dart", "")
	checkFile("pubspec.lock", "dart", "")

	// Scala ecosystem
	checkFile("build.sbt", "scala", "")

	// Haskell ecosystem
	checkFile("*.cabal", "haskell", "")
	checkFile("stack.yaml", "haskell", "")

	// Elixir ecosystem
	checkFile("mix.exs", "elixir", "")

	// Clojure ecosystem
	checkFile("project.clj", "clojure", "")
	checkFile("deps.edn", "clojure", "")

	// R ecosystem
	checkFile("DESCRIPTION", "r", "")

	// Julia ecosystem
	checkFile("Project.toml", "julia", "")

	// Lua ecosystem
	checkFile("*.rockspec", "lua", "")

	// Perl ecosystem
	checkFile("Makefile.PL", "perl", "")
	checkFile("Build.PL", "perl", "")

	// Frameworks - Next.js
	checkFile("next.config.js", "", "next.js")
	checkFile("next.config.ts", "", "next.js")
	checkFile("next.config.mjs", "", "next.js")

	// React
	checkFile("src/App.js", "", "react")
	checkFile("src/App.tsx", "", "react")
	checkFile("src/index.js", "", "react")

	// Vue.js
	checkFile("vue.config.js", "", "vue.js")

	// Angular
	checkFile("angular.json", "", "angular")

	// Django
	checkFile("manage.py", "", "django")
	checkFile("settings.py", "", "django")

	// Flask
	checkFile("app.py", "", "flask")

	// FastAPI
	checkFile("main.py", "", "fastapi")

	// Express.js
	checkFile("app.js", "", "express")

	// Spring Boot
	checkFile("src/main/java", "", "spring-boot")

	// ASP.NET Core
	checkFile("Startup.cs", "", "asp.net-core")
	checkFile("Program.cs", "", "asp.net-core")

	// Rails
	checkFile("config/routes.rb", "", "rails")

	// Laravel
	checkFile("artisan", "", "laravel")

	// Symfony
	checkFile("bin/console", "", "symfony")

	// Docker
	checkFile("Dockerfile", "", "docker")
	checkFile("docker-compose.yml", "", "docker-compose")

	// Kubernetes
	checkFile("k8s", "", "kubernetes")
	checkFile("kubernetes", "", "kubernetes")

	// Terraform
	checkFile("*.tf", "", "terraform")

	// Ansible
	checkFile("ansible.cfg", "", "ansible")
	checkFile("playbook.yml", "", "ansible")

	// Configuration files that indicate frameworks
	checkFile("tailwind.config.js", "", "tailwind-css")
	checkFile("webpack.config.js", "", "webpack")
	checkFile("vite.config.js", "", "vite")
	checkFile("rollup.config.js", "", "rollup")

	return nil
}

// detectByExtensions detects languages by scanning file extensions
func (l *LanguageDetector) detectByExtensions(projectPath string, languageMarkers map[string]bool) error {
	extensions := make(map[string]int)

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

		// Count file extensions
		ext := strings.ToLower(filepath.Ext(path))
		if ext != "" {
			extensions[ext]++
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Map extensions to languages
	extensionMap := map[string]string{
		".go":    "go",
		".js":    "javascript",
		".jsx":   "javascript",
		".ts":    "typescript",
		".tsx":   "typescript",
		".py":    "python",
		".java":  "java",
		".kt":    "kotlin",
		".scala": "scala",
		".cs":    "csharp",
		".fs":    "fsharp",
		".vb":    "vbnet",
		".rs":    "rust",
		".cpp":   "cpp",
		".cc":    "cpp",
		".cxx":   "cpp",
		".c":     "c",
		".h":     "c",
		".hpp":   "cpp",
		".rb":    "ruby",
		".php":   "php",
		".swift": "swift",
		".dart":  "dart",
		".hs":    "haskell",
		".ex":    "elixir",
		".exs":   "elixir",
		".clj":   "clojure",
		".cljs":  "clojure",
		".r":     "r",
		".jl":    "julia",
		".lua":   "lua",
		".pl":    "perl",
		".pm":    "perl",
		".sh":    "shell",
		".bash":  "shell",
		".zsh":   "shell",
		".fish":  "shell",
		".ps1":   "powershell",
		".sql":   "sql",
		".yaml":  "yaml",
		".yml":   "yaml",
		".json":  "json",
		".xml":   "xml",
		".html":  "html",
		".css":   "css",
		".scss":  "scss",
		".sass":  "sass",
		".less":  "less",
		".md":    "markdown",
		".toml":  "toml",
		".ini":   "ini",
		".cfg":   "config",
		".conf":  "config",
	}

	// Detect languages from extensions (only if we have multiple files of that type)
	for ext, count := range extensions {
		if count >= 2 { // Require at least 2 files to avoid false positives
			if lang, exists := extensionMap[ext]; exists {
				languageMarkers[lang] = true
			}
		}
	}

	return nil
}
