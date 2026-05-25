package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/sessionops"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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

	hasSource := strings.TrimSpace(input.Source) != ""
	if !hasSource && input.StartingBranch != "" {
		err := fmt.Errorf("starting_branch requires source")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, CreateSessionOutput{}, err
	}

	createReq, err := sessionops.BuildCreateSessionRequest(sessionops.CreateSessionRequestOptions{
		Prompt:              prompt,
		Title:               input.Title,
		RequirePlanApproval: input.RequirePlanApproval,
		AutomationMode:      input.AutomationMode,
		NoSource:            !hasSource,
		Source:              input.Source,
		StartingBranch:      input.StartingBranch,
	})
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, CreateSessionOutput{}, err
	}

	session, err := client.Sessions().Create(ctx, createReq)
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
			} else if sessionops.HasJulesAgentMessageAfter(output.RecentActivities, cursor) {
				output.WakeReason = "jules_agent_message"
				output.NextAction = "inspect recent activities and reply with send_session_message if needed"
				return nil, output, nil
			}
		}
		if output.IsTerminal || output.NeedsUserAction || (output.Session != nil && len(output.Session.Outputs) > 0) {
			output.WakeReason = sessionops.DefaultWatchWakeReason(sessionops.WatchDecision{
				Kind: watchDecisionKindFromOutput(output),
				Stop: true,
			})
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
	snapshot, err := sessionops.CurrentWatchSnapshot(ctx, client, sessionID, cursor, sessionops.CurrentWatchOptions{
		FetchActivities: true,
	})
	if err != nil {
		return WatchSessionOutput{}, err
	}
	session := snapshot.Session

	output := WatchSessionOutput{
		SessionID:        session.ID,
		State:            string(session.State),
		NeedsUserAction:  session.State.NeedsUserAction(),
		IsTerminal:       session.State.IsTerminal(),
		Session:          session,
		RecentActivities: snapshot.Activities,
		NextAction:       sessionops.MCPNextAction(snapshot),
	}
	if !snapshot.NextCursor.IsZero() {
		output.NextActivityCursor = snapshot.NextCursor.Format(time.RFC3339Nano)
	}
	return output, nil
}

func watchDecisionKindFromOutput(output WatchSessionOutput) sessionops.WatchDecisionKind {
	switch {
	case output.NeedsUserAction:
		return sessionops.WatchDecisionNeedsUserAction
	case output.IsTerminal:
		switch output.State {
		case string(jules.SessionStateFailed):
			return sessionops.WatchDecisionFailed
		case string(jules.SessionStateCompleted):
			return sessionops.WatchDecisionCompletedWithDeliverables
		default:
			return sessionops.WatchDecisionFailed
		}
	case output.Session != nil && len(output.Session.Outputs) > 0:
		return sessionops.WatchDecisionOutputs
	default:
		return sessionops.WatchDecisionContinue
	}
}
