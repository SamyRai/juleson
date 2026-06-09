package jmcp

import (
	"context"

	"github.com/SamyRai/go-jules"
	julessessions "github.com/SamyRai/juleson/internal/jules/sessions"
	"github.com/SamyRai/juleson/internal/jules/workspace"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type artifactsProvider struct {
	clientFactory clientFactory
}

// NewArtifactsProvider creates a ToolProvider for artifacts and activities.
func NewArtifactsProvider(cf clientFactory) ToolProvider {
	return &artifactsProvider{clientFactory: cf}
}

func (p *artifactsProvider) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_activities",
		Description: "List Jules activities for a session.",
	}, p.listActivities)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_activity",
		Description: "Get one Jules activity from a session.",
	}, p.getActivity)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_session_artifacts",
		Description: "Return documented artifact manifests for a Jules session.",
	}, p.listArtifacts)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session_outputs",
		Description: "Return documented Jules session outputs such as pull request links.",
	}, p.getOutputs)
}

type listActivitiesInput struct {
	SessionID string `json:"session_id"`
	PageSize  int    `json:"page_size,omitempty"`
}

func (p *artifactsProvider) listActivities(ctx context.Context, _ *mcp.CallToolRequest, in listActivitiesInput) (*mcp.CallToolResult, *jules.ActivitiesResponse, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, nil, err
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}
	response, err := client.Activities().List(ctx, in.SessionID, &jules.ListActivitiesOptions{PageSize: pageSize})
	return nil, response, wrapAPIError("list activities", err)
}

type getActivityInput struct {
	SessionID  string `json:"session_id"`
	ActivityID string `json:"activity_id"`
}

func (p *artifactsProvider) getActivity(ctx context.Context, _ *mcp.CallToolRequest, in getActivityInput) (*mcp.CallToolResult, *jules.Activity, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, nil, err
	}
	activity, err := client.Activities().Get(ctx, in.SessionID, in.ActivityID)
	return nil, activity, wrapAPIError("get activity", err)
}

type artifactsOutput struct {
	SessionID string                       `json:"session_id"`
	Artifacts []workspace.ArtifactManifest `json:"artifacts"`
}

func (p *artifactsProvider) listArtifacts(ctx context.Context, _ *mcp.CallToolRequest, in sessionIDInput) (*mcp.CallToolResult, artifactsOutput, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, artifactsOutput{}, err
	}
	manifests, err := workspace.ListSessionArtifactManifests(ctx, client, in.SessionID)
	return nil, artifactsOutput{SessionID: in.SessionID, Artifacts: manifests}, wrapAPIError("list session artifacts", err)
}

type outputsOutput struct {
	SessionID string         `json:"session_id"`
	Outputs   []jules.Output `json:"outputs"`
}

func (p *artifactsProvider) getOutputs(ctx context.Context, _ *mcp.CallToolRequest, in sessionIDInput) (*mcp.CallToolResult, outputsOutput, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, outputsOutput{}, err
	}
	session, err := client.Sessions().Get(ctx, in.SessionID)
	if err != nil {
		return nil, outputsOutput{}, wrapAPIError("get session outputs", err)
	}
	return nil, outputsOutput{SessionID: in.SessionID, Outputs: julessessions.DocumentedOutputs(session)}, nil
}
