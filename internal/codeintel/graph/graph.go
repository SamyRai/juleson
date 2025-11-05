package graph

import (
	"github.com/SamyRai/juleson/internal/codeintel"
)

// Graph represents a code call graph
type Graph struct {
	Nodes       []codeintel.GraphNode  `json:"nodes"`
	Edges       []*codeintel.GraphEdge `json:"edges"`
	EntryPoints []string               `json:"entry_points"`
	Cycles      []string               `json:"cycles"`
}

// Complexity calculates the overall graph complexity
func (g *Graph) Complexity() float64 {
	if len(g.Nodes) == 0 {
		return 0
	}

	// McCabe cyclomatic complexity for graphs
	// V(G) = E - N + 2P where E=edges, N=nodes, P=connected components
	e := len(g.Edges)
	n := len(g.Nodes)
	p := len(g.EntryPoints)

	if p == 0 {
		p = 1 // At least one component
	}

	complexity := float64(e - n + 2*p)
	if complexity < 1 {
		complexity = 1
	}

	return complexity
}

// GetCallersOf finds all functions that call the given function
func (g *Graph) GetCallersOf(nodeID string) []codeintel.GraphNode {
	callers := make([]codeintel.GraphNode, 0)
	callerIDs := make(map[string]bool)

	for _, edge := range g.Edges {
		if edge.To == nodeID && edge.Type == codeintel.EdgeTypeCall {
			callerIDs[edge.From] = true
		}
	}

	for _, node := range g.Nodes {
		if callerIDs[node.ID] {
			callers = append(callers, node)
		}
	}

	return callers
}

// GetCalleesOf finds all functions called by the given function
func (g *Graph) GetCalleesOf(nodeID string) []codeintel.GraphNode {
	callees := make([]codeintel.GraphNode, 0)
	calleeIDs := make(map[string]bool)

	for _, edge := range g.Edges {
		if edge.From == nodeID && edge.Type == codeintel.EdgeTypeCall {
			calleeIDs[edge.To] = true
		}
	}

	for _, node := range g.Nodes {
		if calleeIDs[node.ID] {
			callees = append(callees, node)
		}
	}

	return callees
}

// FindNode finds a node by ID
func (g *Graph) FindNode(nodeID string) *codeintel.GraphNode {
	for i := range g.Nodes {
		if g.Nodes[i].ID == nodeID {
			return &g.Nodes[i]
		}
	}
	return nil
}

// Stats returns statistics about the graph
func (g *Graph) Stats() *GraphStats {
	stats := &GraphStats{
		TotalNodes:    len(g.Nodes),
		TotalEdges:    len(g.Edges),
		EntryPoints:   len(g.EntryPoints),
		Cycles:        len(g.Cycles),
		FunctionCount: 0,
		MethodCount:   0,
		TypeCount:     0,
		ExportedNodes: 0,
		AvgOutDegree:  0,
		MaxOutDegree:  0,
	}

	outDegree := make(map[string]int)

	for _, node := range g.Nodes {
		switch node.Type {
		case codeintel.NodeTypeFunction:
			stats.FunctionCount++
		case codeintel.NodeTypeMethod:
			stats.MethodCount++
		case codeintel.NodeTypeType:
			stats.TypeCount++
		}

		if node.Exported {
			stats.ExportedNodes++
		}
	}

	for _, edge := range g.Edges {
		outDegree[edge.From]++
	}

	totalOutDegree := 0
	for _, degree := range outDegree {
		totalOutDegree += degree
		if degree > stats.MaxOutDegree {
			stats.MaxOutDegree = degree
		}
	}

	if len(g.Nodes) > 0 {
		stats.AvgOutDegree = float64(totalOutDegree) / float64(len(g.Nodes))
	}

	return stats
}

// GraphStats represents statistics about a code graph
type GraphStats struct {
	TotalNodes    int     `json:"total_nodes"`
	TotalEdges    int     `json:"total_edges"`
	EntryPoints   int     `json:"entry_points"`
	Cycles        int     `json:"cycles"`
	FunctionCount int     `json:"function_count"`
	MethodCount   int     `json:"method_count"`
	TypeCount     int     `json:"type_count"`
	ExportedNodes int     `json:"exported_nodes"`
	AvgOutDegree  float64 `json:"avg_out_degree"`
	MaxOutDegree  int     `json:"max_out_degree"`
}
