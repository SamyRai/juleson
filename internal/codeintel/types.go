package codeintel

import (
	"go/ast"
	"go/token"
)

// GraphNode represents a node in the code graph (function, method, or type)
type GraphNode struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Package    string   `json:"package"`
	File       string   `json:"file"`
	Line       int      `json:"line"`
	Column     int      `json:"column"`
	Type       NodeType `json:"type"`
	Exported   bool     `json:"exported"`
	Complexity int      `json:"complexity,omitempty"`
}

// NodeType represents the type of a graph node
type NodeType string

const (
	NodeTypeFunction NodeType = "function"
	NodeTypeMethod   NodeType = "method"
	NodeTypeType     NodeType = "type"
	NodeTypePackage  NodeType = "package"
)

// GraphEdge represents an edge in the code graph (function call or dependency)
type GraphEdge struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Type     EdgeType `json:"type"`
	Dynamic  bool     `json:"dynamic"`
	Location Location `json:"location"`
}

// EdgeType represents the type of a graph edge
type EdgeType string

const (
	EdgeTypeCall       EdgeType = "call"
	EdgeTypeDependency EdgeType = "dependency"
	EdgeTypeImport     EdgeType = "import"
)

// Location represents a source location
type Location struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// SymbolInfo represents information about a code symbol
type SymbolInfo struct {
	Name       string     `json:"name"`
	Kind       SymbolKind `json:"kind"`
	Location   Location   `json:"location"`
	Signature  string     `json:"signature,omitempty"`
	Doc        string     `json:"doc,omitempty"`
	Exported   bool       `json:"exported"`
	Complexity int        `json:"complexity,omitempty"`
}

// SymbolKind represents the kind of a symbol
type SymbolKind string

const (
	SymbolKindPackage   SymbolKind = "package"
	SymbolKindImport    SymbolKind = "import"
	SymbolKindConst     SymbolKind = "const"
	SymbolKindVar       SymbolKind = "var"
	SymbolKindType      SymbolKind = "type"
	SymbolKindFunc      SymbolKind = "func"
	SymbolKindMethod    SymbolKind = "method"
	SymbolKindField     SymbolKind = "field"
	SymbolKindInterface SymbolKind = "interface"
)

// ReferenceInfo represents a reference to a symbol
type ReferenceInfo struct {
	Symbol   string   `json:"symbol"`
	Location Location `json:"location"`
	Kind     RefKind  `json:"kind"`
}

// RefKind represents the kind of reference
type RefKind string

const (
	RefKindDefinition RefKind = "definition"
	RefKindReference  RefKind = "reference"
	RefKindCall       RefKind = "call"
)

// DependencyInfo represents a dependency relationship
type DependencyInfo struct {
	From string     `json:"from"`
	To   string     `json:"to"`
	Type DepType    `json:"type"`
	Kind ImportKind `json:"kind,omitempty"`
}

// DepType represents the type of dependency
type DepType string

const (
	DepTypeImport   DepType = "import"
	DepTypeCall     DepType = "call"
	DepTypeEmbedded DepType = "embedded"
)

// ImportKind represents how an import is used
type ImportKind string

const (
	ImportKindNormal ImportKind = "normal"
	ImportKindAlias  ImportKind = "alias"
	ImportKindDot    ImportKind = "dot"
	ImportKindBlank  ImportKind = "blank"
)

// FileInfo represents information about a source file
type FileInfo struct {
	Path       string       `json:"path"`
	Package    string       `json:"package"`
	Imports    []ImportInfo `json:"imports"`
	Functions  int          `json:"functions"`
	Types      int          `json:"types"`
	Lines      int          `json:"lines"`
	Complexity int          `json:"complexity"`
}

// ImportInfo represents an import statement
type ImportInfo struct {
	Path  string     `json:"path"`
	Name  string     `json:"name,omitempty"`
	Alias string     `json:"alias,omitempty"`
	Kind  ImportKind `json:"kind"`
}

// AnalysisIssue represents a static analysis issue
type AnalysisIssue struct {
	Location   Location      `json:"location"`
	Message    string        `json:"message"`
	Category   IssueCategory `json:"category"`
	Severity   IssueSeverity `json:"severity"`
	Suggestion string        `json:"suggestion,omitempty"`
	Code       string        `json:"code,omitempty"`
}

// IssueCategory represents the category of an issue
type IssueCategory string

const (
	IssueCategoryBug         IssueCategory = "bug"
	IssueCategoryCodeSmell   IssueCategory = "code_smell"
	IssueCategorySecurity    IssueCategory = "security"
	IssueCategoryPerformance IssueCategory = "performance"
	IssueCategoryComplexity  IssueCategory = "complexity"
	IssueCategoryUnused      IssueCategory = "unused"
	IssueCategoryDeprecated  IssueCategory = "deprecated"
	IssueCategoryStyle       IssueCategory = "style"
)

// IssueSeverity represents the severity of an issue
type IssueSeverity string

const (
	IssueSeverityInfo    IssueSeverity = "info"
	IssueSeverityWarning IssueSeverity = "warning"
	IssueSeverityError   IssueSeverity = "error"
)

// AnalysisSummary represents a summary of analysis results
type AnalysisSummary struct {
	TotalIssues  int                   `json:"total_issues"`
	BySeverity   map[IssueSeverity]int `json:"by_severity"`
	ByCategory   map[IssueCategory]int `json:"by_category"`
	FilesScanned int                   `json:"files_scanned"`
}

// ComplexityMetrics represents code complexity metrics
type ComplexityMetrics struct {
	Cyclomatic int     `json:"cyclomatic"`
	Cognitive  int     `json:"cognitive"`
	Lines      int     `json:"lines"`
	Functions  int     `json:"functions"`
	AvgPerFunc float64 `json:"avg_per_func"`
}

// FunctionNode represents an AST node for a function
type FunctionNode struct {
	Func *ast.FuncDecl
	Pos  token.Pos
	End  token.Pos
}
