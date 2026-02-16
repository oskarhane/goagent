package main

import (
	"context"
	"strings"
	"testing"

	"github.com/oskarhane/goagent/pkg/agent"
	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// statefulMockProvider allows custom Complete behavior for multi-turn tests
type statefulMockProvider struct {
	completeFunc func(context.Context, *types.CompletionRequest) (*types.CompletionResponse, error)
}

func (m *statefulMockProvider) Complete(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, req)
	}
	return &types.CompletionResponse{
		Message: types.Message{
			Role:    types.RoleAssistant,
			Content: "Default mock response",
		},
		Usage: types.Usage{TotalTokens: 10},
	}, nil
}

func (m *statefulMockProvider) Name() string {
	return "stateful-mock"
}

func (m *statefulMockProvider) DefaultModel() string {
	return "stateful-mock-model"
}

// TestExtractInvestigatedServices verifies service extraction from messages
func TestExtractInvestigatedServices(t *testing.T) {
	tests := []struct {
		name     string
		messages []types.Message
		expected []string
	}{
		{
			name:     "no messages",
			messages: []types.Message{},
			expected: []string{},
		},
		{
			name: "no tool calls",
			messages: []types.Message{
				{Role: types.RoleUser, Content: "investigate this"},
				{Role: types.RoleAssistant, Content: "sure"},
			},
			expected: []string{},
		},
		{
			name: "single investigate_service call",
			messages: []types.Message{
				{
					Role: types.RoleAssistant,
					ToolCalls: []types.ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: types.FunctionCall{
								Name:      "investigate_service",
								Arguments: `{"service_name":"api-service","incident_type":"cascading_failure"}`,
							},
						},
					},
				},
			},
			expected: []string{"api-service"},
		},
		{
			name: "multiple investigate_service calls",
			messages: []types.Message{
				{
					Role: types.RoleAssistant,
					ToolCalls: []types.ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: types.FunctionCall{
								Name:      "investigate_service",
								Arguments: `{"service_name":"database","incident_type":"cascading_failure"}`,
							},
						},
					},
				},
				{
					Role: types.RoleAssistant,
					ToolCalls: []types.ToolCall{
						{
							ID:   "call_2",
							Type: "function",
							Function: types.FunctionCall{
								Name:      "investigate_service",
								Arguments: `{"service_name":"auth-service","incident_type":"cascading_failure"}`,
							},
						},
					},
				},
			},
			expected: []string{"database", "auth-service"},
		},
		{
			name: "mixed tool calls",
			messages: []types.Message{
				{
					Role: types.RoleAssistant,
					ToolCalls: []types.ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: types.FunctionCall{
								Name:      "investigate_service",
								Arguments: `{"service_name":"api-service"}`,
							},
						},
						{
							ID:   "call_2",
							Type: "function",
							Function: types.FunctionCall{
								Name:      "other_tool",
								Arguments: `{"param":"value"}`,
							},
						},
					},
				},
			},
			expected: []string{"api-service"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractInvestigatedServices(tt.messages)
			// Handle nil vs empty slice
			if len(tt.expected) == 0 {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestAgentAsTool is an integration test that verifies the coordinator delegates
// to investigator agents. Uses mock provider to avoid external API dependencies.
func TestAgentAsTool(t *testing.T) {
	// Create a stateful mock provider that tracks call count
	// to return different responses for multi-turn interaction
	callCount := 0

	// Create custom mock provider with stateful Complete function
	mockProvider := &statefulMockProvider{
		completeFunc: func(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error) {
			callCount++
			if callCount == 1 {
				// First call: coordinator decides to delegate
				return &types.CompletionResponse{
					Message: types.Message{
						Role: types.RoleAssistant,
						ToolCalls: []types.ToolCall{
							{
								ID:   "coord-call-1",
								Type: "function",
								Function: types.FunctionCall{
									Name:      "investigate_service",
									Arguments: `{"service_name":"database","incident_type":"cascading_failure"}`,
								},
							},
						},
					},
					Usage: types.Usage{TotalTokens: 50},
				}, nil
			} else if callCount == 2 {
				// Second call: investigator response (inside the tool handler)
				return &types.CompletionResponse{
					Message: types.Message{
						Role:    types.RoleAssistant,
						Content: "Database is experiencing connection pool exhaustion. Active connections: 100/100, 45 queries queued.",
					},
					Usage: types.Usage{TotalTokens: 75},
				}, nil
			}
			// Third+ calls: coordinator final synthesis
			return &types.CompletionResponse{
				Message: types.Message{
					Role:    types.RoleAssistant,
					Content: "Root cause identified: database connection pool saturation causing cascading failures.",
				},
				Usage: types.Usage{TotalTokens: 60},
			}, nil
		},
	}

	// Create mock tools registry for investigator agents
	mockRegistry := tools.NewRegistry()
	mockRegistry.MustRegister(NewMockLogsTool(), NewMockLogsHandler())
	mockRegistry.MustRegister(NewMockMetricsTool(), NewMockMetricsHandler())
	mockRegistry.MustRegister(NewMockServiceStatusTool(), NewMockServiceStatusHandler())

	// Create coordinator tools registry with investigate_service
	coordinatorRegistry := tools.NewRegistry()
	investigateTool := NewInvestigateServiceTool()
	investigateHandler := NewInvestigateServiceHandler(mockProvider, mockRegistry)
	coordinatorRegistry.MustRegister(investigateTool, investigateHandler)

	// Create coordinator agent
	coordinatorPrompt := `You are an SRE coordinator. Investigate incidents by delegating to service-specific agents using investigate_service tool.`
	coordinatorConfig := &agent.Config{
		Provider:      mockProvider,
		SystemPrompt:  coordinatorPrompt,
		Registry:      coordinatorRegistry,
		MaxIterations: 3, // Limit iterations for test
	}

	coordinator, err := agent.NewAgent(coordinatorConfig)
	require.NoError(t, err, "Coordinator agent creation should succeed")

	// Execute investigation
	ctx := context.Background()
	incident := "Database service is experiencing high error rates. Users reporting authentication failures."
	result := coordinator.Run(ctx, incident, nil)

	// Verify execution succeeded
	assert.NoError(t, result.Error, "Investigation should complete without errors")
	assert.NotNil(t, result.Response, "Result should contain response")

	// Verify delegation occurred - check that investigate_service was called
	investigatedServices := extractInvestigatedServices(result.Messages)
	assert.NotEmpty(t, investigatedServices, "Coordinator should have delegated to investigator")
	assert.Contains(t, investigatedServices, "database", "Database service should have been investigated")

	// Verify tool results propagated correctly
	findings := extractInvestigationFindings(result.Messages)
	assert.NotEmpty(t, findings, "Investigation should produce findings")

	// Verify at least one finding contains investigation details
	foundInvestigationResult := false
	for _, finding := range findings {
		if strings.Contains(finding, "Investigation of database") {
			foundInvestigationResult = true
			// Check that finding contains investigation details and stats
			assert.Contains(t, finding, "Investigation of database")
			assert.Contains(t, finding, "Iterations:")
			assert.Contains(t, finding, "Tokens used:")
			break
		}
	}
	assert.True(t, foundInvestigationResult, "Should find investigation result in findings")

	// Verify execution stats
	assert.Greater(t, result.Iterations, 0, "Should have at least one iteration")
	assert.Greater(t, result.TotalTokens, 0, "Should have used tokens")
	assert.Greater(t, result.ExecutionTime.Nanoseconds(), int64(0), "Should have execution time")
}

// TestExtractInvestigationFindings verifies findings extraction from tool results
func TestExtractInvestigationFindings(t *testing.T) {
	tests := []struct {
		name     string
		messages []types.Message
		expected []string
	}{
		{
			name:     "no messages",
			messages: []types.Message{},
			expected: []string{},
		},
		{
			name: "no tool results",
			messages: []types.Message{
				{Role: types.RoleUser, Content: "investigate this"},
				{Role: types.RoleAssistant, Content: "investigating..."},
			},
			expected: []string{},
		},
		{
			name: "single investigate_service result",
			messages: []types.Message{
				{
					Role:       types.RoleTool,
					Name:       "investigate_service",
					ToolCallID: "call_1",
					Content:    "Investigation complete: service is healthy",
				},
			},
			expected: []string{"Investigation complete: service is healthy"},
		},
		{
			name: "multiple investigate_service results",
			messages: []types.Message{
				{
					Role:       types.RoleTool,
					Name:       "investigate_service",
					ToolCallID: "call_1",
					Content:    "Database investigation: connection pool exhausted",
				},
				{
					Role:       types.RoleTool,
					Name:       "investigate_service",
					ToolCallID: "call_2",
					Content:    "Auth service investigation: timeouts detected",
				},
			},
			expected: []string{
				"Database investigation: connection pool exhausted",
				"Auth service investigation: timeouts detected",
			},
		},
		{
			name: "mixed tool results",
			messages: []types.Message{
				{
					Role:       types.RoleTool,
					Name:       "investigate_service",
					ToolCallID: "call_1",
					Content:    "Service investigation result",
				},
				{
					Role:       types.RoleTool,
					Name:       "other_tool",
					ToolCallID: "call_2",
					Content:    "Other tool result",
				},
			},
			expected: []string{"Service investigation result"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractInvestigationFindings(tt.messages)
			// Handle nil vs empty slice
			if len(tt.expected) == 0 {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
