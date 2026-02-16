package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/oskarhane/goagent/pkg/agent"
	"github.com/oskarhane/goagent/pkg/logger"
	"github.com/oskarhane/goagent/pkg/providers/openai"
	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
)

func main() {
	// Create logger with Info level
	log := logger.New(logger.Config{
		Level:   logger.LevelInfo,
		Output:  os.Stderr,
		Enabled: true,
	})

	// Create OpenAI provider with logger
	provider, err := openai.NewProvider(&openai.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Logger: log,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create provider: %v\n", err)
		os.Exit(1)
	}

	// Create tool registry and add a simple calculator
	registry := tools.NewRegistry()
	calculator := tools.NewBuilder("calculator", "Performs basic arithmetic operations").
		StringParam("operation", "Operation to perform: add, subtract, multiply, divide", true).
		NumberParam("a", "First number", true).
		NumberParam("b", "Second number", true).
		Build()

	registry.Register(calculator, func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		// Parse tool arguments
		var args map[string]interface{}
		if err := types.ParseToolArguments(call, &args); err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         err.Error(),
				ExecutionTime: time.Since(start),
			}
		}

		op := args["operation"].(string)
		a := args["a"].(float64)
		b := args["b"].(float64)

		var result float64
		switch op {
		case "add":
			result = a + b
		case "subtract":
			result = a - b
		case "multiply":
			result = a * b
		case "divide":
			if b == 0 {
				return types.ToolResult{
					ToolCallID:    call.ID,
					ToolName:      call.Function.Name,
					Error:         "division by zero",
					ExecutionTime: time.Since(start),
				}
			}
			result = a / b
		default:
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("unknown operation: %s", op),
				ExecutionTime: time.Since(start),
			}
		}

		return types.ToolResult{
			ToolCallID:    call.ID,
			ToolName:      call.Function.Name,
			Content:       fmt.Sprintf("%f", result),
			ExecutionTime: time.Since(start),
		}
	})

	// Create agent with logger
	temperature := 0.7
	ag, err := agent.NewAgent(&agent.Config{
		Provider:    provider,
		Registry:    registry,
		Temperature: &temperature,
		Logger:      log,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create agent: %v\n", err)
		os.Exit(1)
	}

	// Run agent with logging
	log.Info("starting agent execution", map[string]interface{}{
		"prompt": "What is 25 times 4?",
	})

	ctx := context.Background()
	result := ag.Run(ctx, "What is 25 times 4?", nil)

	if result.Error != nil {
		log.Error("agent execution failed", map[string]interface{}{
			"error": result.Error.Error(),
		})
		os.Exit(1)
	}

	log.Info("agent execution completed", map[string]interface{}{
		"iterations":     result.Iterations,
		"total_tokens":   result.TotalTokens,
		"execution_time": result.ExecutionTime.Seconds(),
	})

	fmt.Printf("\nFinal Response: %s\n", result.Response.Content)
}
