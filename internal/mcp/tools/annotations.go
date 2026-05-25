package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

func readOnlyOpenWorldTool(title string) *mcp.ToolAnnotations {
	openWorld := true
	return &mcp.ToolAnnotations{
		Title:         title,
		ReadOnlyHint:  true,
		OpenWorldHint: &openWorld,
	}
}

func mutatingOpenWorldTool(title string, destructive bool, idempotent bool) *mcp.ToolAnnotations {
	openWorld := true
	return &mcp.ToolAnnotations{
		Title:           title,
		DestructiveHint: &destructive,
		IdempotentHint:  idempotent,
		OpenWorldHint:   &openWorld,
	}
}
