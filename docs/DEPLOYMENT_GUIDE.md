# Deployment Guide

Deployment uses release assets, local installs, Docker images, or MCP client
configuration.

## Local Build

```bash
go mod download
go build -o bin/orchestrator ./cmd/orchestrator
./bin/orchestrator build
```

Individual binaries:

```bash
go build -o bin/juleson ./cmd/juleson
go build -o bin/jules-mcp ./cmd/jules-mcp
```

## Install Locally

```bash
./bin/orchestrator install
```

Or:

```bash
juleson dev install --path "$HOME/.local/bin"
```

## Release Assets

The release workflow builds:

- `juleson-linux-amd64.tar.gz`
- `juleson-linux-arm64.tar.gz`
- `juleson-darwin-amd64.tar.gz`
- `juleson-darwin-arm64.tar.gz`
- `juleson-windows-amd64.zip`
- `juleson-windows-arm64.zip`
- matching `jules-mcp` assets for each target
- `install.sh`
- `install.ps1`
- `checksums.txt`

Create a release by pushing a `v*.*.*` tag or using the workflow dispatch input.

## Go Module

After a non-prerelease tag, the release workflow asks the Go module proxy to index:

```bash
go list -m github.com/SamyRai/juleson@VERSION
```

## Docker

Use the repository `Dockerfile` for container builds:

```bash
docker build -t juleson:local .
```

Provide credentials through environment variables for `JULES_API_KEY` or through
mounted config files for full GitHub and Gemini settings. Do not bake API keys
into images.

## MCP Client Deployment

Install `jules-mcp` on the machine running the MCP client. Configure the client
with an absolute binary path. `JULES_API_KEY` can be supplied in the client
environment; GitHub and Gemini MCP tools require `github.token` and
`gemini.api_key` in `juleson.yaml`:

```json
{
  "mcpServers": {
    "juleson": {
      "command": "/usr/local/bin/jules-mcp",
      "env": {
        "JULES_API_KEY": "..."
      }
    }
  }
}
```

## Release Checklist

- `go mod tidy && git diff --exit-code go.mod go.sum`
- `go test ./...`
- `juleson dev build --all`
- `markdownlint '**/*.md'`
- [Changelog](CHANGELOG.md) updated
- Release tag uses semantic version format
