package analyzer

import (
	"os"
	"path/filepath"
	"strings"
)

// ArchitectureAnalyzer analyzes project architecture patterns
type ArchitectureAnalyzer struct{}

// NewArchitectureAnalyzer creates a new architecture analyzer
func NewArchitectureAnalyzer() *ArchitectureAnalyzer {
	return &ArchitectureAnalyzer{}
}

// DetectArchitecture detects the architecture pattern based on file structure and patterns
func (a *ArchitectureAnalyzer) DetectArchitecture(fileStructure map[string]int, projectPath string) string {
	// Analyze directory structure for architectural patterns
	patterns := a.analyzeDirectoryPatterns(projectPath)

	// Analyze file organization patterns
	filePatterns := a.analyzeFilePatterns(fileStructure)

	// Combine patterns to determine architecture
	return a.determineArchitectureType(patterns, filePatterns, fileStructure)
}

// CalculateComplexity calculates project complexity based on multiple factors
func (a *ArchitectureAnalyzer) CalculateComplexity(fileStructure map[string]int, dependencies map[string]string) string {
	totalFiles := 0
	for ext, count := range fileStructure {
		if ext != "no-extension" {
			totalFiles += count
		}
	}

	depCount := len(dependencies)

	// Calculate complexity score based on multiple factors
	score := 0.0

	// File count factor
	if totalFiles > 1000 {
		score += 3
	} else if totalFiles > 500 {
		score += 2
	} else if totalFiles > 100 {
		score += 1
	}

	// Dependency count factor
	if depCount > 50 {
		score += 3
	} else if depCount > 20 {
		score += 2
	} else if depCount > 10 {
		score += 1
	}

	// File type diversity factor
	uniqueExtensions := 0
	for ext := range fileStructure {
		if ext != "no-extension" {
			uniqueExtensions++
		}
	}
	if uniqueExtensions > 10 {
		score += 1
	}

	// Language mixing factor (multiple programming languages)
	langCount := a.countProgrammingLanguages(fileStructure)
	if langCount > 2 {
		score += 1
	}

	switch {
	case score >= 6:
		return "very-high"
	case score >= 4:
		return "high"
	case score >= 2:
		return "medium"
	default:
		return "low"
	}
}

