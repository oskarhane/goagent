package tools

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oskarhane/goagent/pkg/types"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.NotNil(t, r.tools)
	assert.NotNil(t, r.handlers)
}

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		tool    types.Tool
		handler Handler
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid tool",
			tool: types.Tool{
				Type: "function",
				Function: types.FunctionDefinition{
					Name:        "test_tool",
					Description: "A test tool",
					Parameters:  map[string]any{"type": "object"},
				},
			},
			handler: func(_ context.Context, _ types.ToolCall) types.ToolResult {
				return types.ToolResult{Content: "ok"}
			},
			wantErr: false,
		},
		{
			name: "missing name",
			tool: types.Tool{
				Type: "function",
				Function: types.FunctionDefinition{
					Description: "A test tool",
					Parameters:  map[string]any{"type": "object"},
				},
			},
			handler: func(_ context.Context, _ types.ToolCall) types.ToolResult {
				return types.ToolResult{Content: "ok"}
			},
			wantErr: true,
			errMsg:  "tool name cannot be empty",
		},
		{
			name: "missing description",
			tool: types.Tool{
				Type: "function",
				Function: types.FunctionDefinition{
					Name:       "test_tool",
					Parameters: map[string]any{"type": "object"},
				},
			},
			handler: func(_ context.Context, _ types.ToolCall) types.ToolResult {
				return types.ToolResult{Content: "ok"}
			},
			wantErr: true,
			errMsg:  "tool description cannot be empty",
		},
		{
			name: "nil parameters",
			tool: types.Tool{
				Type: "function",
				Function: types.FunctionDefinition{
					Name:        "test_tool",
					Description: "A test tool",
				},
			},
			handler: func(_ context.Context, _ types.ToolCall) types.ToolResult {
				return types.ToolResult{Content: "ok"}
			},
			wantErr: true,
			errMsg:  "tool parameters cannot be nil",
		},
		{
			name: "nil handler",
			tool: types.Tool{
				Type: "function",
				Function: types.FunctionDefinition{
					Name:        "test_tool",
					Description: "A test tool",
					Parameters:  map[string]any{"type": "object"},
				},
			},
			handler: nil,
			wantErr: true,
			errMsg:  "tool handler cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			err := r.Register(tt.tool, tt.handler)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)

				// Verify tool was registered
				tool, exists := r.Get(tt.tool.Function.Name)
				assert.True(t, exists)
				assert.Equal(t, tt.tool.Function.Name, tool.Function.Name)

				// Verify handler was registered
				handler := r.GetHandler(tt.tool.Function.Name)
				assert.NotNil(t, handler)
			}
		})
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()

	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "duplicate",
			Description: "A test tool",
			Parameters:  map[string]any{"type": "object"},
		},
	}
	handler := func(_ context.Context, _ types.ToolCall) types.ToolResult {
		return types.ToolResult{Content: "ok"}
	}

	// First registration should succeed
	err := r.Register(tool, handler)
	require.NoError(t, err)

	// Second registration should fail
	err = r.Register(tool, handler)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_MustRegister(t *testing.T) {
	r := NewRegistry()

	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "must_test",
			Description: "A test tool",
			Parameters:  map[string]any{"type": "object"},
		},
	}
	handler := func(_ context.Context, _ types.ToolCall) types.ToolResult {
		return types.ToolResult{Content: "ok"}
	}

	// Should not panic
	assert.NotPanics(t, func() {
		r.MustRegister(tool, handler)
	})

	// Should panic on invalid tool
	assert.Panics(t, func() {
		r.MustRegister(types.Tool{}, handler)
	})
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()

	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "get_test",
			Description: "A test tool",
			Parameters:  map[string]any{"type": "object"},
		},
	}
	handler := func(_ context.Context, _ types.ToolCall) types.ToolResult {
		return types.ToolResult{Content: "ok"}
	}

	r.MustRegister(tool, handler)

	// Get existing tool
	retrieved, exists := r.Get("get_test")
	assert.True(t, exists)
	assert.Equal(t, tool.Function.Name, retrieved.Function.Name)

	// Get non-existent tool
	_, exists = r.Get("nonexistent")
	assert.False(t, exists)
}

