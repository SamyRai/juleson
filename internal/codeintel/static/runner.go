package static

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"github.com/SamyRai/juleson/internal/codeintel"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
)

// Runner runs static analysis on Go code
type Runner struct {
	fset      *token.FileSet
	analyzers []*analysis.Analyzer
}

// NewRunner creates a new static analysis runner
func NewRunner(analyzerNames []string) *Runner {
	runner := &Runner{
		fset:      token.NewFileSet(),
		analyzers: make([]*analysis.Analyzer, 0),
	}

	// Map of available analyzers
	availableAnalyzers := map[string]*analysis.Analyzer{
		"asmdecl":      asmdecl.Analyzer,
		"assign":       assign.Analyzer,
		"atomic":       atomic.Analyzer,
		"bools":        bools.Analyzer,
		"buildtag":     buildtag.Analyzer,
		"cgocall":      cgocall.Analyzer,
		"composite":    composite.Analyzer,
		"copylock":     copylock.Analyzer,
		"errorsas":     errorsas.Analyzer,
		"httpresponse": httpresponse.Analyzer,
		"loopclosure":  loopclosure.Analyzer,
		"lostcancel":   lostcancel.Analyzer,
		"nilfunc":      nilfunc.Analyzer,
		"printf":       printf.Analyzer,
		"shift":        shift.Analyzer,
		"stdmethods":   stdmethods.Analyzer,
		"structtag":    structtag.Analyzer,
		"tests":        tests.Analyzer,
		"unmarshal":    unmarshal.Analyzer,
		"unreachable":  unreachable.Analyzer,
		"unsafeptr":    unsafeptr.Analyzer,
		"unusedresult": unusedresult.Analyzer,
	}

	// Add requested analyzers
	if len(analyzerNames) == 0 {
		// Use default set
		for _, a := range []string{"assign", "atomic", "bools", "errorsas", "printf", "unreachable"} {
			if analyzer, ok := availableAnalyzers[a]; ok {
				runner.analyzers = append(runner.analyzers, analyzer)
			}
		}
	} else {
		for _, name := range analyzerNames {
			if analyzer, ok := availableAnalyzers[name]; ok {
				runner.analyzers = append(runner.analyzers, analyzer)
			}
		}
	}

	return runner
}

// AnalyzeFile analyzes a single file
func (r *Runner) AnalyzeFile(filePath string, severity codeintel.IssueSeverity) (*AnalysisResult, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	file, err := parser.ParseFile(r.fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	result := &AnalysisResult{
		Issues: make([]codeintel.AnalysisIssue, 0),
		Summary: codeintel.AnalysisSummary{
			BySeverity:   make(map[codeintel.IssueSeverity]int),
			ByCategory:   make(map[codeintel.IssueCategory]int),
			FilesScanned: 1,
		},
	}

	// Run basic checks
	r.checkUnusedVars(file, result)
	r.checkComplexity(file, result)

	// Filter by severity
	filteredIssues := make([]codeintel.AnalysisIssue, 0)
	for _, issue := range result.Issues {
		if r.shouldInclude(issue.Severity, severity) {
			filteredIssues = append(filteredIssues, issue)
			result.Summary.BySeverity[issue.Severity]++
			result.Summary.ByCategory[issue.Category]++
		}
	}

	result.Issues = filteredIssues
	result.Summary.TotalIssues = len(filteredIssues)

	return result, nil
}

// checkUnusedVars checks for unused variables
func (r *Runner) checkUnusedVars(file *ast.File, result *AnalysisResult) {
	declared := make(map[string]token.Pos)
	used := make(map[string]bool)

	// Find declarations
	ast.Inspect(file, func(n ast.Node) bool {
		if spec, ok := n.(*ast.ValueSpec); ok {
			for _, name := range spec.Names {
				if name.Name != "_" {
					declared[name.Name] = name.Pos()
				}
			}
		}
		return true
	})

	// Find usages
	ast.Inspect(file, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			if ident.Obj != nil {
				used[ident.Name] = true
			}
		}
		return true
	})

	// Report unused
	for name, pos := range declared {
		if !used[name] {
			position := r.fset.Position(pos)
			result.Issues = append(result.Issues, codeintel.AnalysisIssue{
				Location: codeintel.Location{
					File:   position.Filename,
					Line:   position.Line,
					Column: position.Column,
				},
				Message:    fmt.Sprintf("variable '%s' declared but not used", name),
				Category:   codeintel.IssueCategoryUnused,
				Severity:   codeintel.IssueSeverityWarning,
				Suggestion: fmt.Sprintf("Remove unused variable '%s' or use it", name),
				Code:       "unused_var",
			})
		}
	}
}

// checkComplexity checks for high complexity functions
func (r *Runner) checkComplexity(file *ast.File, result *AnalysisResult) {
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			complexity := r.calculateCyclomaticComplexity(fn)
			if complexity > 10 {
				position := r.fset.Position(fn.Pos())
				result.Issues = append(result.Issues, codeintel.AnalysisIssue{
					Location: codeintel.Location{
						File:   position.Filename,
						Line:   position.Line,
						Column: position.Column,
					},
					Message:    fmt.Sprintf("function '%s' has high cyclomatic complexity (%d)", fn.Name.Name, complexity),
					Category:   codeintel.IssueCategoryComplexity,
					Severity:   codeintel.IssueSeverityWarning,
					Suggestion: "Consider refactoring to reduce complexity",
					Code:       "high_complexity",
				})
			}
		}
		return true
	})
}

// calculateCyclomaticComplexity calculates cyclomatic complexity
func (r *Runner) calculateCyclomaticComplexity(fn *ast.FuncDecl) int {
	complexity := 1 // Start with 1 for the function itself

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause,
			*ast.CommClause, *ast.BranchStmt:
			complexity++
		case *ast.BinaryExpr:
			// Count logical operators
			if expr, ok := n.(*ast.BinaryExpr); ok {
				if expr.Op == token.LAND || expr.Op == token.LOR {
					complexity++
				}
			}
		}
		return true
	})

	return complexity
}

// shouldInclude determines if an issue should be included based on severity
func (r *Runner) shouldInclude(issueSeverity, minSeverity codeintel.IssueSeverity) bool {
	severityOrder := map[codeintel.IssueSeverity]int{
		codeintel.IssueSeverityInfo:    1,
		codeintel.IssueSeverityWarning: 2,
		codeintel.IssueSeverityError:   3,
	}

	return severityOrder[issueSeverity] >= severityOrder[minSeverity]
}

// AnalysisResult represents the result of static analysis
type AnalysisResult struct {
	Issues  []codeintel.AnalysisIssue `json:"issues"`
	Summary codeintel.AnalysisSummary `json:"summary"`
}
