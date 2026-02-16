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
	"github.com/oskarhane/goagent/pkg/tools/k8s"
)

// K8s Monitoring Agent
// This example creates an agent that monitors Kubernetes cluster health
// and reports on issues with pods, services, deployments, and nodes.

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create provider
	provider, err := openai.NewProvider(&openai.Config{
		APIKey: apiKey,
		Model:  "gpt-5.1",
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Add Kubernetes tool
	k8sTool := k8s.NewTool()
	k8sHandler := k8s.NewHandler(&k8s.Config{
		KubeconfigPath:   os.Getenv("KUBECONFIG"),
		DefaultTimeout:   time.Second * 30,
		DefaultNamespace: "default",
	})
	if err := registry.Register(k8sTool, k8sHandler); err != nil {
		log.Fatalf("Failed to register K8s tool: %v", err)
	}

	// Create logger
	logLevel := logger.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		logLevel = logger.LevelDebug
	}
	l := logger.New(logger.Config{
		Level:  logLevel,
		Output: os.Stdout,
	})

	// Create agent
	systemPrompt := `You are a Kubernetes monitoring assistant. Your job is to check cluster health and report issues.

Check the following:
1. Pods that are not in Running state (across all namespaces)
2. Services that might have issues
3. Deployments with unavailable replicas
4. Nodes that are not Ready

Provide a clear summary of:
- Overall cluster health (healthy/degraded/critical)
- Number of problematic resources found
- Details about each issue
- Recommendations for investigation

Use the k8s_query tool to gather information.`

	cfg := &agent.Config{
		Provider:      provider,
		SystemPrompt:  systemPrompt,
		Registry:      registry,
		MaxIterations: 15,
		Logger:        l,
	}

	a, err := agent.NewAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Run monitoring task
	ctx := context.Background()
	task := "Perform a comprehensive health check of the Kubernetes cluster and report any issues you find."

	fmt.Println("=== Kubernetes Cluster Health Check ===")
	fmt.Printf("Started at: %s\n\n", time.Now().Format(time.RFC3339))

	result := a.Run(ctx, task, nil)
	if result.Error != nil {
		log.Fatalf("Agent execution failed: %v", result.Error)
	}

	// Print results
	fmt.Println("=== Monitoring Report ===")
	for _, msg := range result.Messages {
		if msg.Role == "assistant" && len(msg.ToolCalls) == 0 {
			fmt.Println(msg.Content)
		}
	}

	fmt.Printf("\n=== Execution Stats ===\n")
	fmt.Printf("Iterations: %d\n", result.Iterations)
	fmt.Printf("Total tokens: %d\n", result.TotalTokens)
	fmt.Printf("Duration: %s\n", result.ExecutionTime)
	fmt.Printf("Completed at: %s\n", time.Now().Format(time.RFC3339))
}
