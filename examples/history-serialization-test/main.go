package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/oskarhane/goagent/pkg/types"
)

func main() {
	fmt.Println("=== Message Serialization Test ===")
	fmt.Println()

	// Create a sample conversation with various message types
	messages := []types.Message{
		types.NewSystemMessage("You are a helpful assistant."),
		types.NewUserMessage("What is 2 + 2?"),
		{
			Role:    types.RoleAssistant,
			Content: "Let me calculate that for you.",
			ToolCalls: []types.ToolCall{
				{
					ID:   "call_123",
					Type: "function",
					Function: types.FunctionCall{
						Name:      "calculate",
						Arguments: `{"operation":"add","a":2,"b":2}`,
					},
				},
			},
		},
		types.NewToolMessage("call_123", "calculate", `{"result":4}`),
		types.NewAssistantMessage("The result of 2 + 2 is 4."),
	}

	fmt.Printf("Original messages: %d\n", len(messages))
	fmt.Println()

	// Print message details
	for i, msg := range messages {
		fmt.Printf("Message %d:\n", i+1)
		fmt.Printf("  Role: %s\n", msg.Role)
		fmt.Printf("  Content: %s\n", msg.Content)
		if len(msg.ToolCalls) > 0 {
			fmt.Printf("  ToolCalls: %d\n", len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				fmt.Printf("    [%d] ID=%s, Name=%s\n", j, tc.ID, tc.Function.Name)
			}
		}
		if msg.ToolCallID != "" {
			fmt.Printf("  ToolCallID: %s\n", msg.ToolCallID)
		}
		if msg.Name != "" {
			fmt.Printf("  Name: %s\n", msg.Name)
		}
		fmt.Println()
	}

	// Serialize to JSON
	fmt.Println("Serializing to JSON...")
	jsonData, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		log.Fatalf("Serialization failed: %v", err)
	}

	fmt.Printf("JSON output (%d bytes):\n", len(jsonData))
	fmt.Println(string(jsonData))
	fmt.Println()

	// Deserialize from JSON
	fmt.Println("Deserializing from JSON...")
	var deserialized []types.Message
	if err := json.Unmarshal(jsonData, &deserialized); err != nil {
		log.Fatalf("Deserialization failed: %v", err)
	}

	fmt.Printf("Deserialized messages: %d\n", len(deserialized))
	fmt.Println()

	// Verify integrity
	fmt.Println("Verifying data integrity...")
	if len(messages) != len(deserialized) {
		log.Fatalf("Message count mismatch: %d vs %d", len(messages), len(deserialized))
	}

	allMatch := true
	for i := range messages {
		orig := messages[i]
		deser := deserialized[i]

		if orig.Role != deser.Role {
			fmt.Printf("[FAIL] Message %d: Role mismatch (%s vs %s)\n", i+1, orig.Role, deser.Role)
			allMatch = false
		}
		if orig.Content != deser.Content {
			fmt.Printf("[FAIL] Message %d: Content mismatch\n", i+1)
			allMatch = false
		}
		if len(orig.ToolCalls) != len(deser.ToolCalls) {
			fmt.Printf("[FAIL] Message %d: ToolCalls count mismatch (%d vs %d)\n", i+1, len(orig.ToolCalls), len(deser.ToolCalls))
			allMatch = false
		} else {
			for j := range orig.ToolCalls {
				if orig.ToolCalls[j].ID != deser.ToolCalls[j].ID {
					fmt.Printf("[FAIL] Message %d, ToolCall %d: ID mismatch\n", i+1, j)
					allMatch = false
				}
				if orig.ToolCalls[j].Function.Name != deser.ToolCalls[j].Function.Name {
					fmt.Printf("[FAIL] Message %d, ToolCall %d: Function name mismatch\n", i+1, j)
					allMatch = false
				}
			}
		}
		if orig.ToolCallID != deser.ToolCallID {
			fmt.Printf("[FAIL] Message %d: ToolCallID mismatch (%s vs %s)\n", i+1, orig.ToolCallID, deser.ToolCallID)
			allMatch = false
		}
		if orig.Name != deser.Name {
			fmt.Printf("[FAIL] Message %d: Name mismatch (%s vs %s)\n", i+1, orig.Name, deser.Name)
			allMatch = false
		}
	}

	if allMatch {
		fmt.Println("[PASS] All messages match! Serialization/deserialization working correctly.")
	} else {
		log.Fatal("[FAIL] Data integrity check failed!")
	}

	fmt.Println()
	fmt.Println("=== Test Complete ===")
}
