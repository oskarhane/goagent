# Agent-as-Tool Example

Demonstrates the **agent-as-tool pattern** for hierarchical agent delegation in SRE incident response. A coordinator agent delegates service investigations to specialized investigator agents (wrapped as tools), enabling complex multi-level AI workflows.

## What It Does

This example implements a two-tier agent hierarchy:
- **Coordinator agent**: Receives incident reports, identifies affected services, and orchestrates investigation by delegating to investigator agents
- **Investigator agents**: Wrapped as tools (`investigate_service`), each agent performs focused service-level diagnostics using mock tools (logs, metrics, status)
- **Result synthesis**: Findings from investigator agents bubble back to the coordinator for root cause analysis and recommendations

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Coordinator Agent                         │
│  Role: Triage, delegation, correlation, synthesis           │
│  Tools: investigate_service (agent-as-tool wrapper)         │
│  Max Iterations: 10                                          │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ Delegates via investigate_service tool
                           │
         ┌─────────────────┴─────────────────┐
         │                                   │
         ▼                                   ▼
┌──────────────────────┐          ┌──────────────────────┐
│  Investigator Agent  │          │  Investigator Agent  │
│  (auth-service)      │          │  (api-service)       │
│  Max Iterations: 5   │   ...    │  Max Iterations: 5   │
└────────┬─────────────┘          └────────┬─────────────┘
         │                                  │
         │ Uses diagnostic tools            │
         │                                  │
         ├─→ mock_logs                      ├─→ mock_logs
         ├─→ mock_metrics                   ├─→ mock_metrics
         └─→ mock_service_status            └─→ mock_service_status
```

### Key Components

1. **Mock Diagnostic Tools** (`mocks.go`):
   - `mock_logs`: Returns simulated log entries with errors, warnings, and service-specific patterns
   - `mock_metrics`: Provides time-series data (CPU, memory, request rate, error rate)
   - `mock_service_status`: Returns health check results and dependency information

2. **Investigator Factory** (`investigator.go`):
   - `NewInvestigator(serviceName, provider, registry)`: Creates service-scoped agent
   - Dynamic system prompt includes service name and investigation methodology
   - Max 5 iterations for focused investigation

3. **Agent-as-Tool Wrapper** (`agent_tool.go`):
   - `NewInvestigateServiceTool()`: Defines tool schema with `service_name` and `incident_type` parameters
   - `NewInvestigateServiceHandler()`: Creates inner investigator agent, runs it, converts result to `ToolResult`
   - Propagates errors from inner agent to outer agent cleanly

4. **Coordinator Agent** (`main.go`):
   - Receives incident description (CLI arg or default scenario)
   - System prompt emphasizes triage, delegation, and synthesis
   - Max 10 iterations to allow multiple service investigations

5. **Incident Scenarios** (`scenarios.go`):
   - Pre-built scenarios: cascading failure, memory leak, dependency failure
   - Aligned with mock data for realistic demonstrations

## Prerequisites

- **OpenAI API key** - For LLM reasoning (gpt-4 or compatible model)
- **Go 1.26+**

## Quick Start

### 1. Set up environment

Create `.env` file:

```bash
OPENAI_API_KEY=sk-your-key-here
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Run the agent

**Default scenario (cascading failure):**

```bash
go run main.go
```

**Custom incident description:**

```bash
go run main.go "Cache service memory leak causing timeouts"
```

**Debug logging:**

```bash
DEBUG=true go run main.go
```

## Example Output

> **Note**: Output below is illustrative. Actual token counts, iteration numbers, and analysis may vary based on LLM responses.

