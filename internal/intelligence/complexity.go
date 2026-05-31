package intelligence

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// FunctionComplexity represents the cyclomatic complexity of a single Go function.
type FunctionComplexity struct {
	PkgName    string
	FuncName   string
	FileName   string
	Line       int
	Complexity int
}

// AnalyzeComplexity parses the module at the given path and calculates the
// cyclomatic complexity (McCabe's) of every function.
// It returns a slice of FunctionComplexity sorted from highest to lowest.
func AnalyzeComplexity(ctx context.Context, path string) ([]FunctionComplexity, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedCompiledGoFiles,
		Dir:     path,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}

	var results []FunctionComplexity

	for _, pkg := range pkgs {
		for _, fileAST := range pkg.Syntax {
			for _, decl := range fileAST.Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok {
					complexity := calculateComplexity(funcDecl)
					pos := pkg.Fset.Position(funcDecl.Pos())

					// Build the full function name including receiver if present
					funcName := funcDecl.Name.Name
					if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
						switch recvType := funcDecl.Recv.List[0].Type.(type) {
						case *ast.Ident:
							funcName = fmt.Sprintf("(%s).%s", recvType.Name, funcName)
						case *ast.StarExpr:
							if ident, ok := recvType.X.(*ast.Ident); ok {
								funcName = fmt.Sprintf("(*%s).%s", ident.Name, funcName)
							}
						}
					}

					fileName := pos.Filename
					if i := strings.LastIndex(fileName, "/"); i >= 0 {
						fileName = fileName[i+1:]
					}

					results = append(results, FunctionComplexity{
						PkgName:    pkg.Name,
						FuncName:   funcName,
						FileName:   fileName,
						Line:       pos.Line,
						Complexity: complexity,
					})
				}
			}
		}
	}

	// Sort highest complexity first
	sort.Slice(results, func(i, j int) bool {
		return results[i].Complexity > results[j].Complexity
	})

	return results, nil
}

func calculateComplexity(fn *ast.FuncDecl) int {
	complexity := 1 // Base complexity

	ast.Inspect(fn, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SelectStmt:
			complexity++
		case *ast.CaseClause:
			if n.List != nil { // Don't count default case
				complexity++
			}
		case *ast.CommClause:
			if n.Comm != nil { // Don't count default case
				complexity++
			}
		case *ast.BinaryExpr:
			if n.Op == token.LAND || n.Op == token.LOR {
				complexity++
			}
		}
		return true
	})

	return complexity
}
