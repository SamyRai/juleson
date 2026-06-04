# Testing Guide

## Local Test Commands

Run the standard Go suite:

```bash
go test ./...
```

Run the same command through Juleson:

```bash
juleson dev test
```

Useful variants:

```bash
juleson dev test --short
juleson dev test --cover --coverprofile coverage.out
juleson dev test --run TestMCP
juleson dev test --race --timeout 10m
```

Run quality checks:

```bash
juleson dev check
```

The local pre-commit hook checks `gofmt -s`, `go mod tidy`, and `go vet ./...`.
The pre-push hook runs `go test ./...`.

## Test Types

- Unit tests live beside the package code as `*_test.go`.
- Integration-style tests use local fakes or test servers and should not require
  real credentials by default.
- MCP tests exercise `juleson mcp serve` and the internal MCP server package.
- Installer tests validate shell and PowerShell installer behavior without
  publishing release assets.

## CI Coverage

The GitHub Actions workflow runs:

- `go mod download` and `go mod verify`
- `gofmt -s -l .`
- `go mod tidy && git diff --exit-code go.mod go.sum`
- `go test -v -race -timeout=10m ./...` with `SKIP_E2E=1`
- coverage on Ubuntu stable Go
- MCP command smoke on Ubuntu stable Go
- `go test -v -short -timeout=5m ./...`
- `golangci-lint`
- Gosec and Trivy scans
- builds on Linux, macOS, and Windows

Markdown-only changes are ignored by CI, so documentation edits should be checked
locally with `markdownlint`.

## Writing Tests

- Prefer table-driven tests for input/output behavior.
- Use `t.TempDir()` for filesystem state.
- Keep networked tests behind explicit environment requirements.
- Avoid depending on test execution order.
- Keep fixtures small and local to the package unless multiple packages need them.

## Documentation Checks

```bash
markdownlint '**/*.md'
```

Use a local link checker before moving docs or deleting files. In-repo links
should be relative and should resolve from the file that contains the link.
