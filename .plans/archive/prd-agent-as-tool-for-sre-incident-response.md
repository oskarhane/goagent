Let me first explore the existing examples and tools to understand the patterns in this codebase.
Now let me check the existing incident-response example to understand what's already there:
Now I have a complete understanding of the codebase. Let me generate the PRD for the agent-as-tool example.

---

# PRD: Agent-as-Tool for SRE Incident Response

## Overview

Create an example demonstrating hierarchical agent composition where a specialist "investigator" agent is wrapped as a tool for an outer "coordinator" agent. The coordinator triages and routes incidents while delegating deep investigation to the specialist agent. This pattern enables scoped autonomy—the inner agent has focused tools and context, while the outer agent maintains workflow control.

**Use Case**: SRE incident response where a coordinator agent receives alerts, prioritizes them, and dispatches specialist agents to investigate specific services or subsystems.

## Goals

1. Demonstrate agent-as-tool pattern for hierarchical delegation
2. Provide practical SRE incident response example usable without K8s
3. Show benefits of scoped agent autonomy (inner agent gets constrained toolset)
4. Enable hackathon participants to build multi-agent workflows
5. Work entirely with mock/simulated services for easy demos

## Non-Goals

- Production-ready incident management system
- Real K8s cluster integration (optional for advanced users)
- Persistent conversation history or incident tracking
- Integration with real monitoring systems (Prometheus, Datadog)
- Multi-tenant or multi-team coordination

## Requirements

### Functional Requirements

- **REQ-F-001**: Outer "coordinator" agent receives incident reports and determines investigation strategy
- **REQ-F-002**: Inner "investigator" agent wrapped as tool with scoped access to diagnostic tools
- **REQ-F-003**: Coordinator can invoke investigator tool with specific context (service name, incident type)
- **REQ-F-004**: Investigator agent has access to: mock_logs, mock_metrics, mock_service_status tools
- **REQ-F-005**: Mock tools simulate realistic SRE data without external dependencies
- **REQ-F-006**: Investigator returns structured findings to coordinator
- **REQ-F-007**: Coordinator synthesizes findings and provides recommendations
- **REQ-F-008**: Support multiple investigation scopes: "api-service", "database", "cache", "auth-service"
- **REQ-F-009**: Demo scenario with cascading failure (auth → api → user-facing errors)

### Non-Functional Requirements

- **REQ-NF-001**: Example runs without external dependencies (no K8s, no real APIs)
- **REQ-NF-002**: Execution completes in <60s for demo purposes
- **REQ-NF-003**: Clear logging showing agent delegation and tool execution
- **REQ-NF-004**: Memory usage <100MB during execution
- **REQ-NF-005**: Example code follows existing patterns in `examples/`

## Technical Considerations

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Coordinator Agent                      │
│  System: "You are an SRE coordinator..."                │
│  Tools:                                                  │
│    - investigate_service (wraps Investigator Agent)      │
│    - get_alert_context                                   │
│    - recommend_action                                    │
└─────────────────────────────────────────────────────────┘
                           │
                           │ invoke
                           ▼
┌─────────────────────────────────────────────────────────┐
│                   Investigator Agent                     │
│  System: "You investigate {service_name}..."            │
│  Tools:                                                  │
│    - mock_logs (scoped to service)                      │
│    - mock_metrics (scoped to service)                   │
│    - mock_service_status                                │
└─────────────────────────────────────────────────────────┘
```

### Key Implementation Details

1. **Agent wrapping pattern**:
```go
// Create investigator agent
investigator, _ := agent.NewAgent(&agent.Config{...})

// Wrap as tool for coordinator
investigatorTool := tools.NewBuilder("investigate_service", "...").
    StringParam("service_name", "Service to investigate", true).
    StringParam("incident_type", "Type of issue", true).
    Build()

investigatorHandler := func(ctx context.Context, call types.ToolCall) types.ToolResult {
    var params struct {
        ServiceName  string `json:"service_name"`
        IncidentType string `json:"incident_type"`
    }
    types.ParseToolArguments(call, &params)
    
    // Run inner agent with scoped context
    prompt := fmt.Sprintf("Investigate %s for %s issues", 
        params.ServiceName, params.IncidentType)
    result := investigator.Run(ctx, prompt, nil)
    
    return types.ToolResult{
        ToolCallID: call.ID,
        ToolName:   call.Function.Name,
        Content:    result.Response.Content,
    }
}
```

2. **Mock tools simulate real data**:
```go
// mock_logs returns realistic log entries
// mock_metrics returns time-series-like data
// mock_service_status returns health check results
```

3. **Scoped context**: Inner agent's system prompt dynamically includes service name, limiting investigation scope

### File Structure

```
examples/agent-as-tool/
├── main.go           # Entry point, coordinator setup
├── investigator.go   # Inner agent factory
├── mocks.go          # Mock tools (logs, metrics, status)
├── scenarios.go      # Pre-built incident scenarios
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

### Integration Points

- Uses existing `pkg/agent` for both agents
- Uses existing `pkg/tools` for tool registration
- Uses existing `pkg/providers/openai` for LLM
- Follows `examples/incident-response` patterns

### Potential Challenges

1. **Context limits**: Inner agent's full response becomes tool result—may be large
2. **Token costs**: Two agents = 2x token usage minimum
3. **Error propagation**: Inner agent errors need clean propagation to outer
4. **Demo reliability**: Mock data must be deterministic for reproducible demos

## Acceptance Criteria

- [ ] Example compiles with `go build ./...`
- [ ] Runs successfully with only `OPENAI_API_KEY` env var
- [ ] Coordinator correctly delegates to investigator agent
- [ ] Investigator uses mock tools without external dependencies
- [ ] Log output clearly shows agent hierarchy and tool calls
- [ ] README documents the pattern with architecture diagram
- [ ] Example demonstrates cascading failure investigation
- [ ] Execution completes within 60 seconds
- [ ] Code follows patterns from AGENTS.md (builder, registry, error handling)
- [ ] Tests pass with `go test ./...`

## Out of Scope

- Real Kubernetes integration (covered by existing `incident-response` example)
- WebSocket/streaming responses
- Persistent state between runs
- Multiple concurrent investigators
- Custom LLM providers (OpenAI only for simplicity)
- Authentication/authorization for agents
- Rate limiting or cost tracking

## Open Questions

1. Should mock data be randomized or deterministic per run?
2. Max iterations for inner agent—5 or 10?
3. Include simple unit tests or just integration test?
4. ASCII art for architecture in README or skip?