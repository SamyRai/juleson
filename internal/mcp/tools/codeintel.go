package tools

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/codeintel"
	codeContext "github.com/SamyRai/juleson/internal/codeintel/context"
	"github.com/SamyRai/juleson/internal/codeintel/graph"
	"github.com/SamyRai/juleson/internal/codeintel/static"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterCodeIntelTools registers code intelligence MCP tools
func RegisterCodeIntelTools(server *mcp.Server) {
	// Code graph analysis tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "analyze_code_graph",
		Description: "Analyze code call graph and dependencies with cycle detection",
	}, analyzeCodeGraphHandler)

	// Code context analysis tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "analyze_code_context",
		Description: "Analyze code context including symbols, imports, and dependencies",
	}, analyzeCodeContextHandler)

	// Find symbol references tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_symbol_references",
		Description: "Find all references to a symbol across the project",
	}, findSymbolReferencesHandler)

	// Static analysis tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "run_static_analysis",
		Description: "Run static analysis checks on Go code",
	}, runStaticAnalysisHandler)

	// Complexity analysis tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "analyze_complexity",
		Description: "Analyze code complexity metrics",
	}, analyzeComplexityHandler)
}

// Input/Output types

type AnalyzeCodeGraphInput struct {
	ProjectPath  string `json:"project_path" jsonschema:"Path to the project directory"`
	IncludeTests bool   `json:"include_tests,omitempty" jsonschema:"Include test files in analysis (default: false)"`
	MaxDepth     int    `json:"max_depth,omitempty" jsonschema:"Maximum depth for graph traversal (default: unlimited)"`
	OutputFormat string `json:"output_format,omitempty" jsonschema:"Output format: json, dot, or mermaid (default: json)"`
}

type AnalyzeCodeGraphOutput struct {
	Nodes       []codeintel.GraphNode `json:"nodes"`
	Edges       []codeintel.GraphEdge `json:"edges"`
	EntryPoints []string              `json:"entry_points"`
	Cycles      []string              `json:"cycles"`
	Stats       *graph.GraphStats     `json:"stats"`
	Diagram     string                `json:"diagram,omitempty"`
	Summary     string                `json:"summary"`
}

type AnalyzeCodeContextInput struct {
	FilePath     string `json:"file_path" jsonschema:"Path to the file to analyze"`
	SymbolName   string `json:"symbol_name,omitempty" jsonschema:"Specific symbol to analyze (optional)"`
	ContextLines int    `json:"context_lines,omitempty" jsonschema:"Number of context lines (default: 5)"`
	IncludeRefs  bool   `json:"include_refs,omitempty" jsonschema:"Include references (default: false)"`
}

type AnalyzeCodeContextOutput struct {
	FileInfo     codeintel.FileInfo         `json:"file_info"`
	Symbols      []codeintel.SymbolInfo     `json:"symbols"`
	Imports      []codeintel.ImportInfo     `json:"imports"`
	Dependencies []codeintel.DependencyInfo `json:"dependencies,omitempty"`
	References   []codeintel.ReferenceInfo  `json:"references,omitempty"`
	Summary      string                     `json:"summary"`
}

type FindSymbolReferencesInput struct {
	ProjectPath string `json:"project_path" jsonschema:"Path to the project directory"`
	SymbolName  string `json:"symbol_name" jsonschema:"Symbol name to search for"`
}

type FindSymbolReferencesOutput struct {
	SymbolName string                    `json:"symbol_name"`
	References []codeintel.ReferenceInfo `json:"references"`
	Count      int                       `json:"count"`
	Summary    string                    `json:"summary"`
}

type RunStaticAnalysisInput struct {
	ProjectPath string   `json:"project_path" jsonschema:"Path to the project directory"`
	FilePath    string   `json:"file_path,omitempty" jsonschema:"Specific file to analyze (optional)"`
	Analyzers   []string `json:"analyzers,omitempty" jsonschema:"Analyzers to run: unused, complexity, all (default: all)"`
	Severity    string   `json:"severity,omitempty" jsonschema:"Minimum severity: info, warning, error (default: info)"`
}

type RunStaticAnalysisOutput struct {
	Issues  []codeintel.AnalysisIssue `json:"issues"`
	Summary codeintel.AnalysisSummary `json:"summary"`
}

type AnalyzeComplexityInput struct {
	ProjectPath string `json:"project_path" jsonschema:"Path to the project directory"`
	FilePath    string `json:"file_path,omitempty" jsonschema:"Specific file to analyze (optional)"`
}

type AnalyzeComplexityOutput struct {
	Metrics codeintel.ComplexityMetrics `json:"metrics"`
	Issues  []codeintel.AnalysisIssue   `json:"issues"`
	Summary string                      `json:"summary"`
}

// Handler functions

