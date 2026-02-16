// Package types defines core data structures used throughout the GoAgent SDK.
package types

import (
	"encoding/json"
	"time"
)

// Role represents the role of a message sender in a conversation.
type Role string

const (
	// RoleSystem represents system-level instructions and prompts.
	RoleSystem Role = "system"
	// RoleUser represents messages from the user.
	RoleUser Role = "user"
	// RoleAssistant represents messages from the AI assistant.
	RoleAssistant Role = "assistant"
	// RoleTool represents the result of a tool execution.
	RoleTool Role = "tool"
)

// Message represents a single message in a conversation between user, assistant, and tools.
// It supports both simple text messages and complex tool call interactions.
type Message struct {
	// Role identifies the sender of the message (system, user, assistant, tool).
	Role Role `json:"role"`

	// Content is the text content of the message. May be empty if ToolCalls is present.
	Content string `json:"content,omitempty"`

	// ToolCalls contains tool invocations made by the assistant.
	// Only present in assistant messages when tools are being called.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID links tool result messages back to their corresponding tool call.
	// Only present in tool role messages.
	ToolCallID string `json:"tool_call_id,omitempty"`

	// Name optionally identifies the tool that produced this message.
	// Used primarily with RoleTool messages.
	Name string `json:"name,omitempty"`
}

// ToolCall represents a request from the assistant to execute a tool.
// The structure is compatible with both OpenAI and Vertex AI tool calling patterns.
type ToolCall struct {
	// ID is a unique identifier for this tool call, used to match results.
	ID string `json:"id"`

	// Type indicates the type of tool call. Currently only "function" is supported.
	Type string `json:"type"`

	// Function contains the function name and arguments.
	Function FunctionCall `json:"function"`
}

// FunctionCall represents the details of a function/tool being called.
type FunctionCall struct {
	// Name is the name of the function to call.
	Name string `json:"name"`

	// Arguments contains the JSON-encoded parameters for the function.
	// This is a string rather than map to match OpenAI's API format.
	Arguments string `json:"arguments"`
}

// Tool represents a function that can be called by the agent during execution.
// Tools are defined with JSON Schema for parameter validation.
type Tool struct {
	// Function contains the function definition including name, description, and parameters.
	Function FunctionDefinition `json:"function"`

	// Type indicates the tool type. Currently only "function" is supported.
	Type string `json:"type"`
}

// FunctionDefinition describes a tool function that can be called by the agent.
type FunctionDefinition struct {
	// Parameters defines the expected parameters using JSON Schema.
	// Should be a valid JSON Schema object with "type": "object" and "properties".
	Parameters map[string]any `json:"parameters"`

	// Name is the unique identifier for this function.
	Name string `json:"name"`

	// Description explains what the function does. Used by the LLM to decide when to call it.
	Description string `json:"description"`
}

// CompletionRequest represents a request to an LLM provider for text generation.
// It abstracts the differences between provider APIs into a unified interface.
type CompletionRequest struct {
	// Model specifies which model to use (e.g., "gpt-4", "gemini-pro").
	Model string `json:"model"`

	// Messages contains the conversation history and current prompt.
	Messages []Message `json:"messages"`

	// Tools defines functions the model can choose to call.
	Tools []Tool `json:"tools,omitempty"`

	// Temperature controls randomness in the response (0.0 to 2.0).
	// Lower values make output more focused and deterministic.
	Temperature float64 `json:"temperature,omitempty"`

	// MaxTokens limits the length of the generated response.
	MaxTokens int `json:"max_tokens,omitempty"`

	// TopP controls diversity via nucleus sampling (0.0 to 1.0).
	// Alternative to temperature for controlling randomness.
	TopP float64 `json:"top_p,omitempty"`
}

// CompletionResponse represents the response from an LLM provider.
// It includes the generated message and metadata about the request.
type CompletionResponse struct {
	// Message contains the generated content and any tool calls.
	Message Message `json:"message"`

	// Usage contains token usage statistics for this request.
	Usage Usage `json:"usage"`

	// ID is a unique identifier for this completion.
	ID string `json:"id"`

	// Model indicates which model was used to generate this response.
	Model string `json:"model"`

	// FinishReason indicates why the model stopped generating.
	// Common values: "stop" (natural completion), "tool_calls" (called a tool),
	// "length" (hit max tokens), "content_filter" (filtered by safety).
	FinishReason string `json:"finish_reason"`

	// Created timestamp when the completion was generated (Unix timestamp).
	Created int64 `json:"created"`
}

// Usage tracks token consumption for a completion request.
// Useful for monitoring costs and staying within rate limits.
type Usage struct {
	// PromptTokens is the number of tokens in the input.
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens is the number of tokens in the generated output.
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens is the sum of prompt and completion tokens.
	TotalTokens int `json:"total_tokens"`
}

// ToolResult represents the outcome of executing a tool.
// It captures both successful results and errors in a format the LLM can understand.
type ToolResult struct {
	// ToolCallID matches the ID from the ToolCall that triggered this execution.
	ToolCallID string `json:"tool_call_id"`

	// ToolName is the name of the tool that was executed.
	ToolName string `json:"tool_name"`

	// Content contains the result data as a string.
	// For structured data, this should be JSON-encoded.
	Content string `json:"content"`

	// Error contains error information if the tool execution failed.
	// When present, Content may contain additional error context.
	Error string `json:"error,omitempty"`

	// ExecutionTime tracks how long the tool took to execute.
	ExecutionTime time.Duration `json:"execution_time"`
}

// ParseToolArguments unmarshals tool call arguments from JSON string into a target struct.
// This is a helper for implementing tool handlers that need typed parameters.
func ParseToolArguments(call ToolCall, target any) error {
	return json.Unmarshal([]byte(call.Function.Arguments), target)
}

// NewSystemMessage creates a new message with the system role.
func NewSystemMessage(content string) Message {
	return Message{
		Role:    RoleSystem,
		Content: content,
	}
}

// NewUserMessage creates a new message with the user role.
func NewUserMessage(content string) Message {
	return Message{
		Role:    RoleUser,
		Content: content,
	}
}

// NewAssistantMessage creates a new message with the assistant role.
func NewAssistantMessage(content string) Message {
	return Message{
		Role:    RoleAssistant,
		Content: content,
	}
}

// NewToolMessage creates a new message with the tool role.
func NewToolMessage(toolCallID, toolName, content string) Message {
	return Message{
		Role:       RoleTool,
		Content:    content,
		ToolCallID: toolCallID,
		Name:       toolName,
	}
}

// HasToolCalls returns true if the message contains any tool calls.
func (m *Message) HasToolCalls() bool {
	return len(m.ToolCalls) > 0
}

// IsToolResult returns true if the message is a tool execution result.
func (m *Message) IsToolResult() bool {
	return m.Role == RoleTool && m.ToolCallID != ""
}
