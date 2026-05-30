package tools

import (
	"context"

	"github.com/SamyRai/go-jules"

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
		Annotations: readOnlyOpenWorldTool("List Jules Sessions"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListSessionsInput) (*mcp.CallToolResult, ListSessionsOutput, error) {
		return listSessions(ctx, req, input, julesClient)
	})

	// Get Session Status Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session_status",
		Description: "Get detailed status summary of sessions, including active and user-action states",
		Annotations: readOnlyOpenWorldTool("Get Session Status"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetSessionStatusInput) (*mcp.CallToolResult, GetSessionStatusOutput, error) {
		return getSessionStatus(ctx, req, input, julesClient)
	})

	// Approve Session Plan Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "approve_session_plan",
		Description: "Approve a session plan for execution",
		Annotations: mutatingOpenWorldTool("Approve Session Plan", false, true),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ApproveSessionPlanInput) (*mcp.CallToolResult, ApproveSessionPlanOutput, error) {
		return approveSessionPlan(ctx, req, input, julesClient)
	})

	// Delete Session Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_session",
		Description: "Delete a Jules session after explicit confirmation",
		Annotations: mutatingOpenWorldTool("Delete Session", true, true),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteSessionInput) (*mcp.CallToolResult, DeleteSessionOutput, error) {
		return deleteSession(ctx, req, input, julesClient)
	})

	// Apply Session Patches Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "apply_session_patches",
		Description: "Preview by default or apply session patches to a working directory only when confirm_apply=true; blocks dirty worktrees unless allow_dirty=true.",
		Annotations: mutatingOpenWorldTool("Apply Session Patches", true, false),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ApplySessionPatchesInput) (*mcp.CallToolResult, ApplySessionPatchesOutput, error) {
		return applySessionPatches(ctx, req, input, julesClient)
	})

	// Preview Session Changes Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "preview_session_changes",
		Description: "Preview what changes would be made if session patches were applied; supports activity_id and artifact_index scopes.",
		Annotations: readOnlyOpenWorldTool("Preview Session Changes"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input PreviewSessionChangesInput) (*mcp.CallToolResult, PreviewSessionChangesOutput, error) {
		return previewSessionChanges(ctx, req, input, julesClient)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "review_session",
		Description: "Read-only operator review of session state, latest plan, outputs, artifacts, patch dry-run preview, blockers, and safe next actions.",
		Annotations: readOnlyOpenWorldTool("Review Session"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ReviewSessionInput) (*mcp.CallToolResult, ReviewSessionOutput, error) {
		return reviewSession(ctx, req, input, julesClient)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_session_artifacts",
		Description: "List documented session artifacts as a manifest with activity IDs, artifact indexes, patch metadata, media MIME types, and bash exit codes.",
		Annotations: readOnlyOpenWorldTool("List Session Artifacts"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListSessionArtifactsInput) (*mcp.CallToolResult, ListSessionArtifactsOutput, error) {
		return listSessionArtifacts(ctx, req, input, julesClient)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session_outputs",
		Description: "Get documented session outputs such as pull requests.",
		Annotations: readOnlyOpenWorldTool("Get Session Outputs"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetSessionOutputsInput) (*mcp.CallToolResult, GetSessionOutputsOutput, error) {
		return getSessionOutputs(ctx, req, input, julesClient)
	})

	// Send Session Message Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "send_session_message",
		Description: "Send a message to Jules within a session to request changes or provide feedback",
		Annotations: mutatingOpenWorldTool("Send Session Message", false, false),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SendSessionMessageInput) (*mcp.CallToolResult, SendSessionMessageOutput, error) {
		return sendSessionMessage(ctx, req, input, julesClient)
	})

	// Create Session Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_session",
		Description: "Create a Jules coding session; use list_sources first when the source name is unknown.",
		Annotations: mutatingOpenWorldTool("Create Session", false, false),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateSessionInput) (*mcp.CallToolResult, CreateSessionOutput, error) {
		return createSession(ctx, req, input, julesClient)
	})

	// Get Session Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session",
		Description: "Get detailed information about a specific Jules session",
		Annotations: readOnlyOpenWorldTool("Get Session"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetSessionInput) (*mcp.CallToolResult, GetSessionOutput, error) {
		return getSession(ctx, req, input, julesClient)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "watch_session",
		Description: "Poll a Jules session until it completes, fails, or needs user action such as plan approval or feedback.",
		Annotations: readOnlyOpenWorldTool("Watch Session"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input WatchSessionInput) (*mcp.CallToolResult, WatchSessionOutput, error) {
		return watchSession(ctx, req, input, julesClient)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "verify_session_changes",
		Description: "Run repo-standard verification for applied or previewed session changes; detects Go, Node/Yarn, Python/uv, and Rust, or uses an explicit command.",
		Annotations: readOnlyOpenWorldTool("Verify Session Changes"),
	}, verifySessionChanges)
}
