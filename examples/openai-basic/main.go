package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/oskarhane/goagent/pkg/providers/openai"
	"github.com/oskarhane/goagent/pkg/types"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI provider
	provider, err := openai.NewProvider(openai.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	fmt.Printf("Created OpenAI provider: %s (default model: %s)\n",
		provider.Name(),
		provider.DefaultModel(),
	)

	// Create a simple completion request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &types.CompletionRequest{
		Messages: []types.Message{
			types.NewSystemMessage("You are a helpful assistant."),
			types.NewUserMessage("Say hello!"),
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	fmt.Println("\nSending completion request...")
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		log.Fatalf("Completion failed: %v", err)
	}

	fmt.Printf("\nResponse:\n")
	fmt.Printf("  ID: %s\n", resp.ID)
	fmt.Printf("  Model: %s\n", resp.Model)
	fmt.Printf("  Message: %s\n", resp.Message.Content)
	fmt.Printf("  Finish Reason: %s\n", resp.FinishReason)
	fmt.Printf("  Tokens: %d prompt + %d completion = %d total\n",
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
		resp.Usage.TotalTokens,
	)
}
