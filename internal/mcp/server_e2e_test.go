package mcp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerE2EWithCommandTransport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping command-transport E2E test in short mode")
	}
	if os.Getenv("SKIP_E2E") != "" {
		t.Skip("E2E tests disabled by SKIP_E2E")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	binaryPath := resolveMCPBinary(t, ctx)

	// Create a command transport like VSCode would
	cmd := exec.Command(binaryPath)
	cmd.Dir = repoRoot(t)
	cmd.Env = append(os.Environ(),
		"HOME="+t.TempDir(),
		"JULES_API_KEY=",
	)
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

func resolveMCPBinary(t *testing.T, ctx context.Context) string {
	t.Helper()

	if binaryPath := os.Getenv("JULESON_MCP_BINARY"); binaryPath != "" {
		// #nosec G304,G703 -- tests intentionally accept an explicit local binary path.
		if _, err := os.Stat(binaryPath); err != nil {
			t.Fatalf("JULESON_MCP_BINARY %q is not usable: %v", binaryPath, err)
		}
		return binaryPath
	}

	root := repoRoot(t)
	for _, path := range []string{
		filepath.Join(root, "bin", "jules-mcp"),
		filepath.Join(root, "jules-mcp"),
	} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	binaryPath := filepath.Join(t.TempDir(), "jules-mcp")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, "./cmd/jules-mcp")
	buildCmd.Dir = root
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build jules-mcp E2E binary: %v\n%s", err, output)
	}
	return binaryPath
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root")
		}
		dir = parent
	}
}
