// Package openai provides an OpenAI API client that implements the types.Provider interface.
//
// The provider supports:
//   - Authentication via API key (from environment variables)
//   - Automatic retry with exponential backoff for transient failures
//   - Context-based cancellation and timeout handling
//   - Tool calling (function calling) capabilities
//   - Rate limit handling with automatic retry
//
// Environment Variable Setup:
//
// The API key is read from environment variables. For .env file support, use a package like
// github.com/joho/godotenv to load your .env file before creating the provider:
//
//	import _ "github.com/joho/godotenv/autoload"
//
// Example usage:
//
//	provider, err := openai.NewProvider(&openai.Config{
//	    APIKey: os.Getenv("OPENAI_API_KEY"),
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	resp, err := provider.Complete(ctx, &types.CompletionRequest{
//	    Model: "gpt-5.1",
//	    Messages: []types.Message{
//	        types.NewUserMessage("What is the weather like?"),
//	    },
//	})
package openai
