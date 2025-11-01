package mcp

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/mcp/tools"
	"github.com/SamyRai/juleson/internal/services"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// init configures the default logger to write to stderr to avoid corrupting stdio transport
func init() {
	log.SetOutput(os.Stderr)
}

// Server represents the MCP server using the official SDK
type Server struct {
	container *services.Container
	server    *mcp.Server
	logger    *slog.Logger
}

// NewServer creates a new MCP server using the official SDK
func NewServer(cfg *config.Config) *Server {
	log.Println("Initializing MCP server...")

	container := services.NewContainer(cfg)

	// Create MCP server with options
	opts := &mcp.ServerOptions{
		Instructions: `You are Juleson MCP Server, a powerful tool for project analysis, automation, and session management.

Available capabilities:
- Project Analysis: Analyze project structure, dependencies, and complexity
- Template Management: List, search, create, and execute automation templates
- Session Management: Monitor, approve, cancel, and delete Jules sessions
- Automation Execution: Run templates to automate project tasks

Use the available tools to help users with their automation needs. Always provide clear, actionable results.`,
	}

	server := &Server{
		container: container,
		logger:    slog.Default(),
	}

	// Create the MCP server instance
	server.server = mcp.NewServer(&mcp.Implementation{
		Name:    "github.com/SamyRai/juleson",
		Version: "1.0.0",
	}, opts)

	// Register all tools, resources, and prompts BEFORE running
	log.Println("Registering tools...")
	server.addTools()

	log.Println("Registering resources...")
	server.addResources()

	log.Println("Registering prompts...")
	server.addPrompts()

	log.Println("MCP server created successfully")
	return server
}

// Start starts the MCP server
func (s *Server) Start() error {
	log.Println("Starting Juleson MCP Server on stdio transport...")

	// Use LoggingTransport to write transport logs to stderr, keeping stdout clean for JSON-RPC
	transport := &mcp.LoggingTransport{
		Transport: &mcp.StdioTransport{},
		Writer:    os.Stderr,
	}

	log.Println("Server ready, waiting for client connection...")

	// Run the server on stdio transport with logging
	// The Run method handles connection, initialization, and serving
	return s.server.Run(context.Background(), transport)
}

// Shutdown gracefully shuts down the MCP server
func (s *Server) Shutdown(ctx context.Context) error {
	// Cleanup container resources
	return s.container.Close()
}

// addTools adds all MCP tools to the server using shared services
func (s *Server) addTools() {
	// Register developer tools (always available - no external dependencies)
	log.Println("Registering developer tools...")
	tools.RegisterDevTools(s.server)

	// Register project-related tools (lazy initialization)
	log.Println("Registering project tools...")
	tools.RegisterProjectTools(s.server, s.container)

	// Register template-related tools (lazy initialization)
	log.Println("Registering template tools...")
	tools.RegisterTemplateTools(s.server, s.container)

	// Register session-related tools (only if Jules client is available)
	julesClient := s.container.JulesClient()
	if julesClient != nil {
		log.Println("Registering session tools...")
		tools.RegisterSessionTools(s.server, julesClient)
	} else {
		s.logger.Warn("Jules client not available - session tools will not be registered")
	}

	log.Println("All tools registered successfully")
}

// addResources adds MCP resources to the server
func (s *Server) addResources() {
	// Add a resource for server documentation
	s.server.AddResource(&mcp.Resource{
		Name:        "server-info",
		MIMEType:    "text/plain",
		URI:         "jules://server/info",
		Description: "Information about the Juleson MCP Server capabilities",
	}, s.handleServerInfoResource)

	// Add a resource for configuration template
	s.server.AddResource(&mcp.Resource{
		Name:        "config-template",
		MIMEType:    "application/json",
		URI:         "jules://config/template",
		Description: "Template for Jules configuration",
	}, s.handleConfigTemplateResource)
}

// addPrompts adds MCP prompts to the server
func (s *Server) addPrompts() {
	// Add a prompt for project analysis workflow
	s.server.AddPrompt(&mcp.Prompt{
		Name:        "analyze-project-workflow",
		Description: "Complete workflow for analyzing a project and getting automation recommendations",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "project_path",
				Description: "Path to the project directory to analyze",
				Required:    true,
			},
		},
	}, s.handleAnalyzeProjectWorkflowPrompt)

	// Add a prompt for session management
	s.server.AddPrompt(&mcp.Prompt{
		Name:        "session-management-guide",
		Description: "Guide for managing Jules automation sessions",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "action",
				Description: "Action to perform: list, status, approve, cancel, delete",
				Required:    false,
			},
		},
	}, s.handleSessionManagementGuidePrompt)
}

