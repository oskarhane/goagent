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
	"github.com/oskarhane/goagent/pkg/tools/shell"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create provider
	provider, err := openai.NewProvider(&openai.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Create tool registry with shell tool
	registry := tools.NewRegistry()

	// Register the shell tool with safety configuration
	shellTool := shell.NewTool()
	shellHandler := shell.NewHandler(&shell.Config{
		DefaultTimeout: 30 * time.Second,
		MaxOutputSize:  1024 * 1024, // 1MB
		// Allow common safe commands for demonstration
		AllowedCommands: []string{
			"ls", "cat", "echo", "pwd", "date", "whoami",
			"grep", "wc", "head", "tail", "find", "which",
			"uname", "hostname", "printenv",
		},
		// Additional blocked commands for safety
		BlockedCommands: []string{
			"rm -rf /",
			"mkfs",
			"dd",
			":(){ :|:& };:", // fork bomb
			"sudo",
			"su",
		},
	})
	registry.MustRegister(shellTool, shellHandler)

	// Create agent
	a, err := agent.NewAgent(&agent.Config{
		Provider: provider,
		Registry: registry,
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Example 1: Simple command execution
	fmt.Println("=== Example 1: Get current date and time ===")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel1()

	result1 := a.Run(ctx1, "What is the current date and time on this system?", nil)
	if result1.Error != nil {
		log.Printf("Agent error: %v", result1.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result1.Iterations)
		fmt.Printf("Final response: %s\n\n", result1.Response.Content)
	}

	// Example 2: Working directory usage
	fmt.Println("\n=== Example 2: List files in /tmp directory ===")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel2()

	result2 := a.Run(ctx2, "List all files in the /tmp directory. Tell me how many files are there.", nil)
	if result2.Error != nil {
		log.Printf("Agent error: %v", result2.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result2.Iterations)
		fmt.Printf("Final response: %s\n\n", result2.Response.Content)
	}

	// Example 3: Environment variables
	fmt.Println("\n=== Example 3: Using environment variables ===")
	ctx3, cancel3 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel3()

	result3 := a.Run(ctx3, "Run the command 'printenv MY_VAR' with environment variable MY_VAR set to 'Hello from GoAgent'. What does it output?", nil)
	if result3.Error != nil {
		log.Printf("Agent error: %v", result3.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result3.Iterations)
		fmt.Printf("Final response: %s\n\n", result3.Response.Content)
	}

	// Example 4: Command chaining
	fmt.Println("\n=== Example 4: System information gathering ===")
	ctx4, cancel4 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel4()

	result4 := a.Run(ctx4, "Tell me what operating system this is and what the hostname is.", nil)
	if result4.Error != nil {
		log.Printf("Agent error: %v", result4.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result4.Iterations)
		fmt.Printf("Final response: %s\n\n", result4.Response.Content)
	}

	// Example 5: Error handling - blocked command
	fmt.Println("\n=== Example 5: Safety - blocked command ===")
	ctx5, cancel5 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel5()

	result5 := a.Run(ctx5, "Try to run 'sudo whoami' and tell me what happens.", nil)
	if result5.Error != nil {
		log.Printf("Agent error: %v", result5.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result5.Iterations)
		fmt.Printf("Final response: %s\n\n", result5.Response.Content)
	}

	// Example 6: Multi-step operation
	fmt.Println("\n=== Example 6: Multi-step file analysis ===")
	ctx6, cancel6 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel6()

	result6 := a.Run(ctx6, "Find the go.mod file in the current directory and tell me what Go version it specifies.", nil)
	if result6.Error != nil {
		log.Printf("Agent error: %v", result6.Error)
	} else {
		fmt.Printf("Agent completed in %d iterations\n", result6.Iterations)
		fmt.Printf("Final response: %s\n\n", result6.Response.Content)
	}

	fmt.Println("\n=== All examples completed ===")
}
