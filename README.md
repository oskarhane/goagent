# GoAgent SDK

A lightweight Go library for building AI agents that can interact with cloud infrastructure, execute tools, and automate monitoring and incident response tasks.

## Quick Start

Get your first agent running in **under 5 minutes**:

### 1. Clone and run an example (1 minute)

```bash
git clone https://github.com/oskarhane/goagent.git
cd goagent
```

Create a `.env` file with your API key:

```bash
OPENAI_API_KEY=sk-...
```

Run an example:

```bash
export $(cat .env | xargs) && go run ./examples/agent-basic
```

### 2. Use in your own project

```bash
mkdir myagent && cd myagent
go mod init myagent
go get github.com/oskarhane/goagent
```

### 3. Create your first agent (3 minutes)

Create `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    _ "github.com/joho/godotenv/autoload" // Auto-load .env file
    "github.com/oskarhane/goagent/pkg/agent"
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
    
    // Create agent with HTTP tool
    registry := tools.NewRegistry()
    httpTool := http.NewTool()
    httpHandler := http.NewHandler(nil)
    registry.MustRegister(httpTool, httpHandler)
    
    cfg := &agent.Config{
        Provider:      provider,
        SystemPrompt:  "You are a helpful assistant that can make HTTP requests.",
        Registry:      registry,
        MaxIterations: 10,
    }
    
    a, err := agent.NewAgent(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    // Run agent
    ctx := context.Background()
    result := a.Run(ctx, "Get the current weather from wttr.in/London", nil)
    if result.Error != nil {
        log.Fatal(result.Error)
    }
    
    // Print final response
    for _, msg := range result.Messages {
        if msg.Role == "assistant" && len(msg.ToolCalls) == 0 {
            fmt.Println("Agent:", msg.Content)
        }
    }
}
```

### 4. Run it!

```bash
go mod tidy
export $(cat .env | xargs) && go run main.go
```

**That's it!** You now have a working AI agent that can make HTTP requests.

## Features

- **Rapid Development**: Zero to agent in under 5 minutes
- **Built-in Tools**: HTTP requests, shell execution, Kubernetes queries
- **Multi-Provider**: OpenAI and Vertex AI support
- **Cloud-Native**: Kubernetes-ready with RBAC templates
- **Observability**: Structured logging and OpenTelemetry tracing
- **Production-Ready**: Rate limiting, timeouts, and safety constraints

## Core Concepts

### Providers

Providers are LLM backends. GoAgent supports:

- **OpenAI**: gpt-5.1, gpt-5-mini, gpt-5
- **Vertex AI**: gemini-2.5-pro, gemini-3-flash-preview, gemini-2.5-flash, gemini-flash-latest

```go
// OpenAI
provider, _ := openai.NewProvider(&openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
    Model:  "gpt-5.1",
})

// Vertex AI
provider, _ := vertex.NewProvider(&vertex.Config{
    ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
    Location:  os.Getenv("GOOGLE_CLOUD_LOCATION"),
    Model:     "gemini-2.5-pro",
})
```

### Tools

Tools give agents capabilities. Built-in tools:

```go
import (
    "github.com/oskarhane/goagent/pkg/tools/http"
    "github.com/oskarhane/goagent/pkg/tools/shell"
    "github.com/oskarhane/goagent/pkg/tools/k8s"
)

// HTTP requests
httpRegistry := http.NewHandler(nil)

// Shell commands (with safety constraints)
shellRegistry := shell.NewHandler(&shell.Config{
    AllowedCommands: []string{"ls", "ps", "df"},
    Timeout:         time.Minute * 5,
})

// Kubernetes queries
k8sRegistry := k8s.NewHandler(&k8s.Config{
    KubeconfigPath: os.Getenv("KUBECONFIG"),
})
```

### Agent

The agent orchestrates reasoning and tool execution:

```go
cfg := &agent.Config{
    Provider:      provider,
    SystemPrompt:  "You are a cloud monitoring assistant.",
    Registry:      registry,
    MaxIterations: 10,  // Safety limit
    Temperature:   &temp, // 0.0-1.0
    Logger:        logger,
}

a, _ := agent.NewAgent(cfg)
result, _ := a.Run(ctx, "Check system health", nil)
```

## Examples

### Kubernetes Monitoring

Monitor cluster health and report issues:

```bash
cd examples/k8s-monitoring
go run main.go
```

See [examples/k8s-monitoring/](examples/k8s-monitoring/) for full code.

### Incident Response

Investigate and diagnose incidents:

```bash
cd examples/incident-response
go run main.go
```

See [examples/incident-response/](examples/incident-response/) for full code.

### More Examples

- [Agent basics](examples/agent-basic/) - Simple agent with tools
- [HTTP tool](examples/http-tool/) - Making API requests
- [Shell tool](examples/shell-tool/) - Executing commands
- [K8s tool](examples/k8s-tool/) - Querying clusters
- [Conversation history](examples/conversation-history/) - Multi-turn conversations
- [Logging](examples/logging-basic/) - Structured logging

## Creating Custom Tools

