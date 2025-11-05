# Orchestrator Architecture

## Overview

The Juleson orchestrator provides a unified interface for all project build, test, quality, and deployment operations. It follows a clean, interface-based architecture that eliminates code duplication and enables multiple consumers (CLI and MCP tools).

## Architecture

### Core Components

#### 1. Interface Layer (`internal/orchestrator/orchestrator.go`)

The `Orchestrator` interface defines all available operations:

```go
type Orchestrator interface {
    // Build operations
    BuildAll(ctx context.Context) error
    BuildCLI(ctx context.Context) error
    BuildMCP(ctx context.Context) error

    // Test operations
    Test(ctx context.Context, options TestOptions) error
    Coverage(ctx context.Context) error

    // Quality operations
    Lint(ctx context.Context) error
    Format(ctx context.Context) error
    RunAllChecks(ctx context.Context) error

    // ... and more
}
```

**Benefits:**

- Testable through interface mocking
- Enables dependency injection
- Clean separation of concerns
- Future-proof for new implementations

#### 2. Service Layer

The service layer implements the orchestrator interface across multiple files:

- **`orchestrator.go`** - Core interface, config, service struct, build operations
- **`test.go`** - Test execution and coverage
- **`quality.go`** - Linting, formatting, quality checks
- **`deps.go`** - Dependency management
- **`run.go`** - Runtime operations (install, run CLI/MCP, dev mode)
- **`docker.go`** - Docker operations

**Pattern:**

```go
type Service struct {
    config *Config
    stdout io.Writer
    stderr io.Writer
}

func (s *Service) BuildAll(ctx context.Context) error {
    // Implementation
}
```

#### 3. Configuration

Project-specific configuration is centralized:

```go
type Config struct {
    BinaryCLI    string
    BinaryMCP    string
    BinDir       string
    DockerImage  string
    Version      string
    BuildDate    string
    GitCommit    string
    // ... more fields
}

func DefaultConfig(version, buildDate, gitCommit string) *Config {
    // Returns Juleson-specific defaults
}
```

This pattern allows:

- Easy configuration for different projects
- Version information injection
- Consistent naming across all operations

### Consumer Layers

#### 1. CLI (`cmd/orchestrator/`)

Cobra-based CLI that provides user-facing commands:

```
cmd/orchestrator/
├── main.go              # Entry point, version info
└── commands/
    ├── commands.go      # Root command, service initialization
    ├── build.go         # Build commands
    ├── test.go          # Test commands
    ├── quality.go       # Quality commands
    ├── deps.go          # Dependency commands
    ├── run.go           # Runtime commands
    ├── docker.go        # Docker commands
    └── misc.go          # Misc commands (clean, install, dev, check, version)
```

**Pattern:**

```go
var buildCmd = &cobra.Command{
    Use:   "build",
    Short: "Build all binaries",
    Run: func(cmd *cobra.Command, args []string) {
        ctx := context.Background()
        if err := svc.BuildAll(ctx); err != nil {
            log.Fatal(err)
        }
    },
}
```

Commands are thin wrappers that:

1. Parse user input
2. Call service methods
3. Handle errors
4. Present results

#### 2. MCP Tools (`internal/mcp/tools/orchestrator.go`)

MCP integration layer that exposes orchestrator functionality via Model Context Protocol:

```go
func GetOrchestrator() orchestrator.Orchestrator {
    config := orchestrator.DefaultConfig("dev", time.Now().Format("2006-01-02"), "mcp-tools")
    return orchestrator.NewService(config)
}
```

This enables:

- AI agents to trigger builds
- Automated testing workflows
- Remote orchestration
- Integration with development tools

The existing `dev.go` MCP tools can use `GetOrchestrator()` to delegate operations to the orchestrator service instead of duplicating logic.

## Design Principles

### 1. No Code Duplication

**Before:** CLI commands and MCP tools each implemented their own build/test logic.

**After:** Single source of truth in the service layer, consumed by multiple interfaces.

### 2. Interface-Based Design

```go
// Easy to test with mocks
type MockOrchestrator struct {
    orchestrator.Orchestrator
    BuildAllFunc func(ctx context.Context) error
}

func (m *MockOrchestrator) BuildAll(ctx context.Context) error {
    return m.BuildAllFunc(ctx)
}
```

### 3. Context Support

All operations accept `context.Context` for:

- Cancellation support
- Timeout handling
- Request-scoped values
- Graceful shutdown

### 4. Error Handling

Consistent error wrapping:

```go
if err := s.runCommand(ctx, "go", "build", ...); err != nil {
    return fmt.Errorf("failed to build CLI: %w", err)
}
```

### 5. Dependency Injection

Service dependencies are injected through configuration:

```go
svc := orchestrator.NewService(config)
svc = svc.WithOutput(customStdout, customStderr)
```

## Usage Examples

### CLI Usage

```bash
# Build all binaries
./bin/orchestrator build

# Run tests with coverage
./bin/orchestrator test --cover

# Run quality checks
./bin/orchestrator check

# Clean build artifacts
./bin/orchestrator clean

# Docker operations
./bin/orchestrator docker-build
./bin/orchestrator docker-compose-up
```

### Programmatic Usage

