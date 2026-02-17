# Architecture

ARCHITECTURE PATTERN: Modular layered Go SDK (core agent loop + provider adapters + tool plugins) with thin CLI
DIRECTORY STRUCTURE: Go standard layout `cmd/` CLI, `pkg/` library modules, `examples/`, `deployments/`, `scripts/`, `docs/`; no `src/`, `lib/`, `app/`
DESIGN PATTERNS: Strategy (Provider interface), registry/command for tools, builder for JSON schema tools, adapter for provider API mapping, config-struct DI
DATABASE: None; no ORM/schema
API DESIGN: Provider REST clients (OpenAI/Vertex); no first-party REST/GraphQL endpoints

- Code organization: core in `pkg/agent`, shared types in `pkg/types`, tool system in `pkg/tools`, providers in `pkg/providers/*`, logging in `pkg/logger`
- Config mgmt: env vars + `.env` (examples), K8s ConfigMap/Secret templates in `deployments/k8s/base`, Docker env defaults in `Dockerfile`, Make targets in `Makefile`
- Dependency injection: `agent.Config` injects `Provider`, `Registry`, `Logger`; providers accept `HTTPClient` + config; tool handlers created with config
- Error handling: typed `ProviderError` with retryable flag, retry w/ backoff, context cancel checks, tool arg validation + structured tool error strings
- Logging/monitoring: JSON structured logger + optional OpenTelemetry spans; K8s manifests include Prometheus scrape annotations (no metrics handler shown)
- Security: shell tool allow/blocklist + timeouts, k8s RBAC Role + SA, non-root container + read-only FS + drop caps, secrets for API keys
- Performance: max iterations, history trimming, per-tool timeouts, HTTP response size cap, shell output truncation, exponential backoff retries

---

*This file is part of the AGENTS.md documentation system.*
