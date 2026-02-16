package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
)

func main() {
	fmt.Println("Tool Parameter Validation Examples")
	fmt.Println("===================================")
	fmt.Println()

	// Create a comprehensive tool with various parameter types
	complexTool := tools.NewBuilder(
		"create_user",
		"Create a new user account",
	).
		StringParam("username", "Username (3-20 chars)", true).
		StringParam("email", "Email address", true).
		IntegerParam("age", "User age", true).
		BooleanParam("verified", "Is user verified", false).
		ArrayOfStrings("tags", "User tags", false).
		ObjectParam("address", "User address", false,
			map[string]any{
				"street": map[string]any{
					"type":        "string",
					"description": "Street address",
				},
				"city": map[string]any{
					"type":        "string",
					"description": "City name",
				},
				"zipcode": map[string]any{
					"type":        "string",
					"description": "ZIP code",
				},
			},
			[]string{"street", "city"},
		).
		Build()

	// Test cases
	testCases := []struct {
		name      string
		toolCall  types.ToolCall
		shouldErr bool
	}{
		{
			name: "Valid parameters",
			toolCall: types.ToolCall{
				ID:   "call_1",
				Type: "function",
				Function: types.FunctionCall{
					Name: "create_user",
					Arguments: `{
						"username": "johndoe",
						"email": "john@example.com",
						"age": 25,
						"verified": true,
						"tags": ["developer", "admin"],
						"address": {
							"street": "123 Main St",
							"city": "San Francisco",
							"zipcode": "94102"
						}
					}`,
				},
			},
			shouldErr: false,
		},
		{
			name: "Missing required field",
			toolCall: types.ToolCall{
				ID:   "call_2",
				Type: "function",
				Function: types.FunctionCall{
					Name: "create_user",
					Arguments: `{
						"username": "johndoe",
						"age": 25
					}`,
				},
			},
			shouldErr: true,
		},
		{
			name: "Wrong type for age (string instead of integer)",
			toolCall: types.ToolCall{
				ID:   "call_3",
				Type: "function",
				Function: types.FunctionCall{
					Name: "create_user",
					Arguments: `{
						"username": "johndoe",
						"email": "john@example.com",
						"age": "25"
					}`,
				},
			},
			shouldErr: true,
		},
		{
			name: "Invalid nested object (missing required field)",
			toolCall: types.ToolCall{
				ID:   "call_4",
				Type: "function",
				Function: types.FunctionCall{
					Name: "create_user",
					Arguments: `{
						"username": "johndoe",
						"email": "john@example.com",
						"age": 25,
						"address": {
							"street": "123 Main St"
						}
					}`,
				},
			},
			shouldErr: true,
		},
		{
			name: "Invalid array type",
			toolCall: types.ToolCall{
				ID:   "call_5",
				Type: "function",
				Function: types.FunctionCall{
					Name: "create_user",
					Arguments: `{
						"username": "johndoe",
						"email": "john@example.com",
						"age": 25,
						"tags": "not-an-array"
					}`,
				},
			},
			shouldErr: true,
		},
		{
			name: "Minimal valid (only required fields)",
			toolCall: types.ToolCall{
				ID:   "call_6",
				Type: "function",
				Function: types.FunctionCall{
					Name: "create_user",
					Arguments: `{
						"username": "johndoe",
						"email": "john@example.com",
						"age": 25
					}`,
				},
			},
			shouldErr: false,
		},
	}

	// Run test cases
	passed := 0
	failed := 0

	for _, tc := range testCases {
		fmt.Printf("Test: %s\n", tc.name)
		fmt.Printf("Arguments: %s\n", tc.toolCall.Function.Arguments)

		err := tools.ValidateParameters(complexTool, tc.toolCall)

		if tc.shouldErr {
			if err != nil {
				fmt.Printf("✓ Expected error: %v\n\n", err)
				passed++
			} else {
				fmt.Printf("✗ Expected error but got none\n\n")
				failed++
			}
		} else {
			if err == nil {
				fmt.Printf("✓ Validation passed\n\n")
				passed++
			} else {
				fmt.Printf("✗ Unexpected error: %v\n\n", err)
				failed++
			}
		}
	}

	fmt.Printf("Results: %d passed, %d failed\n\n", passed, failed)

	// Demonstrate registry usage
	fmt.Println("Registry Demo")
	fmt.Println("=============")
	fmt.Println()

	registry := tools.NewRegistry()

	// Create a handler that validates before execution
	handler := func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		// Get tool from registry for validation
		tool, exists := registry.Get(call.Function.Name)
		if !exists {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         "tool not found",
				ExecutionTime: time.Since(start),
			}
		}

		// Validate parameters
		if err := tools.ValidateParameters(tool, call); err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("validation failed: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		// Parse and process (in real implementation)
		return types.ToolResult{
			ToolCallID:    call.ID,
			ToolName:      call.Function.Name,
			Content:       "User created successfully",
			ExecutionTime: time.Since(start),
		}
	}

	// Register tool
	if err := registry.Register(complexTool, handler); err != nil {
		log.Fatalf("Failed to register tool: %v", err)
	}

	fmt.Printf("Registered tool: %s\n", complexTool.Function.Name)
	fmt.Printf("Total tools in registry: %d\n\n", registry.Count())

	// Execute with valid parameters
	ctx := context.Background()
	result := registry.Execute(ctx, testCases[0].toolCall)
	fmt.Printf("Execution result:\n")
	fmt.Printf("  Tool: %s\n", result.ToolName)
	fmt.Printf("  Content: %s\n", result.Content)
	fmt.Printf("  Error: %s\n", result.Error)
	fmt.Printf("  Execution Time: %v\n\n", result.ExecutionTime)

	// Execute with invalid parameters
	result = registry.Execute(ctx, testCases[1].toolCall)
	fmt.Printf("Execution result (invalid):\n")
	fmt.Printf("  Tool: %s\n", result.ToolName)
	fmt.Printf("  Content: %s\n", result.Content)
	fmt.Printf("  Error: %s\n", result.Error)
	fmt.Printf("  Execution Time: %v\n", result.ExecutionTime)

	if failed == 0 {
		fmt.Println("\n✓ All validation tests passed!")
	}
}
