package tools

import (
	"context"
	"fmt"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/internal/sessionops"

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

	response, err := client.Sessions().List(ctx, &jules.ListSessionsOptions{PageSize: limit, PageToken: input.Cursor})
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

	response, err := client.Sessions().List(ctx, &jules.ListSessionsOptions{PageSize: limit})
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get session status: %v", err)},
			},
		}, GetSessionStatusOutput{}, err
	}

	summary := sessionops.SummarizeSessions(response.Sessions, 5)

	output := GetSessionStatusOutput{
		TotalSessions:      summary.TotalSessions,
		StateBreakdown:     summary.StateBreakdown,
		ActiveSessions:     summary.ActiveSessions,
		UserActionSessions: summary.UserActionSessions,
		RecentSessions:     summary.RecentSessions,
		Summary:            summary.Summary,
	}

	return nil, output, nil
}

// approveSessionPlan approves a session plan for execution
func approveSessionPlan(ctx context.Context, req *mcp.CallToolRequest, input ApproveSessionPlanInput, client *jules.Client) (
	*mcp.CallToolResult,
	ApproveSessionPlanOutput,
	error,
) {
	err := client.Sessions().ApprovePlan(ctx, input.SessionID)
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

	if err := client.Sessions().Delete(ctx, input.SessionID); err != nil {
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

// sendSessionMessage sends a message to Jules within a session
func sendSessionMessage(ctx context.Context, req *mcp.CallToolRequest, input SendSessionMessageInput, client *jules.Client) (
	*mcp.CallToolResult,
	SendSessionMessageOutput,
	error,
) {
	sendReq := &jules.SendMessageRequest{
		Prompt: input.Message,
	}

	err := client.Sessions().SendMessage(ctx, input.SessionID, sendReq)
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

// getSession retrieves detailed information about a session
func getSession(ctx context.Context, req *mcp.CallToolRequest, input GetSessionInput, client *jules.Client) (
	*mcp.CallToolResult,
	GetSessionOutput,
	error,
) {
	session, err := client.Sessions().Get(ctx, input.SessionID)
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

func verifySessionChanges(ctx context.Context, req *mcp.CallToolRequest, input VerifySessionChangesInput) (
	*mcp.CallToolResult,
	VerifySessionChangesOutput,
	error,
) {
	result, err := julesops.VerifyProjectChanges(ctx, julesops.VerificationOptions{
		WorkingDir: input.WorkingDir,
		Command:    input.Command,
		Packages:   input.Packages,
		Short:      input.Short,
	})
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		}, VerifySessionChangesOutput{}, err
	}
	output := VerifySessionChangesOutput{
		WorkingDir: result.WorkingDir,
		Success:    result.Success,
		Command:    result.Command,
		Output:     result.Output,
		Summary:    result.Summary,
	}
	if !result.Success {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: output.Summary}},
		}, output, nil
	}
	return nil, output, nil
}

func listSessionArtifacts(ctx context.Context, req *mcp.CallToolRequest, input ListSessionArtifactsInput, client *jules.Client) (
	*mcp.CallToolResult,
	ListSessionArtifactsOutput,
	error,
) {
	artifacts, err := julesops.ListSessionArtifactManifests(ctx, client, input.SessionID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list session artifacts: %v", err)},
			},
		}, ListSessionArtifactsOutput{}, err
	}
	if artifacts == nil {
		artifacts = []julesops.ArtifactManifest{}
	}
	return nil, ListSessionArtifactsOutput{
		SessionID:  input.SessionID,
		Artifacts:  artifacts,
		TotalCount: len(artifacts),
	}, nil
}

func getSessionOutputs(ctx context.Context, req *mcp.CallToolRequest, input GetSessionOutputsInput, client *jules.Client) (
	*mcp.CallToolResult,
	GetSessionOutputsOutput,
	error,
) {
	session, err := client.Sessions().Get(ctx, input.SessionID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get session outputs: %v", err)},
			},
		}, GetSessionOutputsOutput{}, err
	}
	documentedOutputs := sessionops.DocumentedOutputs(session)
	return nil, GetSessionOutputsOutput{
		SessionID:  session.ID,
		Outputs:    documentedOutputs,
		TotalCount: len(documentedOutputs),
	}, nil
}

func reviewSession(ctx context.Context, req *mcp.CallToolRequest, input ReviewSessionInput, client *jules.Client) (
	*mcp.CallToolResult,
	ReviewSessionOutput,
	error,
) {
	artifactIndex := 0
	hasArtifactIndex := false
	if input.ArtifactIndex != nil {
		artifactIndex = *input.ArtifactIndex
		hasArtifactIndex = true
	}
	review, err := sessionops.BuildSessionReview(ctx, client, sessionops.ReviewRequest{
		SessionID:        input.SessionID,
		WorkingDir:       input.WorkingDir,
		ActivityID:       input.ActivityID,
		ArtifactIndex:    artifactIndex,
		HasArtifactIndex: hasArtifactIndex,
	})
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to review session: %v", err)},
			},
		}, ReviewSessionOutput{}, err
	}
	return nil, ReviewSessionOutput{Review: *review}, nil
}
