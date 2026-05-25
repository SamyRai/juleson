package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/pkg/jules"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
		stateCounts[string(session.State)]++
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
func deleteSession(ctx context.Context, req *mcp.CallToolRequest, input DeleteSessionInput, client *jules.Client) (
	*mcp.CallToolResult,
	DeleteSessionOutput,
	error,
) {
	if !input.Confirm {
		err := fmt.Errorf("delete_session requires confirm=true")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, DeleteSessionOutput{}, err
	}

	if err := client.DeleteSession(ctx, input.SessionID); err != nil {
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
		Message:   fmt.Sprintf("Session deleted successfully: %s", input.SessionID),
	}

	return nil, output, nil
}

// applySessionPatches applies git patches from a session to the working directory
func applySessionPatches(ctx context.Context, req *mcp.CallToolRequest, input ApplySessionPatchesInput, client *jules.Client) (
	*mcp.CallToolResult,
	ApplySessionPatchesOutput,
	error,
) {
	options := &julesops.PatchApplicationOptions{
		WorkingDir:   input.WorkingDir,
		DryRun:       input.DryRun,
		Force:        input.Force,
		CreateBackup: input.CreateBackup,
	}

	result, err := julesops.ApplySessionPatches(ctx, client, input.SessionID, options)
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

// previewSessionChanges previews what changes would be made if patches were applied
func previewSessionChanges(ctx context.Context, req *mcp.CallToolRequest, input PreviewSessionChangesInput, client *jules.Client) (
	*mcp.CallToolResult,
	PreviewSessionChangesOutput,
	error,
) {
	changes, err := julesops.PreviewSessionPatches(ctx, client, input.SessionID, input.WorkingDir)

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

// createSession creates a new Jules session
func createSession(ctx context.Context, req *mcp.CallToolRequest, input CreateSessionInput, client *jules.Client) (
	*mcp.CallToolResult,
	CreateSessionOutput,
	error,
) {
	var sourceContext *jules.SourceContext
	if strings.TrimSpace(input.Source) != "" {
		sourceContext = &jules.SourceContext{
			Source: normalizeMCPSourceID(input.Source),
		}
		if input.StartingBranch != "" {
			sourceContext.GithubRepoContext = &jules.GithubRepoContext{
				StartingBranch: input.StartingBranch,
			}
		}
	} else if input.StartingBranch != "" {
		err := fmt.Errorf("starting_branch requires source")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, CreateSessionOutput{}, err
	}

	createReq := &jules.CreateSessionRequest{
		Prompt:              input.Prompt,
		SourceContext:       sourceContext,
		Title:               input.Title,
		RequirePlanApproval: input.RequirePlanApproval,
		AutomationMode:      jules.AutomationMode(input.AutomationMode),
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
func normalizeMCPSourceID(sourceID string) string {
	sourceID = strings.TrimSpace(sourceID)
	if strings.HasPrefix(sourceID, "sources/") {
		return sourceID
	}
	return fmt.Sprintf("sources/%s", sourceID)
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
