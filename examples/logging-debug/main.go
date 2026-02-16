package main

import (
	"context"
	"fmt"
	"os"

	"github.com/oskarhane/goagent/pkg/agent"
	"github.com/oskarhane/goagent/pkg/logger"
	"github.com/oskarhane/goagent/pkg/providers/openai"
	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/tools/http"
)

func main() {
	// Create logger with Debug level to see all execution details
	log := logger.New(logger.Config{
		Level:   logger.LevelDebug,
		Output:  os.Stderr,
		Enabled: true,
	})

	// Create OpenAI provider with logger
	provider, err := openai.NewProvider(&openai.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Logger: log,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create provider: %v\n", err)
		os.Exit(1)
	}

	// Create tool registry with HTTP tool
	registry := tools.NewRegistry()
	httpTool := http.NewTool()
	httpHandler := http.NewHandler(&http.Config{})
	if err := registry.Register(httpTool, httpHandler); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register HTTP tool: %v\n", err)
		os.Exit(1)
	}

	// Create agent with logger
	temperature := 0.7
	ag, err := agent.NewAgent(&agent.Config{
		Provider:    provider,
		Registry:    registry,
		Temperature: &temperature,
		Logger:      log,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create agent: %v\n", err)
		os.Exit(1)
	}

	// Run agent with debug logging - you'll see:
	// - Agent iteration start/end
	// - Provider API calls with parameters
	// - Tool execution details
	// - Token usage tracking
	// - Execution timing
	log.Info("starting agent with debug logging enabled", nil)

	ctx := context.Background()
	result := ag.Run(ctx, "Fetch the current time from worldtimeapi.org for timezone America/New_York", nil)

	if result.Error != nil {
		log.Error("agent execution failed", map[string]interface{}{
			"error": result.Error.Error(),
		})
		os.Exit(1)
	}

	fmt.Printf("\nFinal Response: %s\n", result.Response.Content)
	fmt.Printf("Iterations: %d\n", result.Iterations)
	fmt.Printf("Total Tokens: %d\n", result.TotalTokens)
	fmt.Printf("Execution Time: %.2fs\n", result.ExecutionTime.Seconds())
}
