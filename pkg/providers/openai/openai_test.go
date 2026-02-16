package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oskarhane/goagent/pkg/types"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name:    "missing API key",
			config:  &Config{},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "custom base URL",
			config: &Config{
				APIKey:  "test-key",
				BaseURL: "https://custom.api.com",
			},
			wantErr: false,
		},
		{
			name: "custom model",
			config: &Config{
				APIKey: "test-key",
				Model:  "gpt-5-mini",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestProvider_Name(t *testing.T) {
	provider, err := NewProvider(&Config{APIKey: "test-key"})
	require.NoError(t, err)
	assert.Equal(t, "openai", provider.Name())
}

func TestProvider_DefaultModel(t *testing.T) {
	provider, err := NewProvider(&Config{APIKey: "test-key"})
	require.NoError(t, err)
	assert.Equal(t, DefaultModel, provider.DefaultModel())
}

func TestProvider_Complete_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Return mock response
		response := map[string]any{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"created": 1677652288,
			"model":   "gpt-5.1",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "Hello! How can I help you?",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider, err := NewProvider(&Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	req := &types.CompletionRequest{
		Messages: []types.Message{
			types.NewUserMessage("Hello"),
		},
	}

	resp, err := provider.Complete(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, types.RoleAssistant, resp.Message.Role)
	assert.Equal(t, "Hello! How can I help you?", resp.Message.Content)
	assert.Equal(t, "stop", resp.FinishReason)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 20, resp.Usage.CompletionTokens)
	assert.Equal(t, 30, resp.Usage.TotalTokens)
}

func TestProvider_Complete_WithToolCalls(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]any{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"created": 1677652288,
			"model":   "gpt-5.1",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "",
						"tool_calls": []map[string]any{
							{
								"id":   "call_123",
								"type": "function",
								"function": map[string]any{
									"name":      "get_weather",
									"arguments": `{"location": "NYC"}`,
								},
							},
						},
					},
					"finish_reason": "tool_calls",
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     15,
				"completion_tokens": 25,
				"total_tokens":      40,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider, err := NewProvider(&Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	req := &types.CompletionRequest{
		Messages: []types.Message{
			types.NewUserMessage("What's the weather?"),
		},
	}

	resp, err := provider.Complete(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "tool_calls", resp.FinishReason)
	assert.Len(t, resp.Message.ToolCalls, 1)
	assert.Equal(t, "call_123", resp.Message.ToolCalls[0].ID)
	assert.Equal(t, "get_weather", resp.Message.ToolCalls[0].Function.Name)
}

func TestProvider_Complete_RetryOnError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		if callCount < 2 {
			// First call returns retryable error
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Second call succeeds
		response := map[string]any{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"created": 1677652288,
			"model":   "gpt-5.1",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "Success after retry",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     5,
				"completion_tokens": 10,
				"total_tokens":      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider, err := NewProvider(&Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	req := &types.CompletionRequest{
		Messages: []types.Message{
			types.NewUserMessage("Test"),
		},
	}

	resp, err := provider.Complete(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "Success after retry", resp.Message.Content)
	assert.Equal(t, 2, callCount, "Should retry once")
}

func TestProvider_Complete_MaxRetriesExceeded(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": {"message": "Internal server error"}}`))
	}))
	defer server.Close()

	provider, err := NewProvider(&Config{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		MaxRetries: 2,
	})
	require.NoError(t, err)

	ctx := context.Background()
	req := &types.CompletionRequest{
		Messages: []types.Message{
			types.NewUserMessage("Test"),
		},
	}

	_, err = provider.Complete(ctx, req)

	require.Error(t, err)
	// Error message varies - just check we got an error after retries
	assert.Equal(t, 3, callCount, "Should attempt initial + 2 retries")
}

func TestProvider_Complete_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// Server never responds (simulated by not writing)
	}))
	defer server.Close()

	provider, err := NewProvider(&Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 1, // 1 second timeout
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &types.CompletionRequest{
		Messages: []types.Message{
			types.NewUserMessage("Test"),
		},
	}

	_, err = provider.Complete(ctx, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestProvider_Complete_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	provider, err := NewProvider(&Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	req := &types.CompletionRequest{
		Messages: []types.Message{
			types.NewUserMessage("Test"),
		},
	}

	_, err = provider.Complete(ctx, req)

	require.Error(t, err)
	// Error message varies - just check we got a JSON parsing error
	assert.True(t, strings.Contains(err.Error(), "unmarshal") || strings.Contains(err.Error(), "decode"))
}
