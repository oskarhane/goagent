package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
)

// Mock tool implementations for SRE diagnostics
// These tools return realistic, deterministic data for demonstration purposes

// NewMockLogsHandler creates a handler that returns simulated log entries for a service
func NewMockLogsHandler() func(context.Context, types.ToolCall) types.ToolResult {
	return func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		// Parse parameters
		var params struct {
			ServiceName string `json:"service_name"`
		}
		if err := types.ParseToolArguments(call, &params); err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         err.Error(),
				ExecutionTime: time.Since(start),
			}
		}

		// Generate deterministic mock logs based on service name
		logs := generateMockLogs(params.ServiceName)

		// Return result as JSON
		resultJSON, _ := json.Marshal(map[string]any{
			"service": params.ServiceName,
			"entries": logs,
			"count":   len(logs),
		})

		return types.ToolResult{
			ToolCallID:    call.ID,
			ToolName:      call.Function.Name,
			Content:       string(resultJSON),
			ExecutionTime: time.Since(start),
		}
	}
}

// NewMockMetricsHandler creates a handler that returns time-series metrics for a service
func NewMockMetricsHandler() func(context.Context, types.ToolCall) types.ToolResult {
	return func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		// Parse parameters
		var params struct {
			ServiceName string `json:"service_name"`
		}
		if err := types.ParseToolArguments(call, &params); err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         err.Error(),
				ExecutionTime: time.Since(start),
			}
		}

		// Generate deterministic mock metrics based on service name
		metrics := generateMockMetrics(params.ServiceName)

		// Return result as JSON
		resultJSON, _ := json.Marshal(map[string]any{
			"service": params.ServiceName,
			"metrics": metrics,
		})

		return types.ToolResult{
			ToolCallID:    call.ID,
			ToolName:      call.Function.Name,
			Content:       string(resultJSON),
			ExecutionTime: time.Since(start),
		}
	}
}

// NewMockServiceStatusHandler creates a handler that returns health check results for a service
func NewMockServiceStatusHandler() func(context.Context, types.ToolCall) types.ToolResult {
	return func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		// Parse parameters
		var params struct {
			ServiceName string `json:"service_name"`
		}
		if err := types.ParseToolArguments(call, &params); err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         err.Error(),
				ExecutionTime: time.Since(start),
			}
		}

		// Generate deterministic mock status based on service name
		status := generateMockStatus(params.ServiceName)

		// Return result as JSON
		resultJSON, _ := json.Marshal(status)

		return types.ToolResult{
			ToolCallID:    call.ID,
			ToolName:      call.Function.Name,
			Content:       string(resultJSON),
			ExecutionTime: time.Since(start),
		}
	}
}

// Tool builders

// NewMockLogsTool creates the mock_logs tool definition
func NewMockLogsTool() types.Tool {
	return tools.NewBuilder(
		"mock_logs",
		"Retrieve recent log entries from a service. Returns simulated log data for demo purposes.",
	).
		StringParam("service_name", "Name of the service to retrieve logs from (e.g., api-service, auth-service, database, cache)", true).
		Build()
}

// NewMockMetricsTool creates the mock_metrics tool definition
func NewMockMetricsTool() types.Tool {
	return tools.NewBuilder(
		"mock_metrics",
		"Retrieve time-series metrics for a service. Returns CPU, memory, request rate, and error rate data.",
	).
		StringParam("service_name", "Name of the service to retrieve metrics from", true).
		Build()
}

// NewMockServiceStatusTool creates the mock_service_status tool definition
func NewMockServiceStatusTool() types.Tool {
	return tools.NewBuilder(
		"mock_service_status",
		"Check the health status of a service. Returns current health state, uptime, and dependency status.",
	).
		StringParam("service_name", "Name of the service to check status for", true).
		Build()
}

// Mock data generators

