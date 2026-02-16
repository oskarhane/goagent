/*
Package types defines the core data structures and interfaces used throughout GoAgent.

# Overview

This package provides:
  - Core message types for conversations (Message, Role)
  - Tool definitions and calling structures (Tool, ToolCall, ToolResult)
  - Provider abstraction interface (Provider)
  - Request/response types for LLM completions
  - Common error types for provider implementations

# Message Structure

Messages represent the conversation flow between users, assistants, and tools.
Each message has a Role (system, user, assistant, or tool) and content.

Example:

	msg := types.NewUserMessage("What is the status of my cluster?")
	systemMsg := types.NewSystemMessage("You are a Kubernetes expert")

# Tool Calling

Tools allow agents to execute functions during their reasoning process.
The tool calling flow is:
 1. Assistant generates ToolCall with function name and arguments
 2. Application executes the tool and creates ToolResult
 3. ToolResult is sent back as a Message with RoleTool

Example:

	tool := types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "get_pods",
			Description: "List pods in a namespace",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"namespace": map[string]any{
						"type":        "string",
						"description": "Kubernetes namespace",
					},
				},
				"required": []string{"namespace"},
			},
		},
	}

# Provider Interface

The Provider interface abstracts different LLM providers (OpenAI, Vertex AI)
into a unified API. All providers must implement:
  - Complete(ctx, req) - Send completion request
  - Name() - Provider identifier
  - DefaultModel() - Default model name

This allows writing provider-agnostic agent code.

# Completion Flow

1. Create CompletionRequest with messages and optional tools
2. Call Provider.Complete(ctx, request)
3. Receive CompletionResponse with generated message
4. If response contains tool calls, execute them and repeat

# Error Handling

Provider implementations should return ProviderError for failures,
which includes:
  - Provider name for identification
  - HTTP status code if applicable
  - Retryable flag for exponential backoff logic
  - Wrapped cause error for debugging

All operations respect context cancellation and timeouts.
*/
package types
