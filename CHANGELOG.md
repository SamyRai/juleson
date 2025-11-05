# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **GitHub Integration Refactoring**: Complete overhaul of GitHub client package following SOLID principles
  - Split monolithic `client.go` (900+ lines) into focused service modules
  - Added `ActionsService` for GitHub Actions workflows, runs, jobs, artifacts, and caches
  - Added `RepositoryService` for repository discovery and management
  - Added `PullRequestService` for PR operations and management
  - Added `SessionService` for Jules session creation with GitHub context
  - Added `GitRemoteParser` utility for git remote URL parsing
  - Implemented service-based architecture with dependency injection
  - Added comprehensive domain types in `types.go`
  - Updated all CLI commands to use new service-based architecture
  - Added detailed architecture documentation in `internal/github/README.md`

- **CLI Setup Wizard**: Interactive `juleson setup` command for first-time configuration
  - Auto-detects shell type (bash, zsh, fish, powershell)
  - Guides through Jules API key setup
  - Optional GitHub token configuration
  - Automatic shell completion installation
  - Config validation
  - Non-interactive mode for CI/CD (`--non-interactive`)
- **Shell Completion**: Auto-completion for all shells via `juleson completion <shell>`
  - Bash completion support
  - Zsh completion support
  - Fish completion support
  - PowerShell completion support
- **GitHub Integration**: Complete GitHub CLI commands
  - `juleson github login` - Authenticate with GitHub
  - `juleson github status` - Check authentication and rate limits
  - `juleson github repos` - List accessible repositories
  - `juleson github current` - Show current repository (auto-detected)
- **Pull Request Management**: Full PR workflow
  - `juleson pr list` - List PRs from Jules sessions
  - `juleson pr get` - View PR details
  - `juleson pr merge` - Merge PRs with method selection
  - `juleson pr diff` - Show actual PR diff via GitHub API
- **Config Persistence**: `Config.Save()` method for writing configuration
- **Documentation**: Comprehensive user guides
  - Setup Guide (docs/SETUP_GUIDE.md)
  - CLI Reference (docs/CLI_REFERENCE.md)
  - Updated README with quick start

- **Google Gemini AI Integration**: Advanced AI capabilities for code analysis and automation
  - Migrated from deprecated `generative-ai-go` to new unified `google.golang.org/genai` SDK
  - Added `GeminiConfig` to configuration structure with API key, backend, project, location, model, timeout, and max_tokens settings
  - Created `internal/gemini/client.go` with Gemini client supporting both API and Vertex AI backends
  - Implemented `GenerateContent` and `GenerateContentWithImages` methods for multimodal AI interactions
  - Integrated Gemini client into services container with lazy initialization
  - **Redesigned Gemini MCP tools for high-level project orchestration**:
    - `plan_project_automation` - AI-powered project analysis and comprehensive automation planning
    - `orchestrate_workflow` - Multi-step workflow execution with dependency management
    - `manage_github_project` - Natural language GitHub project management (issues, milestones, projects)
    - `synthesize_session_results` - Jules session analysis with actionable insights and recommendations
  - Conditional tool registration in MCP server when Gemini API key is configured
  - Support for Gemini 2.5 Pro, Flash, and Flash-Lite models with 1M token context windows
  - Multimodal input support for images and structured output capabilities

### Changed

- Enhanced GitHub client with `GetPullRequestDiff()` method
- Improved config loading to support multiple locations
- Updated help text and command descriptions

### Planned Additions

- Interactive CLI mode for guided workflows
- Smart error handling with suggestions
- Progress indicators for long operations
- Config validation command
- Output formatting options (JSON, YAML)
- Advanced dependency analysis
- Test coverage analytics
- Code complexity metrics

## [0.1.0] - 2024-11-01

### Added

- Initial alpha release
- Core automation engine
- Jules API integration
- MCP protocol support
- Basic CLI commands:
  - `init` - Initialize project
  - `analyze` - Analyze project structure
  - `execute` - Execute templates
  - `template` - Manage templates
  - `sessions` - Manage Jules sessions
- Built-in templates:
  - Modular restructure
  - Test generation
- Documentation:
  - README with quick start guide
  - MCP server usage guide
  - Template system documentation
  - Contributing guidelines

[Unreleased]: https://github.com/SamyRai/Juleson/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/SamyRai/Juleson/releases/tag/v0.1.0
