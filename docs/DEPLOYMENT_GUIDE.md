# Deployment Guide

Deployment uses release assets, local installs, Docker images, or MCP client
configuration.

## Local Build

```bash
go mod download
go build -o bin/builder ./cmd/builder
./bin/builder build
```

Individual executable names:

```bash
go build -o bin/juleson ./cmd/juleson
go build -o bin/jsn ./cmd/juleson
```

## Install Locally

```bash
./bin/builder install
```

Or:

```bash
juleson dev install --path "$HOME/.local/bin"
```

## Release Assets

The release workflow builds matching `juleson` and `jsn` assets for Linux,
macOS, and Windows, plus:

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

Provide credentials through environment variables such as `JULES_API_KEY` or
through mounted config files. Do not bake API keys into images.

## MCP Client Deployment

Install `juleson` on the machine running the MCP client. Configure the client
with an absolute binary path and `mcp serve` arguments:

```json
{
  "mcpServers": {
    "juleson": {
      "command": "/usr/local/bin/juleson",
      "args": ["mcp", "serve"],
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
- `juleson mcp serve --version`
- `markdownlint '**/*.md'`
- [Changelog](CHANGELOG.md) updated
- Release tag uses semantic version format
