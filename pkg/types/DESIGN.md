# Types Package Design

## Overview

This document explains how the types package supports both OpenAI and Vertex AI API patterns through a unified interface.

## Provider Compatibility

### OpenAI API Pattern

OpenAI's Chat Completions API uses:
- Messages with roles: `system`, `user`, `assistant`, `tool`
- Tool calls with `id`, `type`, and `function` structure
- Function calls with `name` and `arguments` (JSON string)
- Responses include `finish_reason` and `usage` statistics

Our types directly map to this structure:
```go
Message{
    Role: "assistant",
    ToolCalls: []ToolCall{
        {
            ID: "call_abc123",
            Type: "function",
            Function: FunctionCall{
                Name: "get_weather",
                Arguments: `{"location":"San Francisco"}`,
            },
        },
    },
}
```

### Vertex AI API Pattern

Vertex AI (Gemini) uses:
- Messages with roles: `user`, `model` (maps to assistant)
- Function calling with similar structure to OpenAI
- Different field names but same concepts

Our abstraction handles this by:
1. Provider implementations translate between our types and their API
2. `Role` type supports all needed roles
3. Tool/function structures are compatible with both APIs
4. Response types include all metadata both providers supply

## Key Design Decisions

### 1. String-based Arguments

Tool call arguments are stored as JSON strings (`string`) rather than `map[string]any` to match OpenAI's API exactly. This avoids double-serialization and makes the types work with OpenAI's client libraries directly.

Utility function `ParseToolArguments()` provides type-safe unmarshaling.

### 2. Unified Message Structure

Both providers use message-based conversation history, though with different role names. Our `Message` type includes all fields needed by both:
- `Content` - text content
- `ToolCalls` - assistant tool invocations
- `ToolCallID` - link tool results to calls
- `Name` - identify tools in results

### 3. Provider Interface Abstraction

The `Provider` interface doesn't expose provider-specific details. Each implementation handles:
- Authentication (API keys for OpenAI, service accounts for Vertex)
- Request/response translation between our types and their API
- Rate limiting and retries
- Error handling with `ProviderError`

### 4. Extensible Tool Definitions

Tools use JSON Schema for parameters, which both OpenAI and Vertex AI support:
```go
Tool{
    Type: "function",
    Function: FunctionDefinition{
        Name: "search",
        Description: "Search for information",
        Parameters: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "query": map[string]any{
                    "type": "string",
                    "description": "Search query",
                },
            },
            "required": []string{"query"},
        },
    },
}
```

This structure works with both providers' function calling systems.

## Verification Checklist

- [x] Message roles support both OpenAI (system/user/assistant/tool) and Vertex AI (user/model)
- [x] ToolCall structure matches OpenAI's format exactly
- [x] Tool definitions use JSON Schema compatible with both providers
- [x] CompletionRequest includes all parameters needed by both providers
- [x] CompletionResponse captures finish_reason and usage from both providers
- [x] Provider interface abstracts authentication differences
- [x] ProviderError supports retryable error handling for both
- [x] All types include proper JSON tags for serialization

## Future Considerations

If adding more providers (Anthropic, Cohere, etc.):
1. Ensure their message/role concepts map to our structure
2. Add any missing fields to types with `omitempty` JSON tags
3. Update provider implementations to handle translation
4. Keep the unified interface - no provider-specific types in agent code

The current design is flexible enough to support additional providers without breaking changes.
