package context

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/SamyRai/juleson/internal/codeintel"
	"golang.org/x/tools/go/packages"
)

// Analyzer analyzes code context from Go source files
type Analyzer struct {
	fset *token.FileSet
}

// NewAnalyzer creates a new context analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		fset: token.NewFileSet(),
	}
}

// AnalyzeFile analyzes a single Go source file
func (a *Analyzer) AnalyzeFile(filePath string, contextLines int) (*FileContext, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	file, err := parser.ParseFile(a.fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	ctx := &FileContext{
		FileInfo: codeintel.FileInfo{
			Path:    filePath,
			Package: file.Name.Name,
		},
		Symbols:      make([]codeintel.SymbolInfo, 0),
		Imports:      make([]codeintel.ImportInfo, 0),
		Dependencies: make([]codeintel.DependencyInfo, 0),
	}

	// Extract imports
	for _, imp := range file.Imports {
		importInfo := a.extractImportInfo(imp)
		ctx.Imports = append(ctx.Imports, importInfo)
		ctx.FileInfo.Imports = append(ctx.FileInfo.Imports, importInfo)
	}

	// Extract symbols
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			ctx.Symbols = append(ctx.Symbols, a.extractFuncSymbol(node))
			ctx.FileInfo.Functions++
		case *ast.GenDecl:
			for _, spec := range node.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					ctx.Symbols = append(ctx.Symbols, a.extractTypeSymbol(typeSpec, node.Doc))
					ctx.FileInfo.Types++
				} else if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valueSpec.Names {
						ctx.Symbols = append(ctx.Symbols, a.extractValueSymbol(name, node.Tok, node.Doc))
					}
				}
			}
		}
		return true
	})

	// Calculate file metrics
	lines := strings.Count(string(src), "\n") + 1
	ctx.FileInfo.Lines = lines

	return ctx, nil
}

// AnalyzeSymbol analyzes a specific symbol in a file
func (a *Analyzer) AnalyzeSymbol(filePath, symbolName string) (*SymbolContext, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	file, err := parser.ParseFile(a.fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	ctx := &SymbolContext{
		References:   make([]codeintel.ReferenceInfo, 0),
		Dependencies: make([]codeintel.DependencyInfo, 0),
	}

	// Find the symbol
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name.Name == symbolName {
				ctx.Symbol = a.extractFuncSymbol(node)
				ctx.Definition = a.getNodeLocation(node.Pos())
			}
		case *ast.TypeSpec:
			if node.Name.Name == symbolName {
				ctx.Symbol = a.extractTypeSymbol(node, nil)
				ctx.Definition = a.getNodeLocation(node.Pos())
			}
		case *ast.Ident:
			if node.Name == symbolName {
				ref := codeintel.ReferenceInfo{
					Symbol:   symbolName,
					Location: a.getNodeLocation(node.Pos()),
					Kind:     codeintel.RefKindReference,
				}
				ctx.References = append(ctx.References, ref)
			}
		}
		return true
	})

	return ctx, nil
}

// FindReferences finds all references to a symbol across a project
func (a *Analyzer) FindReferences(projectPath, symbolName string) ([]codeintel.ReferenceInfo, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
		Dir:  projectPath,
		Fset: a.fset,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	references := make([]codeintel.ReferenceInfo, 0)

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if ident, ok := n.(*ast.Ident); ok {
					if ident.Name == symbolName {
						ref := codeintel.ReferenceInfo{
							Symbol:   symbolName,
							Location: a.getNodeLocation(ident.Pos()),
							Kind:     codeintel.RefKindReference,
						}
						references = append(references, ref)
					}
				}
				return true
			})
		}
	}

	return references, nil
}

// extractImportInfo extracts import information
func (a *Analyzer) extractImportInfo(imp *ast.ImportSpec) codeintel.ImportInfo {
	path := strings.Trim(imp.Path.Value, "\"")
	info := codeintel.ImportInfo{
		Path: path,
		Kind: codeintel.ImportKindNormal,
	}

	if imp.Name != nil {
		if imp.Name.Name == "." {
			info.Kind = codeintel.ImportKindDot
		} else if imp.Name.Name == "_" {
			info.Kind = codeintel.ImportKindBlank
		} else {
			info.Alias = imp.Name.Name
			info.Kind = codeintel.ImportKindAlias
		}
		info.Name = imp.Name.Name
	}

	return info
}

