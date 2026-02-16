# Product Requirements Document: GoAgent SDK

## Cloud AI Agent Library for Hackathon

**Version:** 1.0  
**Date:** February 2026  
**Author:** Platform Team  
**Status:** Draft

---

## 1. Executive Summary

GoAgent SDK is a lightweight Go library that enables cloud engineers to rapidly build AI agents capable of monitoring, investigating, and gathering information about cloud deployments and incidents. Designed for a hackathon context, the library prioritizes **ease of use** and **quick time-to-first-agent** over feature completeness.

The library draws inspiration from Vercel's AI SDK (TypeScript), Mastra, and PydanticAI, adapting their developer-friendly patterns to idiomatic Go.

---

## 2. Problem Statement

Cloud engineers need to build AI agents that can:
- Monitor cloud infrastructure and respond to incidents
- Execute tools to gather deployment information
- Reason about complex multi-step investigations

Current options in Go are either:
- Too low-level (raw API clients require boilerplate)
- Too complex (enterprise frameworks with steep learning curves)
- Non-existent (most agent frameworks target Python/TypeScript)

**Hackathon participants need to go from zero to working agent in under 30 minutes.**

---

## 3. Goals & Non-Goals

### Goals
- **5-minute setup**: Install, configure, run first agent
- **30-minute mastery**: Build a custom tool-using agent
- **Minimal boilerplate**: Sensible defaults, convention over configuration
- **Production-aware**: Built-in tracing/logging for debugging and observability
- **Cloud-native**: Easy deployment to Kubernetes

### Non-Goals
- Comprehensive LLM provider support (only OpenAI + Vertex AI)
- Advanced features (RAG, memory, multi-agent orchestration)
- UI/Dashboard components
- Streaming responses (v1 will use simple request/response)

---

## 4. Target Users

**Primary:** Cloud engineers participating in the hackathon
- Comfortable with Go
- Familiar with Kubernetes and cloud platforms
- May have limited AI/LLM experience
- Want to build practical monitoring/incident response tools

---

## 5. Core Abstractions

### 5.1 LLM Providers

A unified interface for interacting with different LLM backends.

```go
// Provider interface - simple and consistent
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    CompleteWithTools(ctx context.Context, req ToolCompletionRequest) (*ToolCompletionResponse, error)
}

// Usage - dead simple
provider := openai.New(os.Getenv("OPENAI_API_KEY"))
// or
provider := vertex.New(vertex.Config{
    ProjectID: "my-project",
    Location:  "us-central1",
})
```

**Supported Providers:**
| Provider | Models | Auth |
|----------|--------|------|
| OpenAI | gpt-4o, gpt-4o-mini | API Key |
| Vertex AI | gemini-2.0-flash, gemini-2.0-pro | Application Default Credentials / Service Account |

### 5.2 Tools

Tools are functions the agent can call. Inspired by AI SDK's tool definition pattern.

```go
// Define a tool with minimal boilerplate
getPodsTool := goagent.Tool{
    Name:        "get_pods",
    Description: "List all pods in a Kubernetes namespace",
    Parameters: goagent.Schema{
        Type: "object",
        Properties: map[string]goagent.Property{
            "namespace": {Type: "string", Description: "Kubernetes namespace"},
            "status":    {Type: "string", Description: "Filter by status (Running, Pending, Failed)", Optional: true},
        },
        Required: []string{"namespace"},
    },
    Execute: func(ctx context.Context, params map[string]any) (any, error) {
        namespace := params["namespace"].(string)
        // Your Kubernetes client logic here
        return pods, nil
    },
}
```

**Built-in Tool Helpers:**
```go
// For common cloud operations
tools.HTTPGet()      // Make HTTP requests
tools.ShellExec()    // Run shell commands (with safety controls)
tools.KubeQuery()    // Query Kubernetes resources
```

### 5.3 Agents (Tool Loop)

