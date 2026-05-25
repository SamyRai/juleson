package analyzer

// CodeQualityAnalyzer orchestrates code quality analysis.
type CodeQualityAnalyzer struct {
	coverage   *coverageAnalyzer
	complexity *complexityAnalyzer
	security   *securityAnalyzer
	smells     *smellAnalyzer
	scorer     *maintainabilityScorer
}

// NewCodeQualityAnalyzer creates a new code quality analyzer
func NewCodeQualityAnalyzer() *CodeQualityAnalyzer {
	return &CodeQualityAnalyzer{
		coverage:   &coverageAnalyzer{},
		complexity: &complexityAnalyzer{},
		security:   &securityAnalyzer{},
		smells:     &smellAnalyzer{},
		scorer:     &maintainabilityScorer{},
	}
}

type coverageAnalyzer struct{}
type complexityAnalyzer struct{}
type securityAnalyzer struct{}
type smellAnalyzer struct{}
type maintainabilityScorer struct{}

// CodeQualityMetrics represents various code quality metrics
type CodeQualityMetrics struct {
	TestCoverage      float64            `json:"test_coverage"`
	CodeComplexity    float64            `json:"code_complexity"`
	Maintainability   float64            `json:"maintainability"`
	DuplicationRate   float64            `json:"duplication_rate"`
	SecurityIssues    []SecurityIssue    `json:"security_issues"`
	CodeSmells        []CodeSmell        `json:"code_smells"`
	PerformanceIssues []PerformanceIssue `json:"performance_issues"`
}

// SecurityIssue represents a security vulnerability
type SecurityIssue struct {
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	CVE         string `json:"cve,omitempty"`
}

// CodeSmell represents a code quality issue
type CodeSmell struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Suggestion  string `json:"suggestion"`
}

// PerformanceIssue represents a performance problem
type PerformanceIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Suggestion  string `json:"suggestion"`
}

// Analyze performs comprehensive code quality analysis
func (c *CodeQualityAnalyzer) Analyze(projectPath string, languages []string) (*CodeQualityMetrics, error) {
	metrics := &CodeQualityMetrics{
		TestCoverage:      0.0,
		CodeComplexity:    0.0,
		Maintainability:   0.0,
		DuplicationRate:   0.0,
		SecurityIssues:    make([]SecurityIssue, 0),
		CodeSmells:        make([]CodeSmell, 0),
		PerformanceIssues: make([]PerformanceIssue, 0),
	}

	// Analyze test coverage
	if coverage, err := c.coverage.analyzeTestCoverage(projectPath, languages); err == nil {
		metrics.TestCoverage = coverage
	}

	// Analyze code complexity
	if complexity, err := c.complexity.analyzeCodeComplexity(projectPath, languages); err == nil {
		metrics.CodeComplexity = complexity
	}

	// Analyze security issues
	if issues, err := c.security.analyzeSecurityIssues(projectPath, languages); err == nil {
		metrics.SecurityIssues = issues
	}

	// Analyze code smells
	if smells, err := c.smells.analyzeCodeSmells(projectPath, languages); err == nil {
		metrics.CodeSmells = smells
	}

	// Calculate maintainability index
	metrics.Maintainability = c.scorer.calculateMaintainabilityIndex(metrics)

	return metrics, nil
}
