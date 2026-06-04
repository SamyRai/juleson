# Changelog

The project is pre-1.0. Changes are tracked here between tagged releases.

## Unreleased

### Added in v0.2.0

- MCP server support is restored through the official Go MCP SDK and served by
  `juleson mcp serve`.
- Release builds now include `jsn`, a short alias for the same CLI.
- MCP tools cover Jules source/session/activity operations, read-only review
  helpers, and repository developer build/test/check commands.

### Changed in v0.2.0

- General GitHub and Actions operations are no longer owned by Juleson. Use
  `gh`, GitHub's CLI, or the official GitHub MCP server for that scope.
- Juleson keeps only Jules-created pull request context through `juleson pr`.
- Release, install, Docker, CI, and docs flows now use `juleson` plus `jsn`
  instead of a separate `jules-mcp` binary.

## v0.1.1 - 2026-05-27

### Added in v0.1.1

- Native session plan inspection through `juleson sessions plans`, including
  full plan steps, approval state, latest-plan filtering, and JSON output.
- Read-only session review through `juleson sessions review`, combining session
  state, plans, outputs, artifact manifests, patch dry-run results, worktree
  blockers, and safe next-action suggestions.
- MCP `review_session` and structured `plans` output on `get_session_plans` for
  parity with the native CLI workflow.

### Changed in v0.1.1

- `activities list` now prints activity IDs and resource names for direct reuse
  with `activities get`, scoped review, and scoped patch apply commands.
- `sessions get` now points operators to `sessions plans` for complete plan
  details while keeping its concise preview behavior.

## v0.1.0 - 2026-05-26

### Added

- Public Go SDK package at `github.com/SamyRai/go-jules` with
  option-based client construction.
- Internal Jules CLI command composer for core Jules commands.
- SDK boundary and CLI composition tests.
- SDK contract fixtures for sessions, sources, activities, timestamps,
  repoless sessions, and embedded artifacts.
- GitHub service split for repositories, pull requests, actions, sessions, and remote parsing.
- GitHub CLI commands for login, status, repository discovery, search, and PR operations.
- Setup wizard with shell completion, Jules API configuration, optional GitHub
  configuration, and non-interactive mode.
- Shell completion generation for Bash, Zsh, Fish, and PowerShell.
- Gemini-backed orchestration tools and MCP tools gated by Jules and Gemini configuration.
- Installer scripts for Linux, macOS, and Windows release assets.
- MCP command-transport E2E coverage and installer tests.
- Jules API session delete support in the client, CLI, and MCP server.
- Repoless Jules session creation through CLI `--no-source` and optional MCP
  `create_session.source`.

### Changed

- The reusable Go SDK moved to the standalone public module
  `github.com/SamyRai/go-jules`; Juleson now consumes it as an external
  dependency.
- Local artifact download and patch application behavior moved out of the SDK
  into internal app operations.
- CLI configuration can load without a Jules API key for local commands such as
  `help` and `version`.
- Release assets package `juleson` and `jules-mcp` separately per OS and architecture.
- CI uses Go module version discovery, race tests, linting, security scans, and release validation.
- Activity filtering uses the documented `createTime` API parameter; legacy
  type, status, plan, and artifact filters are applied client-side.
- SDK timestamps now use `time.Time`, session states and activity originators
  are typed, and undocumented artifact/content/analyze/search network calls were
  removed in favor of documented activity payloads.

### Notes

- Jules API session cancel is not exposed because the API reference used by this
  project does not provide that lifecycle operation.
- The Juleson application module path remains `github.com/SamyRai/juleson`.
