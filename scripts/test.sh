#!/bin/bash
# Test script for goagent

set -e

echo "Running tests for goagent..."

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report if coverage.out exists
if [ -f coverage.out ]; then
    echo "Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    
    # Show coverage percentage
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo "Total coverage: $COVERAGE"
    
    # Check if coverage meets threshold (80%)
    COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
    if (( $(echo "$COVERAGE_NUM >= 80" | bc -l) )); then
        echo "✅ Coverage meets requirement (>= 80%)"
    else
        echo "❌ Coverage below requirement (< 80%)"
        exit 1
    fi
fi

echo "Tests complete!"