# Roadmap

Juleson is currently a pre-1.0 project. This roadmap records planned work and
known gaps without committing to release dates.

## Current Sprint Objective

Restore Juleson MCP support while enforcing a tighter single-responsibility
boundary.

Window: 2026-06-04 through 2026-06-18

Acceptance criteria:

- Serve MCP through `juleson mcp serve` and `jsn mcp serve` using the official
  Go MCP SDK.
- Package `juleson` and `jsn`; do not restore a separate `jules-mcp` binary.
- Remove standalone GitHub and Actions command surfaces from Juleson.
- Keep GitHub support only for Jules-connected source inference and
  Jules-created pull request context.
- Expose Jules-focused MCP tools for sources, sessions, activities, plans,
  reviews, artifacts, outputs, and developer build/test/check workflows.
- Keep mutating MCP tools behind explicit confirmation fields.
- Update CLI, install, deployment, setup, MCP, and testing docs with the code.
- Verify with `go test ./...`, `go vet ./...`, `juleson mcp serve --version`,
  and `juleson dev build --all`.

## Near Term

- Add MCP command-transport tests that call `initialize`, `tools/list`, and a
  representative tool over stdio.
- Continue moving CLI printing code behind structured service functions where
  MCP needs the same behavior.
- Improve Jules API error messages with request context and safe secret
  redaction.
- Expand unit coverage for `internal/github`, `internal/events`, and CLI command
  behavior that remains in Juleson.

## Jules Operator Workflow

- Harden long-running Jules session tracking: status-change wakeups, Jules agent
  message wakeups, resumable activity cursors, and clear next-action reasons.
- Reduce operator risk around patch application: cleaner preview summaries,
  scoped artifact application, and verification guidance before mutation.
- Keep delivery measurable with focused tests for session watches, activity
  filtering, review snapshots, dirty-worktree guards, and base-commit mismatch
  handling.

## GitHub Boundary

- Keep `pr` commands focused on pull requests created by Jules sessions.
- Avoid rebuilding GitHub Actions, repository search, or general PR management.
  Use `gh`, GitHub's CLI, or the official GitHub MCP server for those workflows.
- Keep remote parsing and source inference test coverage strong because that is
  part of the Jules source workflow.

## Documentation

- Keep root Markdown limited to project essentials.
- Keep `docs/README.md` as the documentation index.
- Update docs in the same change as user-facing CLI, config, MCP, install, or
  workflow behavior changes.
