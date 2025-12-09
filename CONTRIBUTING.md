# Contributing to Juleson

Thank you for your interest in contributing to Juleson! This document
provides guidelines and instructions for contributing.

## 🤝 Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## 🚀 Getting Started

### Prerequisites

- Go 1.23 or higher
- Git
- Jules API access (for integration testing)

### Setting Up Development Environment

1. **Fork the repository**

   ```bash
   # Fork on GitHub, then clone your fork
   git clone https://github.com/SamyRai/Juleson.git
   cd Juleson
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Set up configuration**

   ```bash
   cp configs/Juleson.example.yaml configs/Juleson.yaml
   # Edit configs/Juleson.yaml with your settings
   ```

4. **Build the project**

   ```bash
   make build
   # or manually:
   # go build -o bin/juleson cmd/juleson/main.go
   # go build -o bin/jules-mcp cmd/jules-mcp/main.go
   ```

5. **Run tests**

   ```bash
   make test
   # or manually:
   # go test ./...
   ```

## 📝 How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/SamyRai/Juleson/issues)
2. If not, create a new issue with:
   - Clear, descriptive title
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)
   - Relevant logs or screenshots

### Suggesting Features

1. Check [existing feature requests](https://github.com/SamyRai/Juleson/issues?q=is%3Aissue+label%3Aenhancement)
2. If your idea is new, create a feature request issue

### Submitting Pull Requests

1. **Create a branch**

   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes**
   - Write clean, idiomatic Go code
   - Follow existing code style
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**

   ```bash
   # Run all tests
   go test ./...

   # Run tests with coverage
   go test -cover ./...

   # Run linter
   go vet ./...
   ```

4. **Commit your changes**

   ```bash
   git add .
   git commit -m "feat: add new feature X"
   # or
   git commit -m "fix: resolve issue with Y"
   ```

   Follow [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `test:` for test additions/changes
   - `refactor:` for code refactoring
   - `chore:` for maintenance tasks

5. **Push to your fork**

   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request**
   - Go to the original repository on GitHub
   - Click "New Pull Request"
   - Select your fork and branch
   - Fill in the PR template with:
     - Description of changes
     - Related issues (if any)
     - Testing performed
     - Screenshots (if applicable)

## 🏗️ Project Structure

```bash
Juleson/
├── cmd/                    # Entry points for binaries
│   ├── juleson/         # CLI application
│   └── jules-mcp/         # MCP server
├── internal/              # Private application code
│   ├── automation/       # Automation engine
│   ├── cli/              # CLI implementation
│   ├── config/           # Configuration management
│   ├── github/           # GitHub API integration (SOLID architecture)
│   │   ├── client.go     # Main client facade
│   │   ├── actions.go    # GitHub Actions service
│   │   ├── repositories.go # Repository service
│   │   ├── pullrequests.go # Pull request service
│   │   ├── sessions.go   # Session service
│   │   ├── git.go        # Git utilities
│   │   ├── types.go      # Domain models
│   │   └── utils.go      # Helper functions
│   ├── jules/            # Jules API client
│   ├── mcp/              # MCP server implementation
│   └── templates/        # Template management
├── configs/              # Configuration files
├── docs/                 # Documentation
├── scripts/              # Utility scripts
└── templates/            # Template definitions
```

## 📋 Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Use `go vet` for static analysis
- Keep functions small and focused
- Write descriptive variable and function names
- Add comments for exported functions and types

### Testing

- Write unit tests for new functionality
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies
- Test edge cases and error conditions

Example test:

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "expected", false},
        {"invalid input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Documentation

- Update README.md for user-facing changes
- Add/update godoc comments for exported items
- Update docs/ for significant features
- Include examples in documentation

## 🔍 Review Process

1. **Automated Checks**: CI will run tests, linting, and coverage checks
2. **Code Review**: Maintainers will review your code
3. **Feedback**: Address review comments
4. **Approval**: Once approved, maintainers will merge your PR

## 🐛 Debugging Tips

### Running with verbose logging

```bash
./bin/juleson --verbose analyze --project ./test-project
```

### Running specific tests

```bash
go test -v -run TestFunctionName ./internal/package
```

### Checking test coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📦 Releasing

Releases are handled by maintainers. Version numbers follow [Semantic Versioning](https://semver.org/):

- MAJOR: Breaking changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes (backward compatible)

## ❓ Questions?

- Open an [issue](https://github.com/SamyRai/Juleson/issues) for questions
- Check existing [documentation](docs/)
- Review [closed issues] for similar problems

[closed issues]: https://github.com/SamyRai/Juleson/issues?q=is%3Aissue+is%3Aclosed

## 📜 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Juleson! 🎉
