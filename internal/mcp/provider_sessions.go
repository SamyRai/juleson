package jmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SamyRai/go-jules"
	julessessions "github.com/SamyRai/juleson/internal/jules/sessions"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type sessionsProvider struct {
	clientFactory clientFactory
}

// NewSessionsProvider creates a ToolProvider for session management.
func NewSessionsProvider(cf clientFactory) ToolProvider {
	return &sessionsProvider{clientFactory: cf}
}

func (p *sessionsProvider) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sessions",
		Description: "List Jules sessions.",
	}, p.listSessions)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session",
		Description: "Get a Jules session by ID.",
	}, p.getSession)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_session",
		Description: "Create a Jules session. Source-backed sessions require a Jules source ID; repoless sessions set no_source=true.",
	}, p.createSession)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "approve_session_plan",
		Description: "Approve the current plan for a Jules session. Requires confirm=true.",
	}, p.approvePlan)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "send_session_message",
		Description: "Send feedback or a follow-up message to a Jules session.",
	}, p.sendMessage)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_session",
		Description: "Delete a Jules session. Requires confirm=true.",
	}, p.deleteSession)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session_plans",
		Description: "Return generated plans found in a Jules session's activities.",
	}, p.getPlans)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "review_session",
		Description: "Build a read-only operator review with plans, outputs, artifacts, patch preview, worktree state, blockers, and next actions.",
	}, p.reviewSession)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "watch_session",
		Description: "Wait for a Jules session to complete or require human approval, emitting progress notifications. Acts as a long-running task.",
	}, p.watchSession)

	server.AddPrompt(&mcp.Prompt{
		Name:        "review_jules_plan",
		Description: "Prompt template for an agent to review a Jules session plan.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "session_id",
				Description: "Jules session ID to review",
				Required:    true,
			},
		},
	}, p.getReviewPrompt)
}

type listSessionsInput struct {
	Filter    string `json:"filter,omitempty"`
	PageToken string `json:"page_token,omitempty"`
	PageSize  int    `json:"page_size,omitempty"`
}

func (p *sessionsProvider) listSessions(ctx context.Context, _ *mcp.CallToolRequest, in listSessionsInput) (*mcp.CallToolResult, *jules.SessionsResponse, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, nil, err
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 30
	}
	response, err := client.Sessions().List(ctx, &jules.ListSessionsOptions{
		PageSize:  pageSize,
		PageToken: in.PageToken,
		Filter:    in.Filter,
	})
	return nil, response, wrapAPIError("list sessions", err)
}

func (p *sessionsProvider) getSession(ctx context.Context, _ *mcp.CallToolRequest, in sessionIDInput) (*mcp.CallToolResult, *jules.Session, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, nil, err
	}
	session, err := client.Sessions().Get(ctx, in.SessionID)
	return nil, session, wrapAPIError("get session", err)
}

type createSessionInput struct {
	SourceID            *string `json:"source_id,omitempty"`
	Title               *string `json:"title,omitempty"`
	StartingBranch      *string `json:"starting_branch,omitempty"`
	AutomationMode      *string `json:"automation_mode,omitempty"`
	Prompt              string  `json:"prompt"`
	NoSource            bool    `json:"no_source,omitempty"`
	RequirePlanApproval bool    `json:"require_plan_approval,omitempty"`
}

func (p *sessionsProvider) createSession(ctx context.Context, _ *mcp.CallToolRequest, in createSessionInput) (*mcp.CallToolResult, *jules.Session, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, nil, err
	}
	req := &jules.CreateSessionRequest{
		Prompt:              in.Prompt,
		Title:               optionalString(in.Title),
		RequirePlanApproval: in.RequirePlanApproval,
		AutomationMode:      jules.AutomationMode(optionalString(in.AutomationMode)),
	}
	if !in.NoSource {
		sourceID := optionalString(in.SourceID)
		if sourceID == "" {
			return nil, nil, fmt.Errorf("source_id is required unless no_source=true")
		}
		req.SourceContext = &jules.SourceContext{
			Source: sourceID,
		}
		if startingBranch := optionalString(in.StartingBranch); startingBranch != "" {
			req.SourceContext.GithubRepoContext = &jules.GithubRepoContext{StartingBranch: startingBranch}
		}
	}
	session, err := client.Sessions().Create(ctx, req)
	return nil, session, wrapAPIError("create session", err)
}

func (p *sessionsProvider) approvePlan(ctx context.Context, _ *mcp.CallToolRequest, in confirmSessionInput) (*mcp.CallToolResult, actionOutput, error) {
	if err := requireConfirm(in.Confirm, "approve_session_plan"); err != nil {
		return nil, actionOutput{}, err
	}
	client, err := p.clientFactory()
	if err != nil {
		return nil, actionOutput{}, err
	}
	if err := client.Sessions().ApprovePlan(ctx, in.SessionID); err != nil {
		return nil, actionOutput{}, wrapAPIError("approve session plan", err)
	}
	return nil, actionOutput{OK: true, Message: "plan approved"}, nil
}

type sendMessageInput struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

func (p *sessionsProvider) sendMessage(ctx context.Context, _ *mcp.CallToolRequest, in sendMessageInput) (*mcp.CallToolResult, actionOutput, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, actionOutput{}, err
	}
	if err := client.Sessions().SendMessage(ctx, in.SessionID, &jules.SendMessageRequest{Prompt: in.Message}); err != nil {
		return nil, actionOutput{}, wrapAPIError("send session message", err)
	}
	return nil, actionOutput{OK: true, Message: "message sent"}, nil
}

