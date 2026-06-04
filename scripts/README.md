# Scripts

This directory contains release installer scripts and script-level tests. It is a
script-local note; the canonical install documentation lives under `docs/`.

## Installers

- `install.sh`: Linux and macOS installer for `juleson` and the short `jsn` alias.
- `install.ps1`: Windows installer for `juleson` and the short `jsn` alias.

Linux and macOS:

```bash
curl -L https://github.com/SamyRai/juleson/releases/latest/download/install.sh | bash
```

Windows:

```powershell
irm https://github.com/SamyRai/juleson/releases/latest/download/install.ps1 | iex
```

Use [Installation Guide](../docs/INSTALLATION_GUIDE.md) for the canonical install
instructions.
