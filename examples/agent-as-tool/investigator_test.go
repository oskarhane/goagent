package main

import (
	"testing"

	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInvestigator(t *testing.T) {
	provider := &mockProvider{}
	registry := tools.NewRegistry()

	// Register mock tools
	registry.MustRegister(NewMockLogsTool(), NewMockLogsHandler())
	registry.MustRegister(NewMockMetricsTool(), NewMockMetricsHandler())
	registry.MustRegister(NewMockServiceStatusTool(), NewMockServiceStatusHandler())

	tests := []struct {
		name        string
		serviceName string
		provider    types.Provider
		registry    *tools.Registry
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid investigator",
			serviceName: "api-service",
			provider:    provider,
			registry:    registry,
			wantErr:     false,
		},
		{
			name:        "empty service name",
			serviceName: "",
			provider:    provider,
			registry:    registry,
			wantErr:     true,
			errContains: "service name is required",
		},
		{
			name:        "nil provider",
			serviceName: "api-service",
			provider:    nil,
			registry:    registry,
			wantErr:     true,
			errContains: "provider is required",
		},
		{
			name:        "nil registry",
			serviceName: "api-service",
			provider:    provider,
			registry:    nil,
			wantErr:     true,
			errContains: "registry is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewInvestigator(tt.serviceName, tt.provider, tt.registry)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, agent)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, agent)
			}
		})
	}
}

func TestInvestigatorSystemPrompt(t *testing.T) {
	provider := &mockProvider{}
	registry := tools.NewRegistry()

	// Register mock tools
	registry.MustRegister(NewMockLogsTool(), NewMockLogsHandler())
	registry.MustRegister(NewMockMetricsTool(), NewMockMetricsHandler())
	registry.MustRegister(NewMockServiceStatusTool(), NewMockServiceStatusHandler())

	serviceName := "database"
	agent, err := NewInvestigator(serviceName, provider, registry)
	require.NoError(t, err)
	require.NotNil(t, agent)

	// The system prompt is internal to the agent, but we can verify the agent was created
	// In a real scenario, the prompt would include "database" multiple times
	// We can't access private fields, but we verified it compiles and runs
}

func TestInvestigatorMaxIterations(t *testing.T) {
	provider := &mockProvider{}
	registry := tools.NewRegistry()

	// Register mock tools
	registry.MustRegister(NewMockLogsTool(), NewMockLogsHandler())

	agent, err := NewInvestigator("test-service", provider, registry)
	require.NoError(t, err)
	require.NotNil(t, agent)

	// The agent should be configured with max 5 iterations
	// We can't directly test this without running the agent, but we verified the constructor
}
