package tools

import (
	"context"

	"github.com/SamyRai/juleson/internal/services"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterGeminiTools registers all Gemini AI-related MCP tools
func RegisterGeminiTools(server *mcp.Server, container *services.Container) {
	// Project Automation Planning Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "plan_project_automation",
		Description: "Analyze project structure and create comprehensive automation plans using Gemini AI",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input PlanProjectAutomationInput) (*mcp.CallToolResult, PlanProjectAutomationOutput, error) {
		return planProjectAutomation(ctx, req, input, container)
	})

	// Workflow Orchestration Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "orchestrate_workflow",
		Description: "Execute complex multi-step automation workflows based on project analysis",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input OrchestrateWorkflowInput) (*mcp.CallToolResult, OrchestrateWorkflowOutput, error) {
		return orchestrateWorkflow(ctx, req, input, container)
	})

	// GitHub Project Management Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "manage_github_project",
		Description: "Manage GitHub issues, milestones, and projects through natural language commands",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ManageGitHubProjectInput) (*mcp.CallToolResult, ManageGitHubProjectOutput, error) {
		return manageGitHubProject(ctx, req, input, container)
	})

	// Session Results Synthesis Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "synthesize_session_results",
		Description: "Analyze Jules session results and provide actionable insights and recommendations",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SynthesizeSessionResultsInput) (*mcp.CallToolResult, SynthesizeSessionResultsOutput, error) {
		return synthesizeSessionResults(ctx, req, input, container)
	})
}
