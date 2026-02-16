// Package http provides a built-in HTTP client tool for agents.
// It supports common HTTP methods with authentication, timeout, and error handling.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
)

// Config configures the HTTP tool behavior.
type Config struct {
	// DefaultTimeout is the default timeout for HTTP requests.
	// If not set, defaults to 30 seconds.
	DefaultTimeout time.Duration

	// MaxResponseSize limits the response body size to prevent memory issues.
	// If not set, defaults to 10MB.
	MaxResponseSize int64
}

// Params defines the parameters for HTTP requests.
type Params struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Timeout int               `json:"timeout"` // in seconds
}

// Response represents the result of an HTTP request.
type Response struct {
	StatusCode int               `json:"status_code"`
	Status     string            `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Error      string            `json:"error,omitempty"`
}

// NewTool creates a new HTTP tool definition.
// The tool supports GET, POST, PUT, PATCH, DELETE methods with optional
// authentication headers and request bodies.
func NewTool() types.Tool {
	return tools.NewBuilder(
		"http_request",
		"Make HTTP requests to external APIs and services. "+
			"Supports GET, POST, PUT, PATCH, DELETE methods with custom headers and request bodies.",
	).
		StringParam("url", "The full URL to request (must include http:// or https://)", true).
		StringParamWithEnum(
			"method",
			"HTTP method to use",
			true,
			[]string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		).
		ObjectParam(
			"headers",
			"Optional HTTP headers to include in the request "+
				`(e.g., {"Authorization": "Bearer token", "Content-Type": "application/json"})`,
			false,
			map[string]any{},
			[]string{},
		).
		StringParam(
			"body",
			"Optional request body for POST/PUT requests (JSON string or plain text)",
			false,
		).
		IntegerParam("timeout", "Optional timeout in seconds (default: 30, max: 300)", false).
		Build()
}

// NewHandler creates a new HTTP tool handler with the given configuration.
// If config is nil, default values are used.
func NewHandler(config *Config) tools.Handler {
	if config == nil {
		config = &Config{}
	}
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 30 * time.Second
	}
	if config.MaxResponseSize == 0 {
		config.MaxResponseSize = 10 * 1024 * 1024 // 10MB
	}

	return func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		// Parse parameters
		var params Params
		if err := types.ParseToolArguments(call, &params); err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("invalid parameters: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		// Validate URL
		if !strings.HasPrefix(params.URL, "http://") && !strings.HasPrefix(params.URL, "https://") {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         "URL must start with http:// or https://",
				ExecutionTime: time.Since(start),
			}
		}

		// Validate method
		params.Method = strings.ToUpper(params.Method)
		validMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
		}
		if !validMethods[params.Method] {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("unsupported HTTP method: %s", params.Method),
				ExecutionTime: time.Since(start),
			}
		}

		// Determine timeout
		timeout := config.DefaultTimeout
		if params.Timeout > 0 {
			if params.Timeout > 300 {
				params.Timeout = 300 // max 5 minutes
			}
			timeout = time.Duration(params.Timeout) * time.Second
		}

		// Create request context with timeout
		reqCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Create request body if provided
		var bodyReader io.Reader
		if params.Body != "" {
			bodyReader = bytes.NewBufferString(params.Body)
		}

		// Create HTTP request
		req, err := http.NewRequestWithContext(reqCtx, params.Method, params.URL, bodyReader)
		if err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("failed to create request: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		// Set default User-Agent if not provided (many APIs require it)
		if req.Header.Get("User-Agent") == "" {
			req.Header.Set("User-Agent", "GoAgent/1.0")
		}

		// Add custom headers
		if params.Headers != nil {
			for key, value := range params.Headers {
				req.Header.Set(key, value)
			}
		}

		// Set default Content-Type if body is present and not set
		if params.Body != "" && req.Header.Get("Content-Type") == "" {
			// Try to detect if body is JSON
			if isJSON(params.Body) {
				req.Header.Set("Content-Type", "application/json")
			} else {
				req.Header.Set("Content-Type", "text/plain")
			}
		}

		// Execute request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("request failed: %v", err),
				ExecutionTime: time.Since(start),
			}
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		// Read response body with size limit
		limitedReader := io.LimitReader(resp.Body, config.MaxResponseSize)
		bodyBytes, err := io.ReadAll(limitedReader)
		if err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("failed to read response: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		// Extract response headers
		responseHeaders := make(map[string]string)
		for key, values := range resp.Header {
			if len(values) > 0 {
				responseHeaders[key] = values[0]
			}
		}

		// Create response object
		httpResp := Response{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Headers:    responseHeaders,
			Body:       string(bodyBytes),
		}

		// Include error message for non-2xx status codes
		if resp.StatusCode >= 400 {
			httpResp.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		// Marshal response to JSON
		resultJSON, err := json.Marshal(httpResp)
		if err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("failed to marshal response: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		return types.ToolResult{
			ToolCallID:    call.ID,
			ToolName:      call.Function.Name,
			Content:       string(resultJSON),
			ExecutionTime: time.Since(start),
		}
	}
}

// isJSON checks if a string is valid JSON.
func isJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}
