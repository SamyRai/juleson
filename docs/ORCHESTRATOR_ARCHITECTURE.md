# Orchestrator Architecture

The orchestrator package centralizes build, test, quality, dependency, runtime,
install, release, and Docker operations. CLI and MCP tools call this service
instead of duplicating shell command logic.

## Entry Points

- `cmd/orchestrator`: internal CLI wrapper.
- `internal/orchestrator`: service implementation.
- `internal/mcp/tools/dev.go`: MCP developer tools.
- `internal/cli/commands/dev.go`: `juleson dev` commands.

## Service Layout

- `orchestrator.go`: interface, config, constructor, and build operations.
- `test.go`: tests and coverage.
- `quality.go`: linting, formatting, and combined checks.
- `deps.go`: Go module commands.
- `run.go`: install, run, and dev helpers.
- `docker.go`: Docker build and container helpers.

## CLI Mapping

```bash
juleson dev build
juleson dev test
juleson dev lint
juleson dev fmt
juleson dev clean
juleson dev mod tidy
juleson dev check
juleson dev install
juleson dev release
```

## MCP Mapping

- `build_project`
- `run_tests`
- `lint_code`
- `format_code`
- `clean_artifacts`
- `quality_check`
- `module_maintenance`
- `build_release`

## Design Notes

- Keep command construction in the orchestrator service.
- Keep CLI and MCP handlers thin.
- Return errors with enough command context for callers to display useful messages.
- Avoid adding one-off shell command paths outside the orchestrator for existing responsibilities.
