package jmcp

import (
	"context"
	"fmt"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
	"github.com/SamyRai/juleson/pkg/builder"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const ServerName = "juleson"

type ServerOptions struct {
	Config *config.Config
}

func NewServer(options ServerOptions) (*mcp.Server, error) {
	if options.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    ServerName,
		Version: core.Version,
	}, nil)

	cf := func() (*jules.Client, error) {
		if options.Config.Jules.APIKey == "" {
			return nil, fmt.Errorf("jules API key is not configured; set jules.api_key or JULES_API_KEY")
		}
		return core.NewJulesClient(options.Config), nil
	}

	devSvc := builder.NewService(builder.DefaultConfig("dev", "", ""))

	providers := []ToolProvider{
		NewCoreProvider(options.Config),
		NewSessionsProvider(cf),
		NewSourcesProvider(cf),
		NewArtifactsProvider(cf),
		NewDevProvider(devSvc),
	}

	for _, p := range providers {
		p.Register(server)
	}

	return server, nil
}

func RunStdio(ctx context.Context, cfg *config.Config) error {
	server, err := NewServer(ServerOptions{Config: cfg})
	if err != nil {
		return err
	}
	return server.Run(ctx, &mcp.StdioTransport{})
}
