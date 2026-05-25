# Changelog

The project is pre-1.0. Changes are tracked here until a tagged release process is established.

## Unreleased

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
