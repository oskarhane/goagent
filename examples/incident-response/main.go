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
	"github.com/oskarhane/goagent/pkg/tools/k8s"
	"github.com/oskarhane/goagent/pkg/tools/shell"
)

// Incident Response Agent
// This example creates an agent that investigates incidents by querying
// multiple data sources: Kubernetes cluster, HTTP APIs, and system commands.

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create provider with gpt-5.1 for complex reasoning
	provider, err := openai.NewProvider(&openai.Config{
		APIKey: apiKey,
		Model:  "gpt-5.1",
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Add HTTP tool for API queries
	httpTool := http.NewTool()
	httpHandler := http.NewHandler(&http.Config{
		DefaultTimeout:  time.Second * 30,
		MaxResponseSize: 10 * 1024 * 1024, // 10MB
	})
	registry.MustRegister(httpTool, httpHandler)

	// Add Kubernetes tool for cluster queries
	k8sTool := k8s.NewTool()
	k8sHandler := k8s.NewHandler(&k8s.Config{
		KubeconfigPath:   os.Getenv("KUBECONFIG"),
		DefaultTimeout:   time.Second * 30,
		DefaultNamespace: "default",
	})
	registry.MustRegister(k8sTool, k8sHandler)

	// Add shell tool for system diagnostics (with safety constraints)
	shellTool := shell.NewTool()
	shellHandler := shell.NewHandler(&shell.Config{
		// Only allow safe, read-only commands
		AllowedCommands: []string{
			"ps", "top", "df", "free", "uptime",
			"netstat", "ss", "lsof", "vmstat",
			"cat", "tail", "head", "grep", "ls",
		},
		DefaultTimeout:    time.Minute * 2,
		MaxOutputSize:     1024 * 1024, // 1MB
		DefaultWorkingDir: "/tmp",
		BlockedCommands:   []string{"rm", "dd", "mkfs"}, // Extra safety
	})
	registry.MustRegister(shellTool, shellHandler)

	// Create logger
	logLevel := logger.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		logLevel = logger.LevelDebug
	}
	l := logger.New(logger.Config{
		Level:  logLevel,
		Output: os.Stdout,
	})

	// Create agent with incident response prompt
	systemPrompt := `You are an expert incident response engineer. Your job is to investigate and diagnose incidents in production systems.

When investigating an incident, follow this methodology:
1. Gather initial information about the incident
2. Query relevant systems to understand the scope
3. Correlate data from multiple sources (K8s, APIs, system metrics)
4. Identify the root cause
5. Provide actionable recommendations

You have access to:
- k8s_query: Query Kubernetes resources (pods, services, deployments, nodes)
- http_request: Make HTTP API calls to monitoring systems or services
- shell_execute: Run safe diagnostic commands (read-only operations)

Be thorough but efficient. Ask clarifying questions if the incident description is vague.
Provide clear, concise findings with evidence from your investigation.`

	cfg := &agent.Config{
		Provider:      provider,
		SystemPrompt:  systemPrompt,
		Registry:      registry,
		MaxIterations: 20, // More iterations for complex investigations
		Logger:        l,
	}

	a, err := agent.NewAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Get incident description from command line or use default
	incident := "Users are reporting 500 errors from the production API. The errors started 10 minutes ago."
	if len(os.Args) > 1 {
		incident = os.Args[1]
	}

	// Run investigation
	ctx := context.Background()

	fmt.Println("=== Incident Response Investigation ===")
	fmt.Printf("Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Incident: %s\n\n", incident)

	result := a.Run(ctx, fmt.Sprintf("Investigate this incident: %s", incident), nil)
	if result.Error != nil {
		log.Fatalf("Agent execution failed: %v", result.Error)
	}

	// Print investigation report
	fmt.Println("=== Investigation Report ===")
	for _, msg := range result.Messages {
		if msg.Role == "assistant" && len(msg.ToolCalls) == 0 {
			fmt.Println(msg.Content)
		}
	}

	// Print tool execution summary
	fmt.Printf("\n=== Tool Usage Summary ===\n")
	toolCalls := 0
	for _, msg := range result.Messages {
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			toolCalls += len(msg.ToolCalls)
		}
	}
	fmt.Printf("Tools executed: %d\n", toolCalls)

	fmt.Printf("\n=== Execution Stats ===\n")
	fmt.Printf("Iterations: %d\n", result.Iterations)
	fmt.Printf("Total tokens: %d\n", result.TotalTokens)
	fmt.Printf("Duration: %s\n", result.ExecutionTime)
	fmt.Printf("Completed at: %s\n", time.Now().Format(time.RFC3339))
}
