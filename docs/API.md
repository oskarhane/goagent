# GoAgent SDK API Documentation

Complete API reference for the GoAgent SDK.

## Table of Contents

- [Agent](#agent)
- [Providers](#providers)
  - [OpenAI](#openai-provider)
  - [Vertex AI](#vertex-ai-provider)
- [Tools](#tools)
  - [HTTP Tool](#http-tool)
  - [Shell Tool](#shell-tool)
  - [Kubernetes Tool](#kubernetes-tool)
  - [Custom Tools](#custom-tools)
- [Types](#types)
- [Logger](#logger)

## Agent

The Agent orchestrates reasoning and tool execution loops.

### Creating an Agent

```go
import "github.com/oskarhane/goagent/pkg/agent"

cfg := &agent.Config{
    Provider:      provider,       // Required: LLM provider
    SystemPrompt:  "You are...",   // Optional: System instructions
    Registry:      registry,       // Required: Tool registry
    MaxIterations: 10,            // Optional: Safety limit (default: 10)
    Temperature:   &temp,         // Optional: 0.0-1.0 (default: 0.7)
    Logger:        logger,        // Optional: Structured logger
}

a, err := agent.NewAgent(cfg)
```

### Running an Agent

```go
// Simple run
result := a.Run(ctx, "Your task here", nil)

// With options
result := a.Run(ctx, "Your task", &agent.RunOptions{
    Model:             "gpt-5",          // Override provider model
    MaxTokens:         2000,             // Limit response tokens
    History:           previousMessages, // Conversation context
    MaxHistoryMessages: 10,              // Trim old messages
})
```

### RunResult

The result contains execution details:

```go
type RunResult struct {
    Messages      []types.Message // Full conversation
    Iterations    int            // Number of loops
    TotalTokens   int            // Token usage
    ExecutionTime time.Duration  // Total time
    Error         error          // Execution error
}

// Access final response
for _, msg := range result.Messages {
    if msg.Role == "assistant" && len(msg.ToolCalls) == 0 {
        fmt.Println(msg.Content)
    }
}
```

## Providers

Providers are LLM backends. GoAgent abstracts differences between providers.

### OpenAI Provider

```go
import "github.com/oskarhane/goagent/pkg/providers/openai"

provider, err := openai.NewProvider(&openai.Config{
    APIKey:     os.Getenv("OPENAI_API_KEY"), // Required
    Model:      "gpt-5.1",                   // Optional (default)
    MaxRetries: 3,                            // Optional (default: 3)
    BaseURL:    "https://api.openai.com/v1", // Optional
})
```

**Supported Models:**
- `gpt-5.1` (default)
- `gpt-5-mini`
- `gpt-5`

**Environment Variables:**
- `OPENAI_API_KEY` - Your OpenAI API key

**Error Handling:**
- Automatic retry on 429, 500, 502, 503, 504
- Exponential backoff: 1s, 2s, 4s, 8s...
- Context cancellation supported

### Vertex AI Provider

```go
import "github.com/oskarhane/goagent/pkg/providers/vertex"

provider, err := vertex.NewProvider(&vertex.Config{
    ProjectID:  os.Getenv("GOOGLE_CLOUD_PROJECT"),  // Required
    Location:   "us-central1",                       // Required
    Model:      "gemini-2.5-pro",                    // Optional (default)
    MaxRetries: 3,                                   // Optional (default: 3)
})
```

**Supported Models:**
- `gemini-2.5-pro` (default)
- `gemini-3-flash-preview`
- `gemini-2.5-flash`
- `gemini-flash-latest`

**Authentication:**
Uses Google Cloud Application Default Credentials (ADC):
1. `GOOGLE_APPLICATION_CREDENTIALS` env var
2. gcloud CLI credentials
3. In-cluster service account (GKE)

**Environment Variables:**
- `GOOGLE_CLOUD_PROJECT` - GCP project ID
- `GOOGLE_CLOUD_LOCATION` - Region (e.g., us-central1)
- `GOOGLE_APPLICATION_CREDENTIALS` - Path to service account JSON

## Tools

Tools give agents capabilities beyond text generation.

### HTTP Tool

Make HTTP requests to APIs.

```go
import "github.com/oskarhane/goagent/pkg/tools/http"

// Create tool
tool := http.NewTool()
handler := http.NewHandler(&http.Config{
    DefaultTimeout:  time.Second * 30,
    MaxResponseSize: 10 * 1024 * 1024, // 10MB
})

registry.MustRegister(tool, handler)
```

**Tool Parameters:**
- `url` (string, required): Target URL
- `method` (string, required): GET, POST, PUT, PATCH, DELETE
- `headers` (object, optional): Request headers
- `body` (string, optional): Request body

**Response Format:**
```json
{
  "status_code": 200,
  "status": "200 OK",
  "headers": {"content-type": "application/json"},
  "body": "...",
  "error": ""
}
```

**Example Usage by Agent:**
```
Use http_request with:
- url: https://api.example.com/users
- method: GET
- headers: {"Authorization": "Bearer token"}
```

### Shell Tool

Execute shell commands with safety constraints.

```go
import "github.com/oskarhane/goagent/pkg/tools/shell"

tool := shell.NewTool()
handler := shell.NewHandler(&shell.Config{
    AllowedCommands: []string{"ls", "ps", "df"}, // Allowlist
    BlockedCommands: []string{"rm", "dd"},       // Blocklist
    DefaultTimeout:    time.Minute * 5,
    MaxOutputSize:     1024 * 1024,              // 1MB
    DefaultWorkingDir: "/tmp",
})

registry.MustRegister(tool, handler)
```

**Tool Parameters:**
- `command` (string, required): Command to execute
- `working_dir` (string, optional): Working directory
- `env` (object, optional): Environment variables
- `timeout` (integer, optional): Timeout in seconds

**Response Format:**
```json
{
  "exit_code": 0,
  "stdout": "...",
  "stderr": "...",
  "error": "",
  "working_dir": "/tmp",
  "execution_time": 1.23
}
```

**Safety Features:**
- Allowlist/blocklist command filtering
- Output size limiting
- Timeout enforcement
- Commands run via `sh -c` for proper parsing

**Default Blocked Commands:**
- `rm -rf /`
- `mkfs`
- `dd if=/dev/zero`
- `:(){ :|:& };:` (fork bomb)

### Kubernetes Tool

Query Kubernetes cluster resources.

```go
import "github.com/oskarhane/goagent/pkg/tools/k8s"

tool := k8s.NewTool()
handler := k8s.NewHandler(&k8s.Config{
    KubeconfigPath:   os.Getenv("KUBECONFIG"),
    DefaultTimeout:   time.Second * 30,
    DefaultNamespace: "default",
})

registry.MustRegister(tool, handler)
```

**Tool Parameters:**
- `operation` (enum, required): `get` or `list`
- `resource` (enum, required): `pod`, `service`, `deployment`, `node`, etc.
- `name` (string, optional): Resource name (required for `get`)
- `namespace` (string, optional): Target namespace (use `all` for cluster-wide)
- `labels` (string, optional): Label selector (e.g., `app=myapp`)
- `timeout` (integer, optional): Timeout in seconds (max: 300)

**Supported Resources:**
- `pod` / `pods`
- `service` / `services`
- `deployment` / `deployments`
- `configmap` / `configmaps`
- `secret` / `secrets`
- `namespace` / `namespaces`
- `node` / `nodes`

**Response Format:**
```json
{
  "operation": "list",
  "resource": "pods",
  "namespace": "default",
  "data": {...}, // Full Kubernetes object JSON
  "error": ""
}
```

**Authentication:**
Uses kubeconfig with fallback chain:
1. Explicit `KubeconfigPath`
2. `KUBECONFIG` environment variable
3. `~/.kube/config`
4. In-cluster config (when running in K8s)

**RBAC:**
Read-only operations. Create ServiceAccount with appropriate Role:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: goagent
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: goagent-reader
rules:
- apiGroups: ["", "apps"]
  resources: ["pods", "services", "deployments"]
  verbs: ["get", "list"]
```

### Custom Tools

Create custom tools using the builder API:

```go
import (
    "github.com/oskarhane/goagent/pkg/tools"
    "github.com/oskarhane/goagent/pkg/types"
)

// Define tool
tool := tools.NewBuilder(
    "my_tool",
    "Description of what the tool does",
).
    StringParam("input", "Description of input param", true).
    IntegerParam("count", "How many times", false).
    ObjectParam("config", "Configuration object", false).
    Build()

// Define handler
handler := func(ctx context.Context, call types.ToolCall) types.ToolResult {
    start := time.Now()
    
    // Parse parameters
    var params struct {
        Input  string         `json:"input"`
        Count  int            `json:"count"`
        Config map[string]any `json:"config"`
    }
    if err := types.ParseToolArguments(call, &params); err != nil {
        return types.ToolResult{
            ToolCallID:    call.ID,
            ToolName:      call.Function.Name,
            Error:         err.Error(),
            ExecutionTime: time.Since(start),
        }
    }
    
    // Execute logic
    result := doWork(params.Input, params.Count)
    
    // Return result
    resultJSON, _ := json.Marshal(result)
    return types.ToolResult{
        ToolCallID:    call.ID,
        ToolName:      call.Function.Name,
        Content:       string(resultJSON),
        ExecutionTime: time.Since(start),
    }
}

// Register
registry.MustRegister(tool, handler)
```

**Parameter Types:**
- `StringParam(name, desc, required)`
- `StringParamWithEnum(name, desc, required, []string{"opt1", "opt2"})`
- `IntegerParam(name, desc, required)`
- `NumberParam(name, desc, required)` - float64
- `BooleanParam(name, desc, required)`
- `ObjectParam(name, desc, required)`
- `ArrayParam(name, desc, required)`

## Types

### Message

Represents a conversation message:

```go
type Message struct {
    Role       string     // "system", "user", "assistant", "tool"
    Content    string     // Text content
    ToolCalls  []ToolCall // Tool invocations (for assistant)
    ToolCallID string     // Tool call ID (for tool responses)
    Name       string     // Tool name (for tool responses)
}

// Helper constructors
msg := types.NewUserMessage("Hello")
msg := types.NewSystemMessage("You are helpful")
msg := types.NewAssistantMessage("I can help")
```

### Tool

Tool definition with JSON Schema parameters:

```go
type Tool struct {
    Type     string   // Always "function"
    Function Function // Function definition
}

type Function struct {
    Name        string      // Unique tool name
    Description string      // What the tool does
    Parameters  *JSONSchema // JSON Schema for params
}
```

### CompletionRequest / Response

LLM API request/response (internal, rarely used directly):

```go
type CompletionRequest struct {
    Messages    []Message
    Tools       []Tool
    Model       string
    MaxTokens   int
    Temperature float64
}

type CompletionResponse struct {
    Message Message // Assistant response
    Usage   Usage   // Token usage
}
```

## Logger

Structured logging with optional OpenTelemetry tracing.

### Creating a Logger

```go
import "github.com/oskarhane/goagent/pkg/logger"

log := logger.New(logger.Config{
    Level:      logger.LevelInfo,   // Debug, Info, Warn, Error
    Output:     os.Stdout,           // io.Writer
    TracerName: "myagent",          // Optional: enable tracing
})
```

**Log Levels:**
- `LevelDebug` (0) - Verbose debugging
- `LevelInfo` (1) - General information
- `LevelWarn` (2) - Warnings
- `LevelError` (3) - Errors

### Log Output Format

JSON structured logs:

```json
{
  "timestamp": "2026-02-16T14:30:00Z",
  "level": "info",
  "message": "agent iteration started",
  "fields": {
    "iteration": 1,
    "total_tokens": 150
  }
}
```

### OpenTelemetry Integration

Set `TracerName` to enable tracing:

```go
log := logger.New(logger.Config{
    Level:      logger.LevelInfo,
    Output:     os.Stdout,
    TracerName: "myagent", // Enables tracing
})
```

Automatic spans created for:
- `agent.run` - Full agent execution
- `provider.complete` - LLM API calls
- Tool executions

Configure OpenTelemetry exporter separately:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Set up Jaeger exporter
exp, _ := jaeger.New(jaeger.WithCollectorEndpoint(
    jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
))

tp := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exp),
)
otel.SetTracerProvider(tp)
```

### Noop Logger

Disable logging:

```go
log := logger.Noop() // No output
```

## Error Handling

### Provider Errors

```go
result := a.Run(ctx, "task", nil)
if result.Error != nil {
    // Check if retryable
    if providerErr, ok := result.Error.(*types.ProviderError); ok {
        if providerErr.IsRetryable() {
            // Retry with backoff
        }
    }
}
```

### Tool Errors

Tool errors are returned to the agent as `ToolResult.Error`:

```go
return types.ToolResult{
    ToolCallID:    call.ID,
    ToolName:      call.Function.Name,
    Error:         "command not found: kubectl",
    ExecutionTime: time.Since(start),
}
```

The agent can see the error and adjust its strategy.

### Context Cancellation

All operations respect context cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
defer cancel()

result := a.Run(ctx, "long task", nil)
// Automatically stops on timeout
```

## Best Practices

### 1. Set Reasonable Limits

```go
cfg := &agent.Config{
    MaxIterations: 15,              // Prevent infinite loops
    Temperature:   &temp,           // 0.0 = deterministic, 1.0 = creative
    Logger:        logger.New(...), // Always log in production
}
```

### 2. Handle Tool Errors Gracefully

```go
handler := func(ctx context.Context, call types.ToolCall) types.ToolResult {
    // Always return ToolResult, never panic
    if err := validate(params); err != nil {
        return types.ToolResult{
            Error: err.Error(),
            // ...
        }
    }
}
```

### 3. Use Conversation History for Multi-Turn

```go
var history []types.Message

result1 := a.Run(ctx, "First question", nil)
history = result1.Messages

result2 := a.Run(ctx, "Follow-up", &agent.RunOptions{
    History: history,
    MaxHistoryMessages: 10, // Trim old messages
})
```

### 4. Secure Tool Access

```go
// Shell: Use allowlist
shellHandler := shell.NewHandler(&shell.Config{
    AllowedCommands: []string{"ls", "ps", "df"},
    BlockedCommands: []string{"rm", "dd", "mkfs"},
})

// K8s: Use RBAC
// Apply Role with minimum permissions needed
```

### 5. Monitor Token Usage

```go
result := a.Run(ctx, task, nil)
log.Printf("Used %d tokens", result.TotalTokens)

// Approximate cost (GPT-4):
// Input: $0.03 / 1K tokens
// Output: $0.06 / 1K tokens
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/oskarhane/goagent/pkg/agent"
    "github.com/oskarhane/goagent/pkg/logger"
    "github.com/oskarhane/goagent/pkg/providers/openai"
    "github.com/oskarhane/goagent/pkg/tools"
    "github.com/oskarhane/goagent/pkg/tools/http"
)

func main() {
    // Create provider
    provider, err := openai.NewProvider(&openai.Config{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create registry with tools
    registry := tools.NewRegistry()
    
    httpTool := http.NewTool()
    httpHandler := http.NewHandler(nil) // Use defaults
    registry.MustRegister(httpTool, httpHandler)

    // Create logger
    l := logger.New(logger.Config{
        Level:  logger.LevelInfo,
        Output: os.Stdout,
    })

    // Create agent
    cfg := &agent.Config{
        Provider:     provider,
        SystemPrompt: "You are a helpful assistant.",
        Registry:     registry,
        Logger:       l,
    }
    
    a, err := agent.NewAgent(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Run task
    ctx := context.Background()
    result := a.Run(ctx, "Get the weather from wttr.in/London", nil)
    
    if result.Error != nil {
        log.Fatal(result.Error)
    }

    // Print response
    for _, msg := range result.Messages {
        if msg.Role == "assistant" && len(msg.ToolCalls) == 0 {
            fmt.Println(msg.Content)
        }
    }
}
```

## Additional Resources

- [GoDoc](https://pkg.go.dev/github.com/oskarhane/goagent) - Full API reference
- [Examples](../examples/) - Working code examples
- [Deployment Guides](../deployments/) - Docker and Kubernetes
- [GitHub Issues](https://github.com/oskarhane/goagent/issues) - Bug reports and feature requests
