.PHONY: all build clean test coverage lint fmt help install

# Variables
BINARY_CLI=juleson
BINARY_MCP=juleson-mcp
BIN_DIR=bin
CMD_CLI_DIR=cmd/juleson
CMD_MCP_DIR=cmd/jules-mcp
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-trimpath

all: clean lint test build

## build: Build all binaries
build: build-cli build-mcp

## build-cli: Build CLI binary
build-cli:
	@echo "Building CLI binary..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_CLI) ./$(CMD_CLI_DIR)

## build-mcp: Build MCP server binary
build-mcp:
	@echo "Building MCP server binary..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_MCP) ./$(CMD_MCP_DIR)

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

## test-short: Run short tests
test-short:
	@echo "Running short tests..."
	$(GOTEST) -v -short ./...

## coverage: Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

## lint: Run linters
lint:
	@echo "Running linters..."
	$(GOVET) ./...
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

## tidy: Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

## install: Install binaries to $GOPATH/bin
install: build
	@echo "Installing binaries..."
	@cp $(BIN_DIR)/$(BINARY_CLI) $(GOPATH)/bin/
	@cp $(BIN_DIR)/$(BINARY_MCP) $(GOPATH)/bin/
	@echo "Installed to $(GOPATH)/bin/"

## run-cli: Run CLI (requires PROJECT_PATH)
run-cli: build-cli
	@./$(BIN_DIR)/$(BINARY_CLI) $(ARGS)

## run-mcp: Run MCP server
run-mcp: build-mcp
	@./$(BIN_DIR)/$(BINARY_MCP)

## dev: Development mode with live reload (requires air)
dev:
	@which air > /dev/null || (echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest" && exit 1)
	air

## check: Run all checks (lint, test, build)
check: lint test build

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.DEFAULT_GOAL := help
