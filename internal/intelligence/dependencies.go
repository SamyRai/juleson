package intelligence

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/tools/go/packages"
)

// DependencyGraph represents the package-level dependency relationships.
type DependencyGraph struct {
	Nodes []string
	Edges map[string][]string // Maps a package to the list of packages it imports
}

// AnalyzeDependencies parses the module and builds a graph of internal package dependencies.
func AnalyzeDependencies(ctx context.Context, path string) (*DependencyGraph, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedName | packages.NeedImports | packages.NeedDeps | packages.NeedModule,
		Dir:     path,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}

	graph := &DependencyGraph{
		Edges: make(map[string][]string),
	}

	var rootModule string
	for _, pkg := range pkgs {
		if pkg.Module != nil {
			rootModule = pkg.Module.Path
			break
		}
	}

	if rootModule == "" {
		return nil, fmt.Errorf("could not determine root module path")
	}

	for _, pkg := range pkgs {
		pkgName := strings.TrimPrefix(pkg.PkgPath, rootModule+"/")
		if pkgName == pkg.PkgPath {
			pkgName = "root" // Root package
		}

		graph.Nodes = append(graph.Nodes, pkgName)

		for importPath := range pkg.Imports {
			// Only include internal dependencies for the architecture map
			if strings.HasPrefix(importPath, rootModule) {
				cleanImport := strings.TrimPrefix(importPath, rootModule+"/")
				if cleanImport == importPath {
					cleanImport = "root"
				}
				graph.Edges[pkgName] = append(graph.Edges[pkgName], cleanImport)
			}
		}
	}

	return graph, nil
}

// RenderMermaid converts the dependency graph into a Mermaid flowchart format.
func RenderMermaid(graph *DependencyGraph) string {
	var b strings.Builder
	b.WriteString("```mermaid\n")
	b.WriteString("graph TD\n")

	for pkg, imports := range graph.Edges {
		if len(imports) == 0 {
			continue // Skip packages with no internal dependencies to keep graph clean
		}
		for _, imp := range imports {
			b.WriteString(fmt.Sprintf("    \"%s\" --> \"%s\"\n", pkg, imp))
		}
	}

	b.WriteString("```\n")
	return b.String()
}