func TestRegistry_Execute(t *testing.T) {
	r := NewRegistry()

	callReceived := false
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "exec_test",
			Description: "A test tool",
			Parameters:  map[string]any{"type": "object"},
		},
	}
	handler := func(_ context.Context, call types.ToolCall) types.ToolResult {
		callReceived = true
		return types.ToolResult{
			ToolCallID: call.ID,
			ToolName:   call.Function.Name,
			Content:    "success",
		}
	}

	r.MustRegister(tool, handler)

	// Execute registered tool
	ctx := context.Background()
	call := types.ToolCall{
		ID:   "call_123",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "exec_test",
			Arguments: "{}",
		},
	}

	result := r.Execute(ctx, call)
	assert.True(t, callReceived)
	assert.Equal(t, "call_123", result.ToolCallID)
	assert.Equal(t, "exec_test", result.ToolName)
	assert.Equal(t, "success", result.Content)
	assert.Empty(t, result.Error)

	// Execute non-existent tool
	call2 := types.ToolCall{
		ID:   "call_456",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "nonexistent",
			Arguments: "{}",
		},
	}

	result2 := r.Execute(ctx, call2)
	assert.Equal(t, "call_456", result2.ToolCallID)
	assert.Equal(t, "nonexistent", result2.ToolName)
	assert.Contains(t, result2.Error, "not registered")
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()

	// Empty registry
	tools := r.List()
	assert.Empty(t, tools)

	// Add tools
	handler := func(_ context.Context, _ types.ToolCall) types.ToolResult {
		return types.ToolResult{Content: "ok"}
	}

	tool1 := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "tool1",
			Description: "First tool",
			Parameters:  map[string]any{"type": "object"},
		},
	}
	tool2 := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "tool2",
			Description: "Second tool",
			Parameters:  map[string]any{"type": "object"},
		},
	}

	r.MustRegister(tool1, handler)
	r.MustRegister(tool2, handler)

	tools = r.List()
	assert.Len(t, tools, 2)

	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Function.Name] = true
	}
	assert.True(t, names["tool1"])
	assert.True(t, names["tool2"])
}

func TestRegistry_ContextCancellation(t *testing.T) {
	r := NewRegistry()

	handlerCalled := false
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "slow_tool",
			Description: "A slow tool",
			Parameters:  map[string]any{"type": "object"},
		},
	}
	handler := func(ctx context.Context, call types.ToolCall) types.ToolResult {
		handlerCalled = true
		// Check context is passed correctly
		select {
		case <-ctx.Done():
			return types.ToolResult{
				ToolCallID: call.ID,
				ToolName:   call.Function.Name,
				Error:      "context canceled",
			}
		case <-time.After(100 * time.Millisecond):
			return types.ToolResult{
				ToolCallID: call.ID,
				ToolName:   call.Function.Name,
				Content:    "completed",
			}
		}
	}

	r.MustRegister(tool, handler)

	// Test with canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	call := types.ToolCall{
		ID:   "call_123",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "slow_tool",
			Arguments: "{}",
		},
	}

	result := r.Execute(ctx, call)
	assert.True(t, handlerCalled)
	assert.Contains(t, result.Error, "context canceled")
}
func TestValidateParameters_ValidObject(t *testing.T) {
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "test_tool",
			Description: "Test",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
					"age": map[string]any{
						"type": "integer",
					},
				},
				"required": []any{"name"},
			},
		},
	}

	call := types.ToolCall{
		ID:   "call_1",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "test_tool",
			Arguments: `{"name": "John", "age": 30}`,
		},
	}

	err := ValidateParameters(tool, call)
	assert.NoError(t, err)
}

