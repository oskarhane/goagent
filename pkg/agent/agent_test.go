package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
)

// MockProvider implements types.Provider for testing
type MockProvider struct {
	responses []*types.CompletionResponse
	callCount int
	err       error
}

func (m *MockProvider) Complete(_ context.Context, _ *types.CompletionRequest) (*types.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.callCount >= len(m.responses) {
		return nil, errors.New("mock provider: no more responses")
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}

func (m *MockProvider) Name() string {
	return "mock"
}

func (m *MockProvider) DefaultModel() string {
	return "mock-model"
}

func TestNewAgent(t *testing.T) {
	provider := &MockProvider{}
	registry := tools.NewRegistry()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Provider: provider,
				Registry: registry,
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			config: &Config{
				Registry: registry,
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "missing registry",
			config: &Config{
				Provider: provider,
			},
			wantErr: true,
			errMsg:  "registry is required",
		},
		{
			name: "custom max iterations",
			config: &Config{
				Provider:      provider,
				Registry:      registry,
				MaxIterations: 5,
			},
			wantErr: false,
		},
		{
			name: "custom temperature",
			config: &Config{
				Provider:    provider,
				Registry:    registry,
				Temperature: floatPtr(0.5),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewAgent(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, agent)
			}
		})
	}
}

func TestAgent_RunSimple(t *testing.T) {
	provider := &MockProvider{
		responses: []*types.CompletionResponse{
			{
				Message: types.Message{
					Role:    types.RoleAssistant,
					Content: "Hello! The answer is 42.",
				},
				Usage: types.Usage{
					PromptTokens:     10,
					CompletionTokens: 15,
					TotalTokens:      25,
				},
				FinishReason: "stop",
			},
		},
	}

	registry := tools.NewRegistry()
	agent, err := NewAgent(&Config{
		Provider: provider,
		Registry: registry,
	})
	require.NoError(t, err)

	ctx := context.Background()
	result := agent.Run(ctx, "What is the answer?", nil)

	assert.Nil(t, result.Error)
	assert.Equal(t, 1, result.Iterations)
	assert.Equal(t, 25, result.TotalTokens)
	assert.Equal(t, 1, provider.callCount)
	assert.Len(t, result.Messages, 3) // system + user + assistant
}

func TestAgent_RunWithToolCall(t *testing.T) {
	// Setup provider to first call a tool, then provide final answer
	provider := &MockProvider{
		responses: []*types.CompletionResponse{
			{
				Message: types.Message{
					Role: types.RoleAssistant,
					ToolCalls: []types.ToolCall{
						{
							ID:   "call_123",
							Type: "function",
							Function: types.FunctionCall{
								Name:      "get_time",
								Arguments: `{}`,
							},
						},
					},
				},
				Usage: types.Usage{
					PromptTokens:     10,
					CompletionTokens: 5,
					TotalTokens:      15,
				},
				FinishReason: "tool_calls",
			},
			{
				Message: types.Message{
					Role:    types.RoleAssistant,
					Content: "The current time is 12:00 PM.",
				},
				Usage: types.Usage{
					PromptTokens:     20,
					CompletionTokens: 10,
					TotalTokens:      30,
				},
				FinishReason: "stop",
			},
		},
	}

	// Register tool
	registry := tools.NewRegistry()
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "get_time",
			Description: "Get current time",
			Parameters:  map[string]any{"type": "object"},
		},
	}
	handler := func(_ context.Context, call types.ToolCall) types.ToolResult {
		return types.ToolResult{
			ToolCallID: call.ID,
			ToolName:   call.Function.Name,
			Content:    "12:00 PM",
		}
	}
	registry.MustRegister(tool, handler)

	agent, err := NewAgent(&Config{
		Provider: provider,
		Registry: registry,
	})
	require.NoError(t, err)

	ctx := context.Background()
	result := agent.Run(ctx, "What time is it?", nil)

	assert.Nil(t, result.Error)
	assert.Equal(t, 2, result.Iterations)
	assert.Equal(t, 45, result.TotalTokens) // 15 + 30
	assert.Equal(t, 2, provider.callCount)
}

func TestAgent_RunMaxIterations(t *testing.T) {
	// Provider always returns tool calls
	provider := &MockProvider{
		responses: []*types.CompletionResponse{
			{
				Message: types.Message{
					Role: types.RoleAssistant,
					ToolCalls: []types.ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: types.FunctionCall{
								Name:      "test_tool",
								Arguments: `{}`,
							},
						},
					},
				},
				FinishReason: "tool_calls",
			},
		},
	}

	// Repeat the response many times
	for i := 0; i < 20; i++ {
		provider.responses = append(provider.responses, provider.responses[0])
	}

	registry := tools.NewRegistry()
	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "test_tool",
			Description: "Test tool",
			Parameters:  map[string]any{"type": "object"},
		},
	}
	handler := func(_ context.Context, call types.ToolCall) types.ToolResult {
		return types.ToolResult{
			ToolCallID: call.ID,
			ToolName:   call.Function.Name,
			Content:    "ok",
		}
	}
	registry.MustRegister(tool, handler)

	agent, err := NewAgent(&Config{
		Provider:      provider,
		Registry:      registry,
		MaxIterations: 3,
	})
	require.NoError(t, err)

	ctx := context.Background()
	result := agent.Run(ctx, "Test", nil)

	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "maximum iterations")
	assert.Equal(t, 3, result.Iterations)
	assert.Equal(t, 3, provider.callCount)
}

