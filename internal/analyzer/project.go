package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProjectContext contains project analysis context
type ProjectContext struct {
	ProjectPath   string              `json:"project_path"`
	ProjectName   string              `json:"project_name"`
	ProjectType   string              `json:"project_type"`
	Languages     []string            `json:"languages"`
	Frameworks    []string            `json:"frameworks"`
	Dependencies  map[string]string   `json:"dependencies"`
	FileStructure map[string]int      `json:"file_structure"`
	TestCoverage  float64             `json:"test_coverage"`
	Architecture  string              `json:"architecture"`
	Complexity    string              `json:"complexity"`
	LastModified  time.Time           `json:"last_modified"`
	GitStatus     string              `json:"git_status"`
	CustomParams  map[string]string   `json:"custom_params"`
	CodeQuality   *CodeQualityMetrics `json:"code_quality,omitempty"`
}

// ProjectAnalyzer orchestrates all analyzers to build project context
type ProjectAnalyzer struct {
	fileAnalyzer         *FileStructureAnalyzer
	languageDetector     *LanguageDetector
	dependencyAnalyzer   *DependencyAnalyzer
	architectureAnalyzer *ArchitectureAnalyzer
	gitAnalyzer          *GitAnalyzer
	qualityAnalyzer      *CodeQualityAnalyzer
}

// NewProjectAnalyzer creates a new project analyzer with all sub-analyzers
func NewProjectAnalyzer() *ProjectAnalyzer {
	return &ProjectAnalyzer{
		fileAnalyzer:         NewFileStructureAnalyzer(),
		languageDetector:     NewLanguageDetector(),
		dependencyAnalyzer:   NewDependencyAnalyzer(),
		architectureAnalyzer: NewArchitectureAnalyzer(),
		gitAnalyzer:          NewGitAnalyzer(),
		qualityAnalyzer:      NewCodeQualityAnalyzer(),
	}
}

// Analyze performs complete project analysis
func (p *ProjectAnalyzer) Analyze(projectPath string) (*ProjectContext, error) {
	// Validate input
	if projectPath == "" {
		return nil, fmt.Errorf("project path cannot be empty")
	}

	// Check if path exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project path does not exist: %s", projectPath)
	}

	projectName := filepath.Base(projectPath)

	// Analyze file structure
	fileStructure, err := p.fileAnalyzer.Analyze(projectPath)
	if err != nil {
		return nil, fmt.Errorf("file structure analysis failed: %w", err)
	}

	// Detect languages and frameworks
	languages, frameworks, err := p.languageDetector.Detect(projectPath)
	if err != nil {
		return nil, fmt.Errorf("language detection failed: %w", err)
	}

	// Analyze dependencies (non-critical - continue on error)
	dependencies, err := p.dependencyAnalyzer.Analyze(projectPath)
	if err != nil {
		// Log warning but continue with empty dependencies
		dependencies = make(map[string]string)
	}

	// Detect architecture
	architecture := p.architectureAnalyzer.DetectArchitecture(fileStructure, projectPath)

	// Calculate complexity
	complexity := p.architectureAnalyzer.CalculateComplexity(fileStructure, dependencies)

	// Get git status (non-critical - continue on error)
	gitStatus, err := p.gitAnalyzer.GetStatus(projectPath)
	if err != nil {
		gitStatus = "unknown"
	}

	// Determine project type
	projectType := determineProjectType(languages, frameworks)

	// Analyze code quality metrics (optional - continue on error)
	codeQuality, err := p.qualityAnalyzer.Analyze(projectPath, languages)
	if err != nil {
		// Code quality analysis is optional, don't fail
		codeQuality = nil
	}

	return &ProjectContext{
		ProjectPath:   projectPath,
		ProjectName:   projectName,
		ProjectType:   projectType,
		Languages:     languages,
		Frameworks:    frameworks,
		Dependencies:  dependencies,
		FileStructure: fileStructure,
		TestCoverage:  0.0, // Will be set from code quality if available
		Architecture:  architecture,
		Complexity:    complexity,
		LastModified:  time.Now(),
		GitStatus:     gitStatus,
		CustomParams:  make(map[string]string),
		CodeQuality:   codeQuality,
	}, nil
}

func determineProjectType(languages, frameworks []string) string {
	if len(languages) == 0 {
		return "unknown"
	}

	primaryLang := languages[0]

	// Check for specific frameworks
	for _, fw := range frameworks {
		switch fw {
		case "next.js":
			return "web-application"
		case "maven", "gradle":
			return "java-application"
		}
	}

	// Default to language-based type
	switch primaryLang {
	case "go":
		return "go-application"
	case "javascript", "typescript":
		return "javascript-application"
	case "python":
		return "python-application"
	case "java":
		return "java-application"
	case "rust":
		return "rust-application"
	default:
		return primaryLang + "-project"
	}
}