```
=== Agent-as-Tool: Hierarchical SRE Investigation ===

--- Incident Report ---
INCIDENT REPORT:

Type: cascading_failure
Report: Multiple services reporting degraded performance. Users experiencing login failures 
and API timeouts. Started approximately 15 minutes ago with increased error rates on 
authentication endpoints.

Your task is to investigate this incident, determine the root cause, and provide 
recommendations for resolution.

--- Delegation Hierarchy ---
Coordinator → Investigator Agents:
  1. Investigating: database
  2. Investigating: auth-service
  3. Investigating: api-service

--- Investigation Findings ---

Service Investigation #1:
Investigation of database:

The database is experiencing HIGH LOAD conditions that are causing cascading failures:

EVIDENCE:
- Logs show "connection pool 95% utilized" warnings starting 20 minutes ago
- Slow query detected: SELECT * FROM users WHERE ... (avg 2.5s, normally 50ms)
- Metrics show CPU at 89%, memory at 76%
- Health status: degraded, response time 1200ms (normally 100ms)

ROOT CAUSE: Database connection pool saturation due to slow query performance
---

Service Investigation #2:
Investigation of auth-service:

The auth-service is UNHEALTHY and directly impacted by database issues:

EVIDENCE:
- Logs show "database timeout after 1000ms" errors (150+ occurrences in last 10min)
- Error rate: 45% (baseline 0.2%)
- Metrics show request latency spiked from 50ms to 3500ms
- Health status: unhealthy, dependency (database) failing

CONCLUSION: auth-service timeouts are a symptom of upstream database degradation
---

Service Investigation #3:
Investigation of api-service:

The api-service is experiencing CASCADING FAILURES from auth-service:

EVIDENCE:
- Logs show "authentication service unavailable" errors
- 401/403 errors increased 800% in last 15 minutes
- Metrics show error rate at 38% (baseline 0.5%)
- Health status: degraded, auth dependency failing

CONCLUSION: api-service degradation is secondary to auth-service timeouts
---

=== Coordinator Final Analysis ===

ROOT CAUSE IDENTIFIED:
Database connection pool exhaustion caused by slow query performance has created a cascading 
failure affecting auth-service and api-service.

FAILURE CHAIN:
1. Database: Slow queries → connection pool saturation (95% utilized)
2. Auth-service: Database timeouts → 45% error rate
3. API-service: Auth failures → 38% error rate → user-facing errors

AFFECTED SERVICES:
- database (root cause): Degraded, connection pool exhausted
- auth-service (tier 1 impact): Unhealthy, database dependency failing
- api-service (tier 2 impact): Degraded, auth dependency failing

RECOMMENDATIONS:
1. IMMEDIATE: Kill slow query or optimize query plan to restore database performance
2. IMMEDIATE: Scale database connection pool from current limit to handle load
3. SHORT-TERM: Add connection pool monitoring and alerting at 80% threshold
4. SHORT-TERM: Implement circuit breakers in auth-service to fail fast when database is slow
5. LONG-TERM: Review query performance and add indexes to prevent slow queries

EXPECTED RECOVERY:
- Database optimization should reduce connection pool usage within 2-3 minutes
- Auth-service errors should decrease once database responds normally
- API-service should return to baseline within 5 minutes of auth recovery

=== Execution Statistics ===
Coordinator iterations: 7
Total tokens used: 4823
Total execution time: 18.34s
```

## When to Use This Pattern

### ✅ Use agent-as-tool when:

- **Hierarchical decision-making**: Need triage/routing layer before specialized investigation
- **Specialized scopes**: Different agents have distinct responsibilities (service-level, namespace-level, region-level)
- **Complex interpretation**: Tool results require LLM reasoning that benefits from focused context
- **Token efficiency**: Scoping sub-agents to specific contexts reduces token usage vs single large agent
- **Parallel delegation**: Multiple sub-agents can investigate different services concurrently
- **Reusability**: Same sub-agent logic can be called multiple times with different parameters

### ❌ Don't use this pattern when:

- **Simple workflows**: Single agent with multiple tools suffices
- **Deterministic outputs**: Tool results don't need LLM interpretation
- **Latency critical**: Each agent level adds round-trip time (LLM call)
- **Minimal complexity**: Overhead of agent-as-tool wrapper not justified

## Code Walkthrough

### 1. Mock Tools (`mocks.go`)

Simulate diagnostic data for demo purposes. Each tool:
- Accepts `service_name` parameter
- Returns deterministic, realistic data based on service
- Demonstrates cascading failure patterns (database → auth-service → api-service)

```go
// Example: mock_logs returns service-specific log entries
registry.MustRegister(NewMockLogsTool(), NewMockLogsHandler())
```

**Key functions:**
- `generateMockLogs()`: Creates log entries with errors/warnings based on service
- `generateMockMetrics()`: Returns time-series metrics (CPU, memory, error rate)
- `generateMockServiceStatus()`: Provides health status and dependency info

### 2. Investigator Agent Factory (`investigator.go`)

Creates service-scoped agents for focused investigation:

```go
investigator, err := NewInvestigator("auth-service", provider, mockRegistry)
```

**Features:**
- Dynamic system prompt includes service name 5x for contextual grounding
- Max 5 iterations for focused investigation (prevents runaway)
- Access to all mock diagnostic tools
- Defensive validation (empty service name, nil provider, nil registry)

### 3. Agent-as-Tool Wrapper (`agent_tool.go`)

Core pattern implementation - wraps investigator agent as a tool:

```go
// Tool definition
tool := NewInvestigateServiceTool()

// Handler creates and runs inner agent, returns result to outer agent
handler := NewInvestigateServiceHandler(provider, mockRegistry)
```

**How it works:**
1. Coordinator calls `investigate_service(service_name="auth-service", incident_type="cascading_failure")`
2. Handler parses parameters from tool call
3. Handler creates investigator agent scoped to `service_name`
4. Handler runs investigator with prompt: "Investigate auth-service for cascading_failure..."
5. Handler formats investigator's response as `ToolResult` with content and stats
6. Coordinator receives investigation findings as tool result

**Error propagation:**
- Inner agent errors (timeout, LLM failure) converted to `ToolResult.Error`
- Coordinator receives clean error message without crashing

### 4. Coordinator Agent (`main.go`)

Top-level orchestrator:

```go
coordinator, err := agent.NewAgent(&agent.Config{
    Provider:      provider,
    SystemPrompt:  coordinatorPrompt, // Triage and delegation strategy
    Registry:      coordinatorRegistry, // Only has investigate_service tool
    MaxIterations: 10,
    Logger:        l,
})
```

