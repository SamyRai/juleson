# Builder Architecture

The builder package owns build, test, quality, dependency, runtime, install,
release, and Docker command behavior. CLI and MCP tools are adapters: they bind
flags or JSON schemas, map inputs into builder requests, call the owner, and
present the returned result.

## Entry Points

- `cmd/builder`: internal CLI wrapper.
- `pkg/builder`: project workflow service.
- `internal/presentation/cli/dev`: `juleson dev` command adapter.
- `internal/mcp`: MCP developer workflow adapter.

## Service Layout

- `builder.go`: config, constructor, and core project build operations.
- `command_runner.go`: command execution helper for owner-level tests.
- `dev_workflows.go`: shared dev/build/test/lint/format/clean/module/install
  and release workflows used by CLI and MCP.
- `test.go`: tests and coverage.
- `quality.go`: linting, formatting, and combined checks.
- `deps.go`: Go module commands.
- `run.go`: install, run, and dev helpers.
- `docker.go`: project Docker workflows for the internal builder CLI.

## Binary Targets

`juleson dev build --all` builds two executable names from `./cmd/juleson`:

- `juleson`
- `jsn`

MCP is served through `juleson mcp serve`; there is no separate MCP binary.

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

- `dev_build`
- `dev_test`
- `dev_check`

## Design Notes

- Keep command construction in the builder service.
- Keep CLI and MCP handlers thin.
- Do not add one-off `os/exec` paths outside the builder for existing responsibilities.
- Return errors with enough command context for callers to display useful messages.

## Orchestration Boundary

Gemini-backed planning, autonomous file patching, and broader agent loops live
in `go-agent`. Juleson stays focused on Jules API operator workflows, session
artifacts, patch review, and local development helpers.
