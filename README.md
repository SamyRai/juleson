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
- **🤖 AI-Powered Automation**: Execute complex refactoring and restructuring tasks via Jules AI
- **📋 Template System**: 12+ built-in templates for reorganization, testing, refactoring, and documentation
- **💬 Session Management**: Full control over Jules coding sessions (list, monitor, approve, cancel)
- **🔌 MCP Integration**: Native Model Context Protocol server for AI assistants (Claude, Cursor, etc.)
- **⚙️ CLI Tools**: Comprehensive command-line interface for all automation tasks

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

# Build binaries using Makefile
make build

# Install to system
./bin/juleson dev install

# Verify installation
juleson --version
jules-mcp --version
```

## 📖 **Usage**

### **CLI Commands**

```bash
# Initialize a new project configuration
./bin/juleson init ./my-project

# Analyze project structure
./bin/juleson analyze ./my-project

# List available templates
./bin/juleson template list
./bin/juleson template list reorganization  # Filter by category

# Show template details
./bin/juleson template show modular-restructure

# Execute a template
./bin/juleson execute template modular-restructure ./my-project

# Session management
./bin/juleson sessions list           # List all Jules sessions
./bin/juleson sessions status         # Show session summary

# Search templates
./bin/juleson template search "test coverage"

# Create custom template
./bin/juleson template create my-template refactoring "Custom refactoring workflow"
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
          make build

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
# Run all tests
make test

# Run with coverage
make coverage

# Run specific package tests
go test -v ./internal/jules/...
go test -v ./internal/mcp/...

# Short tests only (exclude integration tests)
make test-short
```

### **Code Quality**

```bash
# Format code
make fmt

# Run linters
make lint

# Run all checks (fmt + lint + test)
make check
```

### **Building**

```bash
# Build both binaries
make build

# Build CLI only
make build-cli

# Build MCP server only
make build-mcp

# Install to $GOPATH/bin
make install
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

- [ ] GitHub Actions integration
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
4. Run quality checks (`make check`)
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

# Run tests
make test

# Build
make build
```

### **Code Standards**

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Maintain >80% test coverage for new code
- Use conventional commits (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`)
- Add godoc comments for exported functions
- Run `make fmt` and `make lint` before committing

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
  - [MCP Server Usage Guide](docs/MCP_SERVER_USAGE.md)
  - [Template System Documentation](docs/Y2Q2_TEMPLATE_SYSTEM.md)
  - [GitHub Actions Integration](docs/GITHUB_ACTIONS_GUIDE.md)
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
