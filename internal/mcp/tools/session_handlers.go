package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
	activeCount := stateCounts[string(jules.SessionStateInProgress)] + stateCounts[string(jules.SessionStatePlanning)] + stateCounts[string(jules.SessionStateQueued)]
	userActionCount := stateCounts[string(jules.SessionStateAwaitingPlanApproval)] + stateCounts[string(jules.SessionStateAwaitingUserFeedback)]

	// Recent sessions (last 5)
	recentCount := 5
	if totalSessions < recentCount {
		recentCount = totalSessions
	}
	recentSessions := sessions[:recentCount]

	// Generate summary
	summary := fmt.Sprintf("Found %d total sessions with %d currently active and %d needing user action", totalSessions, activeCount, userActionCount)

	output := GetSessionStatusOutput{
		TotalSessions:      totalSessions,
		StateBreakdown:     stateCounts,
		ActiveSessions:     activeCount,
		UserActionSessions: userActionCount,
		RecentSessions:     recentSessions,
		Summary:            summary,
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
	dryRun := input.DryRun || !input.ConfirmApply
	if !dryRun && !input.AllowDirty {
		clean, status, err := julesops.IsGitWorkingTreeClean(ctx, input.WorkingDir)
		if err != nil {
			return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to inspect working tree: %v", err)},
					},
				}, ApplySessionPatchesOutput{
					SessionID: input.SessionID,
					DryRun:    true,
					Blockers:  []string{err.Error()},
					Message:   "Refusing to apply patches because working tree status could not be checked",
				}, err
		}
		if !clean {
			blocker := "target worktree has local changes; commit/stash them or set allow_dirty=true"
			if status != "" {
				blocker = blocker + ": " + status
			}
			err := fmt.Errorf("%s", blocker)
			return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: blocker},
					},
				}, ApplySessionPatchesOutput{
					SessionID: input.SessionID,
					DryRun:    true,
					Blockers:  []string{blocker},
					Message:   "Refusing to apply patches to a dirty working tree",
				}, err
		}
	}

	options := &julesops.PatchApplicationOptions{
		WorkingDir:        input.WorkingDir,
		DryRun:            dryRun,
		Force:             input.Force,
		CreateBackup:      input.CreateBackup,
		ActivityID:        input.ActivityID,
		AllowBaseMismatch: input.AllowBaseMismatch,
	}
	if input.ArtifactIndex != nil {
		options.ArtifactIndex = *input.ArtifactIndex
		options.HasArtifactIndex = true
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
		SessionID:               input.SessionID,
		PatchesApplied:          result.PatchesApplied,
		PatchesFailed:           result.PatchesFailed,
		FilesModified:           result.FilesModified,
		SuggestedCommitMessages: result.SuggestedCommitMessages,
		Warnings:                result.Warnings,
		BaseCommitMismatches:    result.BaseCommitMismatches,
		Errors:                  result.Errors,
		DryRun:                  result.DryRun,
		Message:                 message,
	}

	return nil, output, nil
}

// previewSessionChanges previews what changes would be made if patches were applied
func previewSessionChanges(ctx context.Context, req *mcp.CallToolRequest, input PreviewSessionChangesInput, client *jules.Client) (
	*mcp.CallToolResult,
	PreviewSessionChangesOutput,
	error,
) {
	options := &julesops.PatchApplicationOptions{
		WorkingDir: input.WorkingDir,
		ActivityID: input.ActivityID,
	}
	if input.ArtifactIndex != nil {
		options.ArtifactIndex = *input.ArtifactIndex
		options.HasArtifactIndex = true
	}
	changes, err := julesops.PreviewSessionPatchesWithOptions(ctx, client, input.SessionID, options)

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
		SessionID:               input.SessionID,
		TotalPatches:            changes.TotalPatches,
		Files:                   changes.Files,
		SuggestedCommitMessages: changes.SuggestedCommitMessages,
		Warnings:                changes.Warnings,
		BaseCommitMismatches:    changes.BaseCommitMismatches,
		CanApply:                canApply,
		Errors:                  errors,
		Summary:                 summary,
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
	prompt := strings.TrimSpace(input.Prompt)
	if input.PromptFile != "" {
		if prompt != "" {
			err := fmt.Errorf("provide either prompt or prompt_file, not both")
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, CreateSessionOutput{}, err
		}
		data, err := os.ReadFile(input.PromptFile)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to read prompt_file: %v", err)},
				},
			}, CreateSessionOutput{}, err
		}
		prompt = strings.TrimSpace(string(data))
	}
	if prompt == "" {
		err := fmt.Errorf("prompt or prompt_file is required")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, CreateSessionOutput{}, err
	}

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
		Prompt:              prompt,
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