// handleServerInfoResource provides server information
func (s *Server) handleServerInfoResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	info := `Juleson MCP Server

Capabilities:
- Project Analysis: Analyze project structure, languages, frameworks, and complexity
- Template Management: Create, list, search, and execute automation templates
- Session Management: Monitor and control Jules automation sessions
- Automation Execution: Run templates to automate development tasks

Tools Available:
- analyze_project: Analyze project structure and dependencies
- sync_project: Sync project with remote repository
- list_templates: List available automation templates
- search_templates: Search templates by keywords
- create_template: Create new custom templates
- execute_template: Execute templates on projects
- list_sessions: List all Jules sessions
- get_session_status: Get detailed session status summary
- approve_session_plan: Approve session plans for execution
- cancel_session: Cancel running sessions
- delete_session: Delete completed sessions

For more information, use the available tools or check the Jules documentation.`

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "text/plain",
				Text:     info,
			},
		},
	}, nil
}

// handleCompletion provides completion suggestions for tool arguments
func (s *Server) handleCompletion(ctx context.Context, req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	switch req.Params.Ref.Type {
	case "ref/tool":
		return s.handleToolCompletion(req)
	case "ref/resource":
		return s.handleResourceCompletion(req)
	default:
		return &mcp.CompleteResult{
			Completion: mcp.CompletionResultDetails{
				Values: []string{},
				Total:  0,
			},
		}, nil
	}
}

// handleToolCompletion provides completions for tool arguments
func (s *Server) handleToolCompletion(req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	toolName := req.Params.Ref.Name
	argName := req.Params.Argument.Name
	argValue := req.Params.Argument.Value

	var suggestions []string

	switch toolName {
	case "analyze_project":
		if argName == "project_path" {
			// Suggest common project paths
			suggestions = []string{"./", "../", "./src", "./app", "./lib"}
		}
	case "execute_template":
		if argName == "template_name" {
			// Suggest template names based on available templates (lazy initialization)
			if templateManager, err := s.container.TemplateManager(); err == nil {
				templates := templateManager.ListTemplates()
				for _, tmpl := range templates {
					if len(argValue) == 0 || strings.Contains(strings.ToLower(tmpl.Name), strings.ToLower(argValue)) {
						suggestions = append(suggestions, tmpl.Name)
					}
				}
			}
		}
	case "list_sessions", "get_session_status":
		if argName == "limit" {
			suggestions = []string{"10", "50", "100"}
		}
	}

	return &mcp.CompleteResult{
		Completion: mcp.CompletionResultDetails{
			Values: suggestions,
			Total:  len(suggestions),
		},
	}, nil
}

// handleResourceCompletion provides completions for resource URIs
func (s *Server) handleResourceCompletion(req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	uri := req.Params.Argument.Value

	var suggestions []string

	// Suggest common resource URIs
	if strings.HasPrefix(uri, "jules://") {
		suggestions = []string{
			"jules://server/info",
			"jules://config/template",
		}
	}

	return &mcp.CompleteResult{
		Completion: mcp.CompletionResultDetails{
			Values: suggestions,
			Total:  len(suggestions),
		},
	}, nil
}

// handleConfigTemplateResource provides a configuration template
func (s *Server) handleConfigTemplateResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	template := `{
  "jules": {
    "api_key": "your-api-key-here",
    "base_url": "https://api.jules.ai",
    "timeout": 30,
    "retry_attempts": 3
  },
  "templates": {
    "builtin_path": "./templates/builtin",
    "custom_path": "./templates/custom"
  },
  "logging": {
    "level": "info",
    "file": "jules.log"
  }
}`

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     template,
			},
		},
	}, nil
}

// handleAnalyzeProjectWorkflowPrompt provides a complete project analysis workflow
func (s *Server) handleAnalyzeProjectWorkflowPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	projectPath := req.Params.Arguments["project_path"]

	return &mcp.GetPromptResult{
		Description: "Complete project analysis and automation workflow",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Please analyze the project at path: %s

Follow this workflow:
1. First, analyze the project structure using the analyze_project tool
2. Based on the analysis, suggest appropriate automation templates
3. If templates are available, execute the most relevant one
4. Monitor the execution and provide feedback

Project path: %s`, projectPath, projectPath)},
			},
		},
	}, nil
}

// handleSessionManagementGuidePrompt provides session management guidance
func (s *Server) handleSessionManagementGuidePrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	action := req.Params.Arguments["action"]

	var content string
	if action == "" {
		content = `Jules Session Management Guide:

Available actions:
- list: Get all current sessions
- status: Get detailed status summary
- approve: Approve a session plan for execution
- cancel: Cancel a running session
- delete: Delete a completed session

Start by listing sessions to see what's available, then use status for detailed information.`
	} else {
		switch action {
		case "list":
			content = "Use the list_sessions tool to see all current Jules sessions with their status and basic information."
		case "status":
			content = "Use the get_session_status tool to get a comprehensive overview of all sessions, including counts by state and recent activity."
		case "approve":
			content = "Use the approve_session_plan tool with a session_id to approve a planned session for execution."
		case "cancel":
			content = "Use the cancel_session tool with a session_id to stop a currently running session."
		case "delete":
			content = "Use the delete_session tool with a session_id to remove a completed or failed session from the system."
		default:
			content = fmt.Sprintf("Unknown action: %s. Valid actions are: list, status, approve, cancel, delete.", action)
		}
	}

	return &mcp.GetPromptResult{
		Description: "Jules session management guidance",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}
