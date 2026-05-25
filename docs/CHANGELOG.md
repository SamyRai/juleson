# Changelog

The project is pre-1.0. Changes are tracked here until a tagged release process is established.

## Unreleased

### Added

- GitHub service split for repositories, pull requests, actions, sessions, and remote parsing.
- GitHub CLI commands for login, status, repository discovery, search, and PR operations.
- Setup wizard with shell completion, Jules API configuration, optional GitHub
  configuration, and non-interactive mode.
- Shell completion generation for Bash, Zsh, Fish, and PowerShell.
- Gemini-backed orchestration tools and MCP tools gated by Gemini configuration.
- Installer scripts for Linux, macOS, and Windows release assets.
- MCP command-transport E2E coverage and installer tests.
- Jules API session delete support in the client, CLI, and MCP server.
- Repoless Jules session creation through CLI `--no-source` and optional MCP
  `create_session.source`.

### Changed

- CLI configuration can load without a Jules API key for local commands such as
  `help` and `version`.
- Release assets package `juleson` and `jules-mcp` separately per OS and architecture.
- CI uses Go module version discovery, race tests, linting, security scans, and release validation.
- Activity filtering uses the documented `createTime` API parameter; legacy
  type, status, plan, and artifact filters are applied client-side.

### Notes

- Jules API session cancel is not exposed because the API reference used by this
  project does not provide that lifecycle operation.
- The public module path is `github.com/SamyRai/juleson`.
