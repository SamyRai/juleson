# CLI Reference

This reference reflects the current Cobra command tree from `cmd/juleson`.
Juleson focuses on Jules sources, sessions, session artifacts, Jules-created pull
request context, templates, local project sync, developer workflows, and the
integrated MCP server.

## Global Usage

```bash
juleson [command]
jsn [command]
```

`jsn` is the short installed alias for the same CLI.

Available commands:

| Command | Purpose |
| --- | --- |
| `activities` | Manage Jules session activities |
| `completion` | Generate shell completion scripts |
| `config` | Manage Juleson configuration |
| `dev` | Build, test, lint, format, and release helpers |
| `init` | Initialize a project for Jules automation |
| `mcp` | Run the Juleson MCP server |
| `official` | Bridge to the official Jules CLI when installed |
| `pr` | Manage pull requests created by Jules sessions |
| `sessions` | Manage Jules sessions |
| `setup` | Run first-time setup |
| `sources` | Manage Jules sources |
| `sync` | Sync a project with a remote repository |
| `template` | Manage templates |
| `version` | Print version information |

## Config And Setup

```bash
juleson config validate
juleson setup [flags]
```

`config validate` validates the effective configuration and reports missing
credentials as warnings. It never prints API keys or other secrets.

Flags:

```text
--non-interactive   Run setup without prompts
--skip-completion   Skip shell completion installation
--skip-github       Skip GitHub token configuration for Jules-created PR context
--skip-jules        Skip Jules API configuration
```

## Sources And Sessions

```bash
juleson sources list
juleson sources get SOURCE_ID

juleson sessions list
juleson sessions status
juleson sessions create SOURCE_ID "Prompt text" --require-plan-approval
juleson sessions create . --prompt-file task.md --title "Fix failing tests"
juleson sessions create --no-source "Prompt text"
juleson sessions batch SOURCE_ID task.md --parallel 3 --batch-id batch-20260525 --group-title "Fix CI"
juleson sessions watch SESSION_ID --follow-activities --since 2026-05-25T10:00:00Z --cursor-output .juleson.cursor
juleson sessions watch SESSION_ID --wake-policy actionable
juleson sessions watch SESSION_ID --wake-on-status-change --initial-state PLANNING
juleson sessions watch SESSION_ID --wake-on-agent-message --since 2026-05-25T10:00:00Z
juleson sessions get SESSION_ID
juleson sessions plans SESSION_ID
juleson sessions plans SESSION_ID --latest --json
juleson sessions review SESSION_ID PROJECT_PATH
juleson sessions review SESSION_ID PROJECT_PATH --activity-id ACTIVITY_ID --artifact-index 0 --json
juleson sessions approve SESSION_ID
juleson sessions message SESSION_ID "Follow-up text"
juleson sessions apply SESSION_ID PROJECT_PATH
juleson sessions apply SESSION_ID PROJECT_PATH --activity-id ACTIVITY_ID --artifact-index 0
juleson sessions apply SESSION_ID PROJECT_PATH --confirm --allow-base-mismatch
juleson sessions artifacts list SESSION_ID
juleson sessions outputs SESSION_ID
juleson sessions delete SESSION_ID --force
juleson sessions preview SESSION_ID
juleson sessions preview-activity SESSION_ID ACTIVITY_ID
juleson sessions download SESSION_ID OUTPUT_DIR
juleson sessions download-activity SESSION_ID ACTIVITY_ID OUTPUT_DIR

juleson activities list SESSION_ID
juleson activities list SESSION_ID --since 2026-05-25T10:00:00Z --cursor-output .juleson.cursor
juleson activities get SESSION_ID ACTIVITY_ID
```

`sessions create` accepts either `github/owner/repo` or
`sources/github/owner/repo`. Passing `.` asks Juleson to infer the connected
Jules source from the local git `origin` remote. `--no-source` creates a
repoless Jules session by omitting `sourceContext`.

`sessions watch` prints observed session status with an update type. By default,
`--wake-policy actionable` returns only when a session needs user action,
completes, fails, or surfaces session outputs. `--wake-on-status-change` remains
a compatibility alias for `--wake-policy any-status`.

`sessions review` is a read-only operator snapshot. It combines session state,
latest plan, documented outputs, artifact manifests, patch dry-run summary,
base-commit warnings, dirty-worktree blockers, verification suggestions, and
safe next actions.

`sessions apply` dry-runs by default. Use `--confirm` to apply patches; dirty
worktrees are blocked unless `--allow-dirty` is passed. If an artifact includes
`baseCommitId`, real apply blocks on mismatch unless `--allow-base-mismatch` is
passed.

## Jules-Created Pull Requests

Juleson keeps pull request support only where the PR is connected to a Jules
session output.

```bash
juleson pr list --limit 10
juleson pr get SESSION_ID
juleson pr diff SESSION_ID
juleson pr merge SESSION_ID --method squash
```

Use `gh`, GitHub's own CLI, or the official GitHub MCP server for general
repository, Actions, and pull request operations.

## MCP

```bash
juleson mcp serve
juleson mcp serve --version
jsn mcp serve
```

The MCP server runs over stdio and exposes Jules session, artifact, review, and
developer workflow tools. See [MCP Server Usage](MCP_SERVER_USAGE.md).

## Templates

```bash
juleson template list [category]
juleson template show TEMPLATE_NAME
juleson template search QUERY
juleson template create TEMPLATE_NAME CATEGORY DESCRIPTION
```

## Project And Git Sync

```bash
juleson init [project-path]
juleson sync [project-path] [remote] --branch main --pull
juleson sync [project-path] [remote] --branch main --push
juleson official remote new --parallel 3
juleson official remote pull SESSION_ID
juleson official tui
```

## Development Commands

```bash
juleson dev build [--all|--cli|--alias] [--race] [--version dev]
juleson dev test [--race] [--cover] [--short] [--run PATTERN]
juleson dev lint [--fix] [--fast] [--timeout 5m]
juleson dev fmt [--gofumpt]
juleson dev clean [--all|--cache|--modcache|--testcache]
juleson dev mod tidy
juleson dev mod download
juleson dev mod verify
juleson dev mod vendor
juleson dev mod graph
juleson dev mod why PACKAGE
juleson dev deps [path]
juleson dev check-complexity [path]
juleson dev check
juleson dev install [--path DIR] [--skip-checks]
juleson dev release --version VERSION
```

## Environment Variables

- `JULES_API_KEY`: accepted directly by config loading and required for Jules API calls.
- `GITHUB_TOKEN`: read by setup and used only for Jules-created PR context.

Other settings should be configured in `juleson.yaml`.
