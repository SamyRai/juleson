# Installation Guide

Juleson ships two executable names:

- `juleson`: the primary CLI.
- `jsn`: a short alias that runs the same CLI.

MCP is served by `juleson mcp serve` or `jsn mcp serve`; there is no separate
`jules-mcp` binary.

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
- `jsn-<os>-<arch>.tar.gz`
- `juleson-windows-<arch>.zip`
- `jsn-windows-<arch>.zip`

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
```

Make sure `$(go env GOPATH)/bin` is on `PATH`. `go install` installs only the
`juleson` executable; release installers provide the `jsn` alias.

## Build From Source

```bash
git clone https://github.com/SamyRai/juleson.git
cd juleson
go mod download
go build -o bin/builder ./cmd/builder
./bin/builder build
```

Or build individual executable names:

```bash
go build -o bin/juleson ./cmd/juleson
go build -o bin/jsn ./cmd/juleson
```

## Verify

```bash
juleson --help
jsn --help
juleson version
juleson mcp serve --version
```

## Uninstall

Remove the installed executables from the directory where they were installed:

```bash
rm /usr/local/bin/juleson
rm /usr/local/bin/jsn
```

Windows:

```powershell
Remove-Item "$env:USERPROFILE\.juleson\bin\juleson.exe"
Remove-Item "$env:USERPROFILE\.juleson\bin\jsn.exe"
```

Remove config files only if you no longer need local settings.
