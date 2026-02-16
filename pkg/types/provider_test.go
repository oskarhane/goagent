package types

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ProviderError
		expected string
	}{
		{
			name: "error with status code",
			err: &ProviderError{
				Provider:   "openai",
				Message:    "request failed",
				StatusCode: 500,
			},
			expected: "openai: request failed (status: 500)",
		},
		{
			name: "error without status code",
			err: &ProviderError{
				Provider: "vertex",
				Message:  "invalid request",
			},
			expected: "vertex: invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestProviderError_Unwrap(t *testing.T) {
	innerErr := fmt.Errorf("inner error")
	providerErr := &ProviderError{
		Provider: "test",
		Message:  "outer error",
		Cause:    innerErr,
	}

	unwrapped := providerErr.Unwrap()
	assert.Equal(t, innerErr, unwrapped)

	// Test with no wrapped error
	providerErr2 := &ProviderError{
		Provider: "test",
		Message:  "standalone error",
	}
	assert.Nil(t, providerErr2.Unwrap())
}

func TestProviderError_IsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      *ProviderError
		expected bool
	}{
		{
			name: "retryable error",
			err: &ProviderError{
				Provider:  "openai",
				Message:   "rate limited",
				Retryable: true,
			},
			expected: true,
		},
		{
			name: "non-retryable error",
			err: &ProviderError{
				Provider:  "openai",
				Message:   "invalid API key",
				Retryable: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.IsRetryable())
		})
	}
}

func TestProviderError_UnwrapChain(t *testing.T) {
	innermost := errors.New("innermost error")
	middle := fmt.Errorf("middle: %w", innermost)
	providerErr := &ProviderError{
		Provider: "test",
		Message:  "outer",
		Cause:    middle,
	}

	// Test that errors.Is works through the unwrap chain
	assert.True(t, errors.Is(providerErr, innermost))
}

func TestNewProviderError(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewProviderError("openai", "test error", 429, true, cause)

	assert.Equal(t, "openai", err.Provider)
	assert.Equal(t, "test error", err.Message)
	assert.Equal(t, 429, err.StatusCode)
	assert.True(t, err.Retryable)
	assert.Equal(t, cause, err.Cause)
}