// extractFuncSymbol extracts function symbol information
func (a *Analyzer) extractFuncSymbol(fn *ast.FuncDecl) codeintel.SymbolInfo {
	pos := a.fset.Position(fn.Pos())

	kind := codeintel.SymbolKindFunc
	if fn.Recv != nil {
		kind = codeintel.SymbolKindMethod
	}

	signature := a.getFuncSignature(fn)
	doc := ""
	if fn.Doc != nil {
		doc = fn.Doc.Text()
	}

	return codeintel.SymbolInfo{
		Name:      fn.Name.Name,
		Kind:      kind,
		Location:  codeintel.Location{File: pos.Filename, Line: pos.Line, Column: pos.Column},
		Signature: signature,
		Doc:       doc,
		Exported:  ast.IsExported(fn.Name.Name),
	}
}

// extractTypeSymbol extracts type symbol information
func (a *Analyzer) extractTypeSymbol(typeSpec *ast.TypeSpec, doc *ast.CommentGroup) codeintel.SymbolInfo {
	pos := a.fset.Position(typeSpec.Pos())

	kind := codeintel.SymbolKindType
	if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
		kind = codeintel.SymbolKindInterface
	}

	docText := ""
	if doc != nil {
		docText = doc.Text()
	}

	return codeintel.SymbolInfo{
		Name:     typeSpec.Name.Name,
		Kind:     kind,
		Location: codeintel.Location{File: pos.Filename, Line: pos.Line, Column: pos.Column},
		Doc:      docText,
		Exported: ast.IsExported(typeSpec.Name.Name),
	}
}

// extractValueSymbol extracts variable/const symbol information
func (a *Analyzer) extractValueSymbol(name *ast.Ident, tok token.Token, doc *ast.CommentGroup) codeintel.SymbolInfo {
	pos := a.fset.Position(name.Pos())

	kind := codeintel.SymbolKindVar
	if tok == token.CONST {
		kind = codeintel.SymbolKindConst
	}

	docText := ""
	if doc != nil {
		docText = doc.Text()
	}

	return codeintel.SymbolInfo{
		Name:     name.Name,
		Kind:     kind,
		Location: codeintel.Location{File: pos.Filename, Line: pos.Line, Column: pos.Column},
		Doc:      docText,
		Exported: ast.IsExported(name.Name),
	}
}

// getFuncSignature generates a function signature string
func (a *Analyzer) getFuncSignature(fn *ast.FuncDecl) string {
	var sig strings.Builder

	sig.WriteString("func ")

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		sig.WriteString("(")
		// Simplified receiver
		sig.WriteString("receiver")
		sig.WriteString(") ")
	}

	sig.WriteString(fn.Name.Name)
	sig.WriteString("(")

	// Parameters
	if fn.Type.Params != nil {
		for i, field := range fn.Type.Params.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			if len(field.Names) > 0 {
				sig.WriteString(field.Names[0].Name)
			} else {
				sig.WriteString("_")
			}
			sig.WriteString(" ")
			sig.WriteString(a.exprToString(field.Type))
		}
	}

	sig.WriteString(")")

	// Results
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		sig.WriteString(" ")
		if len(fn.Type.Results.List) > 1 {
			sig.WriteString("(")
		}
		for i, field := range fn.Type.Results.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			sig.WriteString(a.exprToString(field.Type))
		}
		if len(fn.Type.Results.List) > 1 {
			sig.WriteString(")")
		}
	}

	return sig.String()
}

// exprToString converts an expression to a string
func (a *Analyzer) exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + a.exprToString(e.X)
	case *ast.ArrayType:
		return "[]" + a.exprToString(e.Elt)
	case *ast.SelectorExpr:
		return a.exprToString(e.X) + "." + e.Sel.Name
	default:
		return "type"
	}
}

// getNodeLocation gets the location of an AST node
func (a *Analyzer) getNodeLocation(pos token.Pos) codeintel.Location {
	position := a.fset.Position(pos)
	return codeintel.Location{
		File:   position.Filename,
		Line:   position.Line,
		Column: position.Column,
	}
}

// FileContext represents the context of a file
type FileContext struct {
	FileInfo     codeintel.FileInfo         `json:"file_info"`
	Symbols      []codeintel.SymbolInfo     `json:"symbols"`
	Imports      []codeintel.ImportInfo     `json:"imports"`
	Dependencies []codeintel.DependencyInfo `json:"dependencies"`
	References   []codeintel.ReferenceInfo  `json:"references,omitempty"`
}

// SymbolContext represents the context of a specific symbol
type SymbolContext struct {
	Symbol       codeintel.SymbolInfo       `json:"symbol"`
	Definition   codeintel.Location         `json:"definition"`
	References   []codeintel.ReferenceInfo  `json:"references"`
	Dependencies []codeintel.DependencyInfo `json:"dependencies"`
}