**Execution flow:**
1. Accept incident description (CLI args or default scenario)
2. Run coordinator with 60s timeout (prevents runaway)
3. Coordinator analyzes incident, identifies services to investigate
4. Coordinator calls `investigate_service` tool for each service
5. Coordinator receives investigation findings via tool results
6. Coordinator synthesizes findings, determines root cause, provides recommendations

**Output formatting:**
- Delegation hierarchy: Shows which services were investigated
- Investigation findings: Displays each service investigation result
- Final analysis: Coordinator's synthesis and recommendations
- Execution stats: Iterations, tokens, duration

### 5. Incident Scenarios (`scenarios.go`)

Pre-built scenarios for demonstration:

```go
scenarios := NewScenarios()
scenario := scenarios.Default() // Returns "Database Cascading Failure"
incident := scenario.GetIncidentDescription()
```

**Available scenarios:**
1. **Cascading Failure**: Database → auth-service → api-service
2. **Memory Leak**: Cache service gradual degradation
3. **Dependency Failure**: Auth-service → database network partition

Mock data aligned with scenarios for realistic demonstrations.

## Customization

### Replace Mock Tools with Real Implementations

```go
// Use real Kubernetes tool instead of mocks
k8sTool := k8s.NewTool(&k8s.Config{
    Namespace: "production",
})
mockRegistry.MustRegister(k8sTool, k8s.NewHandler(&k8s.Config{}))

// Use HTTP tool to query real monitoring APIs
httpTool := http.NewTool()
mockRegistry.MustRegister(httpTool, http.NewHandler(&http.Config{}))
```

### Add Additional Agent Levels

**Three-tier hierarchy** (coordinator → triage → investigator):

```go
// Triage agent wrapped as tool for coordinator
triageTool := NewTriageServiceTool()
triageHandler := NewTriageServiceHandler(provider, investigatorRegistry)

// Triage agent has access to investigator agents
triageRegistry := tools.NewRegistry()
triageRegistry.MustRegister(NewInvestigateServiceTool(), NewInvestigateServiceHandler(...))
```

**Parallel investigation** (multiple services concurrently):

Coordinator can make multiple `investigate_service` calls in single iteration. Framework handles parallel tool execution automatically.

### Change LLM Provider

```go
// Use Vertex AI (Gemini) instead of OpenAI
provider := vertex.NewProvider(&vertex.Config{
    ProjectID: "my-project",
    Location:  "us-central1",
    Model:     "gemini-2.0-flash",
})
```

### Adjust Iteration Limits

```go
// Increase coordinator iterations for complex investigations
MaxIterations: 15

// Increase investigator iterations for thorough service analysis
MaxIterations: 8
```

### Add Conversation History

For multi-turn investigation (follow-up questions):

```go
// Initial investigation
result1 := coordinator.Run(ctx, incident, nil)

// Follow-up with context
result2 := coordinator.Run(ctx, "Focus on database query performance", &agent.RunOptions{
    History: result1.Messages,
})
```

## Testing

Run tests:

```bash
go test ./...
```

**Test coverage:**
- `mocks_test.go`: Tool definitions, handlers, deterministic data, cascading failures
- `scenarios_test.go`: Factory, getters, default scenario, mock data alignment
- `investigator_test.go`: Agent creation, validation, error cases
- `agent_tool_test.go`: Tool creation, handler behavior, error propagation
- `main_test.go`: Helper functions for service extraction and findings parsing

## Troubleshooting

**"OPENAI_API_KEY environment variable is required"**
- Create `.env` file with API key
- Or: `export OPENAI_API_KEY=sk-...`

**"context deadline exceeded"**
- Increase timeout in main.go: `context.WithTimeout(ctx, 120*time.Second)`
- Or: Check network connectivity to OpenAI API

**"Failed to create investigator agent"**
- Check service name is not empty
- Verify provider and registry are not nil

**Agent stops after max iterations**
- Increase `MaxIterations` in coordinator config (currently 10)
- Increase `MaxIterations` in investigator config (currently 5)
- Check if LLM is stuck in loop (enable DEBUG=true)

## Related Examples

- [Incident Response](../incident-response/) - Single agent with K8s/HTTP/shell tools
- [K8s Monitoring](../k8s-monitoring/) - Proactive cluster monitoring
- [Agent Basic](../agent-basic/) - Core agent concepts
- [Tools Basic](../tools-basic/) - Tool registration fundamentals

## Performance Characteristics

**Token usage:**
- Coordinator system prompt: ~250 tokens
- Investigator system prompt: ~200 tokens per service
- Tool results: ~300-500 tokens per investigation
- Total: ~2000-5000 tokens for 2-3 service investigations

**Latency:**
- Each investigator agent: 1-3 LLM calls (5-15 seconds)
- Coordinator synthesis: 2-4 LLM calls (10-20 seconds)
- Total: 15-35 seconds for typical incident

**Cost optimization:**
- Use smaller models for investigator agents (gpt-4-turbo or gpt-3.5-turbo)
- Use larger models for coordinator synthesis (gpt-4)
- Limit investigator iterations to reduce token usage
