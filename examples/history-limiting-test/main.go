package main

import (
	"fmt"

	"github.com/oskarhane/goagent/pkg/types"
)

func main() {
	fmt.Println("=== History Limiting Logic Test ===")
	fmt.Println()

	// Create a sample history with 15 messages
	history := []types.Message{
		types.NewSystemMessage("System prompt"),
		types.NewUserMessage("Message 1"),
		types.NewAssistantMessage("Response 1"),
		types.NewUserMessage("Message 2"),
		types.NewAssistantMessage("Response 2"),
		types.NewUserMessage("Message 3"),
		types.NewAssistantMessage("Response 3"),
		types.NewUserMessage("Message 4"),
		types.NewAssistantMessage("Response 4"),
		types.NewUserMessage("Message 5"),
		types.NewAssistantMessage("Response 5"),
		types.NewUserMessage("Message 6"),
		types.NewAssistantMessage("Response 6"),
		types.NewUserMessage("Message 7"),
		types.NewAssistantMessage("Response 7"),
	}

	fmt.Printf("Total history messages: %d\n", len(history))
	fmt.Println()

	// Test 1: No limit (maxHistoryMessages = 0)
	fmt.Println("Test 1: No limit (maxHistoryMessages = 0)")
	maxHistoryMessages := 0
	limited := history
	if maxHistoryMessages > 0 && len(history) > maxHistoryMessages {
		limited = history[len(history)-maxHistoryMessages:]
	}
	fmt.Printf("Result: %d messages (expected: %d)\n", len(limited), len(history))
	if len(limited) != len(history) {
		fmt.Println("[FAIL] Should keep all messages when limit is 0")
	} else {
		fmt.Println("[PASS]")
	}
	fmt.Println()

	// Test 2: Limit to 5 messages
	fmt.Println("Test 2: Limit to 5 messages")
	maxHistoryMessages = 5
	limited = history
	if maxHistoryMessages > 0 && len(history) > maxHistoryMessages {
		limited = history[len(history)-maxHistoryMessages:]
	}
	fmt.Printf("Result: %d messages (expected: %d)\n", len(limited), maxHistoryMessages)
	if len(limited) != maxHistoryMessages {
		fmt.Println("[FAIL] Should keep exactly 5 messages")
	} else {
		fmt.Println("[PASS]")
		fmt.Println("Kept messages:")
		for i, msg := range limited {
			fmt.Printf("  [%d] %s: %s\n", i, msg.Role, msg.Content)
		}
	}
	fmt.Println()

	// Test 3: Limit to 10 messages
	fmt.Println("Test 3: Limit to 10 messages")
	maxHistoryMessages = 10
	limited = history
	if maxHistoryMessages > 0 && len(history) > maxHistoryMessages {
		limited = history[len(history)-maxHistoryMessages:]
	}
	fmt.Printf("Result: %d messages (expected: %d)\n", len(limited), maxHistoryMessages)
	if len(limited) != maxHistoryMessages {
		fmt.Println("[FAIL] Should keep exactly 10 messages")
	} else {
		fmt.Println("[PASS]")
		fmt.Println("First message in limited history:", limited[0].Content)
		fmt.Println("Last message in limited history:", limited[len(limited)-1].Content)
	}
	fmt.Println()

	// Test 4: Limit larger than history
	fmt.Println("Test 4: Limit larger than history (maxHistoryMessages = 20)")
	maxHistoryMessages = 20
	limited = history
	if maxHistoryMessages > 0 && len(history) > maxHistoryMessages {
		limited = history[len(history)-maxHistoryMessages:]
	}
	fmt.Printf("Result: %d messages (expected: %d)\n", len(limited), len(history))
	if len(limited) != len(history) {
		fmt.Println("[FAIL] Should keep all messages when limit exceeds history size")
	} else {
		fmt.Println("[PASS]")
	}
	fmt.Println()

	fmt.Println("=== All Tests Complete ===")
}
