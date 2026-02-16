package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/oskarhane/goagent/pkg/providers/vertex"
	"github.com/oskarhane/goagent/pkg/types"
)

func main() {
	// Get configuration from environment
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT environment variable is required")
	}

	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1" // Default location
	}

	// Create Vertex AI provider
	provider, err := vertex.NewProvider(&vertex.Config{
		ProjectID: projectID,
		Location:  location,
		Model:     "gemini-1.5-pro",
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Create a completion request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &types.CompletionRequest{
		Messages: []types.Message{
			types.NewSystemMessage("You are a helpful assistant that answers questions concisely."),
			types.NewUserMessage("What is the capital of France?"),
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	// Send the request
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		log.Fatalf("Completion failed: %v", err)
	}

	// Print the response
	fmt.Printf("Provider: %s\n", provider.Name())
	fmt.Printf("Model: %s\n", resp.Model)
	fmt.Printf("Response: %s\n", resp.Message.Content)
	fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n",
		resp.Usage.TotalTokens,
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
	)
	fmt.Printf("Finish reason: %s\n", resp.FinishReason)
}
