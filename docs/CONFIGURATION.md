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

gemini:
  api_key: ""
  backend: "gemini-api"
  project: ""
  location: "us-central1"
  model: "gemini-2.0-flash"
  timeout: "30s"
  max_tokens: 8192

mcp:
  server:
    host: "localhost"
    port: 8080
  client:
    timeout: "10s"

automation:
  strategies: ["modular", "layered", "microservices"]
  max_concurrent_tasks: 5
  task_timeout: "300s"

projects:
  default_path: "./projects"
  backup_enabled: true
  git_integration: true

templates:
  builtin_path: "./templates/builtin"
  custom_path: "./templates/custom"
  enable_custom: true
```

## Environment Variables

Common direct environment variables:

- `JULES_API_KEY`: used as a fallback for `jules.api_key`.
- `GITHUB_TOKEN`: read by `juleson setup --non-interactive` and saved into config.
- `GEMINI_API_KEY`: read by `juleson ai-orchestrate` when `--gemini-key` is omitted.

For GitHub CLI commands, GitHub MCP tools, and Gemini MCP tools, store the token
or API key in `juleson.yaml` through `juleson setup`, `juleson github login`, or
manual config editing. Other nested settings such as `jules.base_url` should be
set in `juleson.yaml`.

## Validation

`juleson` uses optional config loading for local commands. Commands that call the
Jules API still require `JULES_API_KEY` or `jules.api_key`. You can validate
the current configuration state safely using `juleson config validate`, which
reports missing credentials as warnings and validates configuration fields like
MCP ports and automation concurrency constraints without exposing secret values.

`jules-mcp` starts with minimal config when the Jules API key is missing. Tools
that require Jules, GitHub, or Gemini configuration are skipped or fail with a
credential error. GitHub and Gemini MCP tool registration is based on values in
the loaded config object.

The Go SDK in `pkg/jules` does not load this configuration directly. Applications
pass credentials and options explicitly with `jules.NewClient` and client
options such as `jules.WithBaseURL`, `jules.WithTimeout`, and
`jules.WithRetryAttempts`. SDK-only options also include retry backoff,
custom `http.Client`, user agent, and sleep injection for deterministic tests.
