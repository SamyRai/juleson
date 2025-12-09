# Testing Guide

This guide covers the testing strategy, practices, and tools used in the Juleson project.

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Test Types](#test-types)
- [Testing Tools](#testing-tools)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
- [Test Coverage](#test-coverage)
- [CI/CD Integration](#cicd-integration)
- [Best Practices](#best-practices)

## Testing Philosophy

Juleson follows a comprehensive testing strategy that ensures:

- **Quality**: High test coverage (>80%) across all packages
- **Reliability**: Tests that are deterministic and fast
- **Maintainability**: Well-structured, readable test code
- **Integration**: Tests that validate end-to-end workflows

### Testing Pyramid

```
End-to-End Tests (E2E)
    │
    ├── Integration Tests
    │       │
    │       ├── Component Tests
    │       │       │
    │       │       ├── Unit Tests
    │       │       │       │
    │       │       │       ├── Static Analysis
    │       │       │
    │       │       └── Linting
    │       │
    │       └── Contract Tests
    │
    └── Manual Testing
```

## Test Types

### 1. Unit Tests

**Purpose**: Test individual functions and methods in isolation

**Location**: `*_test.go` files alongside source code

**Example**:

```go
// internal/jules/client_test.go
func TestClient_GetSession(t *testing.T) {
    // Setup
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"id": "session-123", "status": "completed"}`))
    }))
    defer mockServer.Close()

    client := jules.NewClient("test-key", mockServer.URL, 30*time.Second, 3)

    // Execute
    session, err := client.GetSession(context.Background(), "session-123")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "session-123", session.ID)
    assert.Equal(t, "completed", session.Status)
}
```

### 2. Integration Tests

**Purpose**: Test interactions between components

**Location**: `integration_test.go` or `*_integration_test.go`

**Example**:

```go
// internal/jules/client_integration_test.go
func TestClient_SessionWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // This test requires a real Jules API key
    apiKey := os.Getenv("JULES_API_KEY")
    if apiKey == "" {
        t.Skip("JULES_API_KEY not set")
    }

    client := jules.NewClient(apiKey, "https://jules.googleapis.com/v1alpha", 30*time.Second, 3)

    // Create session
    session, err := client.CreateSession(context.Background(), &jules.CreateSessionRequest{
        Prompt: "Test session",
    })
    assert.NoError(t, err)
    assert.NotEmpty(t, session.ID)

    // Get session
    retrieved, err := client.GetSession(context.Background(), session.ID)
    assert.NoError(t, err)
    assert.Equal(t, session.ID, retrieved.ID)
}
```

### 3. End-to-End Tests

**Purpose**: Test complete user workflows

**Location**: `e2e/` directory or `*_e2e_test.go`

**Example**:

```bash
# scripts/test-e2e.sh
#!/bin/bash

# Build binaries
make build

# Test CLI workflow
./bin/juleson init test-project
cd test-project

# Test template execution
../bin/juleson execute template test-generation .

# Verify results
if [ ! -f "test_file_test.go" ]; then
    echo "E2E test failed: test file not generated"
    exit 1
fi

