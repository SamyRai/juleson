# Scripts

This directory contains utility and automation scripts.

## Installer Scripts

- `install.sh` - Linux/macOS release installer for `juleson` and `jules-mcp`
- `install.ps1` - Windows release installer for `juleson` and `jules-mcp`

## Usage

### Linux/macOS

```bash
curl -L https://github.com/SamyRai/juleson/releases/latest/download/install.sh | bash
```

Use `INSTALL_DIR` to install somewhere other than `/usr/local/bin`:

```bash
INSTALL_DIR="$HOME/.local/bin" bash install.sh
```

### Windows

```powershell
irm https://github.com/SamyRai/juleson/releases/latest/download/install.ps1 | iex
```
