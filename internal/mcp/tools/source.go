package tools

import (
	"context"
	"fmt"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterSourceTools registers Jules source discovery tools.
func RegisterSourceTools(server *mcp.Server, julesClient *jules.Client) {
	if julesClient == nil {
		return
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sources",
		Description: "List connected Jules sources such as GitHub repositories; use before create_session when the source name is unknown.",
		Annotations: readOnlyOpenWorldTool("List Jules Sources"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListSourcesInput) (*mcp.CallToolResult, ListSourcesOutput, error) {
		return listSourcesMCP(ctx, input, julesClient)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_source",
		Description: "Get details for one Jules source by ID or full sources/... name.",
		Annotations: readOnlyOpenWorldTool("Get Jules Source"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetSourceInput) (*mcp.CallToolResult, GetSourceOutput, error) {
		return getSourceMCP(ctx, input, julesClient)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "infer_source_for_project",
		Description: "Infer a connected Jules source from a local project's git origin remote.",
		Annotations: readOnlyOpenWorldTool("Infer Jules Source"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input InferSourceInput) (*mcp.CallToolResult, InferSourceOutput, error) {
		return inferSourceMCP(ctx, input, julesClient)
	})
}

type ListSourcesInput struct {
	PageSize  int    `json:"page_size,omitempty" jsonschema:"Number of sources to return (default: 50, max: 100)"`
	PageToken string `json:"page_token,omitempty" jsonschema:"Cursor for pagination"`
	Filter    string `json:"filter,omitempty" jsonschema:"Optional AIP-160 source filter"`
}

type ListSourcesOutput struct {
	Sources       []jules.Source `json:"sources"`
	NextPageToken string         `json:"next_page_token,omitempty"`
	TotalCount    int            `json:"total_count"`
}

type GetSourceInput struct {
	SourceID string `json:"source_id" jsonschema:"Source ID or full source name, e.g. github/owner/repo or sources/github/owner/repo"`
}

type GetSourceOutput struct {
	SourceID string       `json:"source_id"`
	Source   jules.Source `json:"source"`
}

type InferSourceInput struct {
	ProjectPath string `json:"project_path,omitempty" jsonschema:"Local git project path (default: current directory)"`
}

type InferSourceOutput struct {
	ProjectPath string       `json:"project_path"`
	SourceID    string       `json:"source_id"`
	Source      jules.Source `json:"source"`
}

func listSourcesMCP(ctx context.Context, input ListSourcesInput, client *jules.Client) (*mcp.CallToolResult, ListSourcesOutput, error) {
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	response, err := client.Sources().List(ctx, &jules.ListSourcesOptions{
		PageSize:  pageSize,
		PageToken: input.PageToken,
		Filter:    input.Filter,
	})
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list sources: %v", err)},
			},
		}, ListSourcesOutput{}, err
	}

	sources := response.Sources
	if sources == nil {
		sources = []jules.Source{}
	}
	return nil, ListSourcesOutput{
		Sources:       sources,
		NextPageToken: response.NextPageToken,
		TotalCount:    len(sources),
	}, nil
}

func getSourceMCP(ctx context.Context, input GetSourceInput, client *jules.Client) (*mcp.CallToolResult, GetSourceOutput, error) {
	source, err := client.Sources().Get(ctx, input.SourceID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get source: %v", err)},
			},
		}, GetSourceOutput{}, err
	}

	return nil, GetSourceOutput{
		SourceID: input.SourceID,
		Source:   *source,
	}, nil
}

func inferSourceMCP(ctx context.Context, input InferSourceInput, client *jules.Client) (*mcp.CallToolResult, InferSourceOutput, error) {
	projectPath := input.ProjectPath
	if projectPath == "" {
		projectPath = "."
	}
	source, err := julesops.InferSourceFromGitRemote(ctx, client, projectPath)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to infer source: %v", err)},
			},
		}, InferSourceOutput{}, err
	}
	return nil, InferSourceOutput{
		ProjectPath: projectPath,
		SourceID:    source.Name,
		Source:      *source,
	}, nil
}
