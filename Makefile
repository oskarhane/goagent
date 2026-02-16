# GoAgent SDK Makefile

.PHONY: help build test lint clean install docs dev-setup

# Default target
help: ## Show this help message
	@echo "GoAgent SDK - Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development setup
dev-setup: ## Install development dependencies
	@echo "Installing development dependencies..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/godoc@latest
	@echo "Development dependencies installed!"

# Build
build: ## Build the CLI binary
	@echo "Building goagent..."
	@go build -o bin/goagent ./cmd/goagent
	@echo "Built: bin/goagent"

install: ## Install the CLI binary to $GOPATH/bin
	@echo "Installing goagent..."
	@go install ./cmd/goagent
	@echo "Installed goagent to $(shell go env GOPATH)/bin"

# Testing
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Linting
lint: ## Run linting
	@echo "Running linter..."
	@PATH=$(shell go env GOPATH)/bin:$(PATH) golangci-lint run

lint-fix: ## Run linting with auto-fix
	@echo "Running linter with auto-fix..."
	@PATH=$(shell go env GOPATH)/bin:$(PATH) golangci-lint run --fix

# Code formatting
fmt: ## Format code
	@echo "Formatting code..."
	@PATH=$(shell go env GOPATH)/bin:$(PATH) goimports -w . || go fmt ./...
	@go fmt ./...

# Documentation
docs: ## Generate documentation
	@echo "Starting documentation server..."
	@echo "Visit http://localhost:6060/pkg/github.com/oskarhane/goagent/"
	@godoc -http=:6060

# Clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# CI targets
ci-test: ## Run tests for CI
	@go test -v -race -coverprofile=coverage.out ./...

ci-lint: ## Run linting for CI  
	@PATH=$(shell go env GOPATH)/bin:$(PATH) golangci-lint run --timeout=5m

# Development workflow
dev: fmt lint test build ## Run development workflow (fmt, lint, test, build)
	@echo "Development workflow complete!"