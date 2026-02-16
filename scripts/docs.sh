#!/bin/bash
# Documentation generation script for goagent

set -e

echo "Starting documentation server..."

# Check if godoc is installed
if ! command -v godoc &> /dev/null; then
    echo "godoc not found. Installing..."
    go install golang.org/x/tools/cmd/godoc@latest
fi

# Start the documentation server
echo "Documentation will be available at:"
echo "  http://localhost:6060/pkg/github.com/oskarhane/goagent/"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

godoc -http=:6060