func TestAgent_RunContextCancelled(t *testing.T) {
	provider := &MockProvider{
		responses: []*types.CompletionResponse{
			{
				Message: types.Message{
					Role:    types.RoleAssistant,
					Content: "response",
				},
				FinishReason: "stop",
			},
		},
	}

	registry := tools.NewRegistry()
	agent, err := NewAgent(&Config{
		Provider: provider,
		Registry: registry,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result := agent.Run(ctx, "Test", nil)

	assert.NotNil(t, result.Error)
	assert.Equal(t, context.Canceled, result.Error)
	// Agent may perform 1 iteration before checking context
	assert.LessOrEqual(t, result.Iterations, 1)
}

func TestAgent_RunProviderError(t *testing.T) {
	provider := &MockProvider{
		err: errors.New("provider error"),
	}

	registry := tools.NewRegistry()
	agent, err := NewAgent(&Config{
		Provider: provider,
		Registry: registry,
	})
	require.NoError(t, err)

	ctx := context.Background()
	result := agent.Run(ctx, "Test", nil)

	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "provider error")
}

func TestAgent_RunWithHistory(t *testing.T) {
	provider := &MockProvider{
		responses: []*types.CompletionResponse{
			{
				Message: types.Message{
					Role:    types.RoleAssistant,
					Content: "Based on our previous conversation, yes!",
				},
				Usage: types.Usage{
					PromptTokens:     30, // More tokens due to history
					CompletionTokens: 10,
					TotalTokens:      40,
				},
				FinishReason: "stop",
			},
		},
	}

	registry := tools.NewRegistry()
	agent, err := NewAgent(&Config{
		Provider: provider,
		Registry: registry,
	})
	require.NoError(t, err)

	// Previous conversation
	history := []types.Message{
		types.NewUserMessage("Do you remember me?"),
		types.NewAssistantMessage("Yes, I remember you!"),
	}

	ctx := context.Background()
	opts := &RunOptions{
		History: history,
	}
	result := agent.Run(ctx, "Great!", opts)

	assert.Nil(t, result.Error)
	assert.Equal(t, 1, result.Iterations)
	assert.Equal(t, 40, result.TotalTokens)
}

func TestAgent_RunWithHistoryLimit(t *testing.T) {
	provider := &MockProvider{
		responses: []*types.CompletionResponse{
			{
				Message: types.Message{
					Role:    types.RoleAssistant,
					Content: "response",
				},
				FinishReason: "stop",
			},
		},
	}

	registry := tools.NewRegistry()
	agent, err := NewAgent(&Config{
		Provider: provider,
		Registry: registry,
	})
	require.NoError(t, err)

	// Large history (6 messages)
	history := []types.Message{
		types.NewUserMessage("msg1"),
		types.NewAssistantMessage("resp1"),
		types.NewUserMessage("msg2"),
		types.NewAssistantMessage("resp2"),
		types.NewUserMessage("msg3"),
		types.NewAssistantMessage("resp3"),
	}

	ctx := context.Background()
	opts := &RunOptions{
		History:            history,
		MaxHistoryMessages: 3, // Limit to 3 messages
	}
	result := agent.Run(ctx, "msg4", opts)

	assert.Nil(t, result.Error)
	// History should be trimmed but we can't easily verify internal state
	// Just check it completed successfully
}

func TestAgent_RunOptions(t *testing.T) {
	provider := &MockProvider{
		responses: []*types.CompletionResponse{
			{
				Message: types.Message{
					Role:    types.RoleAssistant,
					Content: "response",
				},
				FinishReason: "stop",
			},
		},
	}

	registry := tools.NewRegistry()
	agent, err := NewAgent(&Config{
		Provider: provider,
		Registry: registry,
	})
	require.NoError(t, err)

	ctx := context.Background()
	opts := &RunOptions{
		Model:     "custom-model",
		MaxTokens: 1000,
	}
	result := agent.Run(ctx, "Test", opts)

	assert.Nil(t, result.Error)
	// Can't easily verify options were passed to provider
	// but at least verify it completed
}

func TestAgent_RunResultTiming(t *testing.T) {
	provider := &MockProvider{
		responses: []*types.CompletionResponse{
			{
				Message: types.Message{
					Role:    types.RoleAssistant,
					Content: "response",
				},
				FinishReason: "stop",
			},
		},
	}

	registry := tools.NewRegistry()
	agent, err := NewAgent(&Config{
		Provider: provider,
		Registry: registry,
	})
	require.NoError(t, err)

	ctx := context.Background()
	start := time.Now()
	result := agent.Run(ctx, "Test", nil)
	elapsed := time.Since(start)

	assert.Nil(t, result.Error)
	assert.True(t, result.ExecutionTime > 0)
	assert.True(t, result.ExecutionTime <= elapsed)
}

// Helper function
func floatPtr(f float64) *float64 {
	return &f
}
