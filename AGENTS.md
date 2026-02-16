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

## API Design

- **Struct field order**: Group logically (identity fields first, then content, then metadata) for API stability
- **JSON tag consistency**: Use omitempty for optional fields; explicit tags prevent refactor breakage
- **Helper constructors**: Provide NewXMessage() functions for common message types (better UX)
## Provider Implementation

- **Request mutation**: Copy pointer params before modification to avoid mutating caller's data
- **.env support**: Document external .env loading (godotenv) rather than embedding it in provider
- **Retryable errors**: 429, 500, 502, 503, 504 are retryable; use exponential backoff (2^n seconds)
- **Context cancellation**: Check ctx.Err() after network failures and in retry loops
