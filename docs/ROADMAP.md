# Roadmap

Juleson is currently a pre-1.0 project. This roadmap records planned work and
known gaps without committing to release dates.

## Near Term

- Keep CLI, setup, installation, and MCP docs aligned with the code.
- Add HTTP transport as an optional MCP server mode.
- Improve MCP and CLI error messages with request context.
- Add debug logging for Jules API requests and responses, with secret redaction.
- Expand unit coverage for `internal/github`, `internal/events`, and CLI command behavior.

## Next Jules Sprint Objective

Add opt-in debug logging for Jules API requests and responses with secret
redaction. Operators should be able to troubleshoot Jules API behavior without
ever exposing API keys, bearer tokens, or other credential material in logs.

Acceptance criteria:

- Keep logging opt-in through configuration, environment, or an explicit client
  option; default behavior should remain quiet.
- Log method, URL path, status code, duration, retry attempt context, and concise
  error details for Jules API calls.
- Redact `X-Goog-Api-Key`, `Authorization`, API key query params, and obvious
  token-like values in request/response metadata before logging.
- Avoid logging full response bodies by default; if body snippets are added,
  bound their size and apply redaction first.
- Reuse existing `slog` patterns where practical and keep ownership close to
  `github.com/SamyRai/go-jules` client request handling.
- Cover redaction, disabled-by-default behavior, enabled request/response
  logging, retry logging, and error-path logging with focused tests.
- Update `docs/CONFIGURATION.md` and `docs/JULES_API.md` if new user-visible
  options or troubleshooting workflows are added.
- Verify with `go test ./...`, `git diff --check`, and a local Juleson command
  that exercises the client without printing secrets.

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
