# Roadmap

Juleson is currently a pre-1.0 project. This roadmap records planned work and
known gaps without committing to release dates.

## Near Term

- Keep CLI, setup, installation, and MCP docs aligned with the code.
- Add HTTP transport as an optional MCP server mode.
- Improve MCP and CLI error messages with request context.
- Add debug logging for Jules API requests and responses, with secret redaction.
- Expand unit coverage for `internal/github`, `internal/events`, and CLI command behavior.
- Add a config validation command.

## Analysis And Code Intelligence

- Add dependency graph reporting and unused dependency detection for Go projects.
- Add license and vulnerability reporting based on Go module data.
- Add Go coverage parsing for package-level gap reporting.
- Add code complexity reports that combine call graph, complexity, and churn data.

## Agent And Orchestration

- Continue tightening the agent loop around explicit goals, constraints, tool
  execution, review, and checkpointing.
- Make dry-run behavior consistent across CLI, MCP, and agent flows.
- Improve plan approval handling and status reporting for long-running Jules sessions.
- Add safer defaults for operations that can modify working trees.

## GitHub Integration

- Add more tests around GitHub service boundaries and remote parsing.
- Add issue and milestone operations where they are already supported by the internal service layer.
- Improve Actions command output for failed workflow runs and job logs.

## Documentation

- Keep root Markdown limited to project essentials.
- Keep `docs/README.md` as the documentation index.
- Update docs in the same change as user-facing CLI, config, or workflow changes.
