package analyzer

import (
	"path/filepath"
	"time"
)

// ProjectContext contains project analysis context
type ProjectContext struct {
	ProjectPath   string            `json:"project_path"`
	ProjectName   string            `json:"project_name"`
	ProjectType   string            `json:"project_type"`
	Languages     []string          `json:"languages"`
	Frameworks    []string          `json:"frameworks"`
	Dependencies  map[string]string `json:"dependencies"`
	FileStructure map[string]int    `json:"file_structure"`
	TestCoverage  float64           `json:"test_coverage"`
	Architecture  string            `json:"architecture"`
	Complexity    string            `json:"complexity"`
	LastModified  time.Time         `json:"last_modified"`
	GitStatus     string            `json:"git_status"`
	CustomParams  map[string]string `json:"custom_params"`
}

// ProjectAnalyzer orchestrates all analyzers to build project context
type ProjectAnalyzer struct {
	fileAnalyzer         *FileStructureAnalyzer
	languageDetector     *LanguageDetector
	dependencyAnalyzer   *DependencyAnalyzer
	architectureAnalyzer *ArchitectureAnalyzer
	gitAnalyzer          *GitAnalyzer
	coverageAnalyzer     *CoverageAnalyzer
}

// NewProjectAnalyzer creates a new project analyzer with all sub-analyzers
func NewProjectAnalyzer() *ProjectAnalyzer {
	return &ProjectAnalyzer{
		fileAnalyzer:         NewFileStructureAnalyzer(),
		languageDetector:     NewLanguageDetector(),
		dependencyAnalyzer:   NewDependencyAnalyzer(),
		architectureAnalyzer: NewArchitectureAnalyzer(),
		gitAnalyzer:          NewGitAnalyzer(),
		coverageAnalyzer:     NewCoverageAnalyzer(),
	}
}

// Analyze performs complete project analysis
func (p *ProjectAnalyzer) Analyze(projectPath string) (*ProjectContext, error) {
	projectName := filepath.Base(projectPath)

	// Analyze file structure
	fileStructure, err := p.fileAnalyzer.Analyze(projectPath)
	if err != nil {
		return nil, err
	}

	// Detect languages and frameworks
	languages, frameworks, err := p.languageDetector.Detect(projectPath)
	if err != nil {
		return nil, err
	}

	// Analyze dependencies
	dependencies, err := p.dependencyAnalyzer.Analyze(projectPath)
	if err != nil {
		return nil, err
	}

	// Detect architecture
	architecture := p.architectureAnalyzer.DetectArchitecture(fileStructure)

	// Calculate complexity
	complexity := p.architectureAnalyzer.CalculateComplexity(fileStructure, dependencies)

	// Get git status
	gitStatus, err := p.gitAnalyzer.GetStatus(projectPath)
	if err != nil {
		gitStatus = "unknown"
	}

	// Get test coverage
	testCoverage, err := p.coverageAnalyzer.Analyze(projectPath)
	if err != nil {
		// For now, we'll log the error but not fail the analysis
		// In a real application, this might be handled more gracefully
		testCoverage = 0.0
	}

	// Determine project type
	projectType := determineProjectType(languages, frameworks)

	return &ProjectContext{
		ProjectPath:   projectPath,
		ProjectName:   projectName,
		ProjectType:   projectType,
		Languages:     languages,
		Frameworks:    frameworks,
		Dependencies:  dependencies,
		FileStructure: fileStructure,
		TestCoverage:  testCoverage,
		Architecture:  architecture,
		Complexity:    complexity,
		LastModified:  time.Now(),
		GitStatus:     gitStatus,
		CustomParams:  make(map[string]string),
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
