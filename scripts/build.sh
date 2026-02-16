#!/bin/bash
# Build script for goagent

set -e

echo "Building goagent..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Build the CLI
go build -ldflags="-s -w" -o bin/goagent ./cmd/goagent

echo "Build complete: bin/goagent"

# Show binary info
if command -v file &> /dev/null; then
    file bin/goagent
fi

if command -v du &> /dev/null; then
    echo "Size: $(du -h bin/goagent | cut -f1)"
fi