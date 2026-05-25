# Setup Guide

## Setup Wizard

```bash
juleson setup
```

The setup command can install shell completion, configure a Jules API key,
configure GitHub integration, and validate the resulting config.

Flags:

```text
--non-interactive   Run setup without prompts
--skip-completion   Skip shell completion installation
--skip-github       Skip GitHub configuration
--skip-jules        Skip Jules API configuration
```

## Non-Interactive Setup

```bash
export JULES_API_KEY="..."
export GITHUB_TOKEN="..."
juleson setup --non-interactive
```

Non-interactive setup reads `JULES_API_KEY` and, when present,
`GITHUB_TOKEN`, then writes those values to `configs/juleson.yaml`.

## Manual Setup

Create `configs/juleson.yaml`:

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
```

Then set `JULES_API_KEY` through the environment or fill `jules.api_key` in the
config file. Put GitHub and Gemini credentials in `juleson.yaml` for commands
and MCP tools that need them:

```bash
export JULES_API_KEY="..."
```

## Shell Completion

```bash
juleson completion bash
juleson completion zsh
juleson completion fish
juleson completion powershell
```

For Zsh:

```bash
source <(juleson completion zsh)
```

For Fish:

```bash
juleson completion fish > ~/.config/fish/completions/juleson.fish
```

## Verify

```bash
juleson version
juleson github status
juleson sources list
```

Commands that call Jules require a Jules API key. GitHub commands require
`github.token` in config.

See [Configuration](CONFIGURATION.md) for all config paths and defaults.
