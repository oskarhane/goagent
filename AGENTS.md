# Agent Learnings

## Go Project Setup

- **Version injection**: Use `var version = "dev"` not `const` for ldflags compatibility
- **Go version**: Use Go 1.25 for broader CI compatibility vs latest 1.26
- **Empty dirs**: Git won't track empty directories; use `.gitkeep` or remove them
- **godoc**: No need for docs/ dir; use `make docs` to serve on-demand

## CI/CD

- **Multi-version testing**: Test against N and N-1 Go versions for compatibility
- **golangci-lint**: Configure via `.golangci.yml` for consistent linting
- **GoReleaser**: Set up early for proper version/ldflags injection

## Build System

- **Makefile targets**: Provide test/lint/build/docs for dev workflow
- **Shell scripts**: Keep build logic in scripts/ for CI reuse
- **Test output**: `[no test files]` is expected until task-015; not an error
