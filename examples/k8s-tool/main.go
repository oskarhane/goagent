package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/oskarhane/goagent/pkg/agent"
	"github.com/oskarhane/goagent/pkg/providers/openai"
	"github.com/oskarhane/goagent/pkg/tools"
	k8sTool "github.com/oskarhane/goagent/pkg/tools/k8s"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create provider
	provider, err := openai.NewProvider(openai.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Create tool registry with Kubernetes tool
	registry := tools.NewRegistry()

	// Register the Kubernetes tool with default configuration
	// Uses KUBECONFIG environment variable or ~/.kube/config by default
	k8sToolDef := k8sTool.NewTool()
	k8sHandler := k8sTool.NewHandler(&k8sTool.Config{
		DefaultTimeout:   60 * time.Second,
		DefaultNamespace: "default",
		// KubeconfigPath: "/path/to/kubeconfig", // Optional: specify custom path
	})
	registry.MustRegister(k8sToolDef, k8sHandler)

	// Create agent
	a, err := agent.NewAgent(agent.Config{
		Provider: provider,
		Registry: registry,
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Example 1: List all pods in the default namespace
	fmt.Println("=== Example 1: List all pods in default namespace ===")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel1()

	result1 := a.Run(ctx1, "List all pods in the default namespace and summarize their status", nil)
	if result1.Error != nil {
		log.Printf("Agent error: %v", result1.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result1.Iterations)
		fmt.Printf("Final response: %s\n\n", result1.Response.Content)
		fmt.Printf("Total tokens used: %d\n\n", result1.TotalTokens)
	}

	// Example 2: Get information about a specific service
	fmt.Println("\n=== Example 2: Get service information ===")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel2()

	result2 := a.Run(ctx2, "Get information about the 'kubernetes' service in the default namespace and tell me its cluster IP and ports", nil)
	if result2.Error != nil {
		log.Printf("Agent error: %v", result2.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result2.Iterations)
		fmt.Printf("Final response: %s\n\n", result2.Response.Content)
		fmt.Printf("Total tokens used: %d\n\n", result2.TotalTokens)
	}

	// Example 3: List deployments with label selector
	fmt.Println("\n=== Example 3: List deployments with label selector ===")
	ctx3, cancel3 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel3()

	result3 := a.Run(ctx3, "List all deployments in the default namespace with label 'app=nginx' and tell me their replica counts", nil)
	if result3.Error != nil {
		log.Printf("Agent error: %v", result3.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result3.Iterations)
		fmt.Printf("Final response: %s\n\n", result3.Response.Content)
		fmt.Printf("Total tokens used: %d\n\n", result3.TotalTokens)
	}

	// Example 4: Check node health
	fmt.Println("\n=== Example 4: Check node health ===")
	ctx4, cancel4 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel4()

	result4 := a.Run(ctx4, "List all nodes in the cluster and tell me which ones are ready and their CPU/memory capacity", nil)
	if result4.Error != nil {
		log.Printf("Agent error: %v", result4.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result4.Iterations)
		fmt.Printf("Final response: %s\n\n", result4.Response.Content)
		fmt.Printf("Total tokens used: %d\n\n", result4.TotalTokens)
	}

	// Example 5: Troubleshooting scenario - find failing pods
	fmt.Println("\n=== Example 5: Troubleshooting - Find failing pods ===")
	ctx5, cancel5 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel5()

	result5 := a.Run(ctx5, "List all pods across all namespaces and identify any that are not in 'Running' status. Tell me their names, namespaces, and current status", nil)
	if result5.Error != nil {
		log.Printf("Agent error: %v", result5.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result5.Iterations)
		fmt.Printf("Final response: %s\n\n", result5.Response.Content)
		fmt.Printf("Total tokens used: %d\n\n", result5.TotalTokens)
	}

	fmt.Println("\n=== All examples completed ===")
	fmt.Println("\nNote: These examples require a valid Kubernetes cluster connection.")
	fmt.Println("Ensure your KUBECONFIG is set or ~/.kube/config is configured.")
}
