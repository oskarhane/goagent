#!/bin/bash
# Lint script for goagent

set -e

echo "Running linter for goagent..."

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint not found. Installing..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# Run linter
golangci-lint run --timeout=5m

echo "Linting complete!"