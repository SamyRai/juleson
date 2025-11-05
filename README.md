# Juleson

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![MCP Protocol](https://img.shields.io/badge/MCP-2024--11--05-blue)](https://modelcontextprotocol.io/)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/SamyRai/Juleson)
> before publishing.

A production-ready automation toolkit that integrates with Google's Jules AI coding agent through
both CLI and MCP (Model Context Protocol) interfaces. Built with the official MCP Go SDK, it
enables seamless AI-assisted project analysis, refactoring, testing, and documentation.

## 🎯 **Overview**

Juleson bridges your development workflow with Google's Jules AI agent, providing:

- **🔍 Project Analysis**: Deep codebase inspection with language, framework, and architecture detection
- **� Code Intelligence**: Advanced code analysis with call graphs, symbol references, and complexity metrics
- **�🤖 AI-Powered Automation**: Execute complex refactoring and restructuring tasks via Jules AI
- **🎯 High-Level Orchestration**: AI-driven project planning and workflow orchestration
- **📋 Template System**: 12+ built-in templates for reorganization, testing, refactoring, and documentation
- **💬 Session Management**: Full control over Jules coding sessions (list, monitor, approve, cancel)
- **🔌 MCP Integration**: Native Model Context Protocol server for AI assistants (Claude, Cursor, etc.)
- **🐙 GitHub Integration**: Repository management, PR workflows, and git-aware session creation
- **⚙️ CLI Tools**: Comprehensive command-line interface for all automation tasks
- **📡 Event System**: Real-time event-driven architecture with persistence and monitoring

## 🏗️ **Architecture**

```text
Juleson/
├── cmd/                          # Application entry points
│   ├── juleson/               # CLI tool for direct usage
│   └── juleson-mcp/               # MCP server for AI assistants
├── internal/
│   ├── jules/                   # Jules API client with full session support
│   │   ├── client.go           # HTTP client & retry logic
│   │   ├── sessions.go         # Session management (CRUD)
│   │   ├── activities.go       # Activity monitoring
│   │   ├── artifacts.go        # Artifact handling
│   │   └── monitor.go          # Real-time session monitoring
│   ├── events/                  # Event-driven architecture
│   │   ├── bus.go             # Pub/sub event bus with middleware
│   │   ├── queue.go           # Asynchronous message queues
│   │   ├── store.go           # Event persistence and replay
│   │   ├── coordinator.go     # Unified event system interface
│   │   ├── circuit_breaker.go # Fault tolerance for external services
│   │   ├── middleware.go      # Event processing middleware
│   │   └── types.go           # Event definitions and data structures
│   ├── mcp/                     # MCP server implementation
│   │   ├── server.go           # Official SDK integration
│   │   └── tools/              # MCP tool implementations
│   │       ├── project.go      # Project analysis tools
│   │       ├── template.go     # Template management tools
│   │       └── session.go      # Session control tools
│   ├── automation/              # Automation engine
│   │   └── engine.go           # Task execution & orchestration
│   ├── templates/               # Template management
│   │   └── manager.go          # Template CRUD & validation
│   ├── cli/                     # CLI implementation
│   │   ├── app.go              # Main CLI app structure
│   │   └── commands/           # Command implementations
│   ├── services/                # Service container and dependency injection
│   │   └── container.go        # Application services management
│   └── config/                  # Configuration management
│       └── config.go           # YAML config + env vars
├── templates/
│   ├── builtin/                # 12 production templates
│   │   ├── reorganization/     # Architecture refactoring
│   │   ├── testing/            # Test generation
│   │   ├── refactoring/        # Code improvement
│   │   └── documentation/      # Doc generation
│   ├── custom/                 # User-defined templates
│   └── registry/               # Template metadata
└── configs/                     # Configuration files
```

## ✨ **Features**

### **Jules API Integration**

- ✅ Full Jules API v1alpha support
- ✅ Session management (create, get, list, approve, send messages)
- ✅ Activity and artifact monitoring with pagination
- ✅ Pagination support for large datasets
- ✅ Automatic retry with exponential backoff
- ✅ Comprehensive error handling
- ✅ Git patch application from sessions

**Note**: Session cancel/delete are not available in API - use [Jules web UI](https://jules.google.com)

### **Automation Engine**

- ✅ Project analysis (languages, frameworks, dependencies, architecture)
- ✅ Template-based task execution
- ✅ Dependency-aware task ordering
- ✅ Context variable interpolation
- ✅ Backup and rollback support
- ✅ Progress tracking and metrics

### **Template System**

**12 Built-in Templates** across 4 categories:

| Category | Templates | Complexity |
|----------|-----------|------------|
| **Reorganization** | Modular Restructure, Layered Architecture, Microservices Split | High |
| **Testing** | Test Generation, Coverage Improvement, Integration Tests | Medium |
| **Refactoring** | Code Cleanup, Dependency Update, API Modernization | Medium |
| **Documentation** | API Docs, README Generation, Architecture Docs | Low |

### **Event-Driven Architecture**

- ✅ **Event Bus**: Pub/sub system with topic-based routing and middleware
- ✅ **Message Queues**: Asynchronous task processing with priority levels
- ✅ **Event Store**: Event persistence for audit trails and replay capabilities
- ✅ **Circuit Breakers**: Fault tolerance for external API calls
- ✅ **Automatic Event Emission**: All Jules API calls emit structured events
- ✅ **Event Monitoring**: Real-time logging, metrics, and error aggregation

### **MCP Server**

- ✅ Official Model Context Protocol (MCP) Go SDK
- ✅ Stdio transport (compatible with Claude Desktop, Cursor)
- ✅ 19 MCP tools for project automation
- ✅ Resource endpoints (server info, config templates)
- ✅ Comprehensive tool descriptions and schemas

## � **Quick Start**

### **Prerequisites**

- Go 1.23 or higher
- Jules API key ([Get one from Google](https://jules.googleapis.com))
- Git (for project analysis features)

### **Installation**

**📚 For detailed installation instructions for all platforms, see [docs/INSTALLATION_GUIDE.md](./docs/INSTALLATION_GUIDE.md)**

#### Quick Install

**Linux/macOS:**

```bash
# Using Go (requires Go 1.23+)
go install github.com/SamyRai/juleson/cmd/juleson@latest
go install github.com/SamyRai/juleson/cmd/jules-mcp@latest
```

**Windows:**

```powershell
# Using Go (requires Go 1.23+)
go install github.com/SamyRai/juleson/cmd/juleson@latest
go install github.com/SamyRai/juleson/cmd/jules-mcp@latest
```

#### Build from Source

```bash
# Clone the repository
git clone https://github.com/SamyRai/Juleson.git
cd Juleson

# Install dependencies
go mod download

# Configure your API key
export JULES_API_KEY="your-jules-api-key-here"

# Build the orchestrator first
go build -o bin/orchestrator ./cmd/orchestrator

# Build binaries using orchestrator
./bin/orchestrator build

# Install to system
./bin/juleson dev install

# Verify installation
juleson --version
jules-mcp --version
```

## 📖 **Usage**

### **Quick Start**

```bash
# First-time setup (recommended)
juleson setup

# Or configure manually
export JULES_API_KEY="your-jules-api-key"
export GITHUB_TOKEN="ghp_your_github_token"  # Optional, for GitHub integration

# Verify setup
juleson github status
juleson sessions list
```

### **CLI Commands**

For complete command reference, see [docs/CLI_REFERENCE.md](docs/CLI_REFERENCE.md)

**Common Commands:**

```bash
# Initialize a new project configuration
juleson init ./my-project

# Analyze project structure
juleson analyze ./my-project

# List available templates
juleson template list
juleson template list reorganization  # Filter by category

# Execute a template
juleson execute template modular-restructure ./my-project

# Session management
juleson sessions list           # List all Jules sessions
juleson sessions status         # Show session summary

# GitHub integration
juleson github repos            # List your repositories
juleson github current          # Show current repo
juleson pr list                # List pull requests
juleson pr merge session-123   # Merge a PR

# Search templates
juleson template search "test coverage"

# Create custom template
juleson template create my-template refactoring "Custom refactoring workflow"
```

### **MCP Server Usage**

Start the MCP server for integration with AI assistants:

```bash
./bin/juleson-mcp
```

#### **Configure with Claude Desktop**

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "Juleson": {
      "command": "/absolute/path/to/Juleson/bin/juleson-mcp",
      "env": {
        "JULES_API_KEY": "your-api-key"
      }
    }
  }
}
```

#### **Configure with Cursor**

Add to Cursor settings JSON:

```json
{
  "mcp.servers": {
    "Juleson": {
      "command": "/absolute/path/to/Juleson/bin/juleson-mcp",
      "env": {
        "JULES_API_KEY": "your-api-key"
      }
    }
  }
}
```

#### **Available MCP Tools**

| Tool | Description |
|------|-------------|
| **Project Analysis** | |
| `analyze_project` | Deep project analysis (languages, frameworks, architecture) |
| `sync_project` | Sync project with remote Git repository |
| **Code Intelligence** | |
| `analyze_code_graph` | Build and analyze call graphs with cycle detection |
| `analyze_code_context` | Extract symbols, imports, and code structure |
| `find_symbol_references` | Find all references to a symbol across the project |
| `run_static_analysis` | Run static analysis checks (unused vars, complexity, etc.) |
| `analyze_complexity` | Calculate cyclomatic and cognitive complexity metrics |
| **Templates** | |
| `execute_template` | Run automation templates with custom parameters |
| `list_templates` | Browse available templates by category |
| `search_templates` | Find templates by keywords or tags |
| `create_template` | Create custom automation templates |
| **Session Management** | |
| `list_sessions` | View all Jules coding sessions |
| `get_session_status` | Detailed session status summary |
| `approve_session_plan` | Approve Jules session plans |
| `apply_session_patches` | Apply git patches from a session to working directory |
| `preview_session_changes` | Preview changes before applying patches (dry-run) |
| **Development Tools** | |
| `build_project` | Build Juleson binaries (CLI and MCP server) |
| `run_tests` | Execute tests with coverage and race detection |
| `lint_code` | Run linters to check code quality |
| `format_code` | Format Go code with gofmt/gofumpt |
| `clean_artifacts` | Clean build artifacts and caches |
| `quality_check` | Run all quality checks (format, lint, test) |
| `module_maintenance` | Go module operations (tidy, download, verify) |
| `build_release` | Build release binaries for all platforms |
| **Docker Management** | |
| `docker_build` | Build Docker images from Dockerfiles |
| `docker_run` | Run Docker containers with custom options |
| `docker_images` | List Docker images |
| `docker_containers` | List Docker containers |
| `docker_stop` | Stop running containers |
| `docker_remove` | Remove containers |
| `docker_rmi` | Remove Docker images |
| `docker_prune` | Clean up Docker system |
| `docker_exec` | Execute commands in running containers |
| **AI-Powered Orchestration** *(requires GEMINI_API_KEY)* | |
| `plan_project_automation` | AI-powered project analysis and comprehensive automation planning |
| `orchestrate_workflow` | Multi-step workflow execution with dependency management |
| `manage_github_project` | Natural language GitHub project management (issues, milestones, projects) |
| `synthesize_session_results` | Jules session analysis with actionable insights and recommendations |

**Note**: `cancel_session` and `delete_session` are not available in Jules API
v1alpha. Use the [Jules web UI](https://jules.google.com) for these operations.

See [MCP_SERVER_USAGE.md](docs/MCP_SERVER_USAGE.md) for detailed API documentation.

## 💡 **Examples**

### **Example 1: Analyze and Refactor a Go Project**

```bash
# Analyze project
./bin/juleson analyze ./my-go-app

# List reorganization templates
./bin/juleson template list reorganization

# Execute modular restructure template
./bin/juleson execute template modular-restructure ./my-go-app
```

### **Example 2: Generate Tests for Low Coverage**

```bash
# Execute test generation template
./bin/juleson execute template test-generation ./my-project

# Or improve existing coverage
./bin/juleson execute template test-coverage-improvement ./my-project
```

### **Example 3: Session Management Workflow**

```bash
# List all active sessions
./bin/juleson sessions list

# Get status summary
./bin/juleson sessions status

# Monitor a specific session (via Jules API)
# The session ID will be in the execute template output
```

### **Example 4: Using MCP Server with Claude**

After configuring Claude Desktop with the MCP server:

**Prompt to Claude:**
> "Use Juleson to analyze my project at /path/to/my-project and suggest
> appropriate refactoring templates"

Claude will use the MCP tools to:

1. Call `analyze_project` to understand your codebase
2. Call `list_templates` to find relevant templates
3. Suggest the best template based on analysis
4. Optionally execute the template with `execute_template`

### **Example 5: Apply Jules Session Patches**

```bash
# Preview what changes a Jules session would make (dry-run)
./bin/juleson sessions preview session-123 ./my-project

# Apply patches from Jules session to your project
./bin/juleson sessions apply session-123 ./my-project

# Apply with backup files (creates .backup files before modifying)
./bin/juleson sessions apply session-123 ./my-project --backup
```

**Using MCP with Claude:**
> "Get the changes from Jules session session-123 and apply them to my project"

Claude will:

1. Call `preview_session_changes` to show you what will be modified
2. Call `apply_session_patches` to apply the git patches
3. Report which files were modified

### **Example 6: Create Custom Template**

```bash
# Create a custom template
./bin/juleson template create api-versioning refactoring \
  "Add API versioning to existing REST endpoints"

# Edit the generated template file
# templates/custom/refactoring/api-versioning.yaml

# Execute your custom template
./bin/juleson execute template api-versioning ./my-api-project
```

### **Example 7: Automated CI/CD Integration**

```yaml
# .github/workflows/Juleson.yml
name: Juleson

on:
  workflow_dispatch:
    inputs:
      template:
        description: 'Template to execute'
        required: true
        default: 'test-generation'
      project_path:
        description: 'Project path'
        required: true
        default: '.'

jobs:
  automate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      steps:
        - name: Checkout
          uses: actions/checkout@v3
        - name: Download Jules CLI
          run: |
          git clone https://github.com/SamyRai/Juleson.git
          cd Juleson
          go build -o bin/orchestrator ./cmd/orchestrator
          ./bin/orchestrator build

      - name: Execute Template
        env:
          JULES_API_KEY: ${{ secrets.JULES_API_KEY }}
        run: |
          ./Juleson/bin/juleson execute template \
            ${{ github.event.inputs.template }} \
            ${{ github.event.inputs.project_path }}
```

### **Configuration File**

Create `configs/Juleson.yaml`:

```yaml
jules:
  api_key: ""  # Or use JULES_API_KEY environment variable
  base_url: "https://jules.googleapis.com/v1alpha"
  timeout: "30s"
  retry_attempts: 3

mcp:
  server:
    port: 8080
    host: "localhost"
  client:
    timeout: "10s"

automation:
  strategies:
    - "modular"
    - "layered"
    - "microservices"
  max_concurrent_tasks: 5
  task_timeout: "300s"

projects:
  default_path: "./projects"
  backup_enabled: true
  git_integration: true
```

### **Environment Variables**

```bash
# Required
export JULES_API_KEY="your-jules-api-key"

# Optional (with defaults)
export JULES_BASE_URL="https://jules.googleapis.com/v1alpha"
export JULES_TIMEOUT="30s"
export JULES_RETRY_ATTEMPTS="3"
```

See [configs/Juleson.example.yaml](configs/Juleson.example.yaml) for full
configuration options.

## 🧪 **Development**

### **Running Tests**

```bash
# Build orchestrator first
go build -o bin/orchestrator ./cmd/orchestrator

# Run all tests
./bin/orchestrator test

# Run with coverage
./bin/orchestrator coverage

# Run specific package tests
go test -v ./internal/jules/...
go test -v ./internal/mcp/...

# Short tests only (exclude integration tests)
./bin/orchestrator test-short
```

### **Code Quality**

```bash
# Format code
./bin/orchestrator fmt

# Run linters
./bin/orchestrator lint

# Run all checks (fmt + lint + test)
./bin/orchestrator check
```

### **Building**

```bash
# Build both binaries
./bin/orchestrator build

# Build CLI only
./bin/orchestrator build-cli

# Build MCP server only
./bin/orchestrator build-mcp

# Install to $GOPATH/bin
./bin/orchestrator install
```

### **Project Statistics**

- **Test Coverage**: 80%+ across core packages
- **Lines of Code**: ~5,000 (excluding tests)
- **Dependencies**: Minimal (cobra, viper, MCP SDK, testify)
- **Go Packages**: 7 internal packages
- **Built-in Templates**: 12

## � **API Reference**

### **Jules Client API**

```go
// Create a Jules client
client := jules.NewClient(apiKey, baseURL, timeout, retryAttempts)

// Session management
session, err := client.CreateSession(ctx, &jules.CreateSessionRequest{
    Prompt: "Refactor this project to use clean architecture",
    Title:  "Architecture Refactoring",
    SourceContext: &jules.SourceContext{Source: "./my-project"},
})

// List sessions with pagination
response, err := client.ListSessionsWithPagination(ctx, 50, "")

// Get session details
session, err := client.GetSession(ctx, sessionID)

// Approve session plan
err := client.ApprovePlan(ctx, sessionID)

// Send message to session
err := client.SendMessage(ctx, sessionID, "Please add error handling")

// Apply patches from session to working directory
result, err := client.ApplySessionPatches(ctx, sessionID, &jules.PatchApplicationOptions{
    WorkingDir:   "./my-project",
    DryRun:       false,
    CreateBackup: true,
})

// Preview session changes (dry-run)
changes, err := client.PreviewSessionPatches(ctx, sessionID, "./my-project")

// Get session changes summary
changes, err := client.GetSessionChanges(ctx, sessionID)

// Activity monitoring
activities, err := client.ListActivities(ctx, sessionID, 100)
```

### **Automation Engine API**

```go
// Create automation engine
engine := automation.NewEngine(julesClient, templateManager)

// Analyze project
context, err := engine.AnalyzeProject("./my-project")

// Execute template
result, err := engine.ExecuteTemplate(ctx, "modular-restructure", map[string]string{
    "target_architecture": "clean",
    "preserve_tests": "true",
})
```

### **Template Manager API**

```go
// Create template manager
manager, err := templates.NewManager("./templates")

// Load template
template, err := manager.LoadTemplate("modular-restructure")

// List all templates
templates := manager.ListTemplates()

// Search templates
results := manager.SearchTemplates("test coverage")

// Create custom template
template, err := manager.CreateTemplate("my-template", "refactoring", "Description")
```

## �️ **Roadmap**

### **v0.2.0 - Enhanced Analysis** (Q1 2025)

- [ ] Advanced dependency graph analysis
- [ ] Test coverage calculation
- [ ] Code complexity metrics
- [ ] Performance profiling integration

### **v0.3.0 - Workflow Automation** (Q2 2025)

- [ ] Multi-step workflow definitions
- [ ] Conditional task execution
- [ ] Parallel task processing
- [ ] Workflow state persistence

### **v0.4.0 - Extended Platform Support** (Q3 2025)

- [x] GitHub API integration (MCP tools for repo management and PR workflows)
- [ ] GitLab CI/CD support
- [ ] Docker containerization
- [ ] VS Code extension

### **v1.0.0 - Production Release** (Q4 2025)

- [ ] Comprehensive template library (50+ templates)
- [ ] Web UI dashboard
- [ ] Team collaboration features
- [ ] Enterprise security features
- [ ] SLA monitoring and alerts

## 🤝 **Contributing**

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### **Quick Contribution Guide**

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run quality checks (`./bin/orchestrator check`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to your fork (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### **Development Setup**

```bash
# Clone your fork
git clone https://github.com/SamyRai/Juleson.git
cd Juleson

# Install dependencies
go mod download

# Build orchestrator
go build -o bin/orchestrator ./cmd/orchestrator

# Run tests
./bin/orchestrator test

# Build
./bin/orchestrator build
```

### **Code Standards**

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Maintain >80% test coverage for new code
- Use conventional commits (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`)
- Add godoc comments for exported functions
- Run `./bin/orchestrator fmt` and `./bin/orchestrator lint` before committing

## � **License**

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

```text
Copyright (c) 2025 Juleson Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software...
```

## 🔗 **Resources**

- **Documentation**: [docs/](docs/)
  - [Setup Guide](docs/SETUP_GUIDE.md) - First-time setup and configuration
  - [CLI Reference](docs/CLI_REFERENCE.md) - Complete command-line reference
  - [MCP Server Usage Guide](docs/MCP_SERVER_USAGE.md) - MCP integration
  - [Code Intelligence](docs/CODE_INTELLIGENCE.md) - Advanced code analysis features
  - [Event System Quick Start](docs/EVENT_SYSTEM_QUICKSTART.md) - Event-driven architecture
  - [Event System Architecture](docs/EVENT_SYSTEM_ARCHITECTURE.md) - Event system design
  - [GitHub Configuration Guide](docs/GITHUB_CONFIGURATION_GUIDE.md) - GitHub setup
  - [Installation Guide](docs/INSTALLATION_GUIDE.md) - Platform-specific installation
  - [Template System Documentation](docs/Y2Q2_TEMPLATE_SYSTEM.md) - Template creation
  - [GitHub Actions Integration](docs/GITHUB_ACTIONS_GUIDE.md) - CI/CD setup
- **Jules API**: [Google Jules API Documentation](https://developers.google.com/jules/api)
- **MCP Protocol**: [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- **Official MCP Go SDK**: [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)

## 📊 **Project Status**

- **Current Version**: 0.1.0 (Alpha)
- **Production Ready**: Yes (with API key)
- **Test Coverage**: 80%+
- **CI/CD**: GitHub Actions (planned)
- **Stability**: Stable API, active development

## ⚠️ **Known Limitations**

- Jules API access requires approved API key from Google
- MCP server requires stdio transport (no HTTP/WebSocket yet)
- Template execution requires active internet connection
- Large projects (>10k files) may have slower analysis
- Session monitoring is polling-based (no webhooks yet)

## 🆘 **Support**

- **Issues**: [GitHub Issues](https://github.com/SamyRai/Juleson/issues)
- **Discussions**: [GitHub Discussions](https://github.com/SamyRai/Juleson/discussions)
- **Security**: See [SECURITY.md](SECURITY.md)
- **Changelog**: See [CHANGELOG.md](CHANGELOG.md)

## 🙏 **Acknowledgments**

- Google Jules team for the amazing AI coding agent
- Model Context Protocol team for the excellent Go SDK
- [Cobra](https://github.com/spf13/cobra) for CLI framework
- [Viper](https://github.com/spf13/viper) for configuration management
- All contributors who help improve this project

---

## 👥 **Community**

Made with ❤️ by the Juleson Community

*Star ⭐ this repository if you find it helpful!*