func generateMockLogs(serviceName string) []map[string]any {
	baseTime := time.Now().Add(-15 * time.Minute)

	// Different log patterns for different services to simulate cascading failures
	switch serviceName {
	case "auth-service":
		return []map[string]any{
			{
				"timestamp": baseTime.Add(1 * time.Minute).Format(time.RFC3339),
				"level":     "ERROR",
				"message":   "Database connection pool exhausted",
				"details":   map[string]any{"active_connections": 100, "max_connections": 100},
			},
			{
				"timestamp": baseTime.Add(2 * time.Minute).Format(time.RFC3339),
				"level":     "WARN",
				"message":   "Authentication requests timing out",
				"details":   map[string]any{"timeout_ms": 5000, "count": 47},
			},
			{
				"timestamp": baseTime.Add(3 * time.Minute).Format(time.RFC3339),
				"level":     "ERROR",
				"message":   "Failed to authenticate user token",
				"details":   map[string]any{"error": "database timeout"},
			},
		}
	case "api-service":
		return []map[string]any{
			{
				"timestamp": baseTime.Add(4 * time.Minute).Format(time.RFC3339),
				"level":     "ERROR",
				"message":   "Auth service returned 503",
				"details":   map[string]any{"endpoint": "/api/v1/validate", "status_code": 503},
			},
			{
				"timestamp": baseTime.Add(5 * time.Minute).Format(time.RFC3339),
				"level":     "ERROR",
				"message":   "Request failed with upstream service unavailable",
				"details":   map[string]any{"service": "auth-service", "retries": 3},
			},
			{
				"timestamp": baseTime.Add(6 * time.Minute).Format(time.RFC3339),
				"level":     "WARN",
				"message":   "Circuit breaker opened for auth-service",
				"details":   map[string]any{"failure_rate": 0.85, "threshold": 0.5},
			},
		}
	case "database":
		return []map[string]any{
			{
				"timestamp": baseTime.Add(0 * time.Minute).Format(time.RFC3339),
				"level":     "WARN",
				"message":   "Slow query detected",
				"details":   map[string]any{"query_time_ms": 8500, "table": "user_sessions"},
			},
			{
				"timestamp": baseTime.Add(1 * time.Minute).Format(time.RFC3339),
				"level":     "ERROR",
				"message":   "Connection pool saturation",
				"details":   map[string]any{"active": 100, "idle": 0, "waiting": 45},
			},
			{
				"timestamp": baseTime.Add(2 * time.Minute).Format(time.RFC3339),
				"level":     "CRIT",
				"message":   "Deadlock detected on user_sessions table",
				"details":   map[string]any{"transaction_ids": []int{12345, 12346}},
			},
		}
	case "cache":
		return []map[string]any{
			{
				"timestamp": baseTime.Format(time.RFC3339),
				"level":     "INFO",
				"message":   "Cache operating normally",
				"details":   map[string]any{"hit_rate": 0.95, "evictions": 12},
			},
		}
	default:
		return []map[string]any{
			{
				"timestamp": baseTime.Format(time.RFC3339),
				"level":     "INFO",
				"message":   fmt.Sprintf("Service %s operating normally", serviceName),
				"details":   map[string]any{},
			},
		}
	}
}

func generateMockMetrics(serviceName string) map[string]any {
	// Different metric patterns for different services
	switch serviceName {
	case "auth-service":
		return map[string]any{
			"cpu_percent":          []float64{45.2, 48.1, 52.3, 58.7, 65.4, 72.1, 78.9, 85.2, 92.3, 95.8},
			"memory_mb":            []float64{512, 518, 525, 534, 548, 567, 589, 612, 645, 678},
			"request_rate_per_sec": []float64{120, 125, 132, 145, 158, 175, 198, 215, 232, 245},
			"error_rate_percent":   []float64{0.1, 0.2, 0.5, 1.2, 3.5, 8.7, 15.3, 25.8, 38.4, 52.1},
			"p95_latency_ms":       []float64{45, 52, 68, 125, 340, 1250, 2800, 4500, 5000, 5000},
		}
	case "api-service":
		return map[string]any{
			"cpu_percent":          []float64{35.1, 37.2, 39.5, 42.1, 45.8, 52.3, 58.7, 65.4, 72.1, 75.8},
			"memory_mb":            []float64{768, 775, 782, 791, 803, 818, 835, 856, 881, 912},
			"request_rate_per_sec": []float64{450, 445, 438, 425, 410, 385, 352, 318, 285, 250},
			"error_rate_percent":   []float64{0.3, 0.5, 1.2, 2.8, 5.5, 12.3, 18.7, 25.4, 32.8, 42.1},
			"p95_latency_ms":       []float64{125, 145, 178, 245, 380, 650, 1200, 1850, 2500, 3200},
		}
	case "database":
		return map[string]any{
			"cpu_percent":          []float64{55.2, 58.3, 62.1, 67.5, 73.8, 81.2, 88.5, 94.2, 98.7, 99.8},
			"memory_mb":            []float64{2048, 2058, 2071, 2089, 2115, 2148, 2189, 2235, 2287, 2345},
			"connections":          []float64{65, 70, 75, 82, 88, 93, 97, 99, 100, 100},
			"query_rate":           []float64{850, 920, 980, 1050, 1120, 1180, 1200, 1150, 1050, 950},
			"slow_queries_per_min": []float64{0, 0, 1, 2, 5, 8, 12, 18, 25, 32},
		}
	case "cache":
		return map[string]any{
			"cpu_percent":    []float64{12.3, 12.5, 12.8, 13.1, 13.5, 13.8, 14.2, 14.5, 14.8, 15.1},
			"memory_mb":      []float64{1024, 1028, 1032, 1036, 1041, 1045, 1050, 1054, 1058, 1062},
			"hit_rate":       []float64{0.95, 0.95, 0.94, 0.94, 0.95, 0.95, 0.94, 0.95, 0.95, 0.95},
			"requests_per_s": []float64{2500, 2520, 2480, 2510, 2490, 2530, 2500, 2520, 2510, 2495},
		}
	default:
		return map[string]any{
			"cpu_percent": []float64{25.0, 26.0, 25.5, 26.5, 25.8, 26.2, 25.9, 26.1, 25.7, 26.0},
			"memory_mb":   []float64{512, 515, 518, 520, 522, 525, 528, 530, 532, 535},
		}
	}
}

