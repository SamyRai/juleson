package jmcp

import (
	"fmt"

	"github.com/SamyRai/go-jules"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolProvider defines an interface for registering MCP tools to the server.
type ToolProvider interface {
	Register(server *mcp.Server)
}

// clientFactory is a function type that returns a Jules client or an error if not configured.
type clientFactory func() (*jules.Client, error)

// requireConfirm is a helper to ensure dangerous actions are confirmed.
func requireConfirm(confirm bool, action string) error {
	if !confirm {
		return fmt.Errorf("%s requires confirm=true", action)
	}
	return nil
}

// wrapAPIError formats Jules API errors consistently.
func wrapAPIError(action string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("Jules API error during %s: %w", action, err)
}

// optionalString is a helper for safely dereferencing string pointers.
func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

type emptyInput struct{}

type sessionIDInput struct {
	SessionID string `json:"session_id" jsonschema:"Jules session ID"`
}

type confirmSessionInput struct {
	SessionID string `json:"session_id"`
	Confirm   bool   `json:"confirm"`
}

type actionOutput struct {
	Message string `json:"message"`
	OK      bool   `json:"ok"`
}