Define custom tools using the builder API:

```go
import "github.com/oskarhane/goagent/pkg/tools"

func MyCustomTool() *types.Tool {
    return tools.NewBuilder("my_tool", "Describes what my tool does").
        StringParam("input", "Description of the input", true).
        Build(func(ctx context.Context, params map[string]any) (any, error) {
            input := params["input"].(string)
            
            // Your logic here
            result := process(input)
            
            return result, nil
        })
}

// Use it
registry := tools.NewRegistry()
registry.Register(MyCustomTool())

cfg := &agent.Config{
    Provider: provider,
    Registry: registry,
    // ...
}
```

## Conversation History

Maintain context across multiple interactions:

```go
var history []types.Message

// First interaction
result1, _ := a.Run(ctx, "What's 5 + 3?", nil)
history = result1.Messages

// Second interaction (with context)
opts := &agent.RunOptions{
    History: history,
}
result2, _ := a.Run(ctx, "Multiply that by 2", opts)
history = result2.Messages
```

## Deployment

### Docker

Build and run as a container:

```bash
docker build -t myagent .
docker run -e OPENAI_API_KEY=sk-... myagent
```

See [deployments/docker/](deployments/docker/) for service and cronjob variants.

### Kubernetes

Deploy as a service or cronjob:

```bash
# Apply RBAC
kubectl apply -f deployments/k8s/base/rbac.yaml

# Deploy as CronJob
kubectl apply -f deployments/k8s/examples/monitoring-agent.yaml

# Deploy as Service
kubectl apply -f deployments/k8s/examples/incident-response.yaml
```

See [deployments/k8s/](deployments/k8s/) for complete manifests and configuration.

## Observability

### Structured Logging

```go
import "github.com/oskarhane/goagent/pkg/logger"

log := logger.New(&logger.Config{
    Level:      logger.LevelInfo,
    Output:     os.Stdout,
    TracerName: "myagent", // Optional OpenTelemetry
})

cfg := &agent.Config{
    Provider: provider,
    Logger:   log,
    // ...
}
```

### OpenTelemetry Tracing

```go
log := logger.New(&logger.Config{
    Level:      logger.LevelInfo,
    Output:     os.Stdout,
    TracerName: "myagent",
})
```

Traces are automatically created for:
- Agent execution (`agent.run`)
- Provider API calls (`provider.complete`)
- Tool execution

## API Documentation

Complete API documentation is available via godoc:

```bash
make docs
# Visit http://localhost:6060/pkg/github.com/oskarhane/goagent/
```

Or view online at: https://pkg.go.dev/github.com/oskarhane/goagent

Key packages:
- [pkg/agent](https://pkg.go.dev/github.com/oskarhane/goagent/pkg/agent) - Agent orchestration
- [pkg/providers/openai](https://pkg.go.dev/github.com/oskarhane/goagent/pkg/providers/openai) - OpenAI provider
- [pkg/providers/vertex](https://pkg.go.dev/github.com/oskarhane/goagent/pkg/providers/vertex) - Vertex AI provider
- [pkg/tools](https://pkg.go.dev/github.com/oskarhane/goagent/pkg/tools) - Tool system
- [pkg/types](https://pkg.go.dev/github.com/oskarhane/goagent/pkg/types) - Core types

## Architecture

```
├── cmd/goagent/          # CLI application
├── pkg/
│   ├── agent/            # Core agent implementation
│   ├── providers/        # LLM provider implementations
│   │   ├── openai/       # OpenAI provider
│   │   └── vertex/       # Vertex AI provider
│   ├── tools/            # Built-in tool implementations
│   │   ├── http/         # HTTP request tool
│   │   ├── shell/        # Shell execution tool
│   │   └── k8s/          # Kubernetes query tool
│   ├── logger/           # Structured logging
│   └── types/            # Core type definitions
├── deployments/          # Deployment configurations
│   ├── docker/           # Docker configurations
│   └── k8s/              # Kubernetes manifests
└── examples/             # Working examples
```

## Development

### Prerequisites

- Go 1.26 or later
- Make

### Setup

```bash
# Clone the repository
git clone https://github.com/oskarhane/goagent.git
cd goagent

# Run tests
make test

# Run linting
make lint

# Build binary
make build
```

### Available Commands

```bash
make test          # Run all tests
make lint          # Run linting
make build         # Build the CLI binary
make docs          # Start documentation server
make clean         # Clean build artifacts
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b oskar/amazing-feature`)
3. Make your changes
4. Run tests and linting (`make test && make lint`)
5. Commit your changes
6. Push to the branch
7. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Status

**Ready for Hackathons!** This library is production-ready with:

- [x] Project structure and build system
- [x] Core provider interface
- [x] OpenAI provider implementation
- [x] Vertex AI provider implementation
- [x] Tool system with JSON Schema validation
- [x] Agent reasoning loop
- [x] Built-in tools (HTTP, shell, Kubernetes)
- [x] Structured logging and tracing
- [x] Docker and Kubernetes deployment
- [x] Documentation and examples
- [ ] Comprehensive test suite (in progress)