# Vertex AI Provider

This package provides a Google Cloud Vertex AI implementation of the GoAgent Provider interface.

## Authentication

The provider uses Google Cloud Application Default Credentials (ADC). Set up authentication using one of these methods:

### 1. Service Account Key File

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
```

### 2. Workload Identity (when running in GKE)

Configure your Kubernetes service account to use Workload Identity.

### 3. Default Credentials

When running on GCE, Cloud Run, or other Google Cloud services, credentials are automatically available.

## Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/oskarhane/goagent/pkg/providers/vertex"
    "github.com/oskarhane/goagent/pkg/types"
)

func main() {
    // Create provider
    provider, err := vertex.NewProvider(vertex.Config{
        ProjectID: "my-gcp-project",
        Location:  "us-central1",     // Optional, defaults to us-central1
        Model:     "gemini-2.5-pro",  // Optional, defaults to gemini-2.5-pro
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Use provider
    resp, err := provider.Complete(context.Background(), &types.CompletionRequest{
        Messages: []types.Message{
            types.NewUserMessage("What is the weather like?"),
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %s", resp.Message.Content)
}
```

## Features

- **Automatic Retry**: Exponential backoff for transient failures (rate limits, timeouts)
- **Context Support**: Respects context cancellation and timeouts
- **Tool Calling**: Full support for function calling with Gemini models
- **Error Handling**: Detailed error messages with retry information

## Configuration Options

- `ProjectID` (required): Your Google Cloud project ID
- `Location` (optional): GCP region, defaults to "us-central1"
- `Model` (optional): Model to use, defaults to "gemini-2.5-pro"
- `MaxRetries` (optional): Number of retry attempts, defaults to 3
- `Timeout` (optional): Request timeout in seconds, defaults to 60
- `HTTPClient` (optional): Custom HTTP client (useful for testing)

## Supported Models

- `gemini-2.5-pro` (default)
- `gemini-3-flash-preview`
- `gemini-2.5-flash`
- `gemini-flash-latest`
- Other Gemini models available in your region

## API Compatibility

This provider implements the same `types.Provider` interface as the OpenAI provider, allowing you to switch between providers without changing your agent code.
