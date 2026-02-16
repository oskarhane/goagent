package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/oskarhane/goagent/pkg/agent"
	"github.com/oskarhane/goagent/pkg/providers/openai"
	"github.com/oskarhane/goagent/pkg/tools"
	httpTool "github.com/oskarhane/goagent/pkg/tools/http"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create provider
	provider, err := openai.NewProvider(openai.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Create tool registry with HTTP tool
	registry := tools.NewRegistry()

	// Register the HTTP tool with default configuration
	httpToolDef := httpTool.NewTool()
	httpHandler := httpTool.NewHandler(&httpTool.Config{
		DefaultTimeout:  30 * time.Second,
		MaxResponseSize: 5 * 1024 * 1024, // 5MB
	})
	registry.MustRegister(httpToolDef, httpHandler)

	// Create agent
	a, err := agent.NewAgent(agent.Config{
		Provider: provider,
		Registry: registry,
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Example 1: Simple GET request
	fmt.Println("=== Example 1: GET request to GitHub API ===")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel1()

	result1 := a.Run(ctx1, "Get information about the GitHub user 'octocat' using the GitHub API", nil)
	if result1.Error != nil {
		log.Printf("Agent error: %v", result1.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result1.Iterations)
		fmt.Printf("Final response: %s\n\n", result1.Response.Content)
		fmt.Printf("Total tokens used: %d\n\n", result1.TotalTokens)
	}

	// Example 2: Using the tool to check an API endpoint's health
	fmt.Println("\n=== Example 2: Check API health status ===")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel2()

	result2 := a.Run(ctx2, "Check if httpbin.org is responding by making a GET request to https://httpbin.org/status/200. Tell me if it's healthy.", nil)
	if result2.Error != nil {
		log.Printf("Agent error: %v", result2.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result2.Iterations)
		fmt.Printf("Final response: %s\n\n", result2.Response.Content)
		fmt.Printf("Total tokens used: %d\n\n", result2.TotalTokens)
	}

	// Example 3: POST request with JSON body
	fmt.Println("\n=== Example 3: POST request with JSON body ===")
	ctx3, cancel3 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel3()

	result3 := a.Run(ctx3, `Make a POST request to https://httpbin.org/post with JSON body {"name": "GoAgent", "version": "1.0"} and tell me what response you got back.`, nil)
	if result3.Error != nil {
		log.Printf("Agent error: %v", result3.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result3.Iterations)
		fmt.Printf("Final response: %s\n\n", result3.Response.Content)
		fmt.Printf("Total tokens used: %d\n\n", result3.TotalTokens)
	}

	// Example 4: Testing error handling with a 404 endpoint
	fmt.Println("\n=== Example 4: Error handling for 404 ===")
	ctx4, cancel4 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel4()

	result4 := a.Run(ctx4, "Make a GET request to https://httpbin.org/status/404 and tell me what happened.", nil)
	if result4.Error != nil {
		log.Printf("Agent error: %v", result4.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result4.Iterations)
		fmt.Printf("Final response: %s\n\n", result4.Response.Content)
		fmt.Printf("Total tokens used: %d\n\n", result4.TotalTokens)
	}

	fmt.Println("\n=== All examples completed ===")
}
