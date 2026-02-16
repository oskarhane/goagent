// Package openai provides an OpenAI API implementation of the Provider interface.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/oskarhane/goagent/pkg/logger"
	"github.com/oskarhane/goagent/pkg/types"
)

const (
	// DefaultBaseURL is the default OpenAI API endpoint.
	DefaultBaseURL = "https://api.openai.com/v1"

	// DefaultModel is the default model used if none is specified.
	DefaultModel = "gpt-5.1"

	// DefaultMaxRetries is the default number of retry attempts.
	DefaultMaxRetries = 3

	// DefaultTimeout is the default request timeout in seconds.
	DefaultTimeout = 60
)

// Provider implements the types.Provider interface for OpenAI.
type Provider struct {
	apiKey     string
	baseURL    string
	model      string
	maxRetries int
	timeout    time.Duration
	client     *http.Client
	logger     *logger.Logger
}

// Config contains configuration options for the OpenAI provider.
type Config struct {
	// APIKey is the OpenAI API key for authentication.
	APIKey string

	// BaseURL allows overriding the API endpoint. Defaults to DefaultBaseURL.
	BaseURL string

	// Model specifies which model to use. Defaults to DefaultModel.
	Model string

	// MaxRetries controls retry attempts for failed requests. Defaults to DefaultMaxRetries.
	MaxRetries int

	// Timeout specifies the request timeout in seconds. Defaults to DefaultTimeout.
	Timeout int

	// HTTPClient allows injecting a custom HTTP client. If nil, a default client is used.
	HTTPClient *http.Client

	// Logger provides structured logging and tracing. If nil, defaults to Noop logger.
	Logger *logger.Logger
}

// NewProvider creates a new OpenAI provider with the given configuration.
func NewProvider(cfg *Config) (*Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai: API key is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	model := cfg.Model
	if model == "" {
		model = DefaultModel
	}

	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = DefaultMaxRetries
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
	}

	log := cfg.Logger
	if log == nil {
		log = logger.Noop()
	}

	return &Provider{
		apiKey:     cfg.APIKey,
		baseURL:    baseURL,
		model:      model,
		maxRetries: maxRetries,
		timeout:    time.Duration(timeout) * time.Second,
		client:     client,
		logger:     log,
	}, nil
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "openai"
}

// DefaultModel returns the default model for this provider.
func (p *Provider) DefaultModel() string {
	return p.model
}

// Complete sends a completion request to OpenAI and returns the response.
func (p *Provider) Complete(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error) {
	// Make a copy to avoid mutating caller's request
	reqCopy := *req
	if reqCopy.Model == "" {
		reqCopy.Model = p.model
	}

	var lastErr error

	// Start tracing span
	ctx, span := p.logger.StartSpan(ctx, "openai.complete",
		attribute.String("model", reqCopy.Model),
		attribute.Int("tools_count", len(reqCopy.Tools)),
	)
	defer func() {
		p.logger.EndSpan(span, lastErr)
	}()

	p.logger.Debug("openai completion request started", map[string]interface{}{
		"model":       reqCopy.Model,
		"tools_count": len(reqCopy.Tools),
		"max_retries": p.maxRetries,
	})

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s...
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			p.logger.Debug("retrying after backoff", map[string]interface{}{
				"attempt":         attempt + 1,
				"backoff_seconds": backoff.Seconds(),
			})
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		resp, err := p.doRequest(ctx, &reqCopy)
		if err == nil {
			p.logger.Debug("openai completion succeeded", map[string]interface{}{
				"attempt":        attempt + 1,
				"tokens_used":    resp.Usage.TotalTokens,
				"has_tool_calls": resp.Message.HasToolCalls(),
			})
			return resp, nil
		}

		lastErr = err

		// Check if error is retryable
		if provErr, ok := err.(*types.ProviderError); ok {
			if !provErr.IsRetryable() {
				p.logger.Warn("openai non-retryable error", map[string]interface{}{
					"attempt":     attempt + 1,
					"error":       err.Error(),
					"status_code": provErr.StatusCode,
				})
				return nil, err
			}
			p.logger.Warn("openai retryable error", map[string]interface{}{
				"attempt":     attempt + 1,
				"error":       err.Error(),
				"status_code": provErr.StatusCode,
			})
		} else {
			// Unknown error, don't retry
			p.logger.Error("openai unknown error", map[string]interface{}{
				"attempt": attempt + 1,
				"error":   err.Error(),
			})
			return nil, err
		}
	}

	p.logger.Error("openai max retries exceeded", map[string]interface{}{
		"max_retries": p.maxRetries,
		"error":       lastErr.Error(),
	})
	return nil, lastErr
}

// doRequest performs a single completion request without retry logic.
func (p *Provider) doRequest(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error) {
	// Convert our types to OpenAI API format
	openAIReq := convertToOpenAIRequest(req)

	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, types.NewProviderError(
			"openai",
			"failed to marshal request",
			0,
			false,
			err,
		)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		p.baseURL+"/chat/completions",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, types.NewProviderError(
			"openai",
			"failed to create request",
			0,
			false,
			err,
		)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		// Check if context was canceled
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		return nil, types.NewProviderError(
			"openai",
			"request failed",
			0,
			true, // Network errors are retryable
			err,
		)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, types.NewProviderError(
			"openai",
			"failed to read response",
			httpResp.StatusCode,
			false,
			err,
		)
	}

	// Handle error responses
	if httpResp.StatusCode != http.StatusOK {
		var errResp openAIErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			retryable := isRetryableStatus(httpResp.StatusCode)
			return nil, types.NewProviderError(
				"openai",
				errResp.Error.Message,
				httpResp.StatusCode,
				retryable,
				nil,
			)
		}

		retryable := isRetryableStatus(httpResp.StatusCode)
		return nil, types.NewProviderError(
			"openai",
			fmt.Sprintf("unexpected status code: %d", httpResp.StatusCode),
			httpResp.StatusCode,
			retryable,
			nil,
		)
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, types.NewProviderError(
			"openai",
			"failed to unmarshal response",
			httpResp.StatusCode,
			false,
			err,
		)
	}

	// Convert OpenAI response to our types
	resp := convertFromOpenAIResponse(&openAIResp)
	return &resp, nil
}

// isRetryableStatus returns true if the HTTP status code indicates a retryable error.
func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}
