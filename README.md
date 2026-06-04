# Juleson

Juleson is a Go command-line tool and MCP server for operating Google's Jules
coding-agent sessions. It provides local commands for source discovery, session
management, Jules-created pull request review, template management, and
development tasks. The reusable Jules REST API client is available as
`github.com/SamyRai/go-jules`.

Release assets install two executable names:

- `juleson`: user-facing CLI.
- `jsn`: short alias for the same CLI.

The repository also contains `builder`, an internal build/test/release helper.
Juleson keeps the Jules API, operator workflow, and MCP server in this module;
general AI orchestration belongs outside this repository.

## Requirements

- Go 1.25.11 or newer for source builds. Release builds use Go 1.26.4.
- A Jules API key for commands that call the Jules API.
- A GitHub token only when inspecting or merging Jules-created pull requests.

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
```

Build locally:

```bash
go build -o bin/builder ./cmd/builder
./bin/builder build
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
environment. GitHub configuration is used only for Jules-created pull request
context.

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

# Jules-created pull request context
juleson pr list

# MCP server
juleson mcp serve --version
juleson mcp serve

# Development workflow for this repository
juleson dev build --all
juleson dev test --short
juleson dev check
```

See [CLI Reference](docs/CLI_REFERENCE.md) for the command map and flags.

## MCP Server

Juleson serves MCP over stdio:

```bash
juleson mcp serve
```

Use `gh`, GitHub's CLI, or the official GitHub MCP server for general GitHub and
Actions operations.

## Documentation

- [Documentation Index](docs/README.md)
- [CLI Reference](docs/CLI_REFERENCE.md)
- [Installation](docs/INSTALLATION_GUIDE.md)
- [Setup](docs/SETUP_GUIDE.md)
- [Configuration](docs/CONFIGURATION.md)
- [Jules API Notes](docs/JULES_API.md)
- [Templates](docs/TEMPLATES.md)
- [GitHub Integration](docs/GITHUB_INTEGRATION.md)
- [Event System](docs/EVENT_SYSTEM_ARCHITECTURE.md)
- [Testing](docs/TESTING_GUIDE.md)
- [Roadmap](docs/ROADMAP.md)
- [Changelog](docs/CHANGELOG.md)

The same documentation set is published to the project Wiki for GitHub-native
navigation. The repository `docs/` directory remains the source of truth.

## Development

```bash
go mod download
go test ./...
go run ./cmd/juleson --help
```

Go applications can import the SDK package directly:

```go
import jules "github.com/SamyRai/go-jules"

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
by CI, so documentation-only changes should be checked locally with
`markdownlint '**/*.md'`.

## License

MIT. See [LICENSE](LICENSE).
