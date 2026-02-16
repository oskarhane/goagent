// Package vertex provides a Google Cloud Vertex AI implementation of the Provider interface.
package vertex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/oskarhane/goagent/pkg/types"
)

const (
	// DefaultModel is the default Gemini model used if none is specified.
	DefaultModel = "gemini-1.5-pro"

	// DefaultMaxRetries is the default number of retry attempts.
	DefaultMaxRetries = 3

	// DefaultTimeout is the default request timeout in seconds.
	DefaultTimeout = 60

	// DefaultLocation is the default Google Cloud region.
	DefaultLocation = "us-central1"
)

// Provider implements the types.Provider interface for Google Cloud Vertex AI.
type Provider struct {
	projectID  string
	location   string
	model      string
	maxRetries int
	timeout    time.Duration
	client     *http.Client
}

// Config contains configuration options for the Vertex AI provider.
type Config struct {
	// ProjectID is the Google Cloud project ID.
	ProjectID string

	// Location is the Google Cloud region (e.g., "us-central1").
	Location string

	// Model specifies which model to use. Defaults to DefaultModel.
	Model string

	// MaxRetries controls retry attempts for failed requests. Defaults to DefaultMaxRetries.
	MaxRetries int

	// Timeout specifies the request timeout in seconds. Defaults to DefaultTimeout.
	Timeout int

	// HTTPClient allows injecting a custom HTTP client. If nil, one is created with ADC auth.
	HTTPClient *http.Client
}

// NewProvider creates a new Vertex AI provider with the given configuration.
func NewProvider(cfg Config) (*Provider, error) {
	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("vertex: project ID is required")
	}

	location := cfg.Location
	if location == "" {
		location = DefaultLocation
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
		// Use Application Default Credentials for authentication
		ctx := context.Background()
		creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return nil, fmt.Errorf("vertex: failed to find default credentials: %w", err)
		}

		client = oauth2.NewClient(ctx, creds.TokenSource)
		client.Timeout = time.Duration(timeout) * time.Second
	}

	return &Provider{
		projectID:  cfg.ProjectID,
		location:   location,
		model:      model,
		maxRetries: maxRetries,
		timeout:    time.Duration(timeout) * time.Second,
		client:     client,
	}, nil
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "vertex-ai"
}

// DefaultModel returns the default model for this provider.
func (p *Provider) DefaultModel() string {
	return p.model
}

// Complete sends a completion request to Vertex AI and returns the response.
func (p *Provider) Complete(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error) {
	// Make a copy to avoid mutating caller's request
	reqCopy := *req
	if reqCopy.Model == "" {
		reqCopy.Model = p.model
	}

	var lastErr error
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s...
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		resp, err := p.doRequest(ctx, &reqCopy)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Check if error is retryable
		if provErr, ok := err.(*types.ProviderError); ok {
			if !provErr.IsRetryable() {
				return nil, err
			}
		} else {
			// Unknown error, don't retry
			return nil, err
		}
	}

	return nil, lastErr
}

// doRequest performs a single completion request without retry logic.
func (p *Provider) doRequest(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error) {
	// Convert our types to Vertex AI API format
	vertexReq, err := convertToVertexRequest(req)
	if err != nil {
		return nil, types.NewProviderError(
			"vertex-ai",
			"failed to convert request",
			0,
			false,
			err,
		)
	}

	reqBody, err := json.Marshal(vertexReq)
	if err != nil {
		return nil, types.NewProviderError(
			"vertex-ai",
			"failed to marshal request",
			0,
			false,
			err,
		)
	}

	// Build the API endpoint URL
	endpoint := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		p.location,
		p.projectID,
		p.location,
		req.Model,
	)

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		endpoint,
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, types.NewProviderError(
			"vertex-ai",
			"failed to create request",
			0,
			false,
			err,
		)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		// Check if context was canceled
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		return nil, types.NewProviderError(
			"vertex-ai",
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
			"vertex-ai",
			"failed to read response",
			httpResp.StatusCode,
			false,
			err,
		)
	}

	// Handle error responses
	if httpResp.StatusCode != http.StatusOK {
		var errResp vertexErrorResponse
		if unmarshalErr := json.Unmarshal(body, &errResp); unmarshalErr == nil && errResp.Error.Message != "" {
			retryable := isRetryableStatus(httpResp.StatusCode)
			return nil, types.NewProviderError(
				"vertex-ai",
				errResp.Error.Message,
				httpResp.StatusCode,
				retryable,
				nil,
			)
		}

		retryable := isRetryableStatus(httpResp.StatusCode)
		return nil, types.NewProviderError(
			"vertex-ai",
			fmt.Sprintf("unexpected status code: %d", httpResp.StatusCode),
			httpResp.StatusCode,
			retryable,
			nil,
		)
	}

	var vertexResp vertexResponse
	if unmarshalErr := json.Unmarshal(body, &vertexResp); unmarshalErr != nil {
		return nil, types.NewProviderError(
			"vertex-ai",
			"failed to unmarshal response",
			httpResp.StatusCode,
			false,
			unmarshalErr,
		)
	}

	// Convert Vertex AI response to our types
	resp, convErr := convertFromVertexResponse(&vertexResp, req.Model)
	if convErr != nil {
		return nil, types.NewProviderError(
			"vertex-ai",
			"failed to convert response",
			httpResp.StatusCode,
			false,
			convErr,
		)
	}

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
