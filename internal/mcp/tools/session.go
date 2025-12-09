package tools

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/jules"

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

	// NOTE: cancel_session and delete_session tools are NOT available
	// The Jules API v1alpha does not support these operations.
	// Users must use the Jules web UI to cancel or delete sessions.
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

	sessions := response.Sessions
	if sessions == nil {
		sessions = []jules.Session{}
	}

	output := ListSessionsOutput{
		Sessions:   sessions,
		NextCursor: response.NextPageToken,
		TotalCount: len(sessions),
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
	if sessions == nil {
		sessions = []jules.Session{}
	}
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

// NOTE: CancelSession and DeleteSession functionality removed
// The Jules API v1alpha does not provide these endpoints.
// Cancel/Delete operations are only available via the Jules web UI.

// ApplySessionPatchesInput represents input for apply_session_patches tool
type ApplySessionPatchesInput struct {
	SessionID    string `json:"session_id" jsonschema:"ID of the session to apply patches from"`
	WorkingDir   string `json:"working_dir,omitempty" jsonschema:"Working directory where patches should be applied (default: current directory)"`
	DryRun       bool   `json:"dry_run,omitempty" jsonschema:"Whether to perform a dry-run without actually applying changes (default: false)"`
	Force        bool   `json:"force,omitempty" jsonschema:"Whether to force application even if some hunks fail (default: false)"`
	CreateBackup bool   `json:"create_backup,omitempty" jsonschema:"Whether to create backup files before applying patches (default: false)"`
}

// ApplySessionPatchesOutput represents output for apply_session_patches tool
type ApplySessionPatchesOutput struct {
	SessionID      string   `json:"session_id"`
	PatchesApplied int      `json:"patches_applied"`
	PatchesFailed  int      `json:"patches_failed"`
	FilesModified  []string `json:"files_modified"`
	Errors         []string `json:"errors,omitempty"`
	DryRun         bool     `json:"dry_run"`
	Message        string   `json:"message"`
}

// applySessionPatches applies git patches from a session to the working directory
func applySessionPatches(ctx context.Context, req *mcp.CallToolRequest, input ApplySessionPatchesInput, client *jules.Client) (
	*mcp.CallToolResult,
	ApplySessionPatchesOutput,
	error,
) {
	options := &jules.PatchApplicationOptions{
		WorkingDir:   input.WorkingDir,
		DryRun:       input.DryRun,
		Force:        input.Force,
		CreateBackup: input.CreateBackup,
	}

	result, err := client.ApplySessionPatches(ctx, input.SessionID, options)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to apply session patches: %v", err)},
			},
		}, ApplySessionPatchesOutput{}, err
	}

	message := fmt.Sprintf("Successfully applied %d patches", result.PatchesApplied)
	if result.DryRun {
		message = fmt.Sprintf("Dry-run: %d patches can be applied", result.PatchesApplied)
	}
	if result.PatchesFailed > 0 {
		message += fmt.Sprintf(", %d patches failed", result.PatchesFailed)
	}

	output := ApplySessionPatchesOutput{
		SessionID:      input.SessionID,
		PatchesApplied: result.PatchesApplied,
		PatchesFailed:  result.PatchesFailed,
		FilesModified:  result.FilesModified,
		Errors:         result.Errors,
		DryRun:         result.DryRun,
		Message:        message,
	}

	return nil, output, nil
}

// PreviewSessionChangesInput represents input for preview_session_changes tool
type PreviewSessionChangesInput struct {
	SessionID  string `json:"session_id" jsonschema:"ID of the session to preview changes for"`
	WorkingDir string `json:"working_dir,omitempty" jsonschema:"Working directory (default: current directory)"`
}

// PreviewSessionChangesOutput represents output for preview_session_changes tool
type PreviewSessionChangesOutput struct {
	SessionID    string             `json:"session_id"`
	TotalPatches int                `json:"total_patches"`
	Files        []jules.FileChange `json:"files"`
	CanApply     bool               `json:"can_apply"`
	Errors       []string           `json:"errors,omitempty"`
	Summary      string             `json:"summary"`
}