func watchSession(ctx context.Context, req *mcp.CallToolRequest, input WatchSessionInput, client *jules.Client) (
	*mcp.CallToolResult,
	WatchSessionOutput,
	error,
) {
	interval := time.Duration(input.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	timeout := time.Duration(input.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}

	watchCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var cursor time.Time
	baselineState := jules.SessionState(strings.TrimSpace(input.InitialState))
	hasStateBaseline := baselineState != ""
	hasActivityBaseline := false
	if input.Since != "" {
		parsed, err := time.Parse(time.RFC3339Nano, input.Since)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Invalid since cursor: %v", err)},
				},
			}, WatchSessionOutput{}, err
		}
		cursor = parsed
		hasActivityBaseline = true
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		output, err := currentWatchSessionOutput(watchCtx, input.SessionID, client, cursor)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to watch session: %v", err)},
				},
			}, WatchSessionOutput{}, err
		}
		currentState := jules.SessionState(output.State)
		if input.ReturnOnStatusChange {
			if !hasStateBaseline {
				baselineState = currentState
				hasStateBaseline = true
			} else if currentState != baselineState {
				output.WakeReason = "status_change"
				output.NextAction = fmt.Sprintf("session state changed from %s to %s; inspect get_session before taking action", baselineState, currentState)
				return nil, output, nil
			}
		}
		if input.ReturnOnJulesAgentMessage {
			if !hasActivityBaseline {
				hasActivityBaseline = true
			} else if hasJulesAgentMessageAfter(output.RecentActivities, cursor) {
				output.WakeReason = "jules_agent_message"
				output.NextAction = "inspect recent activities and reply with send_session_message if needed"
				return nil, output, nil
			}
		}
		if output.IsTerminal || output.NeedsUserAction || (output.Session != nil && len(output.Session.Outputs) > 0) {
			output.WakeReason = defaultWatchWakeReason(output)
			return nil, output, nil
		}
		if output.NextActivityCursor != "" {
			if parsed, err := time.Parse(time.RFC3339Nano, output.NextActivityCursor); err == nil && parsed.After(cursor) {
				cursor = parsed
			}
		}

		select {
		case <-watchCtx.Done():
			output.NextAction = fmt.Sprintf("watch timed out after %s; call watch_session again or inspect get_session", timeout)
			return nil, output, nil
		case <-ticker.C:
		}
	}
}

func currentWatchSessionOutput(ctx context.Context, sessionID string, client *jules.Client, cursor time.Time) (WatchSessionOutput, error) {
	session, err := client.GetSession(ctx, sessionID)
	if err != nil {
		return WatchSessionOutput{}, err
	}

	activities, err := client.ListActivitiesSince(ctx, sessionID, cursor, 25)
	if err != nil {
		activities = nil
	}
	nextCursor := jules.ActivityCursor(activities)

	output := WatchSessionOutput{
		SessionID:        session.ID,
		State:            string(session.State),
		NeedsUserAction:  session.State.NeedsUserAction(),
		IsTerminal:       session.State.IsTerminal(),
		Session:          session,
		RecentActivities: activities,
		NextAction:       "session is still active; keep watching",
	}
	if !nextCursor.IsZero() {
		output.NextActivityCursor = nextCursor.Format(time.RFC3339Nano)
	}
	switch {
	case session.State == jules.SessionStateAwaitingPlanApproval:
		output.NextAction = "inspect get_session_plans, then call approve_session_plan or send_session_message"
	case session.State == jules.SessionStateAwaitingUserFeedback:
		output.NextAction = "inspect recent activities, then call send_session_message"
	case session.State == jules.SessionStateCompleted:
		hasDeliverables, err := julesops.SessionHasDeliverables(ctx, client, session)
		switch {
		case err != nil:
			output.NextAction = "inspect list_session_activities and list_session_artifacts; deliverable check failed"
		case !hasDeliverables:
			output.NextAction = "no retrievable deliverable was produced; inspect list_session_artifacts or create a follow-up session"
		default:
			output.NextAction = "call preview_session_changes, then apply_session_patches with confirm_apply=true if acceptable"
		}
	case session.State == jules.SessionStateFailed:
		output.NextAction = "inspect get_session and list_session_activities for failure details"
	case len(session.Outputs) > 0:
		output.NextAction = "call get_session_outputs to inspect created pull requests or other outputs"
	}
	return output, nil
}

func hasJulesAgentMessageAfter(activities []jules.Activity, cursor time.Time) bool {
	for _, activity := range activities {
		if activity.AgentMessaged == nil {
			continue
		}
		if cursor.IsZero() || activity.CreateTime.After(cursor) {
			return true
		}
	}
	return false
}

func defaultWatchWakeReason(output WatchSessionOutput) string {
	switch {
	case output.NeedsUserAction:
		return "user_action"
	case output.IsTerminal:
		return "terminal_state"
	case output.Session != nil && len(output.Session.Outputs) > 0:
		return "session_outputs"
	default:
		return ""
	}
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
	session, err := client.GetSession(ctx, input.SessionID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get session outputs: %v", err)},
			},
		}, GetSessionOutputsOutput{}, err
	}
	outputs := session.Outputs
	documentedOutputs := make([]jules.Output, 0, len(outputs))
	for _, output := range outputs {
		if output.PullRequest != nil {
			documentedOutputs = append(documentedOutputs, output)
		}
	}
	return nil, GetSessionOutputsOutput{
		SessionID:  session.ID,
		Outputs:    documentedOutputs,
		TotalCount: len(documentedOutputs),
	}, nil
}
