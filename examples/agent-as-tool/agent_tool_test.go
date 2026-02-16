package main

import (
	"context"
	"testing"

	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInvestigateServiceTool(t *testing.T) {
	tool := NewInvestigateServiceTool()

	assert.Equal(t, "investigate_service", tool.Function.Name)
	assert.Contains(t, tool.Function.Description, "Investigate a specific service")
	assert.NotNil(t, tool.Function.Parameters)

	// Verify schema has required parameters
	props, ok := tool.Function.Parameters["properties"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, props, "service_name")
	assert.Contains(t, props, "incident_type")

	required, ok := tool.Function.Parameters["required"].([]string)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"service_name", "incident_type"}, required)
}

func TestInvestigateServiceHandler_ParseError(t *testing.T) {
	// Create mock provider and registry
	mockProvider := &mockProvider{}
	mockRegistry := tools.NewRegistry()

	handler := NewInvestigateServiceHandler(mockProvider, mockRegistry)

	// Call with malformed JSON
	call := types.ToolCall{
		ID:   "test-call-1",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "investigate_service",
			Arguments: `{invalid json}`, // Malformed JSON
		},
	}

	result := handler(context.Background(), call)

	assert.Equal(t, "test-call-1", result.ToolCallID)
	assert.Equal(t, "investigate_service", result.ToolName)
	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "parse")
}

func TestInvestigateServiceHandler_Success(t *testing.T) {
	// Create mock provider that returns a simple response
	mockProvider := &mockProvider{
		response: &types.CompletionResponse{
			Message: types.Message{
				Role:    types.RoleAssistant,
				Content: "The database service is experiencing connection pool exhaustion. Evidence: 100/100 connections active, 45 queries queued.",
			},
			Usage: types.Usage{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
		},
	}

	// Create mock registry with diagnostic tools
	mockRegistry := tools.NewRegistry()
	mockRegistry.MustRegister(NewMockLogsTool(), NewMockLogsHandler())
	mockRegistry.MustRegister(NewMockMetricsTool(), NewMockMetricsHandler())
	mockRegistry.MustRegister(NewMockServiceStatusTool(), NewMockServiceStatusHandler())

	handler := NewInvestigateServiceHandler(mockProvider, mockRegistry)

	// Call with valid arguments
	call := types.ToolCall{
		ID:   "test-call-2",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "investigate_service",
			Arguments: `{"service_name": "database", "incident_type": "cascading_failure"}`,
		},
	}

	result := handler(context.Background(), call)

	assert.Equal(t, "test-call-2", result.ToolCallID)
	assert.Equal(t, "investigate_service", result.ToolName)
	assert.Empty(t, result.Error)
	assert.NotEmpty(t, result.Content)
	assert.Contains(t, result.Content, "Investigation of database")
	assert.Contains(t, result.Content, "Iterations:")
	assert.Contains(t, result.Content, "Tokens used:")
	assert.GreaterOrEqual(t, result.ExecutionTime.Nanoseconds(), int64(0))
}

func TestInvestigateServiceHandler_InvestigatorCreationError(t *testing.T) {
	// Create handler with valid provider but will fail on empty service name
	mockProvider := &mockProvider{}
	mockRegistry := tools.NewRegistry()

	handler := NewInvestigateServiceHandler(mockProvider, mockRegistry)

	// Call with empty service_name to trigger investigator creation error
	call := types.ToolCall{
		ID:   "test-call-3",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "investigate_service",
			Arguments: `{"service_name": "", "incident_type": "test"}`,
		},
	}

	result := handler(context.Background(), call)

	assert.Equal(t, "test-call-3", result.ToolCallID)
	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "failed to create investigator")
}

func TestInvestigateServiceHandler_ContextPropagation(t *testing.T) {
	// Create mock provider with context-aware response
	mockProvider := &mockProvider{
		response: &types.CompletionResponse{
			Message: types.Message{
				Role:    types.RoleAssistant,
				Content: "Investigation complete.",
			},
			Usage: types.Usage{TotalTokens: 100},
		},
	}

	mockRegistry := tools.NewRegistry()
	mockRegistry.MustRegister(NewMockLogsTool(), NewMockLogsHandler())
	mockRegistry.MustRegister(NewMockMetricsTool(), NewMockMetricsHandler())
	mockRegistry.MustRegister(NewMockServiceStatusTool(), NewMockServiceStatusHandler())

	handler := NewInvestigateServiceHandler(mockProvider, mockRegistry)

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	call := types.ToolCall{
		ID:   "test-call-4",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "investigate_service",
			Arguments: `{"service_name": "api-service", "incident_type": "test"}`,
		},
	}

	result := handler(ctx, call)

	// Should complete successfully with context not canceled
	assert.Empty(t, result.Error)
}

func TestAgentAsToolIntegration(t *testing.T) {
	// This test verifies the complete agent-as-tool pattern:
	// Tool definition + Handler + Inner agent creation

	// Setup
	mockProvider := &mockProvider{
		response: &types.CompletionResponse{
			Message: types.Message{
				Role:    types.RoleAssistant,
				Content: "Service investigation found database connection pool at capacity.",
			},
			Usage: types.Usage{TotalTokens: 200},
		},
	}

	mockRegistry := tools.NewRegistry()
	mockRegistry.MustRegister(NewMockLogsTool(), NewMockLogsHandler())
	mockRegistry.MustRegister(NewMockMetricsTool(), NewMockMetricsHandler())
	mockRegistry.MustRegister(NewMockServiceStatusTool(), NewMockServiceStatusHandler())

	// Create the tool and handler
	tool := NewInvestigateServiceTool()
	handler := NewInvestigateServiceHandler(mockProvider, mockRegistry)

	// Register in a coordinator registry
	coordinatorRegistry := tools.NewRegistry()
	coordinatorRegistry.MustRegister(tool, handler)

	// Verify tool is registered
	retrievedTool, exists := coordinatorRegistry.Get("investigate_service")
	require.True(t, exists)
	assert.Equal(t, "investigate_service", retrievedTool.Function.Name)

	// Execute the tool
	call := types.ToolCall{
		ID:   "integration-test",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "investigate_service",
			Arguments: `{"service_name": "database", "incident_type": "cascading_failure"}`,
		},
	}

	result := handler(context.Background(), call)

	// Verify execution
	assert.Empty(t, result.Error, "Handler should execute successfully")
	assert.Contains(t, result.Content, "Investigation of database")
	assert.Contains(t, result.Content, "Iterations: 1") // Mock provider returns immediately
	assert.Contains(t, result.Content, "Tokens used: 200")
}
