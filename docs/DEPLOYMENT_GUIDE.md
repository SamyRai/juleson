# Deployment Guide

> **Guide for deploying and distributing Juleson**

## Overview

This guide covers different deployment scenarios for Juleson, from local development to production release.

## Table of Contents

- [Local Development](#local-development)
- [Binary Distribution](#binary-distribution)
- [Package Managers](#package-managers)
- [Docker Deployment](#docker-deployment)
- [GitHub Actions Integration](#github-actions-integration)
- [Release Process](#release-process)

---

## Local Development

### Build from Source

```bash
# Clone the repository
git clone https://github.com/SamyRai/Juleson.git
cd Juleson

# Install dependencies
go mod download

# Build binaries
make build

# Binaries will be in ./bin/
./bin/juleson --version
./bin/jules-mcp --version
```

### Install Locally

```bash
# Install to $GOPATH/bin
make install

# Or install specific command
go install ./cmd/juleson
go install ./cmd/jules-mcp
```

### Configuration

```bash
# Copy example configuration
cp configs/Juleson.example.yaml configs/Juleson.yaml

# Edit with your API key
export JULES_API_KEY="your-api-key-here"

# Or set in config file
vim configs/Juleson.yaml
```

---

## Binary Distribution

### Build for Multiple Platforms

```bash
# Build for all platforms
make build-all

# Or build for specific platform
GOOS=linux GOARCH=amd64 go build -o juleson-linux-amd64 ./cmd/juleson
GOOS=darwin GOARCH=arm64 go build -o juleson-darwin-arm64 ./cmd/juleson
GOOS=windows GOARCH=amd64 go build -o juleson-windows-amd64.exe ./cmd/juleson
```

### Supported Platforms

| OS | Architecture | Binary Name |
|----|--------------|-------------|
| Linux | amd64 | `juleson-linux-amd64` |
| Linux | arm64 | `juleson-linux-arm64` |
| macOS | amd64 (Intel) | `juleson-darwin-amd64` |
| macOS | arm64 (Apple Silicon) | `juleson-darwin-arm64` |
| Windows | amd64 | `juleson-windows-amd64.exe` |

### Download Pre-built Binaries

```bash
# Download from GitHub Releases
VERSION=v0.1.0
OS=darwin  # or linux, windows
ARCH=arm64 # or amd64

# CLI
curl -L -o juleson \
  "https://github.com/SamyRai/Juleson/releases/download/${VERSION}/juleson-${OS}-${ARCH}"
chmod +x juleson

# MCP Server
curl -L -o jules-mcp \
  "https://github.com/SamyRai/Juleson/releases/download/${VERSION}/jules-mcp-${OS}-${ARCH}"
chmod +x jules-mcp
```

---

## Package Managers

### Homebrew (macOS/Linux)

**Create Homebrew Formula** (future):

```ruby
class JulesAutomation < Formula
  desc "Automate Google Jules AI coding agent workflows"
  homepage "https://github.com/SamyRai/Juleson"
  url "https://github.com/SamyRai/Juleson/archive/v0.1.0.tar.gz"
  sha256 "..."
  license "MIT"

  depends_on "go" => :build

  def install
    system "make", "build"
    bin.install "bin/juleson"
    bin.install "bin/jules-mcp"
  end

  test do
    system "#{bin}/juleson", "--version"
  end
end
```

**Install**:

```bash
brew tap SamyRai/Juleson
brew install Juleson
```

### Go Install

```bash
# Install latest version
go install github.com/SamyRai/Juleson/cmd/juleson@latest
go install github.com/SamyRai/Juleson/cmd/jules-mcp@latest

# Install specific version
go install github.com/SamyRai/Juleson/cmd/juleson@v0.1.0
```

---

## Docker Deployment

### Build Docker Image

**Dockerfile** (future enhancement):

```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o juleson ./cmd/juleson
RUN CGO_ENABLED=0 GOOS=linux go build -o jules-mcp ./cmd/jules-mcp

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/juleson .
COPY --from=builder /app/jules-mcp .
COPY --from=builder /app/configs/Juleson.example.yaml ./configs/

ENV JULES_API_KEY=""

CMD ["./juleson"]
```

### Build and Run

```bash
# Build image
docker build -t Juleson:latest .

# Run CLI
docker run --rm \
  -e JULES_API_KEY="your-key" \
  -v $(pwd):/workspace \
  Juleson:latest \
  juleson analyze /workspace

# Run MCP Server
docker run --rm \
  -e JULES_API_KEY="your-key" \
  -p 3000:3000 \
  Juleson:latest \
  jules-mcp
```

---

## GitHub Actions Integration

### Use in GitHub Workflows

```yaml
name: Juleson

on:
  push:
    branches: [main]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Download Jules CLI
        run: |
          VERSION=v0.1.0
          curl -L -o juleson \
            "https://github.com/SamyRai/Juleson/releases/download/${VERSION}/juleson-linux-amd64"
          chmod +x juleson

      - name: Analyze codebase
        env:
          JULES_API_KEY: ${{ secrets.JULES_API_KEY }}
        run: |
          ./juleson analyze .

      - name: Execute template
        env:
          JULES_API_KEY: ${{ secrets.JULES_API_KEY }}
        run: |
          ./juleson execute \
            --template test-generation \
            --project . \
            --output results.json
```

### GitHub Action (future)

```yaml
- name: Juleson
  uses: SamyRai/Juleson-action@v1
  with:
    command: analyze
    project-path: .
    api-key: ${{ secrets.JULES_API_KEY }}
```

---

## Release Process

### Automated Release with GitHub Actions

Our release workflow automatically:

1. Builds binaries for all platforms
2. Creates GitHub release
3. Uploads binaries as release assets
4. Generates release notes

### Manual Release Steps

1. **Update Version**:

   ```bash
   # Update CHANGELOG.md
   vim CHANGELOG.md

   # Commit changes
   git add CHANGELOG.md
   git commit -m "chore: prepare release v0.2.0"
   ```

2. **Create and Push Tag**:

   ```bash
   git tag -a v0.2.0 -m "Release v0.2.0"
   git push origin v0.2.0
   ```

3. **GitHub Actions will**:
   - Build all binaries
   - Run tests
   - Create GitHub release
   - Upload artifacts

4. **Verify Release**:
   - Check GitHub Releases page
   - Test binary downloads
   - Verify release notes

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (v1.0.0): Incompatible API changes
- **MINOR** (v0.1.0): Backwards-compatible functionality
- **PATCH** (v0.0.1): Backwards-compatible bug fixes

### Release Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped in relevant files
- [ ] Tag created and pushed
- [ ] GitHub release created
- [ ] Binaries uploaded
- [ ] Release notes complete
- [ ] Announced (optional)

---

## Environment Variables

### Required

```bash
export JULES_API_KEY="your-api-key-here"
```

### Optional

```bash
export JULES_CONFIG_PATH="/path/to/config.yaml"
export JULES_LOG_LEVEL="debug"  # debug, info, warn, error
export JULES_API_URL="https://api.jules.ai"  # Custom API endpoint
```

---

## Configuration Files

### System-wide Configuration

```bash
# Linux
/etc/Juleson/config.yaml

# macOS
/Library/Application Support/JulesAutomation/config.yaml

# Windows
C:\ProgramData\JulesAutomation\config.yaml
```

### User Configuration

```bash
# Linux/macOS
~/.config/Juleson/config.yaml

# Windows
%APPDATA%\JulesAutomation\config.yaml
```

### Project Configuration

```bash
# Project root
./configs/Juleson.yaml
./.Juleson.yaml
```

**Priority** (highest to lowest):

1. `JULES_CONFIG_PATH` environment variable
2. Project configuration (`./.Juleson.yaml`)
3. User configuration (`~/.config/Juleson/config.yaml`)
4. System configuration (`/etc/Juleson/config.yaml`)

---

## MCP Server Deployment

### Claude Desktop Integration

1. **Build MCP Server**:

   ```bash
   make build
   ```

2. **Configure Claude Desktop**:

   ```json
   {
     "mcpServers": {
       "Juleson": {
         "command": "/absolute/path/to/jules-mcp",
         "args": [],
         "env": {
           "JULES_API_KEY": "your-api-key-here"
         }
       }
     }
   }
   ```

3. **Restart Claude Desktop**

4. **Verify**:
   - Open Claude Desktop
   - Check MCP tools are available
   - Test with: "List available Jules templates"

### Standalone MCP Server

```bash
# Run MCP server
export JULES_API_KEY="your-key"
./bin/jules-mcp

# Or with config file
./bin/jules-mcp --config configs/Juleson.yaml
```

---

## Troubleshooting

### Binary Won't Execute (macOS)

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine juleson
xattr -d com.apple.quarantine jules-mcp
```

### Permission Denied

```bash
# Make executable
chmod +x juleson
chmod +x jules-mcp
```

### API Key Issues

```bash
# Verify API key is set
echo $JULES_API_KEY

# Test connection
./bin/juleson --version
```

---

## Production Considerations

### Security

- ✅ Never commit API keys
- ✅ Use environment variables or secure key management
- ✅ Rotate API keys regularly
- ✅ Use least-privilege access
- ✅ Enable audit logging

### Performance

- ✅ Cache API responses when possible
- ✅ Use connection pooling
- ✅ Monitor API rate limits
- ✅ Implement retry logic with backoff

### Monitoring

- ✅ Log all API calls
- ✅ Monitor error rates
- ✅ Track API usage metrics
- ✅ Set up alerts for failures

---

## Support

For deployment issues:

- 📖 [Documentation](../README.md)
- 💬 [Discussions](https://github.com/SamyRai/Juleson/discussions)
- 🐛 [Report Issues](https://github.com/SamyRai/Juleson/issues)

---

**Last Updated**: November 1, 2025
