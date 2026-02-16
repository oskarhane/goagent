package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/oskarhane/goagent/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockLogsTool(t *testing.T) {
	tool := NewMockLogsTool()
	assert.Equal(t, "mock_logs", tool.Function.Name)
	assert.Contains(t, tool.Function.Description, "log entries")

	handler := NewMockLogsHandler()

	// Test with auth-service
	call := types.ToolCall{
		ID:   "test-1",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "mock_logs",
			Arguments: `{"service_name": "auth-service"}`,
		},
	}

	result := handler(context.Background(), call)
	assert.Equal(t, "test-1", result.ToolCallID)
	assert.Equal(t, "mock_logs", result.ToolName)
	assert.Empty(t, result.Error)
	assert.NotEmpty(t, result.Content)

	// Parse result
	var data map[string]any
	err := json.Unmarshal([]byte(result.Content), &data)
	require.NoError(t, err)
	assert.Equal(t, "auth-service", data["service"])
	assert.Greater(t, data["count"].(float64), float64(0))
}

func TestMockMetricsTool(t *testing.T) {
	tool := NewMockMetricsTool()
	assert.Equal(t, "mock_metrics", tool.Function.Name)
	assert.Contains(t, tool.Function.Description, "metrics")

	handler := NewMockMetricsHandler()

	call := types.ToolCall{
		ID:   "test-2",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "mock_metrics",
			Arguments: `{"service_name": "database"}`,
		},
	}

	result := handler(context.Background(), call)
	assert.Empty(t, result.Error)

	var data map[string]any
	err := json.Unmarshal([]byte(result.Content), &data)
	require.NoError(t, err)
	assert.Equal(t, "database", data["service"])

	metrics, ok := data["metrics"].(map[string]any)
	require.True(t, ok)
	assert.NotEmpty(t, metrics)
}

func TestMockServiceStatusTool(t *testing.T) {
	tool := NewMockServiceStatusTool()
	assert.Equal(t, "mock_service_status", tool.Function.Name)
	assert.Contains(t, tool.Function.Description, "health")

	handler := NewMockServiceStatusHandler()

	// Test cache (healthy)
	call := types.ToolCall{
		ID:   "test-3",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "mock_service_status",
			Arguments: `{"service_name": "cache"}`,
		},
	}

	result := handler(context.Background(), call)
	assert.Empty(t, result.Error)

	var data map[string]any
	err := json.Unmarshal([]byte(result.Content), &data)
	require.NoError(t, err)
	assert.Equal(t, "cache", data["service"])
	assert.Equal(t, "healthy", data["status"])
	assert.True(t, data["healthy"].(bool))
}

func TestMockDataDeterminism(t *testing.T) {
	// Verify that data is deterministic
	handler := NewMockLogsHandler()

	call := types.ToolCall{
		ID:   "test-4",
		Type: "function",
		Function: types.FunctionCall{
			Name:      "mock_logs",
			Arguments: `{"service_name": "api-service"}`,
		},
	}

	result1 := handler(context.Background(), call)
	result2 := handler(context.Background(), call)

	// Content should be identical (deterministic)
	// Note: timestamps will vary, but the structure should be consistent
	var data1, data2 map[string]any
	json.Unmarshal([]byte(result1.Content), &data1)
	json.Unmarshal([]byte(result2.Content), &data2)

	assert.Equal(t, data1["service"], data2["service"])
	assert.Equal(t, data1["count"], data2["count"])
}

func TestCascadingFailureScenario(t *testing.T) {
	// Verify that different services show patterns of cascading failure
	statusHandler := NewMockServiceStatusHandler()

	services := []string{"database", "auth-service", "api-service", "cache"}
	statuses := make(map[string]string)

	for _, svc := range services {
		call := types.ToolCall{
			ID:   "test-" + svc,
			Type: "function",
			Function: types.FunctionCall{
				Name:      "mock_service_status",
				Arguments: `{"service_name": "` + svc + `"}`,
			},
		}

		result := statusHandler(context.Background(), call)
		var data map[string]any
		json.Unmarshal([]byte(result.Content), &data)
		statuses[svc] = data["status"].(string)
	}

	// Verify cascading failure pattern
	assert.Equal(t, "critical", statuses["database"])     // Root cause
	assert.Equal(t, "degraded", statuses["auth-service"]) // Depends on database
	assert.Equal(t, "degraded", statuses["api-service"])  // Depends on auth
	assert.Equal(t, "healthy", statuses["cache"])         // Independent
}
