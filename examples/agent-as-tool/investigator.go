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
	systemPrompt := fmt.Sprintf(`Investigate %s service diagnostics.

Tools: mock_logs, mock_metrics, mock_service_status for %s

Steps:
1. Check service status
2. Review logs for errors
3. Check 1 key metric

Report: service health (healthy/degraded/unhealthy) with specific evidence. Be concise.`,
		serviceName, serviceName)

	// Configure agent with 3 max iterations for focused investigation
	cfg := &agent.Config{
		Provider:      provider,
		SystemPrompt:  systemPrompt,
		Registry:      registry,
		MaxIterations: 3,
	}

	return agent.NewAgent(cfg)
}
