package context

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SamyRai/juleson/internal/codeintel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "codeintel_context_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	src := `package testpkg

import (
	"fmt"
	m "math"
	_ "os"
)

// MyInterface represents something
type MyInterface interface {
	Do()
}

// MyType is a struct
type MyType struct {
	Field int
}

// MyFunc does something
func MyFunc(a int) error {
	fmt.Println(m.Pi)
	return nil
}

func (m *MyType) Method() {}

var GlobalVar = "test"
const GlobalConst = 1
`
	filePath := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(filePath, []byte(src), 0644)
	require.NoError(t, err)

	analyzer := NewAnalyzer()
	ctx, err := analyzer.AnalyzeFile(filePath, 0)
	require.NoError(t, err)

	// Check file info
	assert.Equal(t, "testpkg", ctx.FileInfo.Package)
	assert.Equal(t, 2, ctx.FileInfo.Functions)
	assert.Equal(t, 2, ctx.FileInfo.Types)
	assert.Len(t, ctx.Imports, 3)

	// Check symbols
	symbolMap := make(map[string]codeintel.SymbolInfo)
	for _, s := range ctx.Symbols {
		symbolMap[s.Name] = s
	}

	assert.Contains(t, symbolMap, "MyInterface")
	assert.Equal(t, codeintel.SymbolKindInterface, symbolMap["MyInterface"].Kind)

	assert.Contains(t, symbolMap, "MyType")
	assert.Equal(t, codeintel.SymbolKindType, symbolMap["MyType"].Kind)

	assert.Contains(t, symbolMap, "MyFunc")
	assert.Equal(t, codeintel.SymbolKindFunc, symbolMap["MyFunc"].Kind)
	assert.True(t, symbolMap["MyFunc"].Exported)

	assert.Contains(t, symbolMap, "Method")
	assert.Equal(t, codeintel.SymbolKindMethod, symbolMap["Method"].Kind)

	assert.Contains(t, symbolMap, "GlobalVar")
	assert.Equal(t, codeintel.SymbolKindVar, symbolMap["GlobalVar"].Kind)

	assert.Contains(t, symbolMap, "GlobalConst")
	assert.Equal(t, codeintel.SymbolKindConst, symbolMap["GlobalConst"].Kind)
}

func TestAnalyzeSymbol(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "codeintel_context_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	src := `package testpkg

func TargetFunc() {}

func caller() {
	TargetFunc()
}
`
	filePath := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(filePath, []byte(src), 0644)
	require.NoError(t, err)

	analyzer := NewAnalyzer()
	ctx, err := analyzer.AnalyzeSymbol(filePath, "TargetFunc")
	require.NoError(t, err)

	assert.Equal(t, "TargetFunc", ctx.Symbol.Name)
	assert.Equal(t, codeintel.SymbolKindFunc, ctx.Symbol.Kind)

	// Two references: definition and call
	assert.Len(t, ctx.References, 2)
}
