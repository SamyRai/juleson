# Juleson CLI Reference

Complete command-line reference for the Juleson automation tool.

## Table of Contents

- [Global Flags](#global-flags)
- [Setup & Configuration](#setup--configuration)
- [Project Analysis](#project-analysis)
- [Session Management](#session-management)
- [Template Management](#template-management)
- [GitHub Integration](#github-integration)
- [Pull Request Management](#pull-request-management)
- [Development Tools](#development-tools)
- [Shell Completion](#shell-completion)

## Global Flags

These flags are available for all commands:

```
--config string      Config file (default searches ./configs, ., $HOME, /etc)
--verbose            Enable verbose output
--json               Output in JSON format
--help               Show help for command
```

## Setup & Configuration

### setup

Interactive setup wizard for first-time configuration.

```bash
juleson setup [flags]
```

**Flags:**

```
--non-interactive       Run without prompts (use env vars)
--skip-completion       Skip shell completion installation
--skip-jules            Skip Jules API configuration
--skip-github           Skip GitHub integration setup
```

**Examples:**

```bash
# Interactive setup (recommended)
juleson setup

# Non-interactive setup for CI/CD
export JULES_API_KEY="your-key"
export GITHUB_TOKEN="your-token"
juleson setup --non-interactive

# Setup without GitHub
juleson setup --skip-github

# Only configure API, skip everything else
juleson setup --skip-completion --skip-github
```

**What it does:**

1. Detects your shell (bash, zsh, fish, powershell)
2. Offers to install shell completion
3. Prompts for Jules API key
4. Optionally configures GitHub integration
5. Validates configuration
6. Saves to `~/.juleson.yaml`

### init

Initialize a new Juleson project.

```bash
juleson init [flags]
```

**Flags:**

```
--name string           Project name
--description string    Project description
--template string       Template to use
```

**Examples:**

```bash
# Interactive initialization
juleson init

# With project details
juleson init --name "my-project" --description "My automation project"
```

## Project Analysis

### analyze

Analyze project structure, dependencies, and complexity.

```bash
juleson analyze [path] [flags]
```

**Flags:**

```
--depth int            Analysis depth (default 3)
--exclude strings      Patterns to exclude
--output string        Output format: text, json, yaml (default "text")
```

**Examples:**

```bash
# Analyze current directory
juleson analyze

# Analyze specific path
juleson analyze /path/to/project

# JSON output for scripting
juleson analyze --output json

# Exclude patterns
juleson analyze --exclude "node_modules,vendor,*.test.go"
```

**Output includes:**

- Project structure overview
- Language distribution
- Dependency analysis
- Code complexity metrics
- Architecture patterns
- Potential improvements

## Session Management

### sessions list

List all Jules automation sessions.

```bash
juleson sessions list [flags]
```

**Flags:**

```
--status string         Filter by status: pending, active, completed, failed
--limit int            Maximum sessions to return (default 50)
--cursor string        Pagination cursor
--show-all             Show all sessions (no limit)
```

**Examples:**

```bash
# List recent sessions
juleson sessions list

# Show only active sessions
juleson sessions list --status active

# Show all completed sessions
juleson sessions list --status completed --show-all

# Get next page
juleson sessions list --cursor "eyJpZCI6IjEyMyJ9"
```

### sessions get

Get detailed information about a session.

```bash
juleson sessions get <session-id> [flags]
```

**Flags:**

```
--show-sources         Show session source code files
--show-artifacts       Show generated artifacts
--show-patches         Show code patches
--output string        Output format: text, json (default "text")
```

**Examples:**

```bash
# Get session details
juleson sessions get abc123

# Include source files
juleson sessions get abc123 --show-sources

# Show all details
juleson sessions get abc123 --show-sources --show-artifacts --show-patches

# JSON output
juleson sessions get abc123 --output json
```

### sessions create

Create a new automation session.

```bash
juleson sessions create <instruction> [flags]
```

**Flags:**

```
--wait                 Wait for session to complete
--timeout duration     Wait timeout (default 5m)
--template string      Template to use
```

**Examples:**

```bash
# Create session with instruction
juleson sessions create "Refactor CLI commands to use Cobra patterns"

# Create and wait for completion
juleson sessions create "Add unit tests for GitHub client" --wait

# Use specific template
juleson sessions create "Improve code quality" --template "code-quality-improvement"

# With custom timeout
juleson sessions create "Migrate to new API" --wait --timeout 10m
```

### sessions approve

Approve a pending session plan.

```bash
juleson sessions approve <session-id>
```

**Examples:**

```bash
# Approve session
juleson sessions approve abc123
```

### sessions cancel

Cancel an active session.

```bash
juleson sessions cancel <session-id>
```

**Examples:**

```bash
# Cancel session
juleson sessions cancel abc123
```

### sessions delete

Delete a completed or failed session.

```bash
juleson sessions delete <session-id> [flags]
```

**Flags:**

```
--force                Delete without confirmation
```

**Examples:**

```bash
# Delete with confirmation
juleson sessions delete abc123

# Force delete
juleson sessions delete abc123 --force
```

### sessions status

Get summary of all sessions.

```bash
juleson sessions status [flags]
```

**Flags:**

```
--limit int            Maximum sessions to analyze (default 100)
```

**Examples:**

```bash
# Get status summary
juleson sessions status

# Analyze more sessions
juleson sessions status --limit 500
```

## Template Management

### template list

List available automation templates.

```bash
juleson template list [flags]
```

**Flags:**

```
--category string      Filter by category: documentation, refactoring, testing, reorganization
--builtin              Show only builtin templates
--custom               Show only custom templates
```

**Examples:**

```bash
# List all templates
juleson template list

# Show only testing templates
juleson template list --category testing

# Show custom templates
juleson template list --custom

# Show builtin templates
juleson template list --builtin
```

### template search

Search templates by query.

```bash
juleson template search <query>
```

**Examples:**

```bash
# Search for test templates
juleson template search "test"

# Search for refactoring templates
juleson template search "refactor"
```

### template create

Create a custom template.

```bash
juleson template create [flags]
```

**Flags:**

```
--name string          Template name (required)
--category string      Template category (required)
--description string   Template description (required)
--file string          Load template from file
```

**Examples:**

```bash
# Interactive creation
juleson template create

# From command line
juleson template create \
  --name "my-template" \
  --category "refactoring" \
  --description "My custom refactoring template"

# From file
juleson template create --file ./my-template.yaml
```

### template execute

Execute a template on a project.

```bash
juleson template execute <template-name> [path] [flags]
```

**Flags:**

```
--param strings        Template parameters (key=value)
--dry-run             Show what would be done
--wait                Wait for completion
```

**Examples:**

```bash
# Execute template on current directory
juleson template execute "test-generation"

# Execute on specific path
juleson template execute "code-cleanup" /path/to/project

# With parameters
juleson template execute "api-documentation" --param "format=markdown" --param "output=docs/"

# Dry run to preview
juleson template execute "refactor" --dry-run

# Execute and wait
juleson template execute "test-coverage-improvement" --wait
```

## GitHub Integration

### github login

Authenticate with GitHub.

```bash
juleson github login [flags]
```

**Flags:**

```
--token string         GitHub Personal Access Token
--save                Save token to config (default true)
```

**Examples:**

```bash
# Interactive login
juleson github login

# Login with token
juleson github login --token "ghp_your_token_here"

# Login without saving
juleson github login --save=false
```

### github status

Check GitHub authentication and connection status.

```bash
juleson github status
```

**Examples:**

```bash
# Check status
juleson github status
```

**Shows:**

- Authentication status
- Connected user
- API rate limits
- Repository access

### github repos

List accessible GitHub repositories.

```bash
juleson github repos [flags]
```

**Flags:**

```
--org string           Filter by organization
--limit int           Maximum repositories (default 30)
--visibility string   Filter by visibility: public, private, all (default "all")
--sort string         Sort by: created, updated, pushed, full_name (default "updated")
```

**Examples:**

```bash
# List your repositories
juleson github repos

# List organization repositories
juleson github repos --org "my-org"

# Show only private repos
juleson github repos --visibility private

# Show recently updated
juleson github repos --sort updated --limit 10

# Show recently created
juleson github repos --sort created --limit 5
```

### github current

Show current repository (auto-detected from git remote).

```bash
juleson github current
```

**Examples:**

```bash
# Show current repository
juleson github current
```

**Shows:**

- Repository name and owner
- Description
- Default branch
- Clone URLs
- Stars, forks, watchers
- Open issues and PRs

## Pull Request Management

### pr list

List pull requests from Jules sessions.

```bash
juleson pr list [flags]
```

**Flags:**

```
--status string        Filter by status: open, closed, merged, all (default "all")
--session string      Filter by session ID
--limit int           Maximum PRs to show (default 20)
```

**Examples:**

```bash
# List all PRs
juleson pr list

# Show only open PRs
juleson pr list --status open

# PRs for specific session
juleson pr list --session abc123

# Recent merged PRs
juleson pr list --status merged --limit 10
```

### pr get

Get detailed PR information.

```bash
juleson pr get <session-id> [flags]
```

**Flags:**

```
--show-diff            Show PR diff
--show-files           Show changed files
```

**Examples:**

```bash
# Get PR details
juleson pr get abc123

# Show with diff
juleson pr get abc123 --show-diff

# Show changed files
juleson pr get abc123 --show-files

# Show everything
juleson pr get abc123 --show-diff --show-files
```

### pr merge

Merge a pull request.

```bash
juleson pr merge <session-id> [flags]
```

**Flags:**

```
--method string        Merge method: merge, squash, rebase (default "squash")
--message string       Commit message
--delete-branch        Delete branch after merge (default true)
--force               Skip confirmation
```

**Examples:**

```bash
# Merge with confirmation
juleson pr merge abc123

# Squash merge
juleson pr merge abc123 --method squash

# Merge and keep branch
juleson pr merge abc123 --delete-branch=false

# Force merge without confirmation
juleson pr merge abc123 --force

# With custom message
juleson pr merge abc123 --message "Merged: Add new feature"
```

### pr diff

Show PR diff.

```bash
juleson pr diff <session-id> [flags]
```

**Flags:**

```
--output string        Output format: patch, unified (default "unified")
--context int         Lines of context (default 3)
```

**Examples:**

```bash
# Show diff
juleson pr diff abc123

# Show as patch file
juleson pr diff abc123 --output patch

# More context
juleson pr diff abc123 --context 5
```

## Development Tools

### dev build

Build Juleson binaries.

```bash
juleson dev build [flags]
```

**Flags:**

```
--target string        Build target: cli, mcp, all (default "all")
--version string       Version to embed (default "dev")
--race                Enable race detection
--goos string         Target OS (default: current)
--goarch string       Target architecture (default: current)
```

**Examples:**

```bash
# Build all binaries
juleson dev build

# Build CLI only
juleson dev build --target cli

# Build with version
juleson dev build --version "v1.0.0"

# Build for Linux
juleson dev build --goos linux --goarch amd64

# Build with race detection
juleson dev build --race
```

### dev test

Run tests with coverage.

```bash
juleson dev test [flags]
```

**Flags:**

```
--verbose              Verbose test output
--race                Enable race detection (default true)
--cover               Enable coverage (default false)
--short               Run short tests only
--package strings     Specific packages to test
```

**Examples:**

```bash
# Run all tests
juleson dev test

# Verbose output
juleson dev test --verbose

# With coverage
juleson dev test --cover

# Short tests only
juleson dev test --short

# Specific package
juleson dev test --package ./internal/github
```

### dev lint

Run linters.

```bash
juleson dev lint [flags]
```

**Flags:**

```
--fix                 Automatically fix issues
--verbose            Verbose output
```

**Examples:**

```bash
# Run linters
juleson dev lint

# Fix issues
juleson dev lint --fix

# Verbose output
juleson dev lint --verbose
```

### dev format

Format Go code.

```bash
juleson dev format [flags]
```

**Flags:**

```
--use-gofumpt         Use gofumpt instead of gofmt
```

**Examples:**

```bash
# Format code
juleson dev format

# Use gofumpt
juleson dev format --use-gofumpt
```

### dev quality

Run all quality checks (format, lint, test).

```bash
juleson dev quality
```

**Examples:**

```bash
# Run all quality checks
juleson dev quality
```

### dev clean

Clean build artifacts.

```bash
juleson dev clean [flags]
```

**Flags:**

```
--all                 Clean everything including caches
--cache              Clean build cache
--modcache           Clean module cache
```

**Examples:**

```bash
# Clean build artifacts
juleson dev clean

# Clean everything
juleson dev clean --all

# Clean module cache
juleson dev clean --modcache
```

### dev release

Build release binaries for all platforms.

```bash
juleson dev release <version>
```

**Examples:**

```bash
# Build release
juleson dev release v1.0.0
```

### dev module

Go module maintenance.

```bash
juleson dev module <operation>
```

**Operations:**

- `tidy` - Clean up go.mod and go.sum
- `download` - Download dependencies
- `verify` - Verify dependencies
- `graph` - Show dependency graph

**Examples:**

```bash
# Tidy modules
juleson dev module tidy

# Download dependencies
juleson dev module download

# Verify dependencies
juleson dev module verify

# Show dependency graph
juleson dev module graph
```

## Shell Completion

### completion

Generate shell completion scripts.

```bash
juleson completion <shell>
```

**Supported shells:**

- `bash` - Bash completion
- `zsh` - Zsh completion
- `fish` - Fish completion
- `powershell` - PowerShell completion

**Examples:**

```bash
# Generate Bash completion
juleson completion bash > /etc/bash_completion.d/juleson

# Generate Zsh completion
juleson completion zsh > ~/.zfunc/_juleson

# Generate Fish completion
juleson completion fish > ~/.config/fish/completions/juleson.fish

# Generate PowerShell completion
juleson completion powershell >> $PROFILE
```

**Or use setup wizard:**

```bash
juleson setup
```

## Environment Variables

Juleson respects these environment variables:

```bash
# Jules API
JULES_API_KEY              Jules API authentication key
JULES_BASE_URL             Jules API base URL (default: https://jules.googleapis.com/v1alpha)

# GitHub
GITHUB_TOKEN               GitHub Personal Access Token
GITHUB_DEFAULT_ORG         Default GitHub organization

# Configuration
JULESON_CONFIG             Path to config file
JULESON_LOG_LEVEL          Log level: debug, info, warn, error

# Shell
SHELL                      Shell type (auto-detected)
```

## Exit Codes

Juleson uses standard exit codes:

```
0   Success
1   General error
2   Invalid usage (wrong flags, missing args)
3   Configuration error
4   Authentication error
5   Network error
```

## See Also

- [Setup Guide](SETUP_GUIDE.md) - Detailed setup instructions
- [GitHub Integration Guide](GITHUB_CONFIGURATION_GUIDE.md) - GitHub setup
- [Configuration Guide](../configs/README.md) - Config file reference
- [MCP Server Usage](MCP_SERVER_USAGE.md) - MCP integration
