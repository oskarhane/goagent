package main

import (
	"context"

	"github.com/oskarhane/goagent/pkg/types"
)

// mockProvider is a configurable mock provider for testing.
// Used across multiple test files to avoid redeclaration.
type mockProvider struct {
	response *types.CompletionResponse
	err      error
}

func (m *mockProvider) Complete(_ context.Context, _ *types.CompletionRequest) (*types.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	// Default response
	return &types.CompletionResponse{
		Message: types.Message{
			Role:    types.RoleAssistant,
			Content: "Default mock response",
		},
		Usage: types.Usage{TotalTokens: 10},
	}, nil
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) DefaultModel() string {
	return "mock-model"
}
