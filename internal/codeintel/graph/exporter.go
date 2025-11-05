package graph

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/codeintel"
)

// Exporter exports code graphs to various formats
type Exporter struct{}

// NewExporter creates a new exporter
func NewExporter() *Exporter {
	return &Exporter{}
}

// ExportToDOT exports a graph to DOT format (Graphviz)
func (e *Exporter) ExportToDOT(graph *Graph) (string, error) {
	var buf bytes.Buffer

	buf.WriteString("digraph CodeGraph {\n")
	buf.WriteString("  rankdir=LR;\n")
	buf.WriteString("  node [shape=box, style=rounded];\n\n")

	// Write nodes
	for _, node := range graph.Nodes {
		label := e.formatNodeLabel(node)
		shape := e.getNodeShape(node.Type)
		color := e.getNodeColor(node)

		buf.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\", shape=%s, fillcolor=\"%s\", style=filled];\n",
			node.ID, label, shape, color))
	}

	buf.WriteString("\n")

	// Write edges
	for _, edge := range graph.Edges {
		style := "solid"
		if edge.Dynamic {
			style = "dashed"
		}

		buf.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [style=%s];\n",
			edge.From, edge.To, style))
	}

	buf.WriteString("}\n")

	return buf.String(), nil
}

// ExportToMermaid exports a graph to Mermaid format
func (e *Exporter) ExportToMermaid(graph *Graph) (string, error) {
	var buf bytes.Buffer

	buf.WriteString("graph LR\n")

	// Write nodes with styles
	for _, node := range graph.Nodes {
		label := e.formatNodeLabel(node)
		mermaidID := e.sanitizeMermaidID(node.ID)

		switch node.Type {
		case codeintel.NodeTypeFunction:
			buf.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", mermaidID, label))
		case codeintel.NodeTypeMethod:
			buf.WriteString(fmt.Sprintf("  %s{\"%s\"}\n", mermaidID, label))
		case codeintel.NodeTypeType:
			buf.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", mermaidID, label))
		}

		// Add style for exported nodes
		if node.Exported {
			buf.WriteString(fmt.Sprintf("  style %s fill:#e1f5e1\n", mermaidID))
		}
	}

	buf.WriteString("\n")

	// Write edges
	for _, edge := range graph.Edges {
		fromID := e.sanitizeMermaidID(edge.From)
		toID := e.sanitizeMermaidID(edge.To)

		if edge.Dynamic {
			buf.WriteString(fmt.Sprintf("  %s -.-> %s\n", fromID, toID))
		} else {
			buf.WriteString(fmt.Sprintf("  %s --> %s\n", fromID, toID))
		}
	}

	return buf.String(), nil
}

// formatNodeLabel creates a label for a node
func (e *Exporter) formatNodeLabel(node codeintel.GraphNode) string {
	label := node.Name
	if node.Exported {
		label = strings.ToUpper(label[:1]) + label[1:]
	}
	return label
}

// getNodeShape returns the DOT shape for a node type
func (e *Exporter) getNodeShape(nodeType codeintel.NodeType) string {
	switch nodeType {
	case codeintel.NodeTypeFunction:
		return "box"
	case codeintel.NodeTypeMethod:
		return "ellipse"
	case codeintel.NodeTypeType:
		return "diamond"
	case codeintel.NodeTypePackage:
		return "folder"
	default:
		return "box"
	}
}

// getNodeColor returns the fill color for a node
func (e *Exporter) getNodeColor(node codeintel.GraphNode) string {
	if node.Exported {
		return "#e1f5e1"
	}
	return "#f0f0f0"
}

// sanitizeMermaidID sanitizes an ID for Mermaid
func (e *Exporter) sanitizeMermaidID(id string) string {
	// Replace dots and slashes with underscores
	id = strings.ReplaceAll(id, ".", "_")
	id = strings.ReplaceAll(id, "/", "_")
	id = strings.ReplaceAll(id, "*", "ptr_")
	return id
}
