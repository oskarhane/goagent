// Package agent provides the core agent execution loop that combines
// LLM reasoning with tool execution to solve complex tasks.
package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
)

const (
	// DefaultMaxIterations is the default maximum number of reasoning loops.
	DefaultMaxIterations = 10

	// DefaultTemperature is the default LLM temperature setting.
	DefaultTemperature = 0.7
)

// Agent represents an AI agent that can reason and execute tools.
// It combines an LLM provider with a tool registry to solve tasks
// through iterative reasoning and action cycles.
type Agent struct {
	provider      types.Provider
	registry      *tools.Registry
	maxIterations int
	temperature   float64
	systemPrompt  string
}

// Config contains configuration for creating an agent.
type Config struct {
	// Provider is the LLM provider for reasoning.
	Provider types.Provider

	// Registry contains the tools available to the agent.
	Registry *tools.Registry

	// MaxIterations limits the number of reasoning cycles. Defaults to DefaultMaxIterations.
	MaxIterations int

	// Temperature controls LLM randomness. Defaults to DefaultTemperature.
	// Use pointer to distinguish between unset (nil) and explicit zero.
	Temperature *float64

	// SystemPrompt is the system-level instruction for the agent.
	// If empty, a default prompt is used.
	SystemPrompt string
}

// NewAgent creates a new agent with the given configuration.
func NewAgent(cfg Config) (*Agent, error) {
	if cfg.Provider == nil {
		return nil, fmt.Errorf("agent: provider is required")
	}
	if cfg.Registry == nil {
		return nil, fmt.Errorf("agent: tool registry is required")
	}

	maxIterations := cfg.MaxIterations
	if maxIterations == 0 {
		maxIterations = DefaultMaxIterations
	}

	temperature := DefaultTemperature
	if cfg.Temperature != nil {
		temperature = *cfg.Temperature
	}

	systemPrompt := cfg.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant with access to tools. " +
			"Use the available tools to complete tasks effectively. " +
			"Provide clear, concise responses."
	}

	return &Agent{
		provider:      cfg.Provider,
		registry:      cfg.Registry,
		maxIterations: maxIterations,
		temperature:   temperature,
		systemPrompt:  systemPrompt,
	}, nil
}

// RunOptions contains options for running the agent.
type RunOptions struct {
	// History provides conversation context from previous interactions.
	// If provided, these messages are prepended to the current request.
	History []types.Message

	// Model overrides the provider's default model.
	Model string

	// MaxTokens limits the response length.
	MaxTokens int
}

// RunResult contains the outcome of an agent execution.
type RunResult struct {
	// Response is the final message from the agent.
	Response types.Message

	// Messages contains the full conversation including tool calls and results.
	Messages []types.Message

	// Iterations is the number of reasoning cycles executed.
	Iterations int

	// TotalTokens is the cumulative token usage across all iterations.
	TotalTokens int

	// ExecutionTime is the total time spent executing the agent.
	ExecutionTime time.Duration

	// Error contains any error that stopped execution.
	Error error
}

// Run executes the agent with the given prompt and options.
// It performs reasoning → tool execution cycles until the task is complete
// or limits are reached.
func (a *Agent) Run(ctx context.Context, prompt string, opts *RunOptions) *RunResult {
	start := time.Now()
	result := &RunResult{
		Messages: []types.Message{},
	}

	// Initialize options
	if opts == nil {
		opts = &RunOptions{}
	}

	// Build message history
	messages := []types.Message{types.NewSystemMessage(a.systemPrompt)}
	if len(opts.History) > 0 {
		messages = append(messages, opts.History...)
	}
	messages = append(messages, types.NewUserMessage(prompt))

	// Execute reasoning loop
	for iteration := 0; iteration < a.maxIterations; iteration++ {
		result.Iterations = iteration + 1

		// Check context cancellation
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
			result.ExecutionTime = time.Since(start)
			return result
		default:
		}

		// Prepare completion request
		req := &types.CompletionRequest{
			Messages:    messages,
			Tools:       a.registry.List(),
			Temperature: a.temperature,
		}
		if opts.Model != "" {
			req.Model = opts.Model
		}
		if opts.MaxTokens > 0 {
			req.MaxTokens = opts.MaxTokens
		}

		// Call LLM provider
		resp, err := a.provider.Complete(ctx, req)
		if err != nil {
			result.Error = fmt.Errorf("provider error at iteration %d: %w", iteration+1, err)
			result.ExecutionTime = time.Since(start)
			return result
		}

		// Track token usage
		result.TotalTokens += resp.Usage.TotalTokens

		// Add assistant response to messages
		messages = append(messages, resp.Message)

		// Check if we're done (no tool calls)
		if !resp.Message.HasToolCalls() {
			result.Response = resp.Message
			result.Messages = messages
			result.ExecutionTime = time.Since(start)
			return result
		}

		// Execute tool calls
		toolResults := make([]types.ToolResult, 0, len(resp.Message.ToolCalls))
		for _, call := range resp.Message.ToolCalls {
			// Validate tool parameters
			tool, exists := a.registry.Get(call.Function.Name)
			if !exists {
				toolResults = append(toolResults, types.ToolResult{
					ToolCallID: call.ID,
					ToolName:   call.Function.Name,
					Error:      fmt.Sprintf("tool %q not found", call.Function.Name),
				})
				continue
			}

			if err := tools.ValidateParameters(tool, call); err != nil {
				toolResults = append(toolResults, types.ToolResult{
					ToolCallID: call.ID,
					ToolName:   call.Function.Name,
					Error:      fmt.Sprintf("parameter validation failed: %v", err),
				})
				continue
			}

			// Execute tool
			toolResult := a.registry.Execute(ctx, call)
			toolResults = append(toolResults, toolResult)
		}

		// Convert tool results to messages
		for _, tr := range toolResults {
			content := tr.Content
			if tr.Error != "" {
				content = fmt.Sprintf("Error: %s", tr.Error)
			}
			messages = append(messages, types.NewToolMessage(tr.ToolCallID, tr.ToolName, content))
		}
	}

	// Max iterations reached
	result.Error = fmt.Errorf("maximum iterations (%d) reached without completion", a.maxIterations)
	result.Messages = messages
	result.ExecutionTime = time.Since(start)
	return result
}
