# GitHub Actions Guide for Juleson

## Overview

This guide covers GitHub Actions best practices, patterns, and specific configurations for the
Juleson project. It combines practical examples with security best practices for
internal development.

## Table of Contents

- [Basic Workflow Structure](#basic-workflow-structure)
- [Security Best Practices](#security-best-practices)
- [Go-Specific Workflows](#go-specific-workflows)
- [Common Patterns](#common-patterns)
- [Juleson CI/CD Setup](#Juleson-cicd-setup)
- [Troubleshooting & Debugging](#troubleshooting--debugging)
- [Reference](#reference)

## Basic Workflow Structure

### Minimal Secure Workflow

```yaml
name: CI
on: [push, pull_request]
permissions:
  contents: read
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v6
        with:
          go-version: '1.23'
          cache: true
      - run: go test ./...
```

### Triggers

```yaml
# Basic triggers
on: [push, pull_request]

# Advanced triggers with filters
on:
  push:
    branches: [main, develop]
    paths:
      - 'cmd/**'
      - 'internal/**'
      - '**.go'
    paths-ignore:
      - '**.md'
      - 'docs/**'

# Scheduled runs
on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly security scan

# Manual workflow dispatch
on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Deployment environment'
        required: true
        type: choice
        options: [dev, staging, prod]
```

## Security Best Practices

### 1. Minimal Permissions

```yaml
# Workflow level - start with minimal
permissions:
  contents: read

# Job level - add only what's needed
jobs:
  security-scan:
    permissions:
      contents: read
      security-events: write

  release:
    permissions:
      contents: write
      packages: write
      id-token: write
```

### 2. Concurrency Control

```yaml
# Cancel outdated runs to save resources
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

# For production deployments (don't cancel)
concurrency:
  group: production-deploy
  cancel-in-progress: false
```

### 3. Pin Actions to Specific Versions

```yaml
# Good - specific version
- uses: actions/checkout@v5
- uses: actions/setup-go@v6

# Better - pin to SHA for critical workflows
- uses: actions/checkout@8ade135a41bc03ea155e62e844d188df1ea18608
```

### 4. Environment Variables and Secrets

```yaml
env:
  GO_VERSION: "1.23"

jobs:
  test:
    env:
      GOOS: linux
      GOARCH: amd64
    steps:
      - name: Use secret
        env:
          API_KEY: ${{ secrets.API_KEY }}
        run: |
          echo "::add-mask::$API_KEY"
          ./test-with-api.sh
```

## Go-Specific Workflows

### Complete CI Workflow for Go Projects

```yaml
name: CI
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

permissions:
  contents: read

env:
  GO_VERSION: "1.23"

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v5

      - name: Setup Go
        uses: actions/setup-go@v6
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Verify dependencies
        run: go mod verify

      - name: Build
        run: go build -v ./...

      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        if: github.event_name == 'push'
        with:
          file: ./coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v6
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6

  security:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v6
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Run Gosec Security Scanner
        uses: securecodewarrior/github-action-gosec@master
        with:
          args: '-fmt sarif -out results.sarif ./...'

      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
```

### Multi-Platform Build Matrix

```yaml
jobs:
  build:
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        goarch: [amd64, arm64]
        exclude:
          - os: windows-latest
            goarch: arm64
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v6
        with:
          go-version: '1.23'
          cache: true

      - name: Build
        env:
          GOOS: ${{ runner.os == 'Linux' && 'linux' || runner.os == 'macOS' && 'darwin' || 'windows' }}
          GOARCH: ${{ matrix.goarch }}
        run: go build -o jules-${{ env.GOOS }}-${{ matrix.goarch }} ./cmd/juleson
```

## Common Patterns

### Conditional Execution

```yaml
# Step-level conditions
- name: Deploy to production
  if: github.ref == 'refs/heads/main' && github.event_name == 'push'
  run: ./deploy.sh

# Job-level conditions
jobs:
  deploy:
    if: startsWith(github.ref, 'refs/tags/v')
    runs-on: ubuntu-latest

# Complex conditions
- name: Conditional step
  if: |
    github.event_name == 'push' &&
    (github.ref == 'refs/heads/main' || startsWith(github.ref, 'refs/tags/'))
```

### Job Dependencies and Outputs

```yaml
jobs:
  build:
    outputs:
      version: ${{ steps.version.outputs.version }}
      artifact-name: ${{ steps.artifact.outputs.name }}
    steps:
      - id: version
        run: echo "version=$(git describe --tags)" >> $GITHUB_OUTPUT
      - id: artifact
        run: echo "name=jules-${{ steps.version.outputs.version }}" >> $GITHUB_OUTPUT

  test:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - run: echo "Testing version ${{ needs.build.outputs.version }}"

  deploy:
    needs: [build, test]
    if: success()
    runs-on: ubuntu-latest
    steps:
      - run: echo "Deploying ${{ needs.build.outputs.artifact-name }}"
```

### Reusable Workflows

```yaml
# .github/workflows/reusable-go-test.yml
name: Reusable Go Test

on:
  workflow_call:
    inputs:
      go-version:
        required: false
        type: string
        default: '1.23'
      working-directory:
        required: false
        type: string
        default: '.'
    secrets:
      codecov-token:
        required: false

jobs:
  test:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: ${{ inputs.working-directory }}
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v6
        with:
          go-version: ${{ inputs.go-version }}
          cache: true
      - run: go test -race -coverprofile=coverage.out ./...
      - uses: codecov/codecov-action@v4
        if: ${{ secrets.codecov-token }}
        with:
          token: ${{ secrets.codecov-token }}
```

## Juleson CI/CD Setup

### Enhanced CI Features

Our improved CI setup includes:

1. **Performance Optimizations:**
   - Built-in Go caching with `setup-go@v6`
   - Path filtering to skip runs on documentation changes
   - Concurrency control to cancel outdated runs

2. **Enhanced Security:**
   - Minimal permissions by default
   - CodeQL security scanning
   - Dependency vulnerability checks
   - SARIF upload for security findings

3. **Quality Assurance:**
   - Multi-platform testing (Linux, macOS, Windows)
   - Race condition detection (`go test -race`)
   - Markdown linting for documentation
   - Code coverage reporting

### Release Automation

```yaml
name: Release

on:
  push:
    tags: ['v*.*.*']
  workflow_dispatch:

permissions:
  contents: write
  packages: write

jobs:
  release:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]

    steps:
      - uses: actions/checkout@v5
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v6
        with:
          go-version: '1.23'

      - name: Build binaries
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          mkdir -p dist/
          for cmd in juleson jules-mcp; do
            go build -ldflags="-s -w" -o dist/${cmd}-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd/${cmd}
          done

      - name: Create release
        if: matrix.goos == 'linux' && matrix.goarch == 'amd64'
        uses: softprops/action-gh-release@v2
        with:
          files: dist/*
          generate_release_notes: true
```

### Dependency Management

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    labels:
      - "dependencies"
      - "go"
    commit-message:
      prefix: "feat"
      prefix-development: "chore"

  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
    labels:
      - "dependencies"
      - "github-actions"
```

## Troubleshooting & Debugging

### Enable Debug Logging

Add repository secrets:

- `ACTIONS_STEP_DEBUG` = `true` (step-level debugging)
- `ACTIONS_RUNNER_DEBUG` = `true` (runner-level debugging)

### Debug Context Information

```yaml
- name: Debug Context
  run: |
    echo "Event: ${{ github.event_name }}"
    echo "Ref: ${{ github.ref }}"
    echo "SHA: ${{ github.sha }}"
    echo "Actor: ${{ github.actor }}"

- name: Dump GitHub context
  run: echo '${{ toJSON(github) }}'

- name: Dump runner context
  run: echo '${{ toJSON(runner) }}'
```

### Common Issues and Solutions

1. **Cache not working:**
   - Ensure `cache: true` in `setup-go` action
   - Check cache key patterns for Go modules

2. **Permission denied:**
   - Add necessary permissions to job or workflow
   - Check token permissions for cross-repository access

3. **Build failures:**
   - Use `go mod verify` before building
   - Check Go version compatibility
   - Verify all dependencies are available

## Reference

### Available Runners

| Runner | OS | Architecture | vCPU | RAM | Storage |
|--------|----|--------------|------|-----|---------|
| `ubuntu-latest` | Ubuntu 22.04 | x64 | 4 | 16 GB | 14 GB |
| `ubuntu-24.04` | Ubuntu 24.04 | x64 | 4 | 16 GB | 14 GB |
| `macos-latest` | macOS 14 | x64 | 3 | 14 GB | 14 GB |
| `macos-14-arm64` | macOS 14 | ARM64 | 3 | 14 GB | 14 GB |
| `windows-latest` | Server 2022 | x64 | 4 | 16 GB | 14 GB |

### Essential Context Variables

```yaml
${{ github.repository }}        # owner/repo-name
${{ github.ref }}              # refs/heads/main or refs/tags/v1.0.0
${{ github.ref_name }}         # main or v1.0.0
${{ github.sha }}              # commit SHA
${{ github.actor }}            # username who triggered workflow
${{ github.event_name }}       # push, pull_request, schedule, etc.
${{ runner.os }}               # Linux, Windows, macOS
${{ runner.arch }}             # X64, ARM64
```

### Useful Actions for Go Projects

- `actions/checkout@v5` - Check out repository
- `actions/setup-go@v6` - Set up Go environment with caching
- `golangci/golangci-lint-action@v6` - Run Go linter
- `codecov/codecov-action@v4` - Upload coverage reports
- `github/codeql-action@v3` - Security analysis
- `actions/upload-artifact@v5` - Upload build artifacts

### Best Practices Checklist

- ✅ Use minimal permissions
- ✅ Pin actions to specific versions
- ✅ Enable concurrency control
- ✅ Use path filters for efficiency
- ✅ Cache dependencies appropriately
- ✅ Set reasonable timeout limits
- ✅ Use environment variables for repeated values
- ✅ Enable security scanning (CodeQL, Gosec)
- ✅ Test on multiple platforms when relevant
- ✅ Upload artifacts for debugging
- ✅ Use secrets for sensitive data
- ✅ Document workflow purpose and usage

## Additional Resources

- [GitHub Actions Security Best Practices](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions)
- [Workflow Syntax Reference](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Go with GitHub Actions](https://docs.github.com/en/actions/automating-builds-and-tests/building-and-testing-go)
- [Dependabot Configuration](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file)
