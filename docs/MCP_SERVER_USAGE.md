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
        "JULES_API_KEY": "...",
        "GITHUB_TOKEN": "...",
        "GEMINI_API_KEY": "..."
      }
    }
  }
}
```

## Tool Registration

Tools are registered based on available configuration:

- Developer, code intelligence, Docker, project, and template tools are always registered.
- Session and activity tools require a Jules client.
- GitHub tools require a Jules client and GitHub token.
- Gemini tools require a Jules client and Gemini API key.

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
- `approve_session_plan`
- `apply_session_patches`
- `preview_session_changes`
- `send_session_message`
- `create_session`
- `get_session`
- `list_session_activities`
- `get_session_activity`
- `search_session_activities`
- `get_session_plans`

The server does not expose cancel/delete session tools because those operations
are not available through the Jules API used by this project.

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
