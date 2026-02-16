// Package vertex provides a Google Cloud Vertex AI client that implements the types.Provider interface.
//
// The provider supports:
//   - Authentication via Google Cloud service accounts
//   - Automatic retry with exponential backoff for transient failures
//   - Context-based cancellation and timeout handling
//   - Tool calling (function calling) capabilities with Gemini models
//   - Rate limit handling with automatic retry
//
// Authentication:
//
// The provider uses Google Cloud Application Default Credentials (ADC). Authentication is configured
// using one of the following methods:
//
//  1. Service Account Key File:
//     export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
//
//  2. Workload Identity (when running in GKE)
//
//  3. Default credentials (when running on GCE, Cloud Run, etc.)
//
// Example usage:
//
//	provider, err := vertex.NewProvider(&vertex.Config{
//	    ProjectID: "my-gcp-project",
//	    Location:  "us-central1",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	resp, err := provider.Complete(ctx, &types.CompletionRequest{
//	    Model: "gemini-2.5-pro",
//	    Messages: []types.Message{
//	        types.NewUserMessage("What is the weather like?"),
//	    },
//	})
package vertex
