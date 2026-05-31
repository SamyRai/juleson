# Roadmap

Juleson is currently a pre-1.0 project. This roadmap records planned work and
known gaps without committing to release dates.

## Near Term

- Keep CLI, setup, installation, and MCP docs aligned with the code.
- Improve CLI error messages with request context.
- Add debug logging for Jules API requests and responses, with secret redaction.
- Expand unit coverage for `internal/github`, `internal/events`, and CLI command behavior.

## Next Jules Sprint Objective

Ship operator-grade Jules session monitoring and patch review workflows.
Long-running sessions should produce useful status reflections without waking
operators unnecessarily, while approval requests, feedback requests, completion,
failures, and outputs should wake the caller with a clear next action.

Window: 2026-05-27 through 2026-06-09

Acceptance criteria:

- Finish the actionable watch policy work across CLI: progress states
  should be reflected without waking by default, while user-action, terminal, and
  output states wake with structured reasons.
- Preserve compatibility for existing `--wake-on-status-change` callers by
  mapping it to an explicit any-status wake policy.
- Add an operator review command or documented flow that combines session
  status, artifacts, outputs, dry-run patch summary, base commit checks, and
  suggested verification commands before any mutation.
- Make timeout behavior resumable by returning the latest state, next activity
  cursor, update type, and next action.
- Keep patch application gated by clean worktree and explicit confirmation.
- Cover CLI and `internal/sessionops` behavior with focused tests for
  progress-only updates, approval wakeups, completion wakeups, failures, outputs,
  agent messages, timeouts, and cursor continuity.
- Update `docs/CLI_REFERENCE.md`, `docs/MCP_SERVER_USAGE.md`, and this roadmap
  alongside user-visible workflow changes.
- Verify with `go test ./...`, `git diff --check`, and a local `juleson sessions
  watch --help` smoke command.

## Next Sprint Track

- Harden long-running Jules session tracking: status-change wakeups, Jules agent
  message wakeups, resumable activity cursors, and clear next-action reasons.
- Finish agent-loop production readiness: dry-run parity, checkpoint resume
  behavior, and consistent plan approval gates across CLI and orchestration.
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