func generateMockStatus(serviceName string) map[string]any {
	// Different status patterns for different services
	switch serviceName {
	case "auth-service":
		return map[string]any{
			"service":    serviceName,
			"status":     "degraded",
			"healthy":    false,
			"uptime_sec": 3600 * 24 * 7, // 7 days
			"checks": map[string]any{
				"http_endpoint": map[string]any{
					"status":        "unhealthy",
					"response_time": "5000ms",
					"error":         "timeout",
				},
				"database": map[string]any{
					"status":        "unhealthy",
					"response_time": "5000ms",
					"error":         "connection pool exhausted",
				},
				"cache": map[string]any{
					"status":        "healthy",
					"response_time": "2ms",
				},
			},
			"dependencies": map[string]string{
				"database": "unhealthy",
				"cache":    "healthy",
			},
		}
	case "api-service":
		return map[string]any{
			"service":    serviceName,
			"status":     "degraded",
			"healthy":    false,
			"uptime_sec": 3600 * 24 * 14, // 14 days
			"checks": map[string]any{
				"http_endpoint": map[string]any{
					"status":        "healthy",
					"response_time": "125ms",
				},
				"auth_service": map[string]any{
					"status":        "unhealthy",
					"response_time": "3000ms",
					"error":         "503 service unavailable",
				},
			},
			"dependencies": map[string]string{
				"auth-service": "unhealthy",
				"database":     "degraded",
				"cache":        "healthy",
			},
		}
	case "database":
		return map[string]any{
			"service":    serviceName,
			"status":     "critical",
			"healthy":    false,
			"uptime_sec": 3600 * 24 * 30, // 30 days
			"checks": map[string]any{
				"connectivity": map[string]any{
					"status":        "degraded",
					"response_time": "2500ms",
				},
				"replication": map[string]any{
					"status": "healthy",
					"lag_ms": 50,
				},
			},
			"details": map[string]any{
				"active_connections": 100,
				"max_connections":    100,
				"queued_queries":     45,
				"slow_queries":       32,
			},
		}
	case "cache":
		return map[string]any{
			"service":    serviceName,
			"status":     "healthy",
			"healthy":    true,
			"uptime_sec": 3600 * 24 * 60, // 60 days
			"checks": map[string]any{
				"connectivity": map[string]any{
					"status":        "healthy",
					"response_time": "2ms",
				},
				"memory": map[string]any{
					"status":       "healthy",
					"used_percent": 65.5,
				},
			},
		}
	default:
		return map[string]any{
			"service":    serviceName,
			"status":     "healthy",
			"healthy":    true,
			"uptime_sec": 3600 * 24,
			"checks": map[string]any{
				"http_endpoint": map[string]any{
					"status":        "healthy",
					"response_time": "50ms",
				},
			},
		}
	}
}
