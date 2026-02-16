# Conversation History Example

This example demonstrates how to maintain conversation context across multiple agent interactions using the conversation history feature.

## Features Demonstrated

1. **Basic Conversation Continuity**: Chain multiple prompts together by passing previous messages as history
2. **History Serialization**: Save and load conversation history to/from JSON files
3. **History Size Limiting**: Control context window size with `MaxHistoryMessages`
4. **Token Tracking**: Monitor token usage across multi-turn conversations

## Running the Example

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="your-api-key-here"

# Run the example
go run examples/conversation-history/main.go
```

## Key Concepts

### Passing History

Use `RunOptions.History` to provide conversation context:

```go
result1 := agent.Run(ctx, "Calculate 10 + 5", nil)

// Continue with history
result2 := agent.Run(ctx, "Now multiply that by 3", &agent.RunOptions{
    History: result1.Messages,
})
```

### History Serialization

Messages can be serialized to JSON for persistence:

```go
// Save to file
historyJSON, _ := json.Marshal(result.Messages)
os.WriteFile("history.json", historyJSON, 0644)

// Load from file
var history []types.Message
data, _ := os.ReadFile("history.json")
json.Unmarshal(data, &history)

// Use loaded history
result := agent.Run(ctx, "Continue", &agent.RunOptions{
    History: history,
})
```

### History Size Limiting

Prevent context window overflow by limiting history:

```go
result := agent.Run(ctx, "Continue conversation", &agent.RunOptions{
    History:            previousMessages,
    MaxHistoryMessages: 20,  // Keep only last 20 messages
})
```

## Additional Test Examples

### Serialization Test

Verify JSON serialization without API calls:

```bash
go run examples/history-serialization-test/main.go
```

### Limiting Logic Test

Verify history limiting logic:

```bash
go run examples/history-limiting-test/main.go
```

## Use Cases

- **Multi-turn conversations**: Build chatbots that remember context
- **Session persistence**: Save conversations and resume later
- **Context management**: Control token usage with history limits
- **Debugging**: Inspect full conversation history in JSON format
