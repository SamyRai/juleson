# Installation Guide

Juleson ships two user binaries: `juleson` and `jules-mcp`.

## Quick Install

Linux and macOS:

```bash
curl -L https://github.com/SamyRai/juleson/releases/latest/download/install.sh | bash
```

Windows PowerShell:

```powershell
irm https://github.com/SamyRai/juleson/releases/latest/download/install.ps1 | iex
```

The installers download release assets named:

- `juleson-<os>-<arch>.tar.gz`
- `jules-mcp-<os>-<arch>.tar.gz`
- `juleson-windows-<arch>.zip`
- `jules-mcp-windows-<arch>.zip`

Supported OS values are `linux`, `darwin`, and `windows`. Supported architecture
values are `amd64` and `arm64`.

## Installer Options

Linux and macOS:

```bash
bash install.sh --version v1.0.0 --install-dir "$HOME/.local/bin"
INSTALL_DIR="$HOME/.local/bin" bash install.sh
```

Windows:

```powershell
.\install.ps1 -Version v1.0.0 -InstallDir "$env:USERPROFILE\bin"
.\install.ps1 -NoPathUpdate
```

`JULESON_INSTALL_BASE_URL` can point installers at a custom asset base URL.

## Go Install

```bash
go install github.com/SamyRai/juleson/cmd/juleson@latest
go install github.com/SamyRai/juleson/cmd/jules-mcp@latest
```

Make sure `$(go env GOPATH)/bin` is on `PATH`.

## Build From Source

```bash
git clone https://github.com/SamyRai/juleson.git
cd juleson
go mod download
go build -o bin/orchestrator ./cmd/orchestrator
./bin/orchestrator build
```

Or build individual binaries:

```bash
go build -o bin/juleson ./cmd/juleson
go build -o bin/jules-mcp ./cmd/jules-mcp
```

## Verify

```bash
juleson --help
juleson version
jules-mcp --help
```

`jules-mcp` uses stdio transport. Running it directly starts the server and waits
for MCP protocol messages.

## Uninstall

Remove the installed binaries from the directory where they were installed:

```bash
rm /usr/local/bin/juleson
rm /usr/local/bin/jules-mcp
```

Windows:

```powershell
Remove-Item "$env:USERPROFILE\.juleson\bin\juleson.exe"
Remove-Item "$env:USERPROFILE\.juleson\bin\jules-mcp.exe"
```

Remove config files only if you no longer need local settings.
