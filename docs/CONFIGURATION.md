# Configuration

Juleson loads configuration with Viper from config files, `.env` files, and
defaults. `JULES_API_KEY` is also accepted directly when a Jules client is
required.

## Search Paths

The config file name is `juleson.yaml`. Search order:

1. `./configs`
2. `.`
3. `$HOME`
4. `/etc/juleson`

Juleson also loads environment files from:

1. `.env`
2. `$HOME/.env`
3. `$HOME/.juleson.env`
4. `/etc/juleson/.env`

## Minimal Config

```yaml
jules:
  api_key: ""
  base_url: "https://jules.googleapis.com/v1alpha"
  timeout: "30s"
  retry_attempts: 3
  debug_log: false

github:
  token: ""
  default_org: ""
  pr:
    default_merge_method: "squash"
    auto_delete_branch: true
  discovery:
    enabled: true
    use_git_remote: true
    cache_ttl: "5m"

projects:
  default_path: "./projects"
  backup_enabled: true
  git_integration: true

templates:
  builtin_path: "./templates/builtin"
  custom_path: "./templates/custom"
  enable_custom: true

diff:
  tool: ""
  force_native: false
```

## Environment Variables

- `JULES_API_KEY`: used as a fallback for `jules.api_key`.
- `GITHUB_TOKEN`: read by `juleson setup --non-interactive` and saved into config.

GitHub configuration is used only for Jules-connected source discovery and
Jules-created pull request context. Use `gh`, GitHub's CLI, or the official
GitHub MCP server for general GitHub operations.

## Validation

`juleson` uses optional config loading for local commands. Commands that call the
Jules API still require `JULES_API_KEY` or `jules.api_key`. You can validate the
current configuration state safely using:

```bash
juleson config validate
```

The MCP server starts with minimal config. Tools that require Jules credentials
return credential errors when the API key is missing; they do not prompt for or
print secrets.

The Go SDK at `github.com/SamyRai/go-jules` does not load this configuration
directly. Applications pass credentials and options explicitly with
`jules.NewClient` and client options such as `jules.WithBaseURL`,
`jules.WithTimeout`, and `jules.WithRetryAttempts`.