func TestValidateParameters_MissingRequired(t *testing.T) {
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "test_tool",
			Description: "Test",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
				},
				"required": []any{"name"},
			},
		},
	}

	call := types.ToolCall{
		ID:   "call_1",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "test_tool",
			Arguments: `{}`,
		},
	}

	err := ValidateParameters(tool, call)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required field")
}

func TestValidateParameters_WrongType(t *testing.T) {
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "test_tool",
			Description: "Test",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"age": map[string]any{
						"type": "integer",
					},
				},
			},
		},
	}

	call := types.ToolCall{
		ID:   "call_1",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "test_tool",
			Arguments: `{"age": "not a number"}`,
		},
	}

	err := ValidateParameters(tool, call)
	require.Error(t, err)
	// Error message should mention integer/string mismatch
	assert.True(t, strings.Contains(err.Error(), "integer") || strings.Contains(err.Error(), "type"))
}

func TestValidateParameters_NestedObject(t *testing.T) {
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "test_tool",
			Description: "Test",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"config": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"debug": map[string]any{
								"type": "boolean",
							},
						},
						"required": []any{"debug"},
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		args      string
		wantError bool
	}{
		{
			name:      "valid nested object",
			args:      `{"config": {"debug": true}}`,
			wantError: false,
		},
		{
			name:      "missing nested required field",
			args:      `{"config": {}}`,
			wantError: true,
		},
		{
			name:      "wrong nested type",
			args:      `{"config": {"debug": "yes"}}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := types.ToolCall{
				ID:   "call_1",
				Type: "function",
				Function: types.FunctionCall{
					Name:      "test_tool",
					Arguments: tt.args,
				},
			}

			err := ValidateParameters(tool, call)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateParameters_Array(t *testing.T) {
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "test_tool",
			Description: "Test",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		args      string
		wantError bool
	}{
		{
			name:      "valid array",
			args:      `{"tags": ["tag1", "tag2"]}`,
			wantError: false,
		},
		{
			name:      "empty array",
			args:      `{"tags": []}`,
			wantError: false,
		},
		{
			name:      "wrong item type",
			args:      `{"tags": [1, 2, 3]}`,
			wantError: true,
		},
		{
			name:      "not an array",
			args:      `{"tags": "not-array"}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := types.ToolCall{
				ID:   "call_1",
				Type: "function",
				Function: types.FunctionCall{
					Name:      "test_tool",
					Arguments: tt.args,
				},
			}

			err := ValidateParameters(tool, call)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateParameters_InvalidJSON(t *testing.T) {
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "test_tool",
			Description: "Test",
			Parameters: map[string]any{
				"type": "object",
			},
		},
	}

	call := types.ToolCall{
		ID:   "call_1",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "test_tool",
			Arguments: `{invalid json}`,
		},
	}

	err := ValidateParameters(tool, call)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestValidateParameters_NoSchema(t *testing.T) {
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "test_tool",
			Description: "Test",
			Parameters:  nil,
		},
	}

	call := types.ToolCall{
		ID:   "call_1",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "test_tool",
			Arguments: `{}`,
		},
	}

	err := ValidateParameters(tool, call)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no parameter schema")
}

func TestValidateParameters_NumericTypes(t *testing.T) {
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "test_tool",
			Description: "Test",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"count": map[string]any{
						"type": "integer",
					},
					"price": map[string]any{
						"type": "number",
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		args      string
		wantError bool
	}{
		{
			name:      "valid integer",
			args:      `{"count": 42}`,
			wantError: false,
		},
		{
			name:      "valid number",
			args:      `{"price": 19.99}`,
			wantError: false,
		},
		{
			name:      "float as integer",
			args:      `{"count": 42.5}`,
			wantError: true,
		},
		{
			name:      "string as number",
			args:      `{"price": "19.99"}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := types.ToolCall{
				ID:   "call_1",
				Type: "function",
				Function: types.FunctionCall{
					Name:      "test_tool",
					Arguments: tt.args,
				},
			}

			err := ValidateParameters(tool, call)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
