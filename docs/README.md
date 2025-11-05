# Documentation Index

Welcome to the Juleson documentation! This directory contains comprehensive documentation for all aspects of the Juleson automation toolkit.

## 📚 **Documentation Overview**

Juleson is a production-ready automation toolkit that integrates with Google's Jules AI coding agent through both CLI and MCP (Model Context Protocol) interfaces.

## 🚀 **Quick Start**

If you're new to Juleson, start here:

1. **[Installation Guide](INSTALLATION_GUIDE.md)** - Platform-specific installation instructions
2. **[Setup Guide](SETUP_GUIDE.md)** - First-time setup and configuration
3. **[CLI Reference](CLI_REFERENCE.md)** - Complete command-line reference

## 📖 **Core Documentation**

### **Architecture & Design**

- **[Event System Quick Start](EVENT_SYSTEM_QUICKSTART.md)** - Event-driven architecture overview
- **[Event System Architecture](EVENT_SYSTEM_ARCHITECTURE.md)** - Detailed event system design
- **[Code Intelligence](CODE_INTELLIGENCE.md)** - Advanced code analysis and intelligence features
- **[MCP Server Usage](MCP_SERVER_USAGE.md)** - Model Context Protocol integration
- **[Template System](Y2Q2_TEMPLATE_SYSTEM.md)** - Template creation and management

### **Integration Guides**

- **[GitHub Configuration](GITHUB_CONFIGURATION_GUIDE.md)** - GitHub API integration setup
- **[GitHub Actions Integration](GITHUB_ACTIONS_GUIDE.md)** - CI/CD automation
- **[GitHub Integration Proposal](GITHUB_INTEGRATION_PROPOSAL.md)** - Comprehensive GitHub features

### **Development & Operations**

- **[Agent Architecture](AGENT_ARCHITECTURE.md)** - Agent system design
- **[Agent Production Features](AGENT_PRODUCTION_FEATURES.md)** - Production deployment features
- **[DX Improvements](DX_IMPROVEMENTS.md)** - Developer experience enhancements
- **[Deployment Guide](DEPLOYMENT_GUIDE.md)** - Production deployment

## 🔧 **API & Technical Reference**

### **Internal Package Documentation**

- [`internal/events/`](../internal/events/README.md) - Event system components
- [`internal/github/`](../internal/github/README.md) - GitHub integration
- [`internal/jules/`](../api/README.md) - Jules API client
- [`configs/`](../configs/README.md) - Configuration management

### **External Resources**

- [Jules API Documentation](https://developers.google.com/jules/api) - Official Jules API
- [Model Context Protocol](https://modelcontextprotocol.io/) - MCP specification
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) - Official Go SDK

## 📋 **Templates & Automation**

Juleson includes 12+ built-in automation templates across 4 categories:

| Category | Templates | Complexity |
|----------|-----------|------------|
| **Reorganization** | Modular Restructure, Layered Architecture, Microservices Split | High |
| **Testing** | Test Generation, Coverage Improvement, Integration Tests | Medium |
| **Refactoring** | Code Cleanup, Dependency Update, API Modernization | Medium |
| **Documentation** | API Docs, README Generation, Architecture Docs | Low |

See [`templates/`](../templates/README.md) for template documentation.

## 🏗️ **Project Structure**

```
Juleson/
├── cmd/                          # Application entry points
│   ├── juleson/                 # CLI tool
│   └── juleson-mcp/             # MCP server
├── internal/                     # Core packages
│   ├── events/                  # Event-driven architecture
│   ├── jules/                   # Jules API client
│   ├── mcp/                     # MCP server implementation
│   ├── automation/              # Automation engine
│   ├── templates/               # Template management
│   ├── cli/                     # CLI implementation
│   ├── services/                # Service container
│   └── config/                  # Configuration management
├── templates/                    # Automation templates
├── configs/                      # Configuration files
├── docs/                         # Documentation (this directory)
└── scripts/                      # Development scripts
```

## 🎯 **Key Features**

### **Event-Driven Architecture**

- **Event Bus**: Pub/sub system with middleware and topic-based routing
- **Message Queues**: Asynchronous processing with priority levels
- **Event Store**: Persistence and replay capabilities
- **Circuit Breakers**: Fault tolerance for external services
- **Automatic Events**: All Jules API calls emit structured events

### **AI-Powered Automation**

- **Jules Integration**: Full API v1alpha support with session management
- **Template System**: 12+ built-in automation templates
- **Project Analysis**: Deep codebase inspection and architecture detection
- **MCP Server**: Native Model Context Protocol support

### **Developer Experience**

- **CLI Tools**: Comprehensive command-line interface
- **GitHub Integration**: Repository management and PR workflows
- **Configuration Management**: YAML configs with environment variable support
- **Development Tools**: Testing, linting, formatting utilities

## 📊 **Project Status**

- **Version**: 0.1.0 (Alpha)
- **Go Version**: 1.24+
- **Test Coverage**: 80%+
- **License**: MIT
- **Status**: Production-ready (requires Jules API key)

## 🆘 **Support & Community**

- **Issues**: [GitHub Issues](https://github.com/SamyRai/Juleson/issues)
- **Discussions**: [GitHub Discussions](https://github.com/SamyRai/Juleson/discussions)
- **Contributing**: See [CONTRIBUTING.md](../CONTRIBUTING.md)
- **Security**: See [SECURITY.md](../SECURITY.md)

## 📝 **Documentation Standards**

This documentation follows these standards:

- **Markdown**: All docs use GitHub-flavored Markdown
- **Cross-references**: Relative links between documents
- **Code examples**: Runnable code snippets with explanations
- **Table of contents**: Auto-generated in longer documents
- **Versioning**: Documentation updated with code changes

## 🔄 **Contributing to Documentation**

When contributing documentation:

1. Follow the existing structure and style
2. Include code examples where helpful
3. Update cross-references when adding/removing files
4. Test all links and code examples
5. Update this index when adding new documents

---

*Last updated: November 2025*</content>
<parameter name="filePath">/Users/damirmukimov/projects/jules-automation/docs/README.md