func (p *sessionsProvider) deleteSession(ctx context.Context, _ *mcp.CallToolRequest, in confirmSessionInput) (*mcp.CallToolResult, actionOutput, error) {
	if err := requireConfirm(in.Confirm, "delete_session"); err != nil {
		return nil, actionOutput{}, err
	}
	client, err := p.clientFactory()
	if err != nil {
		return nil, actionOutput{}, err
	}
	if err := client.Sessions().Delete(ctx, in.SessionID); err != nil {
		return nil, actionOutput{}, wrapAPIError("delete session", err)
	}
	return nil, actionOutput{OK: true, Message: "session deleted"}, nil
}

type getPlansInput struct {
	SessionID  string `json:"session_id"`
	LatestOnly bool   `json:"latest_only,omitempty"`
}

type plansOutput struct {
	SessionID  string                      `json:"session_id"`
	Plans      []julessessions.PlanSummary `json:"plans"`
	TotalCount int                         `json:"total_count"`
}

func (p *sessionsProvider) getPlans(ctx context.Context, _ *mcp.CallToolRequest, in getPlansInput) (*mcp.CallToolResult, plansOutput, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, plansOutput{}, err
	}
	activities, err := client.Activities().ListAll(ctx, in.SessionID, 100)
	if err != nil {
		return nil, plansOutput{}, wrapAPIError("get session plans", err)
	}
	plans := julessessions.ExtractPlanSummaries(activities)
	if in.LatestOnly {
		if latest := julessessions.LatestPlanSummary(plans); latest != nil {
			plans = []julessessions.PlanSummary{*latest}
		} else {
			plans = []julessessions.PlanSummary{}
		}
	}
	return nil, plansOutput{SessionID: in.SessionID, Plans: plans, TotalCount: len(plans)}, nil
}

type reviewSessionInput struct {
	SessionID        string `json:"session_id"`
	ProjectPath      string `json:"project_path,omitempty"`
	ActivityID       string `json:"activity_id,omitempty"`
	ArtifactIndex    int    `json:"artifact_index,omitempty"`
	HasArtifactIndex bool   `json:"has_artifact_index,omitempty"`
}

func (p *sessionsProvider) reviewSession(ctx context.Context, _ *mcp.CallToolRequest, in reviewSessionInput) (*mcp.CallToolResult, *julessessions.SessionReview, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, nil, err
	}
	review, err := julessessions.BuildSessionReview(ctx, client, julessessions.ReviewRequest{
		SessionID:        in.SessionID,
		WorkingDir:       in.ProjectPath,
		ActivityID:       in.ActivityID,
		ArtifactIndex:    in.ArtifactIndex,
		HasArtifactIndex: in.HasArtifactIndex,
	})
	return nil, review, wrapAPIError("review session", err)
}

type watchSessionInput struct {
	SessionID string `json:"session_id"`
}

type watchSessionOutput struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

func (p *sessionsProvider) watchSession(ctx context.Context, req *mcp.CallToolRequest, in watchSessionInput) (*mcp.CallToolResult, watchSessionOutput, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, watchSessionOutput{}, err
	}

	progressToken := req.Params.GetProgressToken()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	progressCount := 0.0

	for {
		select {
		case <-ctx.Done():
			return nil, watchSessionOutput{}, ctx.Err()
		case <-ticker.C:
			progressCount++
			session, err := client.Sessions().Get(ctx, in.SessionID)
			if err != nil {
				return nil, watchSessionOutput{}, wrapAPIError("watch session", err)
			}

			status := string(session.State)
			msg := fmt.Sprintf("Session status: %s", status)

			if progressToken != nil && req.Session != nil {
				_ = req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
					ProgressToken: progressToken,
					Message:       msg,
					Progress:      progressCount,
				})
			}

			if status != string(jules.SessionStateInProgress) && status != string(jules.SessionStatePlanning) && status != string(jules.SessionStateQueued) {
				return nil, watchSessionOutput{
					SessionID: in.SessionID,
					Status:    status,
					Message:   msg,
				}, nil
			}
		}
	}
}

func (p *sessionsProvider) getReviewPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	sessionID, ok := req.Params.Arguments["session_id"]
	if !ok || sessionID == "" {
		return nil, fmt.Errorf("session_id argument is required")
	}

	client, err := p.clientFactory()
	if err != nil {
		return nil, err
	}
	review, err := julessessions.BuildSessionReview(ctx, client, julessessions.ReviewRequest{
		SessionID: sessionID,
	})
	if err != nil {
		return nil, wrapAPIError("review session for prompt", err)
	}

	reviewJSON, err := json.MarshalIndent(review, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to encode review: %w", err)
	}

	return &mcp.GetPromptResult{
		Description: "Review plan for session " + sessionID,
		Messages: []*mcp.PromptMessage{
			{
				Role: mcp.Role("user"),
				Content: &mcp.TextContent{
					Text: fmt.Sprintf("Please review the following plan generated by Jules for session %s.\n\n%s\n\nIf you approve, call the `approve_session_plan` tool. Otherwise, use `send_session_message` to provide feedback.", sessionID, string(reviewJSON)),
				},
			},
		},
	}, nil
}
