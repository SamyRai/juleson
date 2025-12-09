package tools

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/jules"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterActivityTools registers all activity-related MCP tools
func RegisterActivityTools(server *mcp.Server, julesClient *jules.Client) {
	if julesClient == nil {
		return
	}

	// List Activities Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_session_activities",
		Description: "List all activities within a Jules session including messages, plans, and progress updates",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListActivitiesInput) (*mcp.CallToolResult, ListActivitiesOutput, error) {
		return listActivities(ctx, req, input, julesClient)
	})

	// Get Activity Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session_activity",
		Description: "Get detailed information about a specific activity within a session",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetActivityInput) (*mcp.CallToolResult, GetActivityOutput, error) {
		return getActivity(ctx, req, input, julesClient)
	})

	// Search Activities Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_session_activities",
		Description: "Search activities within a session by query text",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchActivitiesInput) (*mcp.CallToolResult, SearchActivitiesOutput, error) {
		return searchActivities(ctx, req, input, julesClient)
	})

	// Get Activities with Plans Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session_plans",
		Description: "Get all activities that generated plans in a session",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetActivitiesWithPlansInput) (*mcp.CallToolResult, GetActivitiesWithPlansOutput, error) {
		return getActivitiesWithPlans(ctx, req, input, julesClient)
	})
}

// ListActivitiesInput represents input for list_session_activities tool
type ListActivitiesInput struct {
	SessionID    string `json:"session_id" jsonschema:"ID of the session to list activities from"`
	PageSize     int    `json:"page_size,omitempty" jsonschema:"Number of activities per page (default: 50, max: 100)"`
	PageToken    string `json:"page_token,omitempty" jsonschema:"Token for pagination"`
	Type         string `json:"type,omitempty" jsonschema:"Filter by activity type (e.g., 'message', 'plan', 'execution')"`
	HasPlan      *bool  `json:"has_plan,omitempty" jsonschema:"Filter activities that have/don't have plans"`
	HasArtifacts *bool  `json:"has_artifacts,omitempty" jsonschema:"Filter activities that have/don't have artifacts"`
}

// ListActivitiesOutput represents output for list_session_activities tool
type ListActivitiesOutput struct {
	SessionID     string           `json:"session_id"`
	Activities    []jules.Activity `json:"activities"`
	TotalCount    int              `json:"total_count"`
	NextPageToken string           `json:"next_page_token,omitempty"`
}

func listActivities(ctx context.Context, req *mcp.CallToolRequest, input ListActivitiesInput, client *jules.Client) (
	*mcp.CallToolResult,
	ListActivitiesOutput,
	error,
) {
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	var activities []jules.Activity
	var err error
	var nextToken string

	// Use filtered list if filters are provided
	if input.Type != "" || input.HasPlan != nil || input.HasArtifacts != nil {
		filter := &jules.ActivityFilter{
			Type:         input.Type,
			HasPlan:      input.HasPlan,
			HasArtifacts: input.HasArtifacts,
		}
		activities, err = client.ListActivitiesFiltered(ctx, input.SessionID, filter)
	} else {
		response, err := client.ListActivitiesWithPagination(ctx, input.SessionID, pageSize, input.PageToken)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to list activities: %v", err)},
				},
			}, ListActivitiesOutput{}, err
		}
		activities = response.Activities
		nextToken = response.NextPageToken
	}

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list activities: %v", err)},
			},
		}, ListActivitiesOutput{}, err
	}

	if activities == nil {
		activities = []jules.Activity{}
	}

	output := ListActivitiesOutput{
		SessionID:     input.SessionID,
		Activities:    activities,
		TotalCount:    len(activities),
		NextPageToken: nextToken,
	}

	return nil, output, nil
}

// GetActivityInput represents input for get_session_activity tool
type GetActivityInput struct {
	SessionID  string `json:"session_id" jsonschema:"ID of the session"`
	ActivityID string `json:"activity_id" jsonschema:"ID of the activity to retrieve"`
}

// GetActivityOutput represents output for get_session_activity tool
type GetActivityOutput struct {
	SessionID  string         `json:"session_id"`
	ActivityID string         `json:"activity_id"`
	Activity   jules.Activity `json:"activity"`
}

func getActivity(ctx context.Context, req *mcp.CallToolRequest, input GetActivityInput, client *jules.Client) (
	*mcp.CallToolResult,
	GetActivityOutput,
	error,
) {
	activity, err := client.GetActivity(ctx, input.SessionID, input.ActivityID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get activity: %v", err)},
			},
		}, GetActivityOutput{}, err
	}

	output := GetActivityOutput{
		SessionID:  input.SessionID,
		ActivityID: input.ActivityID,
		Activity:   *activity,
	}

	return nil, output, nil
}

// SearchActivitiesInput represents input for search_session_activities tool
type SearchActivitiesInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to search activities in"`
	Query     string `json:"query" jsonschema:"Search query text"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of results (default: 20)"`
	Type      string `json:"type,omitempty" jsonschema:"Filter by activity type"`
}

// SearchActivitiesOutput represents output for search_session_activities tool
type SearchActivitiesOutput struct {
	SessionID  string           `json:"session_id"`
	Query      string           `json:"query"`
	Activities []jules.Activity `json:"activities"`
	TotalCount int              `json:"total_count"`
}

func searchActivities(ctx context.Context, req *mcp.CallToolRequest, input SearchActivitiesInput, client *jules.Client) (
	*mcp.CallToolResult,
	SearchActivitiesOutput,
	error,
) {
	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}

	options := &jules.ActivitySearchOptions{
		Query: input.Query,
		Limit: limit,
	}

	if input.Type != "" {
		options.Filter = &jules.ActivityFilter{
			Type: input.Type,
		}
	}

	activities, err := client.SearchActivities(ctx, input.SessionID, options)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to search activities: %v", err)},
			},
		}, SearchActivitiesOutput{}, err
	}

	if activities == nil {
		activities = []jules.Activity{}
	}

	output := SearchActivitiesOutput{
		SessionID:  input.SessionID,
		Query:      input.Query,
		Activities: activities,
		TotalCount: len(activities),
	}

	return nil, output, nil
}

// GetActivitiesWithPlansInput represents input for get_session_plans tool
type GetActivitiesWithPlansInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to get plans from"`
}

// GetActivitiesWithPlansOutput represents output for get_session_plans tool
type GetActivitiesWithPlansOutput struct {
	SessionID  string           `json:"session_id"`
	Activities []jules.Activity `json:"activities"`
	TotalCount int              `json:"total_count"`
}

func getActivitiesWithPlans(ctx context.Context, req *mcp.CallToolRequest, input GetActivitiesWithPlansInput, client *jules.Client) (
	*mcp.CallToolResult,
	GetActivitiesWithPlansOutput,
	error,
) {
	activities, err := client.GetActivitiesWithPlans(ctx, input.SessionID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get activities with plans: %v", err)},
			},
		}, GetActivitiesWithPlansOutput{}, err
	}

	if activities == nil {
		activities = []jules.Activity{}
	}

	output := GetActivitiesWithPlansOutput{
		SessionID:  input.SessionID,
		Activities: activities,
		TotalCount: len(activities),
	}

	return nil, output, nil
}
