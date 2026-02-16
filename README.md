# GoAgent SDK

A lightweight Go library for building AI agents that can interact with cloud infrastructure, execute tools, and automate monitoring and incident response tasks.

## Quick Start

Get your first agent running in under 5 minutes:

```bash
# Install the CLI
go install github.com/oskarhane/goagent/cmd/goagent@latest

# Run your first agent
goagent --help
```

## Features

- **🚀 Rapid Development**: Zero to agent in under 5 minutes
- **🔧 Built-in Tools**: HTTP requests, shell execution, Kubernetes queries
- **🤖 Multi-Provider**: OpenAI and Vertex AI support
- **☁️ Cloud-Native**: Kubernetes-ready with RBAC templates
- **📊 Observability**: Structured logging and OpenTelemetry tracing
- **🛡️ Production-Ready**: Rate limiting, timeouts, and safety constraints

## Installation

### CLI Tool

```bash
go install github.com/oskarhane/goagent/cmd/goagent@latest
```

### As a Library

```bash
go get github.com/oskarhane/goagent
```

## Basic Usage

### Simple Agent

```go
package main

import (
    "context"
    "log"
    
    "github.com/oskarhane/goagent/pkg/agent"
    "github.com/oskarhane/goagent/pkg/providers/openai"
)

func main() {
    // Create a provider
    provider := openai.New("your-api-key")
    
    // Create an agent
    a := agent.New(provider).
        WithSystemPrompt("You are a helpful cloud monitoring assistant")
    
    // Run the agent
    result, err := a.Run(context.Background(), "Check the status of our production pods")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result.Message)
}
```

### Agent with Custom Tools

```go
// Add custom tools to your agent
a := agent.New(provider).
    WithTool(tools.HTTP()).
    WithTool(tools.K8s()).
    WithTool(myCustomTool)
```

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
│   └── types/            # Core type definitions
├── internal/             # Private packages
├── examples/             # Example implementations
└── docs/                 # Documentation
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

# Install development dependencies
make dev-setup

# Run development workflow
make dev
```

### Available Commands

```bash
make help          # Show all available commands
make build         # Build the CLI binary
make test          # Run all tests
make test-coverage # Run tests with coverage
make lint          # Run linting
make docs          # Start documentation server
make clean         # Clean build artifacts
```

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for CI
make ci-test
```

## Documentation

Start the documentation server:

```bash
make docs
```

Then visit: http://localhost:6060/pkg/github.com/oskarhane/goagent/

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run the development workflow (`make dev`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Standards

- All public APIs must have complete GoDoc documentation
- Test coverage must exceed 80%
- All code must pass linting
- Follow conventional commit messages

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Status

🚧 **Work in Progress** - This library is under active development for hackathon use cases.

Current development status:
- [x] Project structure and build system
- [ ] Core provider interface
- [ ] OpenAI provider implementation
- [ ] Vertex AI provider implementation
- [ ] Tool system with JSON Schema validation
- [ ] Agent reasoning loop
- [ ] Built-in tools (HTTP, shell, Kubernetes)
- [ ] Documentation and examples