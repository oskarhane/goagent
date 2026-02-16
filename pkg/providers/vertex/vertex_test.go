package vertex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
			name: "valid config with custom client",
			config: &Config{
				ProjectID:  "test-project",
				HTTPClient: &http.Client{},
			},
			wantErr: false,
		},
		{
			name:    "missing project ID",
			config:  &Config{},
			wantErr: true,
			errMsg:  "project ID is required",
		},
		{
			name: "custom location",
			config: &Config{
				ProjectID:  "test-project",
				Location:   "europe-west1",
				HTTPClient: &http.Client{},
			},
			wantErr: false,
		},
		{
			name: "custom model",
			config: &Config{
				ProjectID:  "test-project",
				Model:      "gemini-2.5-flash",
				HTTPClient: &http.Client{},
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
	provider, err := NewProvider(&Config{
		ProjectID:  "test-project",
		HTTPClient: &http.Client{},
	})
	require.NoError(t, err)
	assert.Equal(t, "vertex-ai", provider.Name())
}

func TestProvider_DefaultModel(t *testing.T) {
	provider, err := NewProvider(&Config{
		ProjectID:  "test-project",
		HTTPClient: &http.Client{},
	})
	require.NoError(t, err)
	assert.Equal(t, DefaultModel, provider.DefaultModel())
}

func TestProvider_Complete_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "generateContent")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Return mock response
		response := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"role": "model",
						"parts": []map[string]any{
							{
								"text": "Hello! How can I assist you today?",
							},
						},
					},
					"finishReason": "STOP",
				},
			},
			"usageMetadata": map[string]any{
				"promptTokenCount":     12,
				"candidatesTokenCount": 18,
				"totalTokenCount":      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create custom HTTP client that routes to our test server
	client := &http.Client{
		Transport: &mockTransport{server: server},
	}

	provider, err := NewProvider(&Config{
		ProjectID:  "test-project",
		HTTPClient: client,
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
	assert.Equal(t, "Hello! How can I assist you today?", resp.Message.Content)
	assert.Equal(t, "stop", resp.FinishReason)
	assert.Equal(t, 12, resp.Usage.PromptTokens)
	assert.Equal(t, 18, resp.Usage.CompletionTokens)
	assert.Equal(t, 30, resp.Usage.TotalTokens)
}

func TestProvider_Complete_WithToolCalls(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"role": "model",
						"parts": []map[string]any{
							{
								"functionCall": map[string]any{
									"name": "get_weather",
									"args": map[string]any{
										"location": "NYC",
									},
								},
							},
						},
					},
					"finishReason": "STOP",
				},
			},
			"usageMetadata": map[string]any{
				"promptTokenCount":     15,
				"candidatesTokenCount": 20,
				"totalTokenCount":      35,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &mockTransport{server: server},
	}

	provider, err := NewProvider(&Config{
		ProjectID:  "test-project",
		HTTPClient: client,
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
	assert.Len(t, resp.Message.ToolCalls, 1)
	assert.Equal(t, "get_weather", resp.Message.ToolCalls[0].Function.Name)
}

func TestProvider_Complete_RetryOnError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		if callCount < 2 {
			// First call returns retryable error
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Second call succeeds
		response := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"role": "model",
						"parts": []map[string]any{
							{
								"text": "Success after retry",
							},
						},
					},
					"finishReason": "STOP",
				},
			},
			"usageMetadata": map[string]any{
				"promptTokenCount":     5,
				"candidatesTokenCount": 8,
				"totalTokenCount":      13,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &mockTransport{server: server},
	}

	provider, err := NewProvider(&Config{
		ProjectID:  "test-project",
		HTTPClient: client,
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

	client := &http.Client{
		Transport: &mockTransport{server: server},
	}

	provider, err := NewProvider(&Config{
		ProjectID:  "test-project",
		HTTPClient: client,
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
		// Server never responds
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &mockTransport{server: server},
	}

	provider, err := NewProvider(&Config{
		ProjectID:  "test-project",
		HTTPClient: client,
		Timeout:    1, // 1 second timeout
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

// mockTransport is a custom RoundTripper for testing that routes requests to our test server
type mockTransport struct {
	server *httptest.Server
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to point to our test server
	req.URL.Scheme = "http"
	req.URL.Host = m.server.Listener.Addr().String()
	return http.DefaultTransport.RoundTrip(req)
}
