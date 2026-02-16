// Package agent implements the core agent execution loop that combines
// LLM reasoning with tool execution to autonomously solve complex tasks.
//
// # Overview
//
// An Agent orchestrates an iterative reasoning loop where the LLM analyzes
// the task, decides which tools to call, executes those tools, and uses
// the results to continue reasoning until the task is complete.
//
// The execution flow is:
//  1. Send prompt + available tools to LLM
//  2. LLM responds with either:
//     a) Final answer (done)
//     b) Tool calls to execute
//  3. Execute requested tools
//  4. Feed tool results back to LLM
//  5. Repeat from step 2 until done or max iterations reached
//
// # Safety Features
//
// Agents include several safety mechanisms:
//   - Maximum iteration limit (default 10) prevents infinite loops
//   - Context cancellation for timeout and user abort
//   - Parameter validation before tool execution
//   - Error tracking and graceful degradation
//
// # Basic Usage
//
//	// Create provider and tools
//	provider, _ := openai.NewProvider(&openai.Config{
//		APIKey: apiKey,
//	})
//
//	registry := tools.NewRegistry()
//	// ... register tools ...
//
//	// Create agent
//	agent, _ := agent.NewAgent(&agent.Config{
//		Provider: provider,
//		Registry: registry,
//	})
//
//	// Run agent
//	ctx := context.Background()
//	result := agent.Run(ctx, "What is 25 * 17?", nil)
//	if result.Error != nil {
//		log.Fatal(result.Error)
//	}
//
//	fmt.Println(result.Response.Content)
//	fmt.Printf("Completed in %d iterations using %d tokens\n",
//		result.Iterations, result.TotalTokens)
//
// # Conversation History
//
// Agents support maintaining context across multiple interactions:
//
//	// First interaction
//	result1 := agent.Run(ctx, "Calculate 10 + 5", nil)
//
//	// Continue conversation with history
//	result2 := agent.Run(ctx, "Now multiply that by 3", &agent.RunOptions{
//		History: result1.Messages,
//	})
//
// # History Size Limiting
//
// To prevent context window overflow and control costs, limit history size:
//
//	result := agent.Run(ctx, "Continue our conversation", &agent.RunOptions{
//		History:            previousMessages,
//		MaxHistoryMessages: 20,  // Keep only last 20 messages
//	})
//
// History can be serialized to JSON for storage and retrieval:
//
//	// Serialize history
//	historyJSON, _ := json.Marshal(result.Messages)
//	_ = os.WriteFile("history.json", historyJSON, 0644)
//
//	// Deserialize history
//	var history []types.Message
//	data, _ := os.ReadFile("history.json")
//	_ = json.Unmarshal(data, &history)
//
//	// Use restored history
//	result := agent.Run(ctx, "Resume our chat", &agent.RunOptions{
//		History: history,
//	})
//
// # Custom Configuration
//
// Agents can be customized with various options:
//
//	temp := 0.3
//	agent, _ := agent.NewAgent(&agent.Config{
//		Provider:      provider,
//		Registry:      registry,
//		MaxIterations: 20,     // Allow more complex reasoning
//		Temperature:   &temp,  // More deterministic responses
//		SystemPrompt:  "You are a specialized math tutor...",
//	})
package agent