// previewSessionChanges previews what changes would be made if patches were applied
func previewSessionChanges(ctx context.Context, req *mcp.CallToolRequest, input PreviewSessionChangesInput, client *jules.Client) (
	*mcp.CallToolResult,
	PreviewSessionChangesOutput,
	error,
) {
	changes, err := client.PreviewSessionPatches(ctx, input.SessionID, input.WorkingDir)

	canApply := true
	var errors []string
	if err != nil {
		canApply = false
		errors = append(errors, err.Error())
	}

	if changes == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to preview session changes: %v", err)},
			},
		}, PreviewSessionChangesOutput{}, err
	}

	totalLinesAdded := 0
	totalLinesRemoved := 0
	for _, file := range changes.Files {
		totalLinesAdded += file.LinesAdded
		totalLinesRemoved += file.LinesRemoved
	}

	summary := fmt.Sprintf("%d patches affecting %d files (+%d -%d lines)",
		changes.TotalPatches, len(changes.Files), totalLinesAdded, totalLinesRemoved)

	if !canApply {
		summary += " - WARNING: Some patches may fail to apply"
	}

	output := PreviewSessionChangesOutput{
		SessionID:    input.SessionID,
		TotalPatches: changes.TotalPatches,
		Files:        changes.Files,
		CanApply:     canApply,
		Errors:       errors,
		Summary:      summary,
	}

	return nil, output, nil
}

// SendSessionMessageInput represents input for send_session_message tool
type SendSessionMessageInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to send message to"`
	Message   string `json:"message" jsonschema:"Message to send to Jules within the session"`
}

// SendSessionMessageOutput represents output for send_session_message tool
type SendSessionMessageOutput struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// sendSessionMessage sends a message to Jules within a session
func sendSessionMessage(ctx context.Context, req *mcp.CallToolRequest, input SendSessionMessageInput, client *jules.Client) (
	*mcp.CallToolResult,
	SendSessionMessageOutput,
	error,
) {
	sendReq := &jules.SendMessageRequest{
		Prompt: input.Message,
	}

	err := client.SendMessage(ctx, input.SessionID, sendReq)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to send message: %v", err)},
			},
		}, SendSessionMessageOutput{}, err
	}

	output := SendSessionMessageOutput{
		SessionID: input.SessionID,
		Status:    "sent",
		Message:   "Message sent successfully to Jules session",
	}

	return nil, output, nil
}

// CreateSessionInput represents input for create_session tool
type CreateSessionInput struct {
	Source              string `json:"source" jsonschema:"Source ID or path (e.g., 'sources/github/owner/repo')"`
	Prompt              string `json:"prompt" jsonschema:"Prompt describing the task for Jules to work on"`
	Title               string `json:"title,omitempty" jsonschema:"Optional title for the session"`
	RequirePlanApproval bool   `json:"require_plan_approval,omitempty" jsonschema:"Whether to require manual approval of plans (default: false)"`
	AutomationMode      string `json:"automation_mode,omitempty" jsonschema:"Automation mode (e.g., 'AUTO_CREATE_PR')"`
	StartingBranch      string `json:"starting_branch,omitempty" jsonschema:"Starting branch for GitHub repos (default: repo's default branch)"`
}

// CreateSessionOutput represents output for create_session tool
type CreateSessionOutput struct {
	SessionID string        `json:"session_id"`
	Session   jules.Session `json:"session"`
	URL       string        `json:"url"`
	Message   string        `json:"message"`
}

// createSession creates a new Jules session
func createSession(ctx context.Context, req *mcp.CallToolRequest, input CreateSessionInput, client *jules.Client) (
	*mcp.CallToolResult,
	CreateSessionOutput,
	error,
) {
	sourceContext := &jules.SourceContext{
		Source: input.Source,
	}

	if input.StartingBranch != "" {
		sourceContext.GithubRepoContext = &jules.GithubRepoContext{
			StartingBranch: input.StartingBranch,
		}
	}

	createReq := &jules.CreateSessionRequest{
		Prompt:              input.Prompt,
		SourceContext:       sourceContext,
		Title:               input.Title,
		RequirePlanApproval: input.RequirePlanApproval,
		AutomationMode:      input.AutomationMode,
	}

	session, err := client.CreateSession(ctx, createReq)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create session: %v", err)},
			},
		}, CreateSessionOutput{}, err
	}

	output := CreateSessionOutput{
		SessionID: session.ID,
		Session:   *session,
		URL:       session.URL,
		Message:   fmt.Sprintf("Session created successfully: %s", session.ID),
	}

	return nil, output, nil
}

// GetSessionInput represents input for get_session tool
type GetSessionInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to retrieve"`
}

// GetSessionOutput represents output for get_session tool
type GetSessionOutput struct {
	SessionID string        `json:"session_id"`
	Session   jules.Session `json:"session"`
	URL       string        `json:"url"`
}

// getSession retrieves detailed information about a session
func getSession(ctx context.Context, req *mcp.CallToolRequest, input GetSessionInput, client *jules.Client) (
	*mcp.CallToolResult,
	GetSessionOutput,
	error,
) {
	session, err := client.GetSession(ctx, input.SessionID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get session: %v", err)},
			},
		}, GetSessionOutput{}, err
	}

	output := GetSessionOutput{
		SessionID: session.ID,
		Session:   *session,
		URL:       session.URL,
	}

	return nil, output, nil
}
