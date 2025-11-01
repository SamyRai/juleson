# Configuration Files

This directory contains **actual YAML configuration files** for the Jules automation system.

## Important Distinction

- **`internal/config/`** - Go code that loads and validates configuration
- **`configs/`** (this directory) - YAML files with actual configuration values

## Files

- `jules-automation.yaml` - Main application configuration (create this file)
- `.env` - Environment variables (in project root, not tracked in git)

## Configuration Structure

The configuration is loaded by `internal/config/config.go` which looks for files named `jules-automation.yaml` in:

1. `./configs/` directory (this directory)
2. Current working directory (`.`)

## Quick Start

1. Copy `jules-automation.example.yaml` to `jules-automation.yaml`
2. Set your `JULES_API_KEY` environment variable or in `.env` file
3. Adjust other settings as needed

## Environment Variables

You can override any configuration value using environment variables with uppercase names and underscores. For example:

- `JULES_API_KEY` - Your Jules API key (required)
- `JULES_BASE_URL` - Jules API base URL
- `MCP_SERVER_PORT` - MCP server port