The core abstraction - an agent that reasons and acts in a loop until the task is complete.

```go
// Create an agent in 3 lines
agent := goagent.New(
    goagent.WithProvider(provider),
    goagent.WithSystemPrompt(`You are a cloud infrastructure assistant. 
        You help engineers investigate incidents and monitor deployments.
        Always verify information before reporting.`),
    goagent.WithTools(getPodsTool, getLogsTool, getMetricsTool),
)

// Run the agent - it loops until done
result, err := agent.Run(ctx, "Why is the payment service showing high latency?")
```

**Tool Loop Behavior (inspired by AI SDK's generateText with tools):**

```
┌─────────────────────────────────────────────────────────────────┐
│                         AGENT LOOP                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. User Message → LLM                                          │
│         ↓                                                       │
│  2. LLM decides: respond OR call tool(s)                        │
│         ↓                                                       │
│  3a. If tool call → Execute tool → Add result to context        │
│         ↓                                                       │
│      Loop back to step 2                                        │
│                                                                 │
│  3b. If response → Return final answer                          │
│                                                                 │
│  Safety: Max iterations (default: 10) prevents infinite loops   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Configuration Options:**
```go
agent := goagent.New(
    goagent.WithProvider(provider),
    goagent.WithSystemPrompt(prompt),
    goagent.WithTools(tools...),
    
    // Optional configurations with sensible defaults
    goagent.WithMaxIterations(10),        // Default: 10
    goagent.WithTimeout(5 * time.Minute), // Default: 5 min
    goagent.WithOnToolCall(func(name string, params map[string]any) {
        log.Printf("Calling tool: %s", name)
    }),
)
```

### 5.4 Tracing & Logging

Built-in observability that works out of the box.

```go
// Automatic structured logging
agent := goagent.New(
    // ... other config
    goagent.WithTracing(goagent.TracingConfig{
        Enabled: true,
        Output:  os.Stdout,                    // or file, or custom writer
        Level:   goagent.TraceLevelDetailed,  // Minimal | Standard | Detailed
    }),
)
```

**Trace Output Example:**
```json
{
  "trace_id": "abc123",
  "timestamp": "2026-02-16T10:30:00Z",
  "agent_run": {
    "input": "Why is payment service slow?",
    "iterations": [
      {
        "step": 1,
        "llm_reasoning": "I should check the pod status first",
        "tool_calls": [{"name": "get_pods", "params": {"namespace": "payments"}}],
        "tool_results": [{"pods": [...]}],
        "duration_ms": 1250
      },
      {
        "step": 2,
        "llm_reasoning": "Pods are healthy, let me check metrics",
        "tool_calls": [{"name": "get_metrics", "params": {"service": "payment-api"}}],
        "tool_results": [{"cpu": "85%", "memory": "70%"}],
        "duration_ms": 890
      }
    ],
    "final_response": "The payment service latency is caused by...",
    "total_duration_ms": 4500,
    "total_tokens": 2340
  }
}
```

**OpenTelemetry Integration (Optional):**
```go
goagent.WithOTelTracing(tracerProvider) // For production observability
```

---

## 6. API Design Principles

### 6.1 Progressive Disclosure

Simple things should be simple. Complex things should be possible.

```go
// Simplest possible agent (3 lines)
agent := goagent.New(goagent.WithProvider(openai.New(apiKey)))
result, _ := agent.Run(ctx, "Hello!")

// Full-featured agent (still readable)
agent := goagent.New(
    goagent.WithProvider(provider),
    goagent.WithSystemPrompt(prompt),
    goagent.WithTools(tool1, tool2),
    goagent.WithMaxIterations(15),
    goagent.WithTracing(tracingConfig),
    goagent.WithOnToolCall(callback),
)
```

### 6.2 Errors are Values

Clear, actionable error types.

```go
result, err := agent.Run(ctx, input)
if err != nil {
    switch {
    case errors.Is(err, goagent.ErrMaxIterations):
        // Agent hit iteration limit
    case errors.Is(err, goagent.ErrToolExecution):
        // A tool failed
    case errors.Is(err, goagent.ErrProviderRateLimit):
        // Hit API rate limit
    }
}
```

### 6.3 Context-First

All operations respect context for cancellation and timeouts.

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

result, err := agent.Run(ctx, "Investigate the outage")
// Automatically cancelled if timeout exceeded
```

---

## 7. Deployment Patterns

### 7.1 Local Development (Docker)

For hackathon development and testing:

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o agent ./cmd/agent

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/agent .
CMD ["./agent"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  agent:
    build: .
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - GOOGLE_APPLICATION_CREDENTIALS=/secrets/sa.json
    volumes:
      - ./secrets:/secrets:ro
      - ~/.kube/config:/root/.kube/config:ro  # For K8s access
    ports:
      - "8080:8080"  # If running as HTTP service
```

**Quick Start Commands:**
```bash
# Run agent locally
docker-compose up --build

# Run one-off investigation
docker-compose run agent ./agent --query "Check pod health in prod"
```

### 7.2 Kubernetes Deployment

**Pattern A: Long-Running Service (HTTP/gRPC API)**

For agents that respond to on-demand requests:

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloud-agent
  labels:
    app: cloud-agent
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cloud-agent
  template:
    metadata:
      labels:
        app: cloud-agent
    spec:
      serviceAccountName: cloud-agent  # For RBAC
      containers:
      - name: agent
        image: your-registry/cloud-agent:latest
        ports:
        - containerPort: 8080
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: llm-credentials
              key: openai-api-key
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: cloud-agent
spec:
  selector:
    app: cloud-agent
  ports:
  - port: 80
    targetPort: 8080
```

**Pattern B: CronJob (Scheduled Monitoring)**

For periodic checks and reports:

```yaml
# k8s/cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: infra-health-check
spec:
  schedule: "*/15 * * * *"  # Every 15 minutes
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: cloud-agent
          containers:
          - name: agent
            image: your-registry/cloud-agent:latest
            command: ["./agent"]
            args: ["--task", "health-check", "--namespace", "production"]
            env:
            - name: OPENAI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: llm-credentials
                  key: openai-api-key
            - name: SLACK_WEBHOOK_URL
              valueFrom:
                secretKeyRef:
                  name: notifications
                  key: slack-webhook
          restartPolicy: OnFailure
```

**Pattern C: Event-Driven (Incident Response)**

Triggered by alerts or webhooks:

```yaml
# Using KEDA for event-driven scaling
apiVersion: keda.sh/v1alpha1
kind: ScaledJob
metadata:
  name: incident-responder
spec:
  jobTargetRef:
    template:
      spec:
        serviceAccountName: cloud-agent
        containers:
        - name: agent
          image: your-registry/cloud-agent:latest
          command: ["./agent"]
          args: ["--mode", "incident-response"]
  triggers:
  - type: prometheus
    metadata:
      serverAddress: http://prometheus:9090
      metricName: alertmanager_alerts_firing
      threshold: "1"
```

### 7.3 RBAC Configuration

Agents need permissions to query cluster resources:

```yaml
# k8s/rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloud-agent
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cloud-agent-reader
rules:
- apiGroups: [""]
  resources: ["pods", "services", "events", "configmaps", "nodes"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "statefulsets"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cloud-agent-reader-binding
subjects:
- kind: ServiceAccount
  name: cloud-agent
  namespace: default
roleRef:
  kind: ClusterRole
  name: cloud-agent-reader
  apiGroup: rbac.authorization.k8s.io
```

### 7.4 Recommended Hackathon Setup

```
┌─────────────────────────────────────────────────────────────────┐
│                    HACKATHON ARCHITECTURE                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Development Machine                                            │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Docker Compose                                          │   │
│  │  ├── agent (your code)                                   │   │
│  │  └── Local testing against shared K8s cluster            │   │
│  └─────────────────────────────────────────────────────────┘   │
│                           │                                     │
│                           ▼                                     │
│  Shared Kubernetes Cluster (provided)                           │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  ├── Namespace: team-{name}                              │   │
│  │  │   ├── Your agent deployment                           │   │
│  │  │   └── ServiceAccount (pre-configured RBAC)            │   │
│  │  │                                                       │   │
│  │  ├── Namespace: sample-apps (read-only)                  │   │
│  │  │   ├── payment-service                                 │   │
│  │  │   ├── inventory-service                               │   │
│  │  │   └── (apps to monitor/investigate)                   │   │
│  │  │                                                       │   │
│  │  └── Shared Services                                     │   │
│  │      ├── Prometheus (metrics)                            │   │
│  │      ├── Loki (logs)                                     │   │
│  │      └── Jaeger (traces)                                 │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 8. Example: Complete Agent

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/yourorg/goagent"
    "github.com/yourorg/goagent/providers/openai"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func main() {
    // Initialize Kubernetes client
    config, _ := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
    k8sClient, _ := kubernetes.NewForConfig(config)

    // Initialize LLM provider
    provider := openai.New(os.Getenv("OPENAI_API_KEY"),
        openai.WithModel("gpt-4o"),
    )

    // Define tools
    getPods := goagent.Tool{
        Name:        "get_pods",
        Description: "List pods in a namespace with their status",
        Parameters: goagent.Schema{
            Type: "object",
            Properties: map[string]goagent.Property{
                "namespace": {Type: "string", Description: "Kubernetes namespace"},
            },
            Required: []string{"namespace"},
        },
        Execute: func(ctx context.Context, params map[string]any) (any, error) {
            ns := params["namespace"].(string)
            pods, err := k8sClient.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
            if err != nil {
                return nil, err
            }
            
            var result []map[string]string
            for _, pod := range pods.Items {
                result = append(result, map[string]string{
                    "name":   pod.Name,
                    "status": string(pod.Status.Phase),
                    "ready":  fmt.Sprintf("%d/%d", readyContainers(pod), len(pod.Spec.Containers)),
                })
            }
            return result, nil
        },
    }

    getLogs := goagent.Tool{
        Name:        "get_logs",
        Description: "Get recent logs from a pod",
        Parameters: goagent.Schema{
            Type: "object",
            Properties: map[string]goagent.Property{
                "namespace": {Type: "string", Description: "Kubernetes namespace"},
                "pod":       {Type: "string", Description: "Pod name"},
                "lines":     {Type: "integer", Description: "Number of lines (default 50)", Optional: true},
            },
            Required: []string{"namespace", "pod"},
        },
        Execute: func(ctx context.Context, params map[string]any) (any, error) {
            ns := params["namespace"].(string)
            pod := params["pod"].(string)
            lines := int64(50)
            if l, ok := params["lines"].(float64); ok {
                lines = int64(l)
            }
            
            logs, err := k8sClient.CoreV1().Pods(ns).GetLogs(pod, &corev1.PodLogOptions{
                TailLines: &lines,
            }).Do(ctx).Raw()
            
            return string(logs), err
        },
    }

    // Create the agent
    agent := goagent.New(
        goagent.WithProvider(provider),
        goagent.WithSystemPrompt(`You are a Kubernetes incident investigator.
            When investigating issues:
            1. First check pod status to identify unhealthy pods
            2. Look at logs for error messages
            3. Summarize findings with specific evidence
            
            Be concise and actionable in your responses.`),
        goagent.WithTools(getPods, getLogs),
        goagent.WithTracing(goagent.TracingConfig{
            Enabled: true,
            Output:  os.Stdout,
            Level:   goagent.TraceLevelStandard,
        }),
    )

    // Run investigation
    ctx := context.Background()
    result, err := agent.Run(ctx, "Check if there are any issues in the payments namespace")
    if err != nil {
        log.Fatalf("Agent failed: %v", err)
    }

    fmt.Println("\n=== Investigation Result ===")
    fmt.Println(result.Response)
}
```

---

## 9. Project Structure

```
goagent/
├── agent.go              # Core Agent type and Run loop
├── options.go            # WithXxx configuration functions
├── tool.go               # Tool definition and execution
├── schema.go             # JSON Schema types for parameters
├── errors.go             # Error types
├── tracing.go            # Tracing/logging implementation
│
├── providers/
│   ├── provider.go       # Provider interface
│   ├── openai/
│   │   └── openai.go     # OpenAI implementation
│   └── vertex/
│       └── vertex.go     # Vertex AI implementation
│
├── tools/                # Optional built-in tool helpers
│   ├── http.go
│   ├── shell.go
│   └── kubernetes.go
│
├── examples/
│   ├── simple/           # Minimal example
│   ├── kubernetes/       # K8s monitoring example
│   └── incident/         # Full incident response example
│
└── cmd/
    └── goagent/          # Optional CLI for testing
```

---

## 10. Success Criteria

### Hackathon Day Metrics

| Metric | Target |
|--------|--------|
| Time to first running agent | < 5 minutes |
| Time to custom tool-using agent | < 30 minutes |
| Participants who deploy to K8s | > 80% |
| Documentation questions during event | < 10 |

### Code Quality Metrics

| Metric | Target |
|--------|--------|
| Test coverage | > 80% |
| GoDoc coverage | 100% of public API |
| Example coverage | Every major feature |

---

## 11. Timeline

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| Core Implementation | 3 days | Agent loop, OpenAI provider, basic tracing |
| Vertex AI + Tools | 2 days | Vertex provider, built-in tool helpers |
| Examples & Docs | 2 days | Kubernetes example, deployment manifests, README |
| Testing & Polish | 1 day | Integration tests, error handling, edge cases |

---

## 12. Open Questions

1. **Streaming Support**: Should v1 support streaming responses, or is request/response sufficient for hackathon?
   - *Recommendation*: Skip for v1, add later if requested

2. **Multi-Model Routing**: Should agents be able to use different models for different tasks?
   - *Recommendation*: Out of scope for v1

3. **Conversation History**: Should agents maintain conversation context across multiple Run() calls?
   - *Recommendation*: Add simple `WithHistory([]Message)` option

4. **Rate Limiting**: Should the library handle provider rate limits automatically?
   - *Recommendation*: Yes, with exponential backoff built-in

---

## 13. Appendix: Quick Reference Card

```go
// ===== PROVIDER SETUP =====
provider := openai.New(apiKey)
provider := openai.New(apiKey, openai.WithModel("gpt-4o-mini"))
provider := vertex.New(vertex.Config{ProjectID: "x", Location: "us-central1"})

// ===== TOOL DEFINITION =====
tool := goagent.Tool{
    Name:        "tool_name",
    Description: "What this tool does",
    Parameters:  goagent.Schema{...},
    Execute:     func(ctx, params) (any, error) { ... },
}

// ===== AGENT CREATION =====
agent := goagent.New(
    goagent.WithProvider(provider),
    goagent.WithSystemPrompt("You are..."),
    goagent.WithTools(tool1, tool2),
    goagent.WithMaxIterations(10),
    goagent.WithTracing(goagent.TracingConfig{Enabled: true}),
)

// ===== RUNNING =====
result, err := agent.Run(ctx, "Your question or task")
fmt.Println(result.Response)      // Final text response
fmt.Println(result.ToolCalls)     // History of tool calls
fmt.Println(result.TokensUsed)    // Total tokens consumed
```

---

*Document maintained by Platform Team. For questions, reach out on #platform-support.*