// analyzeDirectoryPatterns analyzes the directory structure for architectural patterns
func (a *ArchitectureAnalyzer) analyzeDirectoryPatterns(projectPath string) map[string]int {
	patterns := make(map[string]int)

	// Common architectural patterns to look for
	architecturalIndicators := map[string][]string{
		"layered": {
			"presentation", "business", "data", "infrastructure",
			"ui", "services", "repositories", "models",
			"controllers", "views", "middleware",
		},
		"microservices": {
			"services", "api", "gateway", "discovery",
			"config", "registry", "auth", "monitoring",
		},
		"hexagonal": {
			"domain", "application", "infrastructure", "adapters",
			"ports", "core", "interfaces",
		},
		"clean": {
			"usecases", "entities", "interfaces", "frameworks",
			"controllers", "presenters", "gateways",
		},
		"ddd": {
			"domain", "application", "infrastructure", "presentation",
			"entities", "valueobjects", "aggregates", "repositories",
		},
		"modular": {
			"modules", "components", "shared", "common", "utils",
		},
		"monolithic": {
			"src", "lib", "app", "main",
		},
	}

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() || shouldSkipDir(info.Name()) {
			return nil
		}

		dirName := strings.ToLower(info.Name())

		// Check which architectural patterns this directory indicates
		for pattern, indicators := range architecturalIndicators {
			for _, indicator := range indicators {
				if strings.Contains(dirName, indicator) {
					patterns[pattern]++
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		// If directory analysis fails, return empty patterns
		return patterns
	}

	return patterns
}

// analyzeFilePatterns analyzes file organization patterns
func (a *ArchitectureAnalyzer) analyzeFilePatterns(fileStructure map[string]int) map[string]int {
	patterns := make(map[string]int)

	// Analyze file extension distribution
	totalFiles := 0
	for _, count := range fileStructure {
		totalFiles += count
	}

	// Check for test files
	testFiles := 0
	for ext, count := range fileStructure {
		if strings.Contains(ext, "test") || strings.Contains(ext, "spec") {
			testFiles += count
		}
	}
	if testFiles > 0 {
		testRatio := float64(testFiles) / float64(totalFiles)
		if testRatio > 0.3 {
			patterns["test-driven"]++
		} else if testRatio > 0.1 {
			patterns["well-tested"]++
		}
	}

	// Check for configuration files
	configFiles := 0
	configExtensions := []string{".yaml", ".yml", ".json", ".toml", ".ini", ".cfg", ".conf", ".properties"}
	for _, ext := range configExtensions {
		if count, exists := fileStructure[ext]; exists {
			configFiles += count
		}
	}
	if configFiles > 5 {
		patterns["configurable"]++
	}

	// Check for documentation
	docFiles := 0
	docExtensions := []string{".md", ".txt", ".rst", ".adoc"}
	for _, ext := range docExtensions {
		if count, exists := fileStructure[ext]; exists {
			docFiles += count
		}
	}
	if docFiles > 3 {
		patterns["well-documented"]++
	}

	return patterns
}

// determineArchitectureType combines all patterns to determine the primary architecture
func (a *ArchitectureAnalyzer) determineArchitectureType(dirPatterns, filePatterns map[string]int, fileStructure map[string]int) string {
	// Count total files to help determine scale
	totalFiles := 0
	for _, count := range fileStructure {
		totalFiles += count
	}

	// Score different architecture types
	scores := make(map[string]int)

	// Directory-based scoring
	for arch, count := range dirPatterns {
		scores[arch] += count * 2 // Directory matches are strong indicators
	}

	// File pattern-based scoring
	for pattern, count := range filePatterns {
		switch pattern {
		case "test-driven":
			scores["test-driven"] += count * 3
		case "well-tested":
			scores["modular"] += count
		case "configurable":
			scores["microservices"] += count
		case "well-documented":
			scores["layered"] += count
		}
	}

	// Size-based heuristics
	if totalFiles < 50 {
		scores["simple"] += 5
	} else if totalFiles < 200 {
		scores["modular"] += 3
	} else if totalFiles < 1000 {
		scores["layered"] += 2
	} else {
		scores["complex"] += 3
	}

	// Language count heuristic
	langCount := a.countProgrammingLanguages(fileStructure)
	if langCount > 3 {
		scores["polyglot"] += 2
	} else if langCount == 1 {
		scores["monolithic"] += 1
	}

	// Find the architecture with the highest score
	maxScore := 0
	bestArch := "unknown"

	for arch, score := range scores {
		if score > maxScore {
			maxScore = score
			bestArch = arch
		}
	}

	// Special cases and refinements
	if bestArch == "unknown" {
		if totalFiles < 20 {
			return "minimal"
		} else if langCount == 1 {
			return "monolithic"
		} else {
			return "modular"
		}
	}

	// Refine based on combinations
	if bestArch == "layered" && scores["ddd"] > 0 {
		return "domain-driven"
	}

	if bestArch == "microservices" && scores["configurable"] > 2 {
		return "microservices"
	}

	return bestArch
}

// countProgrammingLanguages counts the number of programming languages used
func (a *ArchitectureAnalyzer) countProgrammingLanguages(fileStructure map[string]int) int {
	programmingExtensions := map[string]bool{
		// Go
		".go": true,
		// JavaScript/TypeScript
		".js": true, ".jsx": true, ".ts": true, ".tsx": true,
		// Python
		".py": true,
		// Java
		".java": true,
		// C#
		".cs": true,
		// C/C++
		".c": true, ".cpp": true, ".cc": true, ".cxx": true, ".h": true, ".hpp": true,
		// Rust
		".rs": true,
		// Ruby
		".rb": true,
		// PHP
		".php": true,
		// Swift
		".swift": true,
		// Kotlin
		".kt": true,
		// Scala
		".scala": true,
		// Haskell
		".hs": true,
		// Elixir
		".ex": true, ".exs": true,
		// Clojure
		".clj": true, ".cljs": true,
		// R
		".r": true,
		// Julia
		".jl": true,
		// Lua
		".lua": true,
		// Perl
		".pl": true, ".pm": true,
		// Shell
		".sh": true, ".bash": true, ".zsh": true,
		// PowerShell
		".ps1": true,
	}

	languages := make(map[string]bool)
	for ext := range fileStructure {
		if programmingExtensions[ext] {
			// Group similar extensions
			switch ext {
			case ".js", ".jsx", ".ts", ".tsx":
				languages["javascript/typescript"] = true
			case ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp":
				languages["c/cpp"] = true
			case ".clj", ".cljs":
				languages["clojure"] = true
			case ".ex", ".exs":
				languages["elixir"] = true
			case ".pl", ".pm":
				languages["perl"] = true
			case ".sh", ".bash", ".zsh":
				languages["shell"] = true
			default:
				languages[ext] = true
			}
		}
	}

	return len(languages)
}
