package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
)

// NewInvestigateServiceTool creates the investigate_service tool definition.
// This tool wraps an investigator agent, demonstrating the agent-as-tool pattern
// where one agent (coordinator) can delegate work to another agent (investigator).
func NewInvestigateServiceTool() types.Tool {
	return tools.NewBuilder(
		"investigate_service",
		"Investigate a specific service for the reported incident. Delegates to a specialized investigator agent that will examine logs, metrics, and health status for the service. Returns a detailed analysis including service health, root cause if found, and evidence from diagnostics.",
	).
		StringParam("service_name", "Name of the service to investigate (e.g., api-service, auth-service, database, cache)", true).
		StringParam("incident_type", "Type of incident being investigated (e.g., cascading_failure, memory_leak, dependency_failure)", true).
		Build()
}

// NewInvestigateServiceHandler creates a handler that wraps the investigator agent as a tool.
// This demonstrates the agent-as-tool pattern: the handler creates and runs an inner agent,
// then returns its response as a tool result to the outer (coordinator) agent.
//
// Parameters:
//   - provider: LLM provider to use for the investigator agent
//   - mockRegistry: Tool registry containing mock diagnostic tools (mock_logs, mock_metrics, mock_service_status)
func NewInvestigateServiceHandler(provider types.Provider, mockRegistry *tools.Registry) func(context.Context, types.ToolCall) types.ToolResult {
	return func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		// Parse parameters
		var params struct {
			ServiceName  string `json:"service_name"`
			IncidentType string `json:"incident_type"`
		}
		if err := types.ParseToolArguments(call, &params); err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("failed to parse parameters: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		// Create investigator agent scoped to this service
		investigator, err := NewInvestigator(params.ServiceName, provider, mockRegistry)
		if err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("failed to create investigator agent: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		// Run the investigator agent with scoped context
		prompt := fmt.Sprintf(
			"Investigate %s for a %s incident. Analyze logs, metrics, and service status to determine if this service is healthy, degraded, or the root cause of the incident. Provide specific evidence from your investigation.",
			params.ServiceName,
			params.IncidentType,
		)

		// Run the inner agent
		result := investigator.Run(ctx, prompt, nil)

		// Check for errors from the inner agent
		if result.Error != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("investigator agent failed: %v", result.Error),
				ExecutionTime: time.Since(start),
			}
		}

		// Format the investigation result
		// Include the agent's response content and some metadata about the investigation
		content := fmt.Sprintf(`Investigation of %s:

%s

---
Investigation Stats:
- Iterations: %d
- Tokens used: %d
- Execution time: %.2fs`,
			params.ServiceName,
			result.Response.Content,
			result.Iterations,
			result.TotalTokens,
			result.ExecutionTime.Seconds(),
		)

		return types.ToolResult{
			ToolCallID:    call.ID,
			ToolName:      call.Function.Name,
			Content:       content,
			ExecutionTime: time.Since(start),
		}
	}
}
