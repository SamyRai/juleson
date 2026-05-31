# Builder Architecture

The builder package owns build, test, quality, dependency, runtime,
install, release, and Docker command behavior. CLI and MCP tools are adapters:
they bind flags or JSON schemas, map inputs into builder requests, call the
owner, and present the returned result.

## Entry Points

- `cmd/builder`: internal CLI wrapper.
- `internal/builder`: service implementation.
- `internal/mcp/tools/dev.go`: MCP developer tool adapter.
- `internal/mcp/tools/docker.go`: MCP Docker tool adapter.
- `internal/cli/commands/dev.go`: `juleson dev` command adapter.

## Service Layout

- `builder.go`: interface, config, constructor, and core project build operations.
- `command_runner.go`: command execution seam for owner-level tests.
- `dev_workflows.go`: shared dev/build/test/lint/format/clean/module/install
  and release workflows used by CLI and MCP.
- `test.go`: tests and coverage.
- `quality.go`: linting, formatting, and combined checks.
- `deps.go`: Go module commands.
- `run.go`: install, run, and dev helpers.
- `docker.go`: project Docker workflows for the internal builder CLI.
- `docker_tools.go`: generic Docker tool operations used by MCP Docker handlers.

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
- Do not add `os/exec`, `exec.Command*`, or `internal/build` imports to dev/build/Docker adapters.
- Return errors with enough command context for callers to display useful messages.
- Avoid adding one-off shell command paths outside the builder for existing responsibilities.

## Agent Orchestration Boundary

Agent and automation workflows are separate from build/dev/Docker orchestration.
Their extraction unit is `internal/orchestration`:

- `domain` owns pure orchestration concepts.
- `ports` owns interfaces consumed by orchestration.
- `app` owns state machines, scheduling, progress, and decision routing.
- `adapters` owns Jules, Gemini, analyzer, template, tool, checkpoint, memory, and other concrete systems.

`internal/services.Container` builds the runtime through
`orchestration.NewRuntime`. New CLI and MCP orchestration paths should use that
runtime rather than directly constructing legacy agent or automation internals.
