package main

import (
	"testing"

	"github.com/oskarhane/goagent/pkg/types"
	"github.com/stretchr/testify/assert"
)

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
