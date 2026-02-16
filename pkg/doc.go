/*
Package goagent provides a lightweight Go library for building AI agents
that can interact with cloud infrastructure, execute tools, and automate
monitoring and incident response tasks.

# Quick Start

The simplest way to get started is to create an agent with a provider:

	provider := openai.New("your-api-key")
	a := agent.New(provider).
		WithSystemPrompt("You are a helpful cloud monitoring assistant")

	result, err := a.Run(context.Background(), "Check system status")
	if err != nil {
		log.Fatal(err)
	}

# Architecture

GoAgent is built around several core concepts:

  - Providers: Abstraction layer for different LLM providers (OpenAI, Vertex AI)
  - Tools: Reusable functions that agents can call during execution
  - Agent: Core orchestration that handles the reasoning loop between LLM and tools

# Providers

Currently supported providers:

  - OpenAI: Complete API integration with tool calling support
  - Vertex AI: Google Cloud's AI platform integration

# Built-in Tools

GoAgent includes several built-in tools for common cloud operations:

  - HTTP Tool: Make REST API requests with authentication and validation
  - Shell Tool: Execute shell commands with safety constraints
  - Kubernetes Tool: Query cluster resources with RBAC integration

# Custom Tools

You can easily create custom tools by implementing the Tool interface:

	type Tool interface {
		Name() string
		Description() string
		Parameters() *jsonschema.Schema
		Execute(ctx context.Context, params map[string]any) (*ToolResult, error)
	}

# Configuration

Agents support various configuration options:

  - WithSystemPrompt: Set the system prompt for the agent
  - WithTools: Add custom tools to the agent
  - WithTimeout: Set execution timeout
  - WithMaxIterations: Limit reasoning loops
  - WithHistory: Include conversation history

# Error Handling

All operations respect Go context for cancellation and timeout handling.
Errors are wrapped with context to help with debugging and monitoring.

# Observability

GoAgent includes built-in structured logging and optional OpenTelemetry
integration for production deployments.

For more detailed examples and documentation, see the examples/ directory
and the project README.
*/
package goagent
