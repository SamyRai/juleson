package jmcp

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdioTransport(t *testing.T) {
	// Build the juleson binary
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "juleson")

	buildCmd := exec.Command("go", "build", "-o", binPath, "../../cmd/juleson")
	buildOut, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build juleson: %s", string(buildOut))

	// Start the server
	cmd := exec.Command(binPath, "mcp", "serve")

	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)

	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	err = cmd.Start()
	require.NoError(t, err)

	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	scanner := bufio.NewScanner(stdout)

	// Helper to send a request and read a response
	sendAndReceive := func(request map[string]interface{}) map[string]interface{} {
		reqBytes, err := json.Marshal(request)
		require.NoError(t, err)

		_, err = stdin.Write(append(reqBytes, '\n'))
		require.NoError(t, err)

		// Read response with timeout
		done := make(chan struct{})
		var responseBytes []byte
		var readErr error

		go func() {
			if scanner.Scan() {
				responseBytes = scanner.Bytes()
			} else {
				readErr = scanner.Err()
				if readErr == nil {
					readErr = io.EOF
				}
			}
			close(done)
		}()

		select {
		case <-done:
			require.NoError(t, readErr)
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for response")
		}

		var response map[string]interface{}
		err = json.Unmarshal(responseBytes, &response)
		require.NoError(t, err)

		return response
	}

	// 1. Send Initialize Request
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0",
			},
		},
	}

	initRes := sendAndReceive(initReq)
	assert.Equal(t, float64(1), initRes["id"])
	assert.NotNil(t, initRes["result"])

	resultMap := initRes["result"].(map[string]interface{})
	assert.Equal(t, "2024-11-05", resultMap["protocolVersion"])
	assert.NotNil(t, resultMap["serverInfo"])
	serverInfo := resultMap["serverInfo"].(map[string]interface{})
	assert.Equal(t, "juleson", serverInfo["name"])

	// 1b. Send initialized notification
	initializedReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	reqBytes, _ := json.Marshal(initializedReq)
	_, err = stdin.Write(append(reqBytes, '\n'))
	require.NoError(t, err)

	// 2. Send tools/list Request
	toolsReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	}

	toolsRes := sendAndReceive(toolsReq)
	assert.Equal(t, float64(2), toolsRes["id"])
	assert.NotNil(t, toolsRes["result"])

	toolsResult := toolsRes["result"].(map[string]interface{})
	toolsList := toolsResult["tools"].([]interface{})

	foundVersion := false
	for _, toolInterface := range toolsList {
		tool := toolInterface.(map[string]interface{})
		if tool["name"] == "version" {
			foundVersion = true
			break
		}
	}
	assert.True(t, foundVersion, "Expected 'version' tool to be registered")

	// 3. Send version tool call
	callReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "version",
		},
	}

	callRes := sendAndReceive(callReq)
	assert.Equal(t, float64(3), callRes["id"])
	assert.NotNil(t, callRes["result"])

	callResult := callRes["result"].(map[string]interface{})
	contentList := callResult["content"].([]interface{})
	assert.True(t, len(contentList) > 0)
	content := contentList[0].(map[string]interface{})

	assert.Equal(t, "text", content["type"])

	// Ensure the returned text contains standard json payload we expect
	text := content["text"].(string)
	assert.True(t, strings.Contains(text, "version"))

	// 4. Send malformed request (syntactically invalid json)
	malformedReq := []byte(`{"jsonrpc": "2.0", "method": "invalid",}`)
	_, err = stdin.Write(append(malformedReq, '\n'))
	require.NoError(t, err)

	// Wait for a JSON-RPC error response
	if scanner.Scan() {
		var errResponse map[string]interface{}
		err = json.Unmarshal(scanner.Bytes(), &errResponse)
		require.NoError(t, err)

		// Expect JSON-RPC error format
		assert.Equal(t, "2.0", errResponse["jsonrpc"])
		assert.NotNil(t, errResponse["error"])
		errObj := errResponse["error"].(map[string]interface{})
		assert.Equal(t, float64(-32700), errObj["code"]) // Parse error code
	}
}
