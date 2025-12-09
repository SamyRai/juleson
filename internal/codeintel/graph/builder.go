package graph

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/SamyRai/juleson/internal/codeintel"
	"golang.org/x/tools/go/packages"
)

// Builder constructs code graphs from Go source code
type Builder struct {
	fset  *token.FileSet
	pkgs  []*packages.Package
	nodes map[string]*codeintel.GraphNode
	edges []*codeintel.GraphEdge
}

// NewBuilder creates a new graph builder
func NewBuilder() *Builder {
	return &Builder{
		fset:  token.NewFileSet(),
		nodes: make(map[string]*codeintel.GraphNode),
		edges: make([]*codeintel.GraphEdge, 0),
	}
}

// BuildFromPath builds a code graph from a project path
func (b *Builder) BuildFromPath(projectPath string, includeTests bool) (*Graph, error) {
	// Load packages
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir:   projectPath,
		Fset:  b.fset,
		Tests: includeTests,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("errors loading packages")
	}

	b.pkgs = pkgs

	// Build graph
	for _, pkg := range pkgs {
		if err := b.processPackage(pkg); err != nil {
			return nil, fmt.Errorf("failed to process package %s: %w", pkg.PkgPath, err)
		}
	}

	return &Graph{
		Nodes:       b.getNodeSlice(),
		Edges:       b.edges,
		EntryPoints: b.findEntryPoints(),
		Cycles:      b.detectCycles(),
	}, nil
}

// BuildFromFile builds a code graph from a single file
func (b *Builder) BuildFromFile(filePath string) (*Graph, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	file, err := parser.ParseFile(b.fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Process file
	b.processFile(file, filepath.Base(filepath.Dir(filePath)))

	return &Graph{
		Nodes:       b.getNodeSlice(),
		Edges:       b.edges,
		EntryPoints: b.findEntryPoints(),
		Cycles:      b.detectCycles(),
	}, nil
}

// processPackage processes a single package
func (b *Builder) processPackage(pkg *packages.Package) error {
	for _, file := range pkg.Syntax {
		b.processFile(file, pkg.PkgPath)
	}
	return nil
}

// processFile processes a single file
func (b *Builder) processFile(file *ast.File, pkgPath string) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			b.processFuncDecl(node, pkgPath, file)
		case *ast.GenDecl:
			b.processGenDecl(node, pkgPath, file)
		}
		return true
	})
}

// processFuncDecl processes a function declaration
func (b *Builder) processFuncDecl(fn *ast.FuncDecl, pkgPath string, file *ast.File) {
	pos := b.fset.Position(fn.Pos())

	nodeID := b.getFuncID(fn, pkgPath)
	nodeType := codeintel.NodeTypeFunction
	if fn.Recv != nil {
		nodeType = codeintel.NodeTypeMethod
	}

	node := &codeintel.GraphNode{
		ID:       nodeID,
		Name:     fn.Name.Name,
		Package:  pkgPath,
		File:     pos.Filename,
		Line:     pos.Line,
		Column:   pos.Column,
		Type:     nodeType,
		Exported: ast.IsExported(fn.Name.Name),
	}

	b.nodes[nodeID] = node

	// Find calls within this function
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			b.processCallExpr(call, nodeID)
		}
		return true
	})
}

// processGenDecl processes a general declaration (types, vars, consts)
func (b *Builder) processGenDecl(gen *ast.GenDecl, pkgPath string, file *ast.File) {
	for _, spec := range gen.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			pos := b.fset.Position(typeSpec.Pos())

			nodeID := fmt.Sprintf("%s.%s", pkgPath, typeSpec.Name.Name)
			node := &codeintel.GraphNode{
				ID:       nodeID,
				Name:     typeSpec.Name.Name,
				Package:  pkgPath,
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Type:     codeintel.NodeTypeType,
				Exported: ast.IsExported(typeSpec.Name.Name),
			}

			b.nodes[nodeID] = node
		}
	}
}

