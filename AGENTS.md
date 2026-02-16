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
- **Zero value handling**: Use pointer types (*float64) for numeric config fields when zero is a valid setting
## Provider Implementation

- **Request mutation**: Copy pointer params before modification to avoid mutating caller's data
- **.env support**: Document external .env loading (godotenv) rather than embedding it in provider
- **Retryable errors**: 429, 500, 502, 503, 504 are retryable; use exponential backoff (2^n seconds)
- **Context cancellation**: Check ctx.Err() after network failures and in retry loops
- **Google Cloud auth**: Use oauth2.NewClient with TokenSource for ADC; simpler than custom transport
- **API format mapping**: Role mapping (assistant→model, tool→function) and response format conversion critical for Vertex AI
- **Vertex AI system messages**: Use systemInstruction field, not system role (Gemini doesn't support it)
- **Tool call IDs**: Use crypto/rand for unique IDs; time.Now().UnixNano() can collide in parallel calls

## Tool System

- **JSON Schema validation**: Handle nil values in optional fields separately from required fields
- **Type compatibility**: Support both []string and []any for required field arrays (Go JSON decoding)
- **Thread safety**: Use sync.RWMutex for registries with read-heavy workloads
- **Builder pattern**: Fluent API with StringParam, IntegerParam, etc. improves ergonomics vs raw schema
- **HTTP tool**: Set User-Agent by default; many APIs reject requests without it
- **Method validation**: Use StringParamWithEnum for HTTP method validation (type safety + clear API)
- **Shell tool**: Use "sh -c" for proper shell parsing; default blocked commands for safety (rm -rf /, mkfs, fork bombs)
- **Command safety**: Support both allowlist (AllowedCommands) and blocklist (BlockedCommands); blocklist takes precedence
- **Output limiting**: Truncate command output at configured limit to prevent memory exhaustion
- **Exit codes**: Use -1 for non-exit errors (timeout, command not found); extract actual exit code from exec.ExitError
- **K8s tool**: Use client-go config chain (kubeconfig path → KUBECONFIG env → ~/.kube/config → in-cluster); automatic fallback
- **K8s namespaces**: Support special "all" value (converted to metav1.NamespaceAll) for cluster-wide queries
- **Heavy params**: Pass large structs (>80 bytes) by pointer to query functions to avoid copying overhead
- **String constants**: Extract repeated string literals (especially for enum-like values) as package constants for maintainability

## Agent Loop

- **Iteration limit**: Default 10 max iterations prevents infinite loops; configurable via Config
- **Context checks**: Check ctx.Done() at start of each iteration (not just once at beginning)
- **Tool result conversion**: Convert ToolResult to Message with RoleTool for next LLM call
- **State tracking**: Track iterations, total tokens, messages, execution time in RunResult
- **Nil slice checks**: Use `len(slice) > 0` not `slice != nil && len(slice) > 0` (gosimple S1009)
