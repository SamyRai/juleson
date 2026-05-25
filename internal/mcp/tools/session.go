package tools

import (
	"context"

	"github.com/SamyRai/juleson/pkg/jules"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterSessionTools registers all session-related MCP tools
func RegisterSessionTools(server *mcp.Server, julesClient *jules.Client) {
	// Don't register session tools if client is not available
	if julesClient == nil {
		return
	}

	// List Sessions Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sessions",
		Description: "List all Jules sessions with their current status",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListSessionsInput) (*mcp.CallToolResult, ListSessionsOutput, error) {
		return listSessions(ctx, req, input, julesClient)
	})

	// Get Session Status Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session_status",
		Description: "Get detailed status summary of all sessions",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetSessionStatusInput) (*mcp.CallToolResult, GetSessionStatusOutput, error) {
		return getSessionStatus(ctx, req, input, julesClient)
	})

	// Approve Session Plan Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "approve_session_plan",
		Description: "Approve a session plan for execution",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ApproveSessionPlanInput) (*mcp.CallToolResult, ApproveSessionPlanOutput, error) {
		return approveSessionPlan(ctx, req, input, julesClient)
	})

	// Delete Session Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_session",
		Description: "Delete a Jules session after explicit confirmation",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteSessionInput) (*mcp.CallToolResult, DeleteSessionOutput, error) {
		return deleteSession(ctx, req, input, julesClient)
	})

	// Apply Session Patches Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "apply_session_patches",
		Description: "Apply git patches from a session to the working directory (similar to 'jules remote pull --apply')",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ApplySessionPatchesInput) (*mcp.CallToolResult, ApplySessionPatchesOutput, error) {
		return applySessionPatches(ctx, req, input, julesClient)
	})

	// Preview Session Changes Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "preview_session_changes",
		Description: "Preview what changes would be made if session patches were applied (dry-run)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input PreviewSessionChangesInput) (*mcp.CallToolResult, PreviewSessionChangesOutput, error) {
		return previewSessionChanges(ctx, req, input, julesClient)
	})

	// Send Session Message Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "send_session_message",
		Description: "Send a message to Jules within a session to request changes or provide feedback",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SendSessionMessageInput) (*mcp.CallToolResult, SendSessionMessageOutput, error) {
		return sendSessionMessage(ctx, req, input, julesClient)
	})

	// Create Session Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_session",
		Description: "Create a new Jules coding session with a source and prompt",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateSessionInput) (*mcp.CallToolResult, CreateSessionOutput, error) {
		return createSession(ctx, req, input, julesClient)
	})

	// Get Session Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session",
		Description: "Get detailed information about a specific Jules session",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetSessionInput) (*mcp.CallToolResult, GetSessionOutput, error) {
		return getSession(ctx, req, input, julesClient)
	})
}
