package graph

import (
	"testing"

	"github.com/SamyRai/juleson/internal/codeintel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestGraph() *Graph {
	return &Graph{
		Nodes: []codeintel.GraphNode{
			{ID: "funcA", Type: codeintel.NodeTypeFunction, Exported: true},
			{ID: "funcB", Type: codeintel.NodeTypeMethod, Exported: false},
			{ID: "funcC", Type: codeintel.NodeTypeFunction, Exported: true},
			{ID: "typeA", Type: codeintel.NodeTypeType, Exported: true},
		},
		Edges: []*codeintel.GraphEdge{
			{From: "funcA", To: "funcB", Type: codeintel.EdgeTypeCall},
			{From: "funcA", To: "funcC", Type: codeintel.EdgeTypeCall},
			{From: "funcC", To: "funcB", Type: codeintel.EdgeTypeCall},
		},
		EntryPoints: []string{"funcA"},
		Cycles:      []string{},
	}
}

func TestGraphComplexity(t *testing.T) {
	// Empty graph
	empty := &Graph{}
	assert.Equal(t, 0.0, empty.Complexity())

	// Test graph
	// E=3, N=4, P=1 => 3 - 4 + 2(1) = 1
	g := createTestGraph()
	assert.Equal(t, 1.0, g.Complexity())

	// Graph with more edges
	g.Edges = append(g.Edges, &codeintel.GraphEdge{From: "funcB", To: "funcA", Type: codeintel.EdgeTypeCall})
	// E=4, N=4, P=1 => 4 - 4 + 2(1) = 2
	assert.Equal(t, 2.0, g.Complexity())
}

func TestGraphGetCallersOf(t *testing.T) {
	g := createTestGraph()

	callersB := g.GetCallersOf("funcB")
	require.Len(t, callersB, 2)

	ids := map[string]bool{callersB[0].ID: true, callersB[1].ID: true}
	assert.True(t, ids["funcA"])
	assert.True(t, ids["funcC"])

	callersA := g.GetCallersOf("funcA")
	assert.Len(t, callersA, 0)
}

func TestGraphGetCalleesOf(t *testing.T) {
	g := createTestGraph()

	calleesA := g.GetCalleesOf("funcA")
	require.Len(t, calleesA, 2)

	ids := map[string]bool{calleesA[0].ID: true, calleesA[1].ID: true}
	assert.True(t, ids["funcB"])
	assert.True(t, ids["funcC"])

	calleesB := g.GetCalleesOf("funcB")
	assert.Len(t, calleesB, 0)
}

func TestGraphFindNode(t *testing.T) {
	g := createTestGraph()

	node := g.FindNode("funcB")
	require.NotNil(t, node)
	assert.Equal(t, "funcB", node.ID)

	notFound := g.FindNode("missing")
	assert.Nil(t, notFound)
}

func TestGraphStats(t *testing.T) {
	g := createTestGraph()

	stats := g.Stats()
	assert.Equal(t, 4, stats.TotalNodes)
	assert.Equal(t, 3, stats.TotalEdges)
	assert.Equal(t, 1, stats.EntryPoints)
	assert.Equal(t, 0, stats.Cycles)

	assert.Equal(t, 2, stats.FunctionCount)
	assert.Equal(t, 1, stats.MethodCount)
	assert.Equal(t, 1, stats.TypeCount)
	assert.Equal(t, 3, stats.ExportedNodes)

	// funcA has out-degree 2, funcC has out-degree 1, funcB has 0, typeA has 0
	// Max = 2
	assert.Equal(t, 2, stats.MaxOutDegree)

	// Avg = 3 edges / 4 nodes = 0.75
	assert.Equal(t, 0.75, stats.AvgOutDegree)
}
