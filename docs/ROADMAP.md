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

## Next Jules Sprint Objective

Implement `juleson config validate` as a low-risk operator safety command.
The command should validate the effective Juleson configuration without printing
secret values, report clear next steps for missing Jules, GitHub, or Gemini
credentials, and distinguish hard validation failures from optional integration
warnings.

Acceptance criteria:

- Add the command to the existing Cobra command tree without breaking current
  `setup`, `sessions`, `sources`, or local `dev` commands.
- Reuse the existing `internal/config` loading and validation behavior where
  possible; keep validation ownership in the config package or a directly
  adjacent CLI handler.
- Never print API keys, tokens, or config file secret values.
- Cover success, missing optional credentials, invalid MCP port, and invalid
  automation concurrency with focused tests.
- Update `docs/CLI_REFERENCE.md` and `docs/CONFIGURATION.md` with the new
  command semantics.
- Verify with `go test ./...`, `go run ./cmd/juleson config validate`, and
  `go run ./cmd/juleson --help`.

## Next Sprint Track

- Harden long-running Jules session tracking: status-change wakeups, Jules agent
  message wakeups, resumable activity cursors, and clear next-action reasons.
- Finish agent-loop production readiness: dry-run parity, checkpoint resume
  behavior, and consistent plan approval gates across CLI, MCP, and orchestration.
- Reduce operator risk around patch application: cleaner preview summaries,
  scoped artifact application, and verification guidance before mutation.
- Keep delivery measurable with focused tests for session watches, activity
  filtering, agent dry-runs, checkpoint persistence, and dirty-worktree guards.

## Analysis And Code Intelligence

- Add dependency graph reporting and unused dependency detection for Go projects.
- Add license and vulnerability reporting based on Go module data.
- Add Go coverage parsing for package-level gap reporting.
- Add code complexity reports that combine call graph, complexity, and churn data.

## Agent And Orchestration

- Continue tightening the agent loop around explicit goals, constraints, review,
  memory, and persistent checkpoint adapters.
- Make dry-run behavior consistent across MCP and remaining non-agent flows.
- Improve status reporting for long-running Jules sessions.
- Add safer defaults for operations that can modify working trees beyond
  session-backed agent execution.

## GitHub Integration

- Add more tests around GitHub service boundaries and remote parsing.
- Add issue and milestone operations where they are already supported by the internal service layer.
- Improve Actions command output for failed workflow runs and job logs.

## Documentation

- Keep root Markdown limited to project essentials.
- Keep `docs/README.md` as the documentation index.
- Update docs in the same change as user-facing CLI, config, or workflow changes.
