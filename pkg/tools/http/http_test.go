package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oskarhane/goagent/pkg/types"
)

func TestNewTool(t *testing.T) {
	tool := NewTool()

	assert.NotNil(t, tool)
	assert.Equal(t, "http_request", tool.Function.Name)
	assert.NotEmpty(t, tool.Function.Description)
	assert.NotNil(t, tool.Function.Parameters)
}

func TestNewHandler_GET(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	handler := NewHandler(nil)
	ctx := context.Background()

	call := types.ToolCall{
		ID:   "call_1",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "http_request",
			Arguments: `{"method": "GET", "url": "` + server.URL + `"}`,
		},
	}

	result := handler(ctx, call)

	assert.Empty(t, result.Error)
	assert.Contains(t, result.Content, `"status_code":200`)
	assert.Contains(t, result.Content, `"status":"200 OK"`)
	assert.Contains(t, result.Content, `"body":"{\"status\": \"ok\"}"`)
}

func TestNewHandler_POST(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"created": true}`))
	}))
	defer server.Close()

	handler := NewHandler(nil)
	ctx := context.Background()

	call := types.ToolCall{
		ID:   "call_2",
		Type: "function",
		Function: types.FunctionCall{
			Name: "http_request",
			Arguments: `{"method": "POST", "url": "` + server.URL + `", "body": "{\"test\": \"data\"}", ` +
				`"headers": {"Content-Type": "application/json"}}`,
		},
	}

	result := handler(ctx, call)

	assert.Empty(t, result.Error)
	assert.Contains(t, result.Content, `"status_code":201`)
	assert.Contains(t, result.Content, `"body":"{\"created\": true}"`)
}

func TestNewHandler_CustomHeaders(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	handler := NewHandler(nil)
	ctx := context.Background()

	call := types.ToolCall{
		ID:   "call_3",
		Type: "function",
		Function: types.FunctionCall{
			Name: "http_request",
			Arguments: `{"method": "GET", "url": "` + server.URL + `", ` +
				`"headers": {"Authorization": "Bearer token123", "X-Custom-Header": "custom-value"}}`,
		},
	}

	result := handler(ctx, call)

	assert.Empty(t, result.Error)
	assert.Contains(t, result.Content, `"status_code":200`)
}

func TestNewHandler_ErrorResponse(t *testing.T) {
	// Create test server returning error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	handler := NewHandler(nil)
	ctx := context.Background()

	call := types.ToolCall{
		ID:   "call_4",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "http_request",
			Arguments: `{"method": "GET", "url": "` + server.URL + `"}`,
		},
	}

	result := handler(ctx, call)

	// Should still succeed but include error in response
	assert.Empty(t, result.Error)
	assert.Contains(t, result.Content, `"status_code":404`)
	assert.Contains(t, result.Content, `"error":"HTTP 404:`)
}

func TestNewHandler_InvalidURL(t *testing.T) {
	handler := NewHandler(nil)
	ctx := context.Background()

	call := types.ToolCall{
		ID:   "call_5",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "http_request",
			Arguments: `{"method": "GET", "url": "not-a-valid-url"}`,
		},
	}

	result := handler(ctx, call)

	assert.NotEmpty(t, result.Error)
}

func TestNewHandler_AllMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, method, r.Method)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			}))
			defer server.Close()

			handler := NewHandler(nil)
			ctx := context.Background()

			call := types.ToolCall{
				ID:   "call_" + method,
				Type: "function",
				Function: types.FunctionCall{
					Name:      "http_request",
					Arguments: `{"method": "` + method + `", "url": "` + server.URL + `"}`,
				},
			}

			result := handler(ctx, call)

			assert.Empty(t, result.Error, "Method %s should succeed", method)
			assert.Contains(t, result.Content, `"status_code":200`)
		})
	}
}

func TestNewHandler_UserAgent(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		assert.NotEmpty(t, userAgent)
		assert.Contains(t, userAgent, "GoAgent")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handler := NewHandler(nil)
	ctx := context.Background()

	call := types.ToolCall{
		ID:   "call_ua",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "http_request",
			Arguments: `{"method": "GET", "url": "` + server.URL + `"}`,
		},
	}

	result := handler(ctx, call)
	assert.Empty(t, result.Error)
}

func TestNewHandler_ParseError(t *testing.T) {
	handler := NewHandler(nil)
	ctx := context.Background()

	call := types.ToolCall{
		ID:   "call_parse",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "http_request",
			Arguments: `{invalid json}`,
		},
	}

	result := handler(ctx, call)

	assert.NotEmpty(t, result.Error)
	require.Contains(t, result.Error, "invalid")
}
