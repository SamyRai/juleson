package jmcp

import (
	"context"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type coreProvider struct {
	cfg *config.Config
}

// NewCoreProvider creates a ToolProvider for core system information.
func NewCoreProvider(cfg *config.Config) ToolProvider {
	return &coreProvider{cfg: cfg}
}

func (p *coreProvider) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "version",
		Description: "Return Juleson build and runtime version information.",
	}, p.version)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "config_status",
		Description: "Report whether Jules and GitHub credentials are configured without exposing secret values.",
	}, p.configStatus)
}

func (p *coreProvider) version(context.Context, *mcp.CallToolRequest, emptyInput) (*mcp.CallToolResult, core.VersionInfo, error) {
	return nil, core.GetVersionInfo(), nil
}

type configStatusOutput struct {
	JulesBaseURL          string `json:"jules_base_url"`
	JulesAPIKeyConfigured bool   `json:"jules_api_key_configured"`
	GitHubTokenConfigured bool   `json:"github_token_configured"`
}

func (p *coreProvider) configStatus(context.Context, *mcp.CallToolRequest, emptyInput) (*mcp.CallToolResult, configStatusOutput, error) {
	return nil, configStatusOutput{
		JulesAPIKeyConfigured: p.cfg.Jules.APIKey != "",
		JulesBaseURL:          p.cfg.Jules.BaseURL,
		GitHubTokenConfigured: p.cfg.GitHub.Token != "",
	}, nil
}
