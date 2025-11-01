package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerStdioIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create minimal config
	cfg := &config.Config{
		Jules: config.JulesConfig{
			APIKey:        "",
			BaseURL:       "https://jules.googleapis.com/v1alpha",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
		},
		MCP: config.MCPConfig{
			Server: config.MCPServerConfig{
				Port: 8080,
				Host: "localhost",
			},
			Client: config.MCPClientConfig{
				Timeout: 10 * time.Second,
			},
		},
		Automation: config.AutomationConfig{
			Strategies:         []string{"modular", "layered", "microservices"},
			MaxConcurrentTasks: 5,
			TaskTimeout:        300 * time.Second,
		},
		Projects: config.ProjectsConfig{
			DefaultPath:    "./projects",
			BackupEnabled:  true,
			GitIntegration: true,
		},
		Templates: config.TemplatesConfig{
			BuiltinPath:  "../../templates/builtin",
			CustomPath:   "../../templates/custom",
			EnableCustom: true,
		},
	}

	// Create server
	server := NewServer(cfg)
	assert.NotNil(t, server)

	// Create in-memory transports for testing
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Start server in background
	serverDone := make(chan error, 1)
	go func() {
		// The server is already configured with tools, resources, and prompts
		// Just connect it to the transport
		session, err := server.server.Connect(ctx, serverTransport, nil)
		if err != nil {
			serverDone <- err
			return
		}
		serverDone <- session.Wait()
	}()

	// Create client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	// Connect client
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err, "Client should connect successfully")
	defer clientSession.Close()

	// List tools to verify server is responding
	toolsResult, err := clientSession.ListTools(ctx, nil)
	require.NoError(t, err, "Should list tools successfully")
	assert.NotEmpty(t, toolsResult.Tools, "Should have registered tools")

	// Check that developer tools are registered
	hasAnalyzeProject := false
	for _, tool := range toolsResult.Tools {
		if tool.Name == "analyze_project" {
			hasAnalyzeProject = true
			break
		}
	}
	assert.True(t, hasAnalyzeProject, "Should have analyze_project tool")

	// Close client and wait for server
	clientSession.Close()

	select {
	case err := <-serverDone:
		assert.NoError(t, err, "Server should exit cleanly")
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not exit within timeout")
	}
}

func TestServerStdioTransport(t *testing.T) {
	// This test verifies the server can be created and configured for stdio
	cfg := &config.Config{
		Jules: config.JulesConfig{
			APIKey:        "",
			BaseURL:       "https://jules.googleapis.com/v1alpha",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
		},
		MCP: config.MCPConfig{
			Server: config.MCPServerConfig{
				Port: 8080,
				Host: "localhost",
			},
		},
		Automation: config.AutomationConfig{
			Strategies:         []string{"modular"},
			MaxConcurrentTasks: 5,
			TaskTimeout:        300 * time.Second,
		},
		Projects: config.ProjectsConfig{
			DefaultPath:    "./projects",
			BackupEnabled:  true,
			GitIntegration: true,
		},
		Templates: config.TemplatesConfig{
			BuiltinPath:  "../../templates/builtin",
			CustomPath:   "../../templates/custom",
			EnableCustom: true,
		},
	}

	server := NewServer(cfg)
	require.NotNil(t, server)

	// Verify server has been initialized with tools
	assert.NotNil(t, server.server, "Server should have MCP server instance")
}
