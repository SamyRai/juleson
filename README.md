# Juleson

Juleson is a Go command-line tool and MCP server for working with Google's Jules
coding agent. It provides local commands for session management, source discovery,
GitHub integration, template execution, project analysis, and development tasks.
The reusable Jules REST API client is available as `github.com/SamyRai/juleson/pkg/jules`.

The repository contains three binaries:

- `juleson`: user-facing CLI.
- `jules-mcp`: Model Context Protocol server over stdio.
- `orchestrator`: internal build/test/release helper used by development workflows.

## Requirements

- Go 1.25 or newer for source builds.
- A Jules API key for commands that call the Jules API.
- A GitHub token for GitHub and Actions commands.
- A Gemini API key for Gemini-backed orchestration commands.

## Install

Linux and macOS:

```bash
curl -L https://github.com/SamyRai/juleson/releases/latest/download/install.sh | bash
```

Windows PowerShell:

```powershell
irm https://github.com/SamyRai/juleson/releases/latest/download/install.ps1 | iex
```

Install from source:

```bash
go install github.com/SamyRai/juleson/cmd/juleson@latest
go install github.com/SamyRai/juleson/cmd/jules-mcp@latest
```

Build locally:

```bash
go build -o bin/orchestrator ./cmd/orchestrator
./bin/orchestrator build
```

See [Installation](docs/INSTALLATION_GUIDE.md) for platform-specific options.

## Configure

Run the setup wizard:

```bash
juleson setup
```

For non-interactive setup, set environment variables first. The setup command
reads them and writes the resulting values to `configs/juleson.yaml`:

```bash
export JULES_API_KEY="..."
export GITHUB_TOKEN="..."
juleson setup --non-interactive
```

Juleson looks for `juleson.yaml` in `./configs`, the current directory, `$HOME`,
and `/etc/juleson`. It also loads `.env`, `$HOME/.env`, `$HOME/.juleson.env`,
and `/etc/juleson/.env`. `JULES_API_KEY` is accepted directly from the
environment. GitHub and Gemini commands read their saved config values, except
where a command documents a flag or setup-specific environment fallback.

See [Configuration](docs/CONFIGURATION.md) and [Setup](docs/SETUP_GUIDE.md).

## Common Commands

```bash
# Inspect available commands
juleson --help

# Work with Jules sources and sessions
juleson sources list
juleson sessions list
juleson sessions create sources/github/owner/repo "Fix failing tests"
juleson sessions create . --prompt-file task.md
juleson sessions create --no-source "Draft a migration plan"
juleson sessions batch sources/github/owner/repo task.md --parallel 3 --group-title "Fix CI"
juleson sessions watch SESSION_ID --follow-activities --cursor-output .juleson.cursor
juleson sessions approve SESSION_ID
juleson sessions artifacts list SESSION_ID
juleson sessions outputs SESSION_ID
juleson sessions preview SESSION_ID
juleson sessions apply SESSION_ID ./path/to/project --activity-id ACTIVITY_ID --artifact-index 0
juleson sessions apply SESSION_ID ./path/to/project --confirm
juleson official remote pull SESSION_ID

# Manage templates
juleson template list
juleson template show test-generation
juleson execute template test-generation ./path/to/project

# GitHub integration
juleson github status
juleson github current
juleson github repos
juleson pr list

# Development workflow for this repository
juleson dev build --all
juleson dev test --short
juleson dev check
```

See [CLI Reference](docs/CLI_REFERENCE.md) for the command map and flags.

## MCP Server

Start the MCP server:

```bash
jules-mcp
```

The server uses stdio transport. Configure clients with the absolute path to the
`jules-mcp` binary. Put credentials in `juleson.yaml`; `JULES_API_KEY` may also
be supplied through the client environment.

See [MCP Server Usage](docs/MCP_SERVER_USAGE.md).

## Documentation

- [Documentation Index](docs/README.md)
- [CLI Reference](docs/CLI_REFERENCE.md)
- [Installation](docs/INSTALLATION_GUIDE.md)
- [Setup](docs/SETUP_GUIDE.md)
- [Configuration](docs/CONFIGURATION.md)
- [MCP Server Usage](docs/MCP_SERVER_USAGE.md)
- [Jules API Notes](docs/JULES_API.md)
- [Templates](docs/TEMPLATES.md)
- [GitHub Integration](docs/GITHUB_INTEGRATION.md)
- [Event System](docs/EVENT_SYSTEM_ARCHITECTURE.md)
- [Testing](docs/TESTING_GUIDE.md)
- [Roadmap](docs/ROADMAP.md)
- [Changelog](docs/CHANGELOG.md)

## Development

```bash
go mod download
go test ./...
go run ./cmd/juleson --help
```

Go applications can import the SDK package directly:

```go
client := jules.NewClient(
    "api-key",
    jules.WithBaseURL("https://jules.googleapis.com/v1alpha"),
    jules.WithRetryAttempts(3),
)
```

The SDK uses typed Jules states and `time.Time` timestamps, follows documented
session/source/activity endpoints, and keeps local artifact writing or patch
application in internal app code.

The CI workflow runs formatting, module consistency, tests, linting, security
scans, and builds on Linux, macOS, and Windows. Markdown-only changes are ignored
by CI.

## License

MIT. See [LICENSE](LICENSE).
