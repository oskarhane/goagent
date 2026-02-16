// Package logger provides structured logging and optional OpenTelemetry tracing
// for monitoring and debugging agent execution.
//
// # Basic Usage
//
// Create a logger with default settings:
//
//	logger := logger.Default()
//	logger.Info("agent started", map[string]interface{}{
//		"provider": "openai",
//		"model": "gpt-5.1",
//	})
//
// # Log Levels
//
// Four log levels are supported (from least to most severe):
//   - Debug: Detailed debugging information
//   - Info: General informational messages
//   - Warn: Warning messages
//   - Error: Error messages
//
// Configure the minimum level to control output:
//
//	logger := logger.New(logger.Config{
//		Level: logger.LevelDebug,
//		Output: os.Stdout,
//		Enabled: true,
//	})
//
// # Structured Output
//
// All log entries are JSON-formatted with timestamp, level, message, and optional fields:
//
//	{"timestamp":"2026-02-16T13:45:30.123Z","level":"INFO","message":"agent started","fields":{"provider":"openai"}}
//
// # OpenTelemetry Tracing
//
// Enable distributed tracing by providing a tracer name:
//
//	logger := logger.New(logger.Config{
//		Level: logger.LevelInfo,
//		Enabled: true,
//		TracerName: "goagent",
//	})
//
//	ctx, span := logger.StartSpan(ctx, "agent.run",
//		attribute.String("provider", "openai"),
//		attribute.Int("iteration", 1),
//	)
//	defer logger.EndSpan(span, err)
//
// # Integration
//
// Pass logger to Agent and Provider configurations:
//
//	agent, err := agent.NewAgent(&agent.Config{
//		Provider: provider,
//		Registry: registry,
//		Logger: logger,
//	})
//
// The logger will automatically capture:
//   - Agent execution lifecycle (start, iterations, completion)
//   - Provider API calls (requests, responses, retries, errors)
//   - Tool execution (calls, results, errors)
//   - Execution context (token usage, timing, iterations)
package logger
