package types

import (
	"context"
	"fmt"
)

// Provider defines the interface for LLM provider implementations.
// This abstraction allows the same agent code to work with different
// providers (OpenAI, Vertex AI, etc.) by standardizing the completion API.
//
// Implementations must handle:
//   - Authentication with their respective APIs
//   - Request/response format translation
//   - Rate limiting and exponential backoff
//   - Context cancellation and timeout handling
type Provider interface {
	// Complete sends a completion request to the LLM provider and returns the response.
	// The method blocks until the response is received or an error occurs.
	//
	// The context is used for cancellation and timeout handling. Implementations
	// should respect context cancellation and return context.Canceled or
	// context.DeadlineExceeded as appropriate.
	//
	// Errors returned should provide clear, actionable information about what
	// went wrong (authentication, rate limits, invalid requests, etc.).
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Name returns a human-readable identifier for this provider.
	// Used in logging and error messages to identify which provider is in use.
	// Examples: "openai", "vertex-ai", "anthropic"
	Name() string

	// DefaultModel returns the default model identifier for this provider.
	// This is used when no model is explicitly specified in the request.
	// Examples: "gpt-5.1", "gemini-2.5-pro", "claude-3-opus"
	DefaultModel() string
}

// ProviderConfig contains common configuration options shared across providers.
// Provider implementations can embed this struct and add provider-specific fields.
type ProviderConfig struct {
	// APIKey is the authentication key for the provider API.
	// Some providers (like Vertex AI) may use alternative auth methods.
	APIKey string

	// BaseURL allows overriding the default API endpoint.
	// Useful for testing, proxies, or alternative endpoints.
	BaseURL string

	// DefaultModel specifies which model to use when not specified in requests.
	DefaultModel string

	// MaxRetries controls how many times to retry failed requests.
	// Set to 0 to disable retries. Default implementations typically use 3.
	MaxRetries int

	// Timeout specifies the default timeout for API requests.
	// Can be overridden by context deadlines in individual requests.
	Timeout int // in seconds
}

// ProviderError represents errors returned by provider implementations.
// It includes structured information to help with debugging and error handling.
type ProviderError struct {
	// Cause is the underlying error that triggered this error.
	Cause error

	// Provider identifies which provider returned the error.
	Provider string

	// Message is a human-readable error description.
	Message string

	// StatusCode is the HTTP status code if applicable.
	// Set to 0 for non-HTTP errors.
	StatusCode int

	// Retryable indicates whether the request can be retried.
	// True for rate limits, timeouts, and transient failures.
	// False for authentication errors, invalid requests, etc.
	Retryable bool
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s: %s (status: %d)", e.Provider, e.Message, e.StatusCode)
	}
	return e.Provider + ": " + e.Message
}

// Unwrap returns the underlying error for error chain inspection.
func (e *ProviderError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns true if the error represents a transient failure
// that can be retried with exponential backoff.
func (e *ProviderError) IsRetryable() bool {
	return e.Retryable
}

// NewProviderError creates a new provider error with the given details.
func NewProviderError(provider, message string, statusCode int, retryable bool, cause error) *ProviderError {
	return &ProviderError{
		Provider:   provider,
		StatusCode: statusCode,
		Message:    message,
		Retryable:  retryable,
		Cause:      cause,
	}
}
