# Agent-as-Tool Example

This example demonstrates the agent-as-tool pattern for hierarchical agent delegation in SRE incident response scenarios.

## What It Does

The example implements a coordinator-investigator pattern where:
- A coordinator agent receives incident reports and delegates investigation
- Individual investigator agents (wrapped as tools) perform service-level diagnostics
- Results bubble back up through the tool call hierarchy

## Architecture

```
Coordinator Agent
    |
    +--> investigate_service tool
            |
            +--> Investigator Agent (service-scoped)
                    |
                    +--> mock_logs tool
                    +--> mock_metrics tool
                    +--> mock_service_status tool
```

## Prerequisites

1. **OpenAI API key** - Set in environment variable
2. **Go 1.25+**

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

With default scenario:

```bash
go run main.go
```

With custom incident description:

```bash
go run main.go "Users reporting auth failures cascading to API errors"
```

For debug output:

```bash
DEBUG=true go run main.go
```

## Example Output

```
=== SRE Incident Response - Agent-as-Tool Pattern ===

Incident: Users reporting authentication failures...

[Coordinator] Delegating investigation to service agents...
[Coordinator] Tool call: investigate_service(service_name=auth-service)
  [Investigator:auth-service] Checking logs...
  [Investigator:auth-service] Analyzing metrics...
  [Investigator:auth-service] Result: Rate limiting triggered, upstream dependency slow
[Coordinator] Tool call: investigate_service(service_name=api-service)
  [Investigator:api-service] Checking service status...
  [Investigator:api-service] Result: Cascading failures from auth timeouts

=== Investigation Summary ===
Root cause: Auth service rate limiting due to upstream dependency latency
Affected services: auth-service, api-service
Recommendations: [...]

=== Execution Stats ===
Iterations: 8
Total tokens: 3421
Duration: 15.2s
```

## When to Use This Pattern

Use agent-as-tool delegation when:
- You need hierarchical decision-making (triage → investigation)
- Different agents have specialized scopes (service-level, namespace-level)
- Tool results require complex analysis that benefits from LLM reasoning
- You want to limit token usage by scoping sub-agents to specific contexts

Don't use this pattern when:
- A single agent with multiple tools suffices
- Tool outputs are deterministic and don't need interpretation
- Latency is critical (each agent adds round-trip time)

## Code Walkthrough

**Note**: Full implementation coming in subsequent tasks. Overview:

### Mock Tools (`mocks.go`)
Simulated diagnostic tools that return realistic data for demo purposes.

### Investigator Agent (`investigator.go`)
Service-scoped agent with access to mock tools. Creates targeted investigation prompts.

### Agent-as-Tool Wrapper (`investigate_service.go`)
Wraps investigator agent as a tool callable by coordinator. Handles parameter parsing and result conversion.

### Coordinator Agent (`main.go`)
Top-level agent that receives incidents, delegates to investigator tools, and synthesizes findings.

## Customization

### Add Real Tools

Replace mock tools with real implementations:

```go
// Instead of mock_logs
registry.MustRegister(k8s.NewTool(), k8s.NewHandler(&k8s.Config{...}))
```

### Adjust Agent Hierarchy

Add additional levels (coordinator → triage → investigator) or parallel agents (multiple services investigated concurrently).

### Change LLM Provider

Replace OpenAI with Vertex AI or Anthropic:

```go
provider := vertex.NewProvider(...)
```

## Related Examples

- [Incident Response](../incident-response/) - Single agent with multiple tools
- [K8s Monitoring](../k8s-monitoring/) - Proactive monitoring patterns
- [Tools Basic](../tools-basic/) - Tool registration fundamentals
