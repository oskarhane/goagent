package main

import (
	"fmt"

	"github.com/oskarhane/goagent/pkg/agent"
	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
)

// NewInvestigator creates an agent configured for service-level investigation.
// The investigator agent is designed to be called by a coordinator agent as a tool,
// demonstrating the agent-as-tool pattern for hierarchical delegation.
//
// Parameters:
//   - serviceName: The name of the service to investigate (e.g., "api-service", "database")
//   - provider: The LLM provider for reasoning
//   - registry: Tool registry containing mock diagnostic tools (mock_logs, mock_metrics, mock_service_status)
//
// Returns an agent configured with:
//   - Dynamic system prompt scoped to the service
//   - Max 5 iterations for focused investigation
//   - Access to mock diagnostic tools
func NewInvestigator(serviceName string, provider types.Provider, registry *tools.Registry) (*agent.Agent, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("investigator: service name is required")
	}
	if provider == nil {
		return nil, fmt.Errorf("investigator: provider is required")
	}
	if registry == nil {
		return nil, fmt.Errorf("investigator: tool registry is required")
	}

	// Create dynamic system prompt scoped to the service
	systemPrompt := fmt.Sprintf(`You are a service-level investigator for %s.

Your role is to perform a focused investigation of this specific service. You have access to:
- mock_logs: Retrieve recent log entries from %s
- mock_metrics: Retrieve time-series metrics (CPU, memory, request rate, error rate) for %s
- mock_service_status: Check health status and dependency information for %s

Investigation methodology:
1. Start by checking the service status to understand its overall health
2. Review logs to identify error patterns and anomalies
3. Analyze metrics to spot performance degradation or resource issues
4. Correlate findings to determine if this service is the root cause or affected by dependencies
5. Provide a clear conclusion: healthy, degraded, or unhealthy with specific evidence

IMPORTANT: Your investigation is scoped to %s only. Focus on this service's behavior and its immediate dependencies.
Provide concrete evidence from logs, metrics, and status checks. Be concise and actionable.`,
		serviceName, serviceName, serviceName, serviceName, serviceName)

	// Configure agent with 5 max iterations for focused investigation
	cfg := &agent.Config{
		Provider:      provider,
		SystemPrompt:  systemPrompt,
		Registry:      registry,
		MaxIterations: 5,
	}

	return agent.NewAgent(cfg)
}