func analyzeCodeGraphHandler(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeCodeGraphInput) (
	*mcp.CallToolResult,
	AnalyzeCodeGraphOutput,
	error,
) {
	builder := graph.NewBuilder()

	codeGraph, err := builder.BuildFromPath(input.ProjectPath, input.IncludeTests)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to build code graph: %v", err)},
			},
		}, AnalyzeCodeGraphOutput{}, err
	}

	stats := codeGraph.Stats()

	// Convert edges from pointers to values
	edges := make([]codeintel.GraphEdge, len(codeGraph.Edges))
	for i, edge := range codeGraph.Edges {
		edges[i] = *edge
	}

	output := AnalyzeCodeGraphOutput{
		Nodes:       codeGraph.Nodes,
		Edges:       edges,
		EntryPoints: codeGraph.EntryPoints,
		Cycles:      codeGraph.Cycles,
		Stats:       stats,
		Summary: fmt.Sprintf("Analyzed %d nodes, %d edges. Found %d entry points and %d cycles.",
			stats.TotalNodes, stats.TotalEdges, stats.EntryPoints, stats.Cycles),
	}

	// Generate diagram if requested
	if input.OutputFormat == "dot" || input.OutputFormat == "mermaid" {
		exporter := graph.NewExporter()
		var diagram string
		var err error

		if input.OutputFormat == "dot" {
			diagram, err = exporter.ExportToDOT(codeGraph)
		} else {
			diagram, err = exporter.ExportToMermaid(codeGraph)
		}

		if err == nil {
			output.Diagram = diagram
		}
	}

	return nil, output, nil
}

func analyzeCodeContextHandler(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeCodeContextInput) (
	*mcp.CallToolResult,
	AnalyzeCodeContextOutput,
	error,
) {
	analyzer := codeContext.NewAnalyzer()

	if input.ContextLines == 0 {
		input.ContextLines = 5
	}

	fileCtx, err := analyzer.AnalyzeFile(input.FilePath, input.ContextLines)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to analyze file: %v", err)},
			},
		}, AnalyzeCodeContextOutput{}, err
	}

	output := AnalyzeCodeContextOutput{
		FileInfo: fileCtx.FileInfo,
		Symbols:  fileCtx.Symbols,
		Imports:  fileCtx.Imports,
		Summary: fmt.Sprintf("File has %d symbols, %d imports, %d lines",
			len(fileCtx.Symbols), len(fileCtx.Imports), fileCtx.FileInfo.Lines),
	}

	// Analyze specific symbol if requested
	if input.SymbolName != "" {
		symbolCtx, err := analyzer.AnalyzeSymbol(input.FilePath, input.SymbolName)
		if err == nil {
			output.References = symbolCtx.References
			output.Dependencies = symbolCtx.Dependencies
		}
	}

	return nil, output, nil
}

func findSymbolReferencesHandler(ctx context.Context, req *mcp.CallToolRequest, input FindSymbolReferencesInput) (
	*mcp.CallToolResult,
	FindSymbolReferencesOutput,
	error,
) {
	analyzer := codeContext.NewAnalyzer()

	refs, err := analyzer.FindReferences(input.ProjectPath, input.SymbolName)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to find references: %v", err)},
			},
		}, FindSymbolReferencesOutput{}, err
	}

	output := FindSymbolReferencesOutput{
		SymbolName: input.SymbolName,
		References: refs,
		Count:      len(refs),
		Summary:    fmt.Sprintf("Found %d references to '%s'", len(refs), input.SymbolName),
	}

	return nil, output, nil
}

func runStaticAnalysisHandler(ctx context.Context, req *mcp.CallToolRequest, input RunStaticAnalysisInput) (
	*mcp.CallToolResult,
	RunStaticAnalysisOutput,
	error,
) {
	severity := codeintel.IssueSeverityInfo
	if input.Severity == "warning" {
		severity = codeintel.IssueSeverityWarning
	} else if input.Severity == "error" {
		severity = codeintel.IssueSeverityError
	}

	runner := static.NewRunner(input.Analyzers)

	// Analyze specific file or all files
	var result *static.AnalysisResult
	var err error

	if input.FilePath != "" {
		result, err = runner.AnalyzeFile(input.FilePath, severity)
	} else {
		result, err = runner.AnalyzeFile(input.ProjectPath, severity)
	}

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to run static analysis: %v", err)},
			},
		}, RunStaticAnalysisOutput{}, err
	}

	output := RunStaticAnalysisOutput{
		Issues:  result.Issues,
		Summary: result.Summary,
	}

	return nil, output, nil
}

func analyzeComplexityHandler(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeComplexityInput) (
	*mcp.CallToolResult,
	AnalyzeComplexityOutput,
	error,
) {
	runner := static.NewRunner([]string{"complexity"})

	filePath := input.FilePath
	if filePath == "" {
		filePath = input.ProjectPath
	}

	result, err := runner.AnalyzeFile(filePath, codeintel.IssueSeverityInfo)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to analyze complexity: %v", err)},
			},
		}, AnalyzeComplexityOutput{}, err
	}

	// Calculate metrics from issues
	metrics := codeintel.ComplexityMetrics{
		Functions: len(result.Issues),
	}

	for _, issue := range result.Issues {
		if issue.Category == codeintel.IssueCategoryComplexity {
			metrics.Cyclomatic++
		}
	}

	output := AnalyzeComplexityOutput{
		Metrics: metrics,
		Issues:  result.Issues,
		Summary: fmt.Sprintf("Found %d complexity issues", len(result.Issues)),
	}

	return nil, output, nil
}
