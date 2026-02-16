package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

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
	provider, err := openai.NewProvider(&openai.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Define a simple calculator tool
	calcTool := tools.NewBuilder(
		"calculate",
		"Perform basic arithmetic operations",
	).
		StringParam("operation", "The operation to perform: add, subtract, multiply, divide", true).
		NumberParam("a", "First number", true).
		NumberParam("b", "Second number", true).
		Build()

	// Register calculator handler
	calcHandler := func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		// Parse parameters
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

		// Perform calculation
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

		// Return result as JSON
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
	}

	if err := registry.Register(calcTool, calcHandler); err != nil {
		log.Fatalf("Failed to register tool: %v", err)
	}

	fmt.Printf("Registered %d tool(s)\n", registry.Count())

	// Make a request that should trigger tool usage
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &types.CompletionRequest{
		Messages: []types.Message{
			types.NewSystemMessage("You are a helpful math assistant. Use the calculate tool when needed."),
			types.NewUserMessage("What is 15 multiplied by 23?"),
		},
		Tools:       registry.List(),
		Temperature: 0.7,
	}

	fmt.Println("\nSending completion request with tools...")
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		log.Fatalf("Completion failed: %v", err)
	}

	fmt.Printf("\nResponse:\n")
	fmt.Printf("  Finish Reason: %s\n", resp.FinishReason)

	// Check if tools were called
	if resp.Message.HasToolCalls() {
		fmt.Printf("  Tool Calls: %d\n", len(resp.Message.ToolCalls))

		for i, call := range resp.Message.ToolCalls {
			fmt.Printf("\n  Tool Call %d:\n", i+1)
			fmt.Printf("    ID: %s\n", call.ID)
			fmt.Printf("    Function: %s\n", call.Function.Name)
			fmt.Printf("    Arguments: %s\n", call.Function.Arguments)

			// Validate parameters
			tool, _ := registry.Get(call.Function.Name)
			if err := tools.ValidateParameters(tool, call); err != nil {
				fmt.Printf("    Validation Error: %v\n", err)
				continue
			}
			fmt.Printf("    Validation: ✓ passed\n")

			// Execute the tool
			result := registry.Execute(ctx, call)
			fmt.Printf("    Execution Time: %v\n", result.ExecutionTime)
			if result.Error != "" {
				fmt.Printf("    Error: %s\n", result.Error)
			} else {
				fmt.Printf("    Result: %s\n", result.Content)
			}
		}
	} else {
		fmt.Printf("  Message: %s\n", resp.Message.Content)
	}

	fmt.Printf("\n  Tokens: %d prompt + %d completion = %d total\n",
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
		resp.Usage.TotalTokens,
	)
}
