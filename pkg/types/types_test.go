package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name    string
		message Message
	}{
		{
			name: "user message",
			message: Message{
				Role:    RoleUser,
				Content: "Hello, world!",
			},
		},
		{
			name: "assistant message with tool calls",
			message: Message{
				Role: RoleAssistant,
				ToolCalls: []ToolCall{
					{
						ID:   "call_123",
						Type: "function",
						Function: FunctionCall{
							Name:      "get_weather",
							Arguments: `{"location": "San Francisco"}`,
						},
					},
				},
			},
		},
		{
			name: "tool result message",
			message: Message{
				Role:       RoleTool,
				Content:    `{"temperature": 72}`,
				ToolCallID: "call_123",
				Name:       "get_weather",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.message)
			require.NoError(t, err)

			// Unmarshal back
			var decoded Message
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			// Compare
			assert.Equal(t, tt.message, decoded)
		})
	}
}

func TestNewUserMessage(t *testing.T) {
	content := "Test message"
	msg := NewUserMessage(content)

	assert.Equal(t, RoleUser, msg.Role)
	assert.Equal(t, content, msg.Content)
	assert.Empty(t, msg.ToolCalls)
	assert.Empty(t, msg.ToolCallID)
	assert.Empty(t, msg.Name)
}

func TestNewSystemMessage(t *testing.T) {
	content := "System instructions"
	msg := NewSystemMessage(content)

	assert.Equal(t, RoleSystem, msg.Role)
	assert.Equal(t, content, msg.Content)
	assert.Empty(t, msg.ToolCalls)
	assert.Empty(t, msg.ToolCallID)
	assert.Empty(t, msg.Name)
}

func TestNewAssistantMessage(t *testing.T) {
	content := "Assistant response"
	msg := NewAssistantMessage(content)

	assert.Equal(t, RoleAssistant, msg.Role)
	assert.Equal(t, content, msg.Content)
	assert.Empty(t, msg.ToolCalls)
	assert.Empty(t, msg.ToolCallID)
	assert.Empty(t, msg.Name)
}

func TestNewToolMessage(t *testing.T) {
	toolCallID := "call_123"
	toolName := "test_tool"
	content := "result content"

	msg := NewToolMessage(toolCallID, toolName, content)

	assert.Equal(t, RoleTool, msg.Role)
	assert.Equal(t, content, msg.Content)
	assert.Equal(t, toolCallID, msg.ToolCallID)
	assert.Equal(t, toolName, msg.Name)
	assert.Empty(t, msg.ToolCalls)
}

func TestMessage_HasToolCalls(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		expected bool
	}{
		{
			name:     "no tool calls",
			message:  Message{Role: RoleUser, Content: "hello"},
			expected: false,
		},
		{
			name: "has tool calls",
			message: Message{
				Role: RoleAssistant,
				ToolCalls: []ToolCall{
					{ID: "call_1", Type: "function", Function: FunctionCall{Name: "test"}},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.message.HasToolCalls())
		})
	}
}

func TestMessage_IsToolResult(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		expected bool
	}{
		{
			name:     "user message",
			message:  Message{Role: RoleUser, Content: "hello"},
			expected: false,
		},
		{
			name:     "tool result",
			message:  Message{Role: RoleTool, ToolCallID: "call_1", Content: "result"},
			expected: true,
		},
		{
			name:     "tool message without call ID",
			message:  Message{Role: RoleTool, Content: "result"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.message.IsToolResult())
		})
	}
}

func TestParseToolArguments(t *testing.T) {
	tests := []struct {
		name     string
		call     ToolCall
		wantErr  bool
		validate func(t *testing.T, result map[string]any)
	}{
		{
			name: "valid JSON object",
			call: ToolCall{
				ID:   "call_1",
				Type: "function",
				Function: FunctionCall{
					Name:      "test",
					Arguments: `{"location": "NYC", "units": "celsius"}`,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result map[string]any) {
				assert.Equal(t, "NYC", result["location"])
				assert.Equal(t, "celsius", result["units"])
			},
		},
		{
			name: "empty object",
			call: ToolCall{
				ID:   "call_2",
				Type: "function",
				Function: FunctionCall{
					Name:      "test",
					Arguments: `{}`,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result map[string]any) {
				assert.Empty(t, result)
			},
		},
		{
			name: "invalid JSON",
			call: ToolCall{
				ID:   "call_3",
				Type: "function",
				Function: FunctionCall{
					Name:      "test",
					Arguments: `{invalid json}`,
				},
			},
			wantErr: true,
		},
		{
			name: "nested object",
			call: ToolCall{
				ID:   "call_4",
				Type: "function",
				Function: FunctionCall{
					Name:      "test",
					Arguments: `{"config": {"debug": true, "level": 2}}`,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result map[string]any) {
				config, ok := result["config"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, true, config["debug"])
				assert.Equal(t, float64(2), config["level"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]any
			err := ParseToolArguments(tt.call, &result)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestToolResult_JSONSerialization(t *testing.T) {
	result := &ToolResult{
		ToolCallID:    "call_123",
		ToolName:      "test_tool",
		Content:       "test output",
		Error:         "test error",
		ExecutionTime: 100,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded ToolResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.ToolCallID, decoded.ToolCallID)
	assert.Equal(t, result.ToolName, decoded.ToolName)
	assert.Equal(t, result.Content, decoded.Content)
	assert.Equal(t, result.Error, decoded.Error)
	assert.Equal(t, result.ExecutionTime, decoded.ExecutionTime)
}

func TestCompletionResponse_JSONSerialization(t *testing.T) {
	response := &CompletionResponse{
		Message: Message{
			Role:    RoleAssistant,
			Content: "test content",
		},
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		FinishReason: "stop",
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var decoded CompletionResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.Message, decoded.Message)
	assert.Equal(t, response.Usage, decoded.Usage)
	assert.Equal(t, response.FinishReason, decoded.FinishReason)
}