```go
package main

import (
    "context"
    "github.com/SamyRai/juleson/internal/orchestrator"
)

func main() {
    config := orchestrator.DefaultConfig("v1.0.0", "2024-01-01", "abc123")
    svc := orchestrator.NewService(config)

    ctx := context.Background()

    // Build
    if err := svc.BuildAll(ctx); err != nil {
        panic(err)
    }

    // Test
    opts := orchestrator.TestOptions{
        Verbose: true,
        Race:    true,
    }
    if err := svc.Test(ctx, opts); err != nil {
        panic(err)
    }

    // Quality checks
    if err := svc.RunAllChecks(ctx); err != nil {
        panic(err)
    }
}
```

### MCP Integration

MCP tools can use the orchestrator:

```go
import "github.com/SamyRai/juleson/internal/mcp/tools"

func buildHandler(ctx context.Context, ...) {
    svc := tools.GetOrchestrator()
    if err := svc.BuildAll(ctx); err != nil {
        // Handle error
    }
}
```

## Benefits

### For Developers

- **Single Command**: `./bin/orchestrator <command>` replaces all `make` commands
- **Consistent Interface**: Same commands work locally and in CI
- **Better Error Messages**: Go error handling vs make's cryptic errors
- **Type Safety**: Compile-time checking vs runtime shell errors

### For CI/CD

- **Reproducible**: No make-specific quirks or shell variations
- **Portable**: Pure Go, works anywhere Go runs
- **Fast**: No shell overhead, parallel builds
- **Observable**: Structured logging, proper exit codes

### For Automation

- **MCP Integration**: AI agents can trigger operations
- **API-First**: Interface-based design enables any consumer
- **Testable**: Easy to test with mocks and stubs
- **Extensible**: Add new operations without breaking existing code

## Migration from Makefile

The orchestrator **completely replaces** the Makefile. All functionality has been migrated:

| Makefile Target | Orchestrator Command | Status |
|----------------|---------------------|--------|
| `make all` | `orchestrator all` | ✅ Complete |
| `make build` | `orchestrator build` | ✅ Complete |
| `make build-cli` | `orchestrator build-cli` | ✅ Complete |
| `make build-mcp` | `orchestrator build-mcp` | ✅ Complete |
| `make test` | `orchestrator test` | ✅ Complete |
| `make test-short` | `orchestrator test --short` | ✅ Complete |
| `make test-coverage` | `orchestrator coverage` | ✅ Complete |
| `make lint` | `orchestrator lint` | ✅ Complete |
| `make fmt` | `orchestrator fmt` | ✅ Complete |
| `make clean` | `orchestrator clean` | ✅ Complete |
| `make deps` | `orchestrator deps` | ✅ Complete |
| `make tidy` | `orchestrator tidy` | ✅ Complete |
| `make install` | `orchestrator install` | ✅ Complete |
| `make run-cli` | `orchestrator run-cli` | ✅ Complete |
| `make run-mcp` | `orchestrator run-mcp` | ✅ Complete |
| `make dev` | `orchestrator dev` | ✅ Complete |
| `make check` | `orchestrator check` | ✅ Complete |
| `make docker-build` | `orchestrator docker-build` | ✅ Complete |
| `make docker-run` | `orchestrator docker-run` | ✅ Complete |
| `make docker-compose-up` | `orchestrator docker-compose-up` | ✅ Complete |
| `make docker-compose-down` | `orchestrator docker-compose-down` | ✅ Complete |

**The Makefile has been completely removed.**

## Future Enhancements

### Planned Features

1. **Parallel Builds**: Leverage goroutines for concurrent operations
2. **Build Cache**: Smart incremental builds
3. **Watch Mode**: Auto-rebuild on file changes
4. **Metrics**: Build time tracking, test performance
5. **Notifications**: Slack/Discord integration
6. **Remote Execution**: Distribute builds across machines

### Extensibility

Adding new operations is straightforward:

1. Add method to `Orchestrator` interface
2. Implement in `Service`
3. Add CLI command that calls the method
4. Optionally expose via MCP tools

Example:

```go
// 1. Interface
type Orchestrator interface {
    // ... existing methods
    Benchmark(ctx context.Context) error
}

// 2. Implementation
func (s *Service) Benchmark(ctx context.Context) error {
    return s.runCommand(ctx, "go", "test", "-bench=.", "./...")
}

// 3. CLI Command
var benchCmd = &cobra.Command{
    Use:   "bench",
    Short: "Run benchmarks",
    Run: func(cmd *cobra.Command, args []string) {
        if err := svc.Benchmark(context.Background()); err != nil {
            log.Fatal(err)
        }
    },
}
```

## Testing

The interface-based design enables comprehensive testing:

```go
func TestBuildAll(t *testing.T) {
    config := orchestrator.DefaultConfig("test", "2024-01-01", "test")
    svc := orchestrator.NewService(config)

    var stdout, stderr bytes.Buffer
    svc = svc.WithOutput(&stdout, &stderr)

    err := svc.BuildAll(context.Background())
    assert.NoError(t, err)
    assert.Contains(t, stdout.String(), "built successfully")
}
```

## Summary

The orchestrator architecture provides:

- ✅ **Clean Interface Design**: Single source of truth, no duplication
- ✅ **Multiple Consumers**: CLI, MCP tools, programmatic usage
- ✅ **Production Ready**: Context support, error handling, logging
- ✅ **Extensible**: Easy to add new operations
- ✅ **Testable**: Interface-based design enables mocking
- ✅ **Maintainable**: Organized code, clear separation of concerns

This architecture replaces the Makefile completely while providing a foundation for future enhancements and integrations.
