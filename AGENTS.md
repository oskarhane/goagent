# Agent Learnings

## Go Project Setup

- **Version injection**: Use `var version = "dev"` not `const` for ldflags compatibility
- **Go version**: Use Go 1.25 for broader CI compatibility vs latest 1.26
- **Empty dirs**: Git won't track empty directories; use `.gitkeep` or remove them
- **godoc**: No need for docs/ dir; use `make docs` to serve on-demand

## CI/CD

- **Multi-version testing**: Test against N and N-1 Go versions for compatibility
- **golangci-lint**: Configure via `.golangci.yml` for consistent linting; use `colored-line-number` format (not deprecated `github-actions`)
- **Go version alignment**: Ensure all CI jobs (test, lint, security) use same Go version as go.mod to avoid tooling incompatibilities
- **Gosec scanner**: Use `securego/gosec@master` action (not `securecodewarrior/github-action-gosec` which doesn't exist)
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
- **OpenAI token limits**: gpt-5+ models use max_completion_tokens instead of max_tokens; detect model version and use appropriate field

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
- **Conversation history**: Pass previous Messages via RunOptions.History to maintain context across interactions
- **History limiting**: Use MaxHistoryMessages to trim history from oldest messages when limit exceeded
- **History serialization**: Messages serialize/deserialize to JSON cleanly for persistence (no custom marshaling needed)
- **Tool message integrity**: When trimming history, preserve assistant+tool_calls with their corresponding tool messages to avoid API errors

## Logging and Tracing

- **Logger injection**: Pass logger to Agent and Provider configs; defaults to Noop logger if nil
- **Log levels**: Debug (iteration details), Info (major events), Warn (retries, errors), Error (failures)
- **Structured output**: All logs JSON-formatted with timestamp, level, message, fields
- **OpenTelemetry**: Optional tracing via TracerName in logger config; creates spans for agent.run, provider.complete
- **Span management**: Use defer pattern for EndSpan to capture errors; check span != nil before recording events
- **Config pointers**: Use pointer receivers for Config structs >80 bytes to avoid copying overhead (hugeParam)

## Docker Deployment

- **Multi-stage builds**: Separate builder and runtime stages for minimal image size (~20MB final)
- **Non-root user**: Always run as UID 1000; create user in Dockerfile with adduser/addgroup
- **Health checks**: Use `--version` flag as basic health indicator; adjust interval/timeout per use case
- **Base images**: Use golang:1.25-alpine for build, alpine:latest for runtime
- **Security**: Include ca-certificates for HTTPS, never include secrets in image layers
- **Image variants**: Provide specialized Dockerfiles for different deployment patterns (service, cronjob)
- **Build args**: Use VERSION build arg for ldflags injection during docker build
- **.dockerignore**: Exclude examples/, .plans/, .env files, and test artifacts to reduce context size
- **CMD placeholder**: Use `["--help"]` as CMD for unimplemented modes (service/cronjob) rather than non-existent subcommands

## Kubernetes Deployment

- **RBAC minimum**: Start with namespace-scoped Role; only use ClusterRole for cross-namespace queries
- **RBAC namespace**: Don't hardcode namespace in RoleBinding subjects; let it inherit from metadata
- **Security context**: Always set runAsNonRoot, readOnlyRootFilesystem, drop ALL capabilities, seccompProfile RuntimeDefault
- **Resource limits**: Set both requests and limits; CronJobs typically need less than Deployments
- **Secret management**: Use stringData for templates, never commit real values; prefer external secret operators
- **Config separation**: ConfigMap for non-sensitive settings, Secret for credentials
- **Volume mounts**: Mount secrets read-only; use emptyDir for /tmp with readOnlyRootFilesystem
- **Health checks**: Use exec probes with --version until HTTP endpoints implemented
- **Job settings**: Set activeDeadlineSeconds for CronJobs to prevent runaway execution; use Forbid concurrencyPolicy
- **Observability**: Include Prometheus annotations and optional ServiceMonitor/PodMonitor
- **PodMonitor caveat**: CronJobs don't expose metrics endpoints by default; PodMonitor for future push-based metrics
- **Examples**: Provide working examples for common patterns (monitoring, incident response)

## Documentation

- **Quickstart accuracy**: Verify code examples compile; common mistakes: tools.Handler vs *tools.Registry type mismatch
- **Return values**: agent.Run() returns *RunResult (no error tuple); check result.Error not err
- **Registry pattern**: tools.NewRegistry() + registry.MustRegister(tool, handler) is correct setup
- **Example completeness**: Include .env.example files, READMEs, go.mod for each example
- **Godotenv pattern**: Use _ "github.com/joho/godotenv/autoload" for easy .env loading in examples

## Testing

- **Test framework**: Use testify/assert and testify/require for assertions
- **Mock providers**: Create simple mock providers implementing types.Provider interface for agent testing
- **Test structure**: Organize tests with table-driven tests for parameter validation
- **Coverage focus**: Prioritize core logic (types, agent, tools) over provider implementations
- **HTTP testing**: Use httptest.NewServer for testing HTTP tool without external dependencies
- **JSON assertions**: JSON marshaling may not include spaces; use Contains for partial matching
- **Context cancellation**: Test context cancellation but allow 1 iteration before check (agent checks at loop start)
- **Builder API**: Test builder pattern ensures proper JSON Schema structure with properties and required arrays
- **Thread safety**: Test concurrent tool execution and logging to verify mutex usage

## Testing

- **Test framework**: Use testify/assert and testify/require for assertions
- **Mock providers**: Create simple mock providers implementing types.Provider interface for agent testing
- **Test structure**: Organize tests with table-driven tests for parameter validation
- **Coverage focus**: Prioritize core logic (types, agent, tools) over provider implementations
- **HTTP testing**: Use httptest.NewServer for testing HTTP tool without external dependencies
- **JSON assertions**: JSON marshaling may not include spaces; use Contains for partial matching
- **Context cancellation**: Test context cancellation but allow 1 iteration before check (agent checks at loop start)
- **Builder API**: Test builder pattern ensures proper JSON Schema structure with properties and required arrays
- **Thread safety**: Test concurrent tool execution and logging to verify mutex usage
- **Provider mocking**: Use httptest for provider tests; mock server routes requests properly
- **Error assertions**: Use flexible checks (Contains/True) for error messages that vary by implementation
- **Shell tests**: Use safe commands only (echo, pwd, ls); avoid platform-specific behavior
- **Flaky tests**: Skip timing-sensitive tests (context cancel) that can't be made deterministic
- **Lint compliance**: Fix unused params with `_`, add missing imports, split long lines, use American spelling
- **Complex algorithm tests**: Add dedicated unit tests for complex functions (like trimming with constraints); test edge cases thoroughly to prevent regression
