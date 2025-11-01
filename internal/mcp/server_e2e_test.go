package mcp

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerE2EWithCommandTransport(t *testing.T) {
	// Use the binary from go/bin
	binaryPath := "/Users/damirmukimov/go/bin/juleson-mcp"
	if _, err := os.Stat(binaryPath); err != nil {
		t.Skip("juleson-mcp binary not found, run 'make build-mcp && make install-mcp' first")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a command transport like VSCode would
	cmd := exec.Command(binaryPath)
	cmd.Dir = "../../" // Run from project root
	transport := &mcp.CommandTransport{Command: cmd}

	// Create client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	// Connect to server
	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err, "Should connect to server successfully")
	defer session.Close()

	t.Log("Connected to server successfully")

	// Test 1: List tools
	t.Run("ListTools", func(t *testing.T) {
		toolsResult, err := session.ListTools(ctx, nil)
		require.NoError(t, err, "Should list tools")
		assert.NotEmpty(t, toolsResult.Tools, "Should have tools")

		// Check for expected tools
		toolNames := make(map[string]bool)
		for _, tool := range toolsResult.Tools {
			toolNames[tool.Name] = true
			t.Logf("Found tool: %s - %s", tool.Name, tool.Description)
		}

		assert.True(t, toolNames["analyze_project"], "Should have analyze_project tool")
	})

	// Test 2: List prompts
	t.Run("ListPrompts", func(t *testing.T) {
		promptsResult, err := session.ListPrompts(ctx, nil)
		require.NoError(t, err, "Should list prompts")

		for _, prompt := range promptsResult.Prompts {
			t.Logf("Found prompt: %s - %s", prompt.Name, prompt.Description)
		}
	})

	// Test 3: List resources
	t.Run("ListResources", func(t *testing.T) {
		resourcesResult, err := session.ListResources(ctx, nil)
		require.NoError(t, err, "Should list resources")

		for _, resource := range resourcesResult.Resources {
			t.Logf("Found resource: %s - %s", resource.Name, resource.Description)
		}
	})
}
