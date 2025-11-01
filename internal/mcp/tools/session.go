package tools

import (
	"context"
	"fmt"

	"jules-automation/internal/jules"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterSessionTools registers all session-related MCP tools
func RegisterSessionTools(server *mcp.Server, julesClient *jules.Client) {
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

	// Cancel Session Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "cancel_session",
		Description: "Cancel a running session",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CancelSessionInput) (*mcp.CallToolResult, CancelSessionOutput, error) {
		return cancelSession(ctx, req, input, julesClient)
	})

	// Delete Session Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_session",
		Description: "Delete a completed or failed session",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteSessionInput) (*mcp.CallToolResult, DeleteSessionOutput, error) {
		return deleteSession(ctx, req, input, julesClient)
	})
}

// ListSessionsInput represents input for list_sessions tool
type ListSessionsInput struct {
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of sessions to return (default: 50)"`
	Cursor string `json:"cursor,omitempty" jsonschema:"Cursor for pagination"`
}

// ListSessionsOutput represents output for list_sessions tool
type ListSessionsOutput struct {
	Sessions   []jules.Session `json:"sessions"`
	NextCursor string          `json:"next_cursor,omitempty"`
	TotalCount int             `json:"total_count"`
}

// listSessions lists all Jules sessions
func listSessions(ctx context.Context, req *mcp.CallToolRequest, input ListSessionsInput, client *jules.Client) (
	*mcp.CallToolResult,
	ListSessionsOutput,
	error,
) {
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}

	response, err := client.ListSessionsWithPagination(ctx, limit, input.Cursor)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list sessions: %v", err)},
			},
		}, ListSessionsOutput{}, err
	}

	output := ListSessionsOutput{
		Sessions:   response.Sessions,
		NextCursor: response.NextPageToken,
		TotalCount: len(response.Sessions),
	}

	return nil, output, nil
}

// GetSessionStatusInput represents input for get_session_status tool
type GetSessionStatusInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"Maximum number of sessions to analyze (default: 100)"`
}

// GetSessionStatusOutput represents output for get_session_status tool
type GetSessionStatusOutput struct {
	TotalSessions  int             `json:"total_sessions"`
	StateBreakdown map[string]int  `json:"state_breakdown"`
	ActiveSessions int             `json:"active_sessions"`
	RecentSessions []jules.Session `json:"recent_sessions"`
	Summary        string          `json:"summary"`
}

// getSessionStatus gets detailed status summary of all sessions
func getSessionStatus(ctx context.Context, req *mcp.CallToolRequest, input GetSessionStatusInput, client *jules.Client) (
	*mcp.CallToolResult,
	GetSessionStatusOutput,
	error,
) {
	limit := input.Limit
	if limit <= 0 {
		limit = 100
	}

	response, err := client.ListSessionsWithPagination(ctx, limit, "")
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get session status: %v", err)},
			},
		}, GetSessionStatusOutput{}, err
	}

	sessions := response.Sessions
	totalSessions := len(sessions)

	// Count sessions by state
	stateCounts := make(map[string]int)
	for _, session := range sessions {
		stateCounts[session.State]++
	}

	// Active sessions count
	activeCount := stateCounts["IN_PROGRESS"] + stateCounts["PLANNING"]

	// Recent sessions (last 5)
	recentCount := 5
	if totalSessions < recentCount {
		recentCount = totalSessions
	}
	recentSessions := sessions[:recentCount]

	// Generate summary
	summary := fmt.Sprintf("Found %d total sessions with %d currently active", totalSessions, activeCount)

	output := GetSessionStatusOutput{
		TotalSessions:  totalSessions,
		StateBreakdown: stateCounts,
		ActiveSessions: activeCount,
		RecentSessions: recentSessions,
		Summary:        summary,
	}

	return nil, output, nil
}

// ApproveSessionPlanInput represents input for approve_session_plan tool
type ApproveSessionPlanInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to approve"`
}

// ApproveSessionPlanOutput represents output for approve_session_plan tool
type ApproveSessionPlanOutput struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// approveSessionPlan approves a session plan for execution
func approveSessionPlan(ctx context.Context, req *mcp.CallToolRequest, input ApproveSessionPlanInput, client *jules.Client) (
	*mcp.CallToolResult,
	ApproveSessionPlanOutput,
	error,
) {
	err := client.ApprovePlan(ctx, input.SessionID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to approve session plan: %v", err)},
			},
		}, ApproveSessionPlanOutput{}, err
	}

	output := ApproveSessionPlanOutput{
		SessionID: input.SessionID,
		Status:    "approved",
		Message:   "Session plan approved successfully",
	}

	return nil, output, nil
}

// CancelSessionInput represents input for cancel_session tool
type CancelSessionInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to cancel"`
}

// CancelSessionOutput represents output for cancel_session tool
type CancelSessionOutput struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// cancelSession cancels a running session
func cancelSession(ctx context.Context, req *mcp.CallToolRequest, input CancelSessionInput, client *jules.Client) (
	*mcp.CallToolResult,
	CancelSessionOutput,
	error,
) {
	err := client.CancelSession(ctx, input.SessionID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to cancel session: %v", err)},
			},
		}, CancelSessionOutput{}, err
	}

	output := CancelSessionOutput{
		SessionID: input.SessionID,
		Status:    "cancelled",
		Message:   "Session cancelled successfully",
	}

	return nil, output, nil
}

// DeleteSessionInput represents input for delete_session tool
type DeleteSessionInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to delete"`
}

// DeleteSessionOutput represents output for delete_session tool
type DeleteSessionOutput struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// deleteSession deletes a completed or failed session
func deleteSession(ctx context.Context, req *mcp.CallToolRequest, input DeleteSessionInput, client *jules.Client) (
	*mcp.CallToolResult,
	DeleteSessionOutput,
	error,
) {
	err := client.DeleteSession(ctx, input.SessionID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to delete session: %v", err)},
			},
		}, DeleteSessionOutput{}, err
	}

	output := DeleteSessionOutput{
		SessionID: input.SessionID,
		Status:    "deleted",
		Message:   "Session deleted successfully",
	}

	return nil, output, nil
}
