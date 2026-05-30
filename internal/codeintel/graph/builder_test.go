package graph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SamyRai/juleson/internal/codeintel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildFromFile(t *testing.T) {
	// Create a temporary directory and a go source file
	tmpDir, err := os.MkdirTemp("", "codeintel_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	src := `package testpkg

func ExportedFunc() {
	internalFunc()
}

func internalFunc() {
	// do something
}

type MyType struct {}

func (m *MyType) Method() {
	ExportedFunc()
}
`
	filePath := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(filePath, []byte(src), 0644)
	require.NoError(t, err)

	builder := NewBuilder()
	g, err := builder.BuildFromFile(filePath)
	require.NoError(t, err)

	// Check graph structure
	// Nodes: ExportedFunc, internalFunc, MyType, MyType.Method
	assert.Len(t, g.Nodes, 4)

	nodeByName := make(map[string]*codeintel.GraphNode)
	for i := range g.Nodes {
		nodeByName[g.Nodes[i].Name] = &g.Nodes[i]
	}

	assert.NotNil(t, nodeByName["ExportedFunc"])
	if nodeByName["ExportedFunc"] != nil {
		assert.True(t, nodeByName["ExportedFunc"].Exported)
	}

	assert.NotNil(t, nodeByName["internalFunc"])
	if nodeByName["internalFunc"] != nil {
		assert.False(t, nodeByName["internalFunc"].Exported)
	}

	assert.NotNil(t, nodeByName["MyType"])
	if nodeByName["MyType"] != nil {
		assert.True(t, nodeByName["MyType"].Exported)
	}

	assert.NotNil(t, nodeByName["Method"])
	if nodeByName["Method"] != nil {
		assert.True(t, nodeByName["Method"].Exported)
	}

	// Edges: ExportedFunc -> internalFunc, MyType.Method -> ExportedFunc
	assert.Len(t, g.Edges, 2)
}

func TestDetectCycles(t *testing.T) {
	b := NewBuilder()

	// Manually add nodes and edges to test the cycle detection logic
	b.nodes["a"] = &codeintel.GraphNode{ID: "a"}
	b.nodes["b"] = &codeintel.GraphNode{ID: "b"}
	b.nodes["c"] = &codeintel.GraphNode{ID: "c"}

	b.edges = append(b.edges, &codeintel.GraphEdge{From: "a", To: "b"})
	b.edges = append(b.edges, &codeintel.GraphEdge{From: "b", To: "c"})
	b.edges = append(b.edges, &codeintel.GraphEdge{From: "c", To: "a"})

	cycles := b.detectCycles()
	assert.NotEmpty(t, cycles)
}
