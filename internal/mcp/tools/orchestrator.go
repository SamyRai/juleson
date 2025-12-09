package tools

import (
	"time"

	"github.com/SamyRai/juleson/internal/orchestrator"
)

// GetOrchestrator returns a shared orchestrator instance for MCP tools
// This can be used by other MCP tool handlers to delegate to the orchestrator service
func GetOrchestrator() orchestrator.Orchestrator {
	config := orchestrator.DefaultConfig("dev", time.Now().Format("2006-01-02"), "mcp-tools")
	return orchestrator.NewService(config)
}
