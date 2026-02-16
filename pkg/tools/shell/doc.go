// Package shell provides a built-in shell execution tool for GoAgent.
//
// The shell tool allows agents to execute shell commands with configurable
// safety constraints, timeout control, and output capture. It's useful for
// system operations, file manipulation, and running CLI tools.
//
// # Basic Usage
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//	    "log"
//	    "time"
//
//	    "github.com/oskarhane/goagent/pkg/agent"
//	    "github.com/oskarhane/goagent/pkg/providers/openai"
//	    "github.com/oskarhane/goagent/pkg/tools"
//	    "github.com/oskarhane/goagent/pkg/tools/shell"
//	)
//
//	func main() {
//	    // Create provider
//	    provider, _ := openai.NewProvider(openai.Config{
//	        APIKey: "your-api-key",
//	    })
//
//	    // Create tool registry with shell tool
//	    registry := tools.NewRegistry()
//
//	    // Register shell tool with safety configuration
//	    shellTool := shell.NewTool()
//	    shellHandler := shell.NewHandler(&shell.Config{
//	        DefaultTimeout:    30 * time.Second,
//	        MaxOutputSize:     1024 * 1024, // 1MB
//	        AllowedCommands:   []string{"ls", "cat", "grep", "echo", "date"},
//	        DefaultWorkingDir: "/tmp",
//	    })
//	    registry.MustRegister(shellTool, shellHandler)
//
//	    // Create and use agent
//	    a, _ := agent.NewAgent(agent.Config{
//	        Provider: provider,
//	        Registry: registry,
//	    })
//
//	    result := a.Run(context.Background(), "List all files in the current directory", nil)
//	    fmt.Println(result.Response.Content)
//	}
//
// # Safety Considerations
//
// The shell tool includes several safety features:
//
//   - Command timeout (default 30s, max 5 minutes)
//   - Output size limiting to prevent memory exhaustion
//   - Blocked commands list (e.g., "rm -rf /", "mkfs", fork bombs)
//   - Optional allowed commands whitelist
//   - Working directory constraints
//
// For production use, always configure AllowedCommands to limit
// what the agent can execute:
//
//	config := &shell.Config{
//	    AllowedCommands: []string{"git", "npm", "docker"},
//	    BlockedCommands: []string{"rm -rf", "sudo", "curl | sh"},
//	}
//
// # Environment Variables
//
// You can pass environment variables to commands:
//
//	// Agent will receive this capability
//	"Execute 'printenv DEBUG' with DEBUG=true environment variable set"
//
// The agent can use the env parameter to set variables like:
//
//	{
//	    "command": "printenv DEBUG",
//	    "env": {"DEBUG": "true"}
//	}
//
// # Working Directory
//
// Commands can run in specific directories:
//
//	{
//	    "command": "ls -la",
//	    "working_dir": "/var/log"
//	}
//
// If DefaultWorkingDir is set in Config, it will be used when
// working_dir is not specified in the command parameters.
package shell
