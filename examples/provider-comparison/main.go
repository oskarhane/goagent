package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/oskarhane/goagent/pkg/providers/openai"
	"github.com/oskarhane/goagent/pkg/providers/vertex"
	"github.com/oskarhane/goagent/pkg/types"
)

// demonstrateProvider shows how the same code works with any provider
func demonstrateProvider(provider types.Provider) {
	fmt.Printf("\n=== Using %s Provider ===\n", provider.Name())
	fmt.Printf("Default Model: %s\n\n", provider.DefaultModel())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &types.CompletionRequest{
		Messages: []types.Message{
			types.NewSystemMessage("You are a helpful assistant."),
			types.NewUserMessage("Say 'Hello from " + provider.Name() + "!' in a friendly way."),
		},
		Temperature: 0.7,
		MaxTokens:   50,
	}

	resp, err := provider.Complete(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.Message.Content)
	fmt.Printf("Tokens: %d (prompt: %d, completion: %d)\n",
		resp.Usage.TotalTokens,
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
	)
	fmt.Printf("Finish Reason: %s\n", resp.FinishReason)
}

func main() {
	fmt.Println("GoAgent Provider Comparison Example")
	fmt.Println("====================================")
	fmt.Println("This example demonstrates how both OpenAI and Vertex AI")
	fmt.Println("providers implement the same interface.")
	fmt.Println()

	// Try OpenAI provider
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		provider, err := openai.NewProvider(openai.Config{
			APIKey: apiKey,
		})
		if err != nil {
			log.Fatalf("Failed to create OpenAI provider: %v", err)
		}
		demonstrateProvider(provider)
	} else {
		fmt.Println("=== OpenAI Provider ===")
		fmt.Println("Skipped: OPENAI_API_KEY not set")
	}

	// Try Vertex AI provider
	if projectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); projectID != "" {
		// For this demo, we'll create the provider but won't call it without credentials
		provider, err := vertex.NewProvider(vertex.Config{
			ProjectID: projectID,
			Location:  getEnvOrDefault("GOOGLE_CLOUD_LOCATION", "us-central1"),
		})
		if err != nil {
			// If ADC is not configured, use a dummy client for demonstration
			provider, err = vertex.NewProvider(vertex.Config{
				ProjectID:  projectID,
				HTTPClient: &http.Client{},
			})
			if err != nil {
				log.Fatalf("Failed to create Vertex AI provider: %v", err)
			}
		}
		demonstrateProvider(provider)
	} else {
		fmt.Println("\n=== Vertex AI Provider ===")
		fmt.Println("Skipped: GOOGLE_CLOUD_PROJECT not set")
	}

	fmt.Println("\n✓ Example complete!")
	fmt.Println("\nNote: Both providers implement the types.Provider interface,")
	fmt.Println("allowing you to write provider-agnostic agent code.")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