echo "E2E test passed"
```

### 4. MCP Server Tests

**Purpose**: Test MCP protocol implementation

**Location**: `internal/mcp/*_test.go`

**Example**:

```go
// internal/mcp/server_test.go
func TestServer_Initialize(t *testing.T) {
    server := mcp.NewServer()

    request := &mcp.InitializeRequest{
        ProtocolVersion: "2024-11-05",
        Capabilities:    &mcp.ClientCapabilities{},
    }

    response, err := server.Initialize(context.Background(), request)

    assert.NoError(t, err)
    assert.Equal(t, "2024-11-05", response.ProtocolVersion)
    assert.NotNil(t, response.Capabilities)
}
```

## Testing Tools

### Go Testing Tools

- **`go test`**: Standard Go testing framework
- **`testify`**: Enhanced assertions and mocking
- **`httptest`**: HTTP testing utilities
- **`sqlmock`**: Database testing (if needed)

### Code Quality Tools

- **`golangci-lint`**: Comprehensive Go linter
- **`go vet`**: Go static analysis
- **`gosec`**: Security linter
- **`gocyclo`**: Cyclomatic complexity checker

### Coverage Tools

- **`go tool cover`**: Built-in coverage analysis
- **Codecov**: Online coverage reporting
- **Coveralls**: Alternative coverage service

### Mocking Tools

- **Testify mocks**: Interface mocking
- **GoMock**: Advanced mocking framework
- **HTTP mocks**: `httptest.Server`

## Running Tests

### Basic Test Execution

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests in short mode (skip integration tests)
go test -short ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./internal/jules/...

# Run specific test function
go test -run TestClient_GetSession ./internal/jules/
```

### Makefile Commands

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run linter
make lint

# Run all quality checks
make check

# Run short tests only
make test-short
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open in browser (macOS)
open coverage.html
```

### Integration Tests

```bash
# Run integration tests only
go test -tags=integration ./...

# Run with real API (requires API key)
JULES_API_KEY=your-key go test -tags=integration ./internal/jules/
```

## Writing Tests

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Setup phase
    // - Create test data
    // - Initialize mocks
    // - Set up test fixtures

    // Execute phase
    // - Call the function under test
    // - Capture results

    // Assert phase
    // - Verify expected behavior
    // - Check error conditions
    // - Validate state changes
}
```

### Table-Driven Tests

```go
func TestValidateConfig(t *testing.T) {
    tests := []struct {
        name    string
        config  *Config
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid config",
            config: &Config{
                Jules: JulesConfig{APIKey: "test-key"},
            },
            wantErr: false,
        },
        {
            name: "missing API key",
            config:  &Config{},
            wantErr: true,
            errMsg:  "API key is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateConfig(tt.config)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Mocking Dependencies

```go
// Using testify/mock
type MockJulesClient struct {
    mock.Mock
}

func (m *MockJulesClient) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*Session), args.Error(1)
}

func TestGitHubService_CreateSession(t *testing.T) {
    mockClient := &MockJulesClient{}
    service := NewSessionService(mockClient)

    expectedSession := &Session{ID: "session-123"}
    mockClient.On("CreateSession", mock.Anything, mock.Anything).Return(expectedSession, nil)

    session, err := service.CreateSessionFromCurrentRepo(context.Background(), "test prompt")

    assert.NoError(t, err)
    assert.Equal(t, "session-123", session.ID)
    mockClient.AssertExpectations(t)
}
```

### Test Helpers

```go
// testhelpers/helpers.go
func CreateTestConfig() *config.Config {
    return &config.Config{
        Jules: config.JulesConfig{
            APIKey:    "test-key",
            BaseURL:   "https://test.api.com",
            Timeout:   30 * time.Second,
            RetryAttempts: 3,
        },
    }
}

func WithTempDir(t *testing.T, fn func(dir string)) {
    dir, err := os.MkdirTemp("", "juleson-test-*")
    assert.NoError(t, err)
    defer os.RemoveAll(dir)

    fn(dir)
}
```

## Test Coverage

### Coverage Goals

- **Overall**: >80% coverage
- **Critical packages**: >90% coverage
  - `internal/jules/` (API client)
  - `internal/mcp/` (MCP server)
  - `internal/github/` (GitHub integration)
- **New code**: 100% coverage required

### Coverage Report

```bash
# Generate detailed coverage report
make coverage

# View coverage by package
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | sort -k3 -nr

# Focus on low coverage areas
go tool cover -func=coverage.out | awk '$3 < 80 {print}'
```

### Coverage Badges

```markdown
<!-- README.md -->
[![Coverage](https://codecov.io/gh/SamyRai/Juleson/branch/main/graph/badge.svg)](https://codecov.io/gh/SamyRai/Juleson)
```

## CI/CD Integration

### GitHub Actions Testing

```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
```

### Quality Gates

```yaml
# Require coverage threshold
- name: Check coverage
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
    if (( $(echo "$coverage < 80" | bc -l) )); then
      echo "Coverage $coverage% is below 80% threshold"
      exit 1
    fi
```

## Best Practices

### Test Naming

```go
// Good: descriptive and follows convention
func TestClient_GetSession(t *testing.T)
func TestClient_GetSession_NotFound(t *testing.T)
func TestClient_GetSession_Timeout(t *testing.T)

// Bad: unclear purpose
func TestGet(t *testing.T)
func TestClient(t *testing.T)
```

### Test Organization

```go
// Group related tests
func TestClient_SessionManagement(t *testing.T) {
    t.Run("create session", testClientCreateSession)
    t.Run("get session", testClientGetSession)
    t.Run("list sessions", testClientListSessions)
}

func testClientCreateSession(t *testing.T) {
    // Test implementation
}
```

### Assertions

```go
// Use descriptive assertion messages
assert.Equal(t, expectedStatus, actualStatus, "session status should match expected value")

// Check for errors first
result, err := someFunction()
assert.NoError(t, err)
assert.NotNil(t, result)

// Use assert vs require appropriately
// assert: continue test on failure
// require: stop test on failure
require.NoError(t, err)  // Stop if setup fails
assert.Equal(t, expected, actual)  // Continue to check other things
```

### Test Data Management

```go
// Use constants for test data
const (
    testAPIKey = "test-api-key-123"
    testSessionID = "session-123"
)

// Create test fixtures
func createTestSession() *Session {
    return &Session{
        ID:      testSessionID,
        Status:  "completed",
        Created: time.Now(),
    }
}
```

### Parallel Tests

```go
// Run tests in parallel when safe
func TestClient_GetSession(t *testing.T) {
    t.Parallel()  // Safe for unit tests

    // Test implementation
}

// Avoid parallel for shared state
func TestDatabaseOperations(t *testing.T) {
    // No t.Parallel() - modifies shared database
}
```

### Benchmark Tests

```go
func BenchmarkClient_GetSession(b *testing.B) {
    client := createTestClient()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.GetSession(context.Background(), testSessionID)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Debugging Test Failures

### Common Issues

1. **Race Conditions**:

   ```bash
   go test -race ./...
   ```

2. **Flaky Tests**:

   ```bash
   # Run test multiple times
   for i in {1..10}; do go test -run TestFlaky; done
   ```

3. **Test Timeouts**:

   ```go
   // Add timeouts to prevent hanging
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

4. **Resource Leaks**:

   ```go
   // Ensure cleanup
   defer func() {
       if err := cleanup(); err != nil {
           t.Errorf("cleanup failed: %v", err)
       }
   }()
   ```

### Test Debugging Tools

```bash
# Verbose test output
go test -v -run TestFailing

# Debug with delve
dlv test -- -test.run TestFailing

# Profile test performance
go test -cpuprofile cpu.prof -memprofile mem.prof ./...
go tool pprof cpu.prof
```

## Contributing Test Improvements

When contributing:

1. **Add tests** for new functionality
2. **Update tests** when changing existing code
3. **Maintain coverage** above 80%
4. **Follow conventions** established in the codebase
5. **Document complex tests** with comments

### Test Checklist

- [ ] Tests pass: `make test`
- [ ] Coverage maintained: `make coverage`
- [ ] Linting passes: `make lint`
- [ ] No race conditions: `go test -race`
- [ ] Integration tests work: `go test -tags=integration`
- [ ] Documentation updated if needed

---

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://pkg.go.dev/github.com/stretchr/testify)
- [Go Testing Best Practices](https://github.com/golang/go/wiki/TestComments)
- [Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

---

**Last Updated**: November 3, 2025
