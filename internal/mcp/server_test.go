package jmcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServerRegistersCoreToolsAndVersion(t *testing.T) {
	oldVersion := core.Version
	core.Version = "test-version"
	t.Cleanup(func() {
		core.Version = oldVersion
	})

	server, err := NewServer(ServerOptions{Config: &config.Config{}})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = clientSession.Close() }()

	tools := map[string]bool{}
	for tool, err := range clientSession.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("list tools: %v", err)
		}
		tools[tool.Name] = true
	}
	for _, name := range []string{"version", "list_sources", "get_session_plans", "review_session", "dev_build"} {
		if !tools[name] {
			t.Fatalf("expected tool %q to be registered; got %#v", name, tools)
		}
	}

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "version"})
	if err != nil {
		t.Fatalf("call version: %v", err)
	}
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	var output core.VersionInfo
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatalf("decode version output: %v", err)
	}
	if output.Version != "test-version" {
		t.Fatalf("version = %q, want test-version", output.Version)
	}
}
