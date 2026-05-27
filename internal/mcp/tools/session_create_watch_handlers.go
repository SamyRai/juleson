package tools

import (
	"context"
	"errors"
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
	wakePolicy, err := sessionops.ParseWakePolicy(input.WakePolicy)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, WatchSessionOutput{}, err
	}
	if input.ReturnOnStatusChange {
		wakePolicy = sessionops.WakePolicyAnyStatus
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
	var lastOutput WatchSessionOutput

	for {
		output, err := currentWatchSessionOutput(watchCtx, input.SessionID, client, cursor)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || watchCtx.Err() != nil {
				return nil, timeoutWatchSessionOutput(lastOutput, input.SessionID, timeout), nil
			}
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to watch session: %v", err)},
				},
			}, WatchSessionOutput{}, err
		}
		lastOutput = output
		currentState := jules.SessionState(output.State)
		stateChanged := false
		if !hasStateBaseline {
			baselineState = currentState
			hasStateBaseline = true
		} else if currentState != baselineState {
			stateChanged = true
		}
		wake := sessionops.EvaluateWatchWake(wakePolicy, sessionops.WatchUpdateType(output.UpdateType), stateChanged)
		if input.ReturnOnJulesAgentMessage {
			if !hasActivityBaseline {
				hasActivityBaseline = true
			} else if sessionops.HasJulesAgentMessageAfter(output.RecentActivities, cursor) {
				wake = sessionops.EvaluateWatchWake(wakePolicy, sessionops.WatchUpdateAgentMessage, false)
				output.UpdateType = string(sessionops.WatchUpdateAgentMessage)
			}
		}
		if output.NextActivityCursor != "" {
			if parsed, err := time.Parse(time.RFC3339Nano, output.NextActivityCursor); err == nil && parsed.After(cursor) {
				cursor = parsed
			}
		}
		if wake.ShouldWake {
			output.ShouldWake = true
			output.WakeReason = wake.WakeReason
			if wake.WakeReason == "status_change" {
				output.WakeReason = "status_change"
				output.NextAction = fmt.Sprintf("session state changed from %s to %s; inspect get_session before taking action", baselineState, currentState)
			}
			if wake.UpdateType == sessionops.WatchUpdateAgentMessage {
				output.NextAction = "inspect recent activities and reply with send_session_message if needed"
			}
			return nil, output, nil
		}
		if stateChanged {
			baselineState = currentState
		}

		select {
		case <-watchCtx.Done():
			return nil, timeoutWatchSessionOutput(output, input.SessionID, timeout), nil
		case <-ticker.C:
		}
	}
}

func timeoutWatchSessionOutput(output WatchSessionOutput, sessionID string, timeout time.Duration) WatchSessionOutput {
	if output.SessionID == "" {
		output.SessionID = sessionID
	}
	output.ShouldWake = false
	output.WakeReason = ""
	output.NextAction = fmt.Sprintf("watch timed out after %s; call watch_session again or inspect get_session", timeout)
	return output
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
		UpdateType:       string(sessionops.WatchUpdateTypeForDecision(snapshot.Decision)),
		NeedsUserAction:  session.State.NeedsUserAction(),
		IsTerminal:       session.State.IsTerminal(),
		Session:          session,
		RecentActivities: snapshot.Activities,
		NextAction:       snapshot.NextAction,
		WakeReason:       snapshot.WakeReason,
	}
	if !snapshot.NextCursor.IsZero() {
		output.NextActivityCursor = snapshot.NextCursor.Format(time.RFC3339Nano)
	}
	return output, nil
}
