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
	provider, err := openai.NewProvider(&openai.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Create tool registry with a simple calculator
	registry := tools.NewRegistry()

	calcTool := tools.NewBuilder(
		"calculate",
		"Perform arithmetic operations (add, subtract, multiply, divide)",
	).
		StringParam("operation", "Operation: add, subtract, multiply, or divide", true).
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

	// Create agent
	temperature := 0.7
	agentInstance, err := agent.NewAgent(&agent.Config{
		Provider:      provider,
		Registry:      registry,
		MaxIterations: 10,
		Temperature:   &temperature,
		SystemPrompt:  "You are a math assistant. Use the calculator tool to perform calculations accurately.",
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	ctx := context.Background()

	fmt.Println("=== Conversation History Demo ===")
	fmt.Println()

	// === Example 1: Basic conversation continuity ===
	fmt.Println("Example 1: Basic Conversation Continuity")
	fmt.Println("-----------------------------------------")

	prompt1 := "Calculate 50 + 25"
	fmt.Printf("Turn 1: %s\n", prompt1)
	result1 := agentInstance.Run(ctx, prompt1, nil)
	if result1.Error != nil {
		log.Fatalf("Turn 1 failed: %v", result1.Error)
	}
	fmt.Printf("Response: %s\n", result1.Response.Content)
	fmt.Printf("Messages in history: %d\n\n", len(result1.Messages))

	prompt2 := "Now multiply that result by 4"
	fmt.Printf("Turn 2 (with history): %s\n", prompt2)
	result2 := agentInstance.Run(ctx, prompt2, &agent.RunOptions{
		History: result1.Messages,
	})
	if result2.Error != nil {
		log.Fatalf("Turn 2 failed: %v", result2.Error)
	}
	fmt.Printf("Response: %s\n", result2.Response.Content)
	fmt.Printf("Messages in history: %d\n\n", len(result2.Messages))

	prompt3 := "Finally, subtract 100 from that"
	fmt.Printf("Turn 3 (with history): %s\n", prompt3)
	result3 := agentInstance.Run(ctx, prompt3, &agent.RunOptions{
		History: result2.Messages,
	})
	if result3.Error != nil {
		log.Fatalf("Turn 3 failed: %v", result3.Error)
	}
	fmt.Printf("Response: %s\n", result3.Response.Content)
	fmt.Printf("Messages in history: %d\n\n", len(result3.Messages))

	// === Example 2: History serialization ===
	fmt.Println("\nExample 2: History Serialization/Deserialization")
	fmt.Println("------------------------------------------------")

	// Serialize history to JSON
	historyJSON, err := json.MarshalIndent(result3.Messages, "", "  ")
	if err != nil {
		log.Fatalf("Failed to serialize history: %v", err)
	}

	fmt.Printf("Serialized history to JSON (%d bytes)\n", len(historyJSON))

	// Save to file
	tmpFile := "/tmp/goagent-history.json"
	if err := os.WriteFile(tmpFile, historyJSON, 0644); err != nil {
		log.Fatalf("Failed to write history file: %v", err)
	}
	fmt.Printf("Saved history to: %s\n\n", tmpFile)

	// Deserialize from file
	loadedData, err := os.ReadFile(tmpFile)
	if err != nil {
		log.Fatalf("Failed to read history file: %v", err)
	}

	var loadedHistory []types.Message
	if err := json.Unmarshal(loadedData, &loadedHistory); err != nil {
		log.Fatalf("Failed to deserialize history: %v", err)
	}

	fmt.Printf("Loaded history from file (%d messages)\n", len(loadedHistory))

	// Continue conversation with loaded history
	prompt4 := "What was the final result of all our calculations?"
	fmt.Printf("Turn 4 (with loaded history): %s\n", prompt4)
	result4 := agentInstance.Run(ctx, prompt4, &agent.RunOptions{
		History: loadedHistory,
	})
	if result4.Error != nil {
		log.Fatalf("Turn 4 failed: %v", result4.Error)
	}
	fmt.Printf("Response: %s\n", result4.Response.Content)
	fmt.Printf("Messages in history: %d\n\n", len(result4.Messages))

	// === Example 3: History size limiting ===
	fmt.Println("\nExample 3: History Size Limiting")
	fmt.Println("--------------------------------")

	// Build up a long conversation
	longHistory := []types.Message{}
	for i := 1; i <= 5; i++ {
		prompt := fmt.Sprintf("Calculate %d times %d", i, i+1)
		fmt.Printf("Turn %d: %s\n", i, prompt)

		result := agentInstance.Run(ctx, prompt, &agent.RunOptions{
			History: longHistory,
		})
		if result.Error != nil {
			log.Fatalf("Turn %d failed: %v", i, result.Error)
		}

		fmt.Printf("Response: %s (history size: %d messages)\n", result.Response.Content, len(result.Messages))
		longHistory = result.Messages
	}

	fmt.Printf("\nTotal messages in history: %d\n", len(longHistory))

	// Now use history limiting
	maxHistoryMessages := 10
	fmt.Printf("\nApplying MaxHistoryMessages=%d limit...\n", maxHistoryMessages)

	prompt5 := "What was the first calculation we did?"
	fmt.Printf("Turn 6 (with limited history): %s\n", prompt5)
	result5 := agentInstance.Run(ctx, prompt5, &agent.RunOptions{
		History:            longHistory,
		MaxHistoryMessages: maxHistoryMessages,
	})
	if result5.Error != nil {
		log.Fatalf("Turn 6 failed: %v", result5.Error)
	}
	fmt.Printf("Response: %s\n", result5.Response.Content)
	fmt.Printf("(Agent only sees last %d messages, may not remember earliest turns)\n\n", maxHistoryMessages)

	// === Example 4: Token tracking with history ===
	fmt.Println("\nExample 4: Token Usage Tracking")
	fmt.Println("-------------------------------")

	fmt.Printf("Turn 1 tokens: %d\n", result1.TotalTokens)
	fmt.Printf("Turn 2 tokens: %d (cumulative with history)\n", result2.TotalTokens)
	fmt.Printf("Turn 3 tokens: %d (cumulative with history)\n", result3.TotalTokens)
	fmt.Printf("Turn 4 tokens: %d (loaded from file)\n", result4.TotalTokens)
	fmt.Printf("Turn 5 tokens: %d (with limited history)\n", result5.TotalTokens)

	totalTokens := result1.TotalTokens + result2.TotalTokens + result3.TotalTokens + result4.TotalTokens + result5.TotalTokens
	fmt.Printf("\nTotal tokens across all turns: %d\n", totalTokens)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Printf("History file saved at: %s\n", tmpFile)
	fmt.Println("You can inspect the JSON to see the full conversation structure.")
}
