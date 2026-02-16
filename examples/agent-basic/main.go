package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/oskarhane/goagent/pkg/agent"
	"github.com/oskarhane/goagent/pkg/providers/openai"
	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create provider
	provider, err := openai.NewProvider(openai.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Register calculator tool
	calcTool := tools.NewBuilder(
		"calculate",
		"Perform basic arithmetic operations (add, subtract, multiply, divide)",
	).
		StringParam("operation", "The operation: add, subtract, multiply, or divide", true).
		NumberParam("a", "First number", true).
		NumberParam("b", "Second number", true).
		Build()

	registry.MustRegister(calcTool, func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		var params struct {
			Operation string  `json:"operation"`
			A         float64 `json:"a"`
			B         float64 `json:"b"`
		}
		if err := types.ParseToolArguments(call, &params); err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         err.Error(),
				ExecutionTime: time.Since(start),
			}
		}

		var result float64
		switch params.Operation {
		case "add":
			result = params.A + params.B
		case "subtract":
			result = params.A - params.B
		case "multiply":
			result = params.A * params.B
		case "divide":
			if params.B == 0 {
				return types.ToolResult{
					ToolCallID:    call.ID,
					ToolName:      call.Function.Name,
					Error:         "division by zero",
					ExecutionTime: time.Since(start),
				}
			}
			result = params.A / params.B
		default:
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("unknown operation: %s", params.Operation),
				ExecutionTime: time.Since(start),
			}
		}

		resultJSON, _ := json.Marshal(map[string]any{
			"operation": params.Operation,
			"a":         params.A,
			"b":         params.B,
			"result":    result,
		})

		return types.ToolResult{
			ToolCallID:    call.ID,
			ToolName:      call.Function.Name,
			Content:       string(resultJSON),
			ExecutionTime: time.Since(start),
		}
	})

	// Register get_time tool
	timeTool := tools.NewBuilder(
		"get_time",
		"Get the current date and time",
	).Build()

	registry.MustRegister(timeTool, func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		now := time.Now()
		resultJSON, _ := json.Marshal(map[string]any{
			"timestamp": now.Unix(),
			"formatted": now.Format(time.RFC3339),
			"timezone":  now.Location().String(),
		})

		return types.ToolResult{
			ToolCallID:    call.ID,
			ToolName:      call.Function.Name,
			Content:       string(resultJSON),
			ExecutionTime: time.Since(start),
		}
	})

	// Create agent
	temperature := 0.7
	agentInstance, err := agent.NewAgent(agent.Config{
		Provider:      provider,
		Registry:      registry,
		MaxIterations: 10,
		Temperature:   &temperature,
		SystemPrompt:  "You are a helpful AI assistant. Use the available tools to complete tasks accurately. Always show your reasoning.",
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	fmt.Println("=== GoAgent Demo ===")
	fmt.Printf("Provider: %s (%s)\n", provider.Name(), provider.DefaultModel())
	fmt.Printf("Available tools: %d\n", registry.Count())
	fmt.Println()

	// Example 1: Simple calculation
	fmt.Println("Example 1: Simple Calculation")
	fmt.Println("Prompt: Calculate 123 * 456 and tell me the result")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result1 := agentInstance.Run(ctx, "Calculate 123 * 456 and tell me the result", nil)
	printResult(result1)

	// Example 2: Multi-step reasoning
	fmt.Println("\nExample 2: Multi-Step Reasoning")
	fmt.Println("Prompt: What's the current time? Then calculate how many seconds are in 24 hours.")
	fmt.Println()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel2()

	result2 := agentInstance.Run(ctx2, "What's the current time? Then calculate how many seconds are in 24 hours.", nil)
	printResult(result2)

	// Example 3: Conversation with history
	fmt.Println("\nExample 3: Conversation with History")
	fmt.Println("Prompt 1: Calculate 100 divided by 4")

	ctx3, cancel3 := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel3()

	result3 := agentInstance.Run(ctx3, "Calculate 100 divided by 4", nil)
	printResult(result3)

	fmt.Println("\nPrompt 2: Now multiply that result by 8")
	ctx4, cancel4 := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel4()

	result4 := agentInstance.Run(ctx4, "Now multiply that result by 8", &agent.RunOptions{
		History: result3.Messages,
	})
	printResult(result4)
}

func printResult(result *agent.RunResult) {
	if result.Error != nil {
		fmt.Printf("❌ Error: %v\n", result.Error)
		return
	}

	fmt.Printf("✅ Success!\n")
	fmt.Printf("Response: %s\n", result.Response.Content)
	fmt.Printf("\nMetrics:\n")
	fmt.Printf("  Iterations: %d\n", result.Iterations)
	fmt.Printf("  Total Tokens: %d\n", result.TotalTokens)
	fmt.Printf("  Execution Time: %v\n", result.ExecutionTime)
	fmt.Printf("  Message Count: %d\n", len(result.Messages))
}
