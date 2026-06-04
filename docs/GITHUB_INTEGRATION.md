# GitHub Boundary

Juleson is not a general GitHub or Actions client. Use `gh`, GitHub's CLI, or
the official GitHub MCP server for repository search, workflow runs, jobs, logs,
artifacts, caches, and general pull request management.

Juleson keeps GitHub support only where it directly belongs to a Jules workflow:

- inferring a connected Jules source from the local git `origin` remote
- listing or inspecting pull requests created by Jules sessions
- merging a Jules-created pull request when the session output identifies one
- showing GitHub links embedded in Jules session outputs

## Configure

Set a GitHub token through setup or by writing `github.token` in
`juleson.yaml`:

```bash
export GITHUB_TOKEN="..."
juleson setup --non-interactive
```

Pull request commands require repository access to the target Jules-created PR.

## CLI Commands

```bash
juleson pr list
juleson pr get SESSION_ID
juleson pr diff SESSION_ID
juleson pr merge SESSION_ID --method squash
```

## Package Layout

`internal/github` is scoped to Jules workflow context:

- `client.go`: client facade and shared dependencies.
- `repositories.go`: repository metadata used by source/session helpers.
- `pullrequests.go`: Jules-created PR lookup, diff, and merge operations.
- `sessions.go`: Jules session helpers with GitHub context.
- `git.go`: remote URL parsing.
- `types.go`: domain types.
