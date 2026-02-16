package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScenarios(t *testing.T) {
	scenarios := NewScenarios()
	require.NotNil(t, scenarios)

	list := scenarios.List()
	assert.GreaterOrEqual(t, len(list), 2, "should have at least 2-3 scenarios")
}

func TestScenariosGet(t *testing.T) {
	scenarios := NewScenarios()

	tests := []struct {
		name      string
		index     int
		wantError bool
	}{
		{
			name:      "valid first scenario",
			index:     0,
			wantError: false,
		},
		{
			name:      "valid second scenario",
			index:     1,
			wantError: false,
		},
		{
			name:      "invalid negative index",
			index:     -1,
			wantError: true,
		},
		{
			name:      "invalid out of range index",
			index:     100,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario, err := scenarios.Get(tt.index)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, scenario.Name)
				assert.NotEmpty(t, scenario.IncidentType)
				assert.NotEmpty(t, scenario.InitialReport)
				assert.NotEmpty(t, scenario.AffectedServices)
			}
		})
	}
}

func TestScenariosGetByName(t *testing.T) {
	scenarios := NewScenarios()

	t.Run("find existing scenario", func(t *testing.T) {
		scenario, err := scenarios.GetByName("Database Cascading Failure")
		assert.NoError(t, err)
		assert.Equal(t, "Database Cascading Failure", scenario.Name)
		assert.Equal(t, "cascading_failure", scenario.IncidentType)
		assert.Contains(t, scenario.AffectedServices, "database")
		assert.Contains(t, scenario.AffectedServices, "auth-service")
	})

	t.Run("nonexistent scenario", func(t *testing.T) {
		_, err := scenarios.GetByName("Nonexistent Scenario")
		assert.Error(t, err)
	})
}

func TestScenariosDefault(t *testing.T) {
	scenarios := NewScenarios()
	defaultScenario := scenarios.Default()

	assert.NotEmpty(t, defaultScenario.Name)
	assert.Equal(t, "cascading_failure", defaultScenario.IncidentType)
	assert.NotEmpty(t, defaultScenario.InitialReport)
	assert.GreaterOrEqual(t, len(defaultScenario.AffectedServices), 2, "cascading failure should affect multiple services")
}

func TestScenarioGetIncidentDescription(t *testing.T) {
	scenarios := NewScenarios()
	scenario := scenarios.Default()

	description := scenario.GetIncidentDescription()

	assert.Contains(t, description, "INCIDENT")
	assert.Contains(t, description, scenario.IncidentType)
	assert.Contains(t, description, scenario.InitialReport)
	assert.Contains(t, description, "root cause")
}

func TestScenarioGetExpectedServices(t *testing.T) {
	scenarios := NewScenarios()
	scenario := scenarios.Default()

	services := scenario.GetExpectedServices()

	assert.NotEmpty(t, services)
	assert.GreaterOrEqual(t, len(services), 2, "cascading failure should list multiple services")
}

func TestCascadingFailureScenarioStructure(t *testing.T) {
	scenarios := NewScenarios()
	scenario, err := scenarios.GetByName("Database Cascading Failure")
	require.NoError(t, err)

	// Verify cascading failure structure
	assert.Equal(t, "Database Cascading Failure", scenario.Name)
	assert.Equal(t, "cascading_failure", scenario.IncidentType)
	assert.Contains(t, scenario.InitialReport, "login failures")

	// Verify the cascade order: database → auth-service
	assert.Contains(t, scenario.AffectedServices, "database")
	assert.Contains(t, scenario.AffectedServices, "auth-service")

	// Verify root cause explains the cascade
	assert.Contains(t, scenario.RootCause, "Database")
	assert.Contains(t, scenario.RootCause, "auth-service")
	assert.NotEmpty(t, scenario.Description)
}

func TestMultipleDistinctScenarios(t *testing.T) {
	scenarios := NewScenarios()
	list := scenarios.List()

	require.GreaterOrEqual(t, len(list), 2, "should have at least 2-3 distinct scenarios")

	// Verify scenarios are distinct
	names := make(map[string]bool)
	types := make(map[string]bool)

	for _, scenario := range list {
		// Check no duplicate names
		assert.False(t, names[scenario.Name], "duplicate scenario name: %s", scenario.Name)
		names[scenario.Name] = true

		// Track incident types (some overlap is okay)
		types[scenario.IncidentType] = true

		// Verify all required fields populated
		assert.NotEmpty(t, scenario.Name)
		assert.NotEmpty(t, scenario.IncidentType)
		assert.NotEmpty(t, scenario.InitialReport)
		assert.NotEmpty(t, scenario.AffectedServices)
		assert.NotEmpty(t, scenario.RootCause)
		assert.NotEmpty(t, scenario.Description)
	}

	// Should have some variety in incident types
	assert.GreaterOrEqual(t, len(types), 2, "should have at least 2 different incident types")
}

func TestScenarioMockDataAlignment(t *testing.T) {
	scenarios := NewScenarios()

	// Test that cascading failure scenario services align with mock data
	scenario, err := scenarios.GetByName("Database Cascading Failure")
	require.NoError(t, err)

	// These services should have corresponding mock data in mocks.go
	knownServices := map[string]bool{
		"database":     true,
		"auth-service": true,
		"api-service":  true,
		"cache":        true,
	}

	for _, service := range scenario.AffectedServices {
		assert.True(t, knownServices[service],
			"scenario references service '%s' which should have mock data", service)
	}
}
