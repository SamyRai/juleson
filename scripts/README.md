# Scripts

This directory contains utility and automation scripts.

## Planned Scripts

- `build.sh` - Build script for all binaries
- `test.sh` - Run all tests
- `install.sh` - Installation script
- `deploy.sh` - Deployment automation
- `release.sh` - Release preparation

## Usage

Scripts should be executable and documented with comments explaining their purpose and usage.

### Example: build.sh

```bash
#!/bin/bash
# Build all binaries

echo "Building juleson..."
go build -o bin/juleson ./cmd/juleson

echo "Building jules-mcp..."
go build -o bin/jules-mcp ./cmd/jules-mcp

echo "Build complete!"
```
