package jmcp

import (
	"context"

	"github.com/SamyRai/go-jules"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type sourcesProvider struct {
	clientFactory clientFactory
}

// NewSourcesProvider creates a ToolProvider for source commands.
func NewSourcesProvider(cf clientFactory) ToolProvider {
	return &sourcesProvider{clientFactory: cf}
}

func (p *sourcesProvider) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sources",
		Description: "List connected Jules repository sources.",
	}, p.listSources)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_source",
		Description: "Get one connected Jules source by ID or resource name.",
	}, p.getSource)
}

type listSourcesInput struct {
	Filter   string `json:"filter,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

func (p *sourcesProvider) listSources(ctx context.Context, _ *mcp.CallToolRequest, in listSourcesInput) (*mcp.CallToolResult, *jules.SourcesResponse, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, nil, err
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}
	response, err := client.Sources().List(ctx, &jules.ListSourcesOptions{PageSize: pageSize, Filter: in.Filter})
	return nil, response, wrapAPIError("list sources", err)
}

type getSourceInput struct {
	SourceID string `json:"source_id" jsonschema:"Jules source ID or resource name"`
}

func (p *sourcesProvider) getSource(ctx context.Context, _ *mcp.CallToolRequest, in getSourceInput) (*mcp.CallToolResult, *jules.Source, error) {
	client, err := p.clientFactory()
	if err != nil {
		return nil, nil, err
	}
	source, err := client.Sources().Get(ctx, in.SourceID)
	return nil, source, wrapAPIError("get source", err)
}
