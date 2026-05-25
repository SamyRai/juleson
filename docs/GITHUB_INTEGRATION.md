# GitHub Integration

Juleson uses `github.com/google/go-github/v76` for repository, pull request,
Actions, and session-related GitHub operations.

## Configure

Set a GitHub token:

```bash
export GITHUB_TOKEN="..."
juleson github status
```

The token is used by GitHub CLI commands and by MCP GitHub tools. Required scopes
depend on the operation:

- repository read access for repository discovery
- pull request write access for merge operations
- workflow access for Actions commands

## CLI Commands

```bash
juleson github login
juleson github status
juleson github repos
juleson github current
juleson github search "org:example language:go"

juleson pr list
juleson pr get SESSION_ID
juleson pr diff SESSION_ID
juleson pr merge SESSION_ID --method squash

juleson actions workflows list owner/repo
juleson actions runs list owner/repo
juleson actions jobs list RUN_ID owner/repo
juleson actions artifacts list owner/repo
juleson actions cache list owner/repo
```

## Package Layout

`internal/github` is split by responsibility:

- `client.go`: client facade and shared dependencies.
- `repositories.go`: repository listing, search, and current-repo detection.
- `pullrequests.go`: PR lookup, diff, and merge operations.
- `actions.go`: workflows, runs, jobs, artifacts, and caches.
- `sessions.go`: Jules session helpers with GitHub context.
- `git.go`: remote URL parsing.
- `types.go`: domain types.

## MCP Tools

GitHub MCP tools are registered only when both the Jules client and GitHub token are available:

- `github_create_session_from_repo`
- `github_merge_session_pr`
- `github_list_repos`
- `github_current_repo`
- `github_list_connected_repos`
- `github_search_repos`