// processCallExpr processes a function call expression
func (b *Builder) processCallExpr(call *ast.CallExpr, fromID string) {
	var toID string
	isDynamic := false

	switch fun := call.Fun.(type) {
	case *ast.Ident:
		// Simple function call: foo()
		toID = fun.Name
	case *ast.SelectorExpr:
		// Method call or package function: pkg.Foo() or obj.Method()
		if ident, ok := fun.X.(*ast.Ident); ok {
			toID = fmt.Sprintf("%s.%s", ident.Name, fun.Sel.Name)
		} else {
			// Dynamic call through interface or complex expression
			toID = fun.Sel.Name
			isDynamic = true
		}
	default:
		// Function literal or other complex expression
		isDynamic = true
		toID = "dynamic"
	}

	if toID != "" {
		pos := b.fset.Position(call.Pos())
		edge := &codeintel.GraphEdge{
			From:    fromID,
			To:      toID,
			Type:    codeintel.EdgeTypeCall,
			Dynamic: isDynamic,
			Location: codeintel.Location{
				File:   pos.Filename,
				Line:   pos.Line,
				Column: pos.Column,
			},
		}
		b.edges = append(b.edges, edge)
	}
}

// getFuncID generates a unique ID for a function
func (b *Builder) getFuncID(fn *ast.FuncDecl, pkgPath string) string {
	if fn.Recv == nil {
		// Regular function
		return fmt.Sprintf("%s.%s", pkgPath, fn.Name.Name)
	}

	// Method - extract receiver type
	var recvType string
	switch t := fn.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			recvType = ident.Name
		}
	case *ast.Ident:
		recvType = t.Name
	}

	return fmt.Sprintf("%s.%s.%s", pkgPath, recvType, fn.Name.Name)
}

// getNodeSlice converts the nodes map to a slice
func (b *Builder) getNodeSlice() []codeintel.GraphNode {
	nodes := make([]codeintel.GraphNode, 0, len(b.nodes))
	for _, node := range b.nodes {
		nodes = append(nodes, *node)
	}
	return nodes
}

// findEntryPoints finds entry points in the graph (main functions, exported functions)
func (b *Builder) findEntryPoints() []string {
	entryPoints := make([]string, 0)
	for id, node := range b.nodes {
		// Main functions
		if node.Name == "main" && strings.HasSuffix(node.Package, "main") {
			entryPoints = append(entryPoints, id)
			continue
		}

		// Exported functions with no callers
		if node.Exported {
			hasCallers := false
			for _, edge := range b.edges {
				if edge.To == id {
					hasCallers = true
					break
				}
			}
			if !hasCallers {
				entryPoints = append(entryPoints, id)
			}
		}
	}
	return entryPoints
}

// detectCycles detects cycles in the call graph
func (b *Builder) detectCycles() []string {
	cycles := make([]string, 0)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var detectCycle func(nodeID string, path []string) bool
	detectCycle = func(nodeID string, path []string) bool {
		visited[nodeID] = true
		recStack[nodeID] = true
		path = append(path, nodeID)

		// Find all outgoing edges
		for _, edge := range b.edges {
			if edge.From == nodeID {
				if !visited[edge.To] {
					if detectCycle(edge.To, path) {
						return true
					}
				} else if recStack[edge.To] {
					// Found cycle
					cycleStart := -1
					for i, id := range path {
						if id == edge.To {
							cycleStart = i
							break
						}
					}
					if cycleStart >= 0 {
						cyclePath := strings.Join(path[cycleStart:], " -> ")
						cycles = append(cycles, cyclePath+" -> "+edge.To)
					}
					return true
				}
			}
		}

		recStack[nodeID] = false
		return false
	}

	for id := range b.nodes {
		if !visited[id] {
			detectCycle(id, []string{})
		}
	}

	return cycles
}
