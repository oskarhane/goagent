// This example directly tests the shell tool without requiring an LLM provider
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/tools/shell"
	"github.com/oskarhane/goagent/pkg/types"
)

func main() {
	fmt.Println("=== Direct Shell Tool Testing ===")

	// Create tool registry with shell tool
	registry := tools.NewRegistry()

	// Register the shell tool with safety configuration
	shellTool := shell.NewTool()
	shellHandler := shell.NewHandler(&shell.Config{
		DefaultTimeout: 10 * time.Second,
		MaxOutputSize:  1024 * 1024, // 1MB
		AllowedCommands: []string{
			"echo", "date", "pwd", "ls", "whoami", "uname",
		},
		BlockedCommands: []string{
			"rm -rf /",
			"sudo",
		},
	})
	registry.MustRegister(shellTool, shellHandler)

	// Test 1: Simple echo command
	fmt.Println("Test 1: Simple echo command")
	testCommand(registry, "echo 'Hello from GoAgent'", "", nil)

	// Test 2: Get current date
	fmt.Println("\nTest 2: Get current date")
	testCommand(registry, "date", "", nil)

	// Test 3: Environment variable
	fmt.Println("\nTest 3: Environment variable")
	testCommand(registry, "echo $TEST_VAR", "", map[string]string{"TEST_VAR": "success"})

	// Test 4: Working directory
	fmt.Println("\nTest 4: Working directory")
	testCommand(registry, "pwd", "/tmp", nil)

	// Test 5: Blocked command
	fmt.Println("\nTest 5: Blocked command (should fail)")
	testCommand(registry, "sudo whoami", "", nil)

	// Test 6: Command not in allowlist
	fmt.Println("\nTest 6: Command not in allowlist (should fail)")
	testCommand(registry, "curl http://example.com", "", nil)

	// Test 7: Exit code handling
	fmt.Println("\nTest 7: Exit code handling (command that fails)")
	testCommand(registry, "ls /nonexistent-directory-xyz", "", nil)

	fmt.Println("\n=== All tests completed ===")
}

func testCommand(registry *tools.Registry, command, workingDir string, env map[string]string) {
	// Build arguments
	args := map[string]any{
		"command": command,
	}
	if workingDir != "" {
		args["working_dir"] = workingDir
	}
	if env != nil {
		args["env"] = env
	}

	argsJSON, _ := json.Marshal(args)

	// Create tool call
	call := types.ToolCall{
		ID: "test-call-1",
		Function: types.FunctionCall{
			Name:      "shell_exec",
			Arguments: string(argsJSON),
		},
	}

	// Execute tool
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := registry.Execute(ctx, call)

	// Parse response
	var response shell.Response
	if result.Content != "" {
		if err := json.Unmarshal([]byte(result.Content), &response); err != nil {
			fmt.Printf("  Error parsing response: %v\n", err)
			return
		}
	}

	// Print results
	fmt.Printf("  Command: %s\n", command)
	if result.Error != "" {
		fmt.Printf("  Tool Error: %s\n", result.Error)
	} else {
		fmt.Printf("  Exit Code: %d\n", response.ExitCode)
		if response.Stdout != "" {
			fmt.Printf("  Stdout: %s\n", response.Stdout)
		}
		if response.Stderr != "" {
			fmt.Printf("  Stderr: %s\n", response.Stderr)
		}
		if response.Error != "" {
			fmt.Printf("  Execution Error: %s\n", response.Error)
		}
		if response.WorkingDir != "" {
			fmt.Printf("  Working Dir: %s\n", response.WorkingDir)
		}
	}
	fmt.Printf("  Execution Time: %v\n", result.ExecutionTime)
}
