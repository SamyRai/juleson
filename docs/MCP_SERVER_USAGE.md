# MCP Server Usage

`jules-mcp` starts a Model Context Protocol server over stdio using the official Go SDK.

```bash
jules-mcp
```

Logs are written to stderr. MCP JSON-RPC messages use stdout.

## Client Configuration

Use an absolute binary path in MCP clients.

Claude Desktop example:

```json
{
  "mcpServers": {
    "juleson": {
      "command": "/absolute/path/to/jules-mcp",
      "env": {
        "JULES_API_KEY": "..."
      }
    }
  }
}
```

GitHub and Gemini MCP tools are registered from `github.token` and
`gemini.api_key` in `juleson.yaml`. Run setup or edit the config file before
starting the MCP client when those tools are needed.

## Tool Registration

Tools are registered based on available configuration:

- Developer, code intelligence, Docker, project, and template tools are always registered.
- Session and activity tools require a Jules client.
- GitHub tools require a Jules client and `github.token` in config.
- Gemini tools require a Jules client and `gemini.api_key` in config.

## Always-Registered Tools

Development:

- `build_project`
- `run_tests`
- `lint_code`
- `format_code`
- `clean_artifacts`
- `quality_check`
- `module_maintenance`
- `build_release`

Code intelligence:

- `analyze_code_graph`
- `analyze_code_context`
- `find_symbol_references`
- `run_static_analysis`
- `analyze_complexity`

Docker:

- `docker_build`
- `docker_run`
- `docker_images`
- `docker_containers`
- `docker_stop`
- `docker_remove`
- `docker_rmi`
- `docker_prune`
- `docker_exec`

Project and templates:

- `analyze_project`
- `sync_project`
- `execute_template`
- `list_templates`
- `search_templates`
- `create_template`

## Jules Tools

Registered when the Jules client is available:

- `list_sessions`
- `get_session_status`
- `list_sources`
- `get_source`
- `infer_source_for_project`
- `approve_session_plan`
- `apply_session_patches`
- `preview_session_changes`
- `list_session_artifacts`
- `get_session_outputs`
- `watch_session`
- `verify_session_changes`
- `send_session_message`
- `create_session`
- `get_session`
- `delete_session`
- `list_session_activities`
- `get_session_activity`
- `search_session_activities`
- `get_session_plans`

`create_session.source` is optional. Omit it for a repoless session, or pass a
source such as `sources/github/owner/repo` for a source-backed session.
`create_session.prompt_file` reads the prompt from a local file and is mutually
exclusive with `prompt`. `infer_source_for_project` can resolve a local git
`origin` remote to a connected Jules source before session creation.
For GitHub-backed sources, `starting_branch` is optional; when omitted, Juleson
reads source metadata and uses the connected repository's default branch.

`apply_session_patches` dry-runs unless `confirm_apply=true`. Actual apply also
checks for a clean worktree unless `allow_dirty=true` is passed. Use
`preview_session_changes` first and `verify_session_changes` after applying.
Both preview and apply accept `activity_id` and `artifact_index` to scope a
single changeset. Patch artifacts with `baseCommitId` warn during dry-run and
block mutation on mismatch unless `allow_base_mismatch=true`.

`watch_session.since` accepts an RFC3339 activity `createTime` cursor, filters
activities client-side, and returns `next_activity_cursor` for resumable watches.
It returns when a session reaches a terminal state, needs user action, times out,
or exposes outputs.
Set `return_on_status_change=true` with optional `initial_state` to return on the
next session state transition. Set `return_on_jules_agent_message=true` to return
when Jules posts a new agent message after `since`; without `since`, the first
poll establishes the activity baseline. Returned watch wakeups include
`wake_reason`.

`list_session_artifacts` returns an artifact manifest containing activity ID,
artifact index, type, file count, changed files, base commit, suggested commit
message, media MIME type, bash command, and bash exit code.
`get_session_outputs` returns documented session outputs such as pull requests.
When a pull request output exists, inspect GitHub Actions with the existing
GitHub tools instead of duplicating Jules CI-fix behavior.

`verify_session_changes` detects Go, Node/Yarn, Python/uv, and Rust project
files and runs conservative repo-standard checks. Pass `command` only when the
caller explicitly wants a custom verification command.

`delete_session` requires `confirm=true` and calls the documented Jules API
delete endpoint. Session cancel is not exposed by the Jules API v1alpha
reference used by this project.

## GitHub Tools

Registered when GitHub config is available:

- `github_create_session_from_repo`
- `github_merge_session_pr`
- `github_list_repos`
- `github_current_repo`
- `github_list_connected_repos`
- `github_search_repos`

## Gemini Tools

Registered when Gemini config is available:

- `plan_project_automation`
- `orchestrate_workflow`
- `manage_github_project`
- `synthesize_session_results`

## Resources And Prompts

The server registers:

- `server-info`
- `config-template`
- `analyze-project-workflow`
- `session-management-guide`

Use these for client-side discovery and guided workflows.
