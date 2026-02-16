// Package shell provides a built-in shell execution tool for agents.
// It supports running shell commands with safety constraints, output capture, and timeout control.
package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
)

// Config configures the shell tool behavior.
type Config struct {
	// DefaultTimeout is the default timeout for command execution.
	// If not set, defaults to 30 seconds.
	DefaultTimeout time.Duration

	// MaxOutputSize limits the command output size to prevent memory issues.
	// If not set, defaults to 1MB.
	MaxOutputSize int

	// AllowedCommands restricts which commands can be executed.
	// If empty, all commands are allowed (use with caution).
	AllowedCommands []string

	// BlockedCommands prevents specific commands from running.
	// Takes precedence over AllowedCommands.
	BlockedCommands []string

	// DefaultWorkingDir is the default working directory for commands.
	// If empty, uses the current process working directory.
	DefaultWorkingDir string
}

// Params defines the parameters for shell command execution.
type Params struct {
	Command    string            `json:"command"`
	WorkingDir string            `json:"working_dir"`
	Env        map[string]string `json:"env"`
	Timeout    int               `json:"timeout"` // in seconds
}

// Response represents the result of a shell command execution.
type Response struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	Error      string `json:"error,omitempty"`
	Truncated  bool   `json:"truncated,omitempty"`
	WorkingDir string `json:"working_dir"`
}

// NewTool creates a new shell execution tool definition.
// The tool supports running shell commands with configurable working directory,
// environment variables, and timeout.
func NewTool() types.Tool {
	return tools.NewBuilder(
		"shell_exec",
		"Execute shell commands with output capture. "+
			"Supports configurable working directory, environment variables, and timeout. "+
			"Use for system operations, file manipulation, and running CLI tools.",
	).
		StringParam("command", "The shell command to execute (e.g., 'ls -la', 'echo hello')", true).
		StringParam(
			"working_dir",
			"Optional working directory for command execution (defaults to current directory)",
			false,
		).
		ObjectParam(
			"env",
			"Optional environment variables to set "+
				`(e.g., {"PATH": "/usr/bin", "DEBUG": "true"})`,
			false,
			map[string]any{},
			[]string{},
		).
		IntegerParam("timeout", "Optional timeout in seconds (default: 30, max: 300)", false).
		Build()
}

// NewHandler creates a new shell tool handler with the given configuration.
// If config is nil, default values are used with minimal safety restrictions.
func NewHandler(config *Config) tools.Handler {
	if config == nil {
		config = &Config{}
	}
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 30 * time.Second
	}
	if config.MaxOutputSize == 0 {
		config.MaxOutputSize = 1024 * 1024 // 1MB
	}

	// Default blocked commands for safety
	if len(config.BlockedCommands) == 0 {
		config.BlockedCommands = []string{
			"rm -rf /",
			"mkfs",
			"dd",
			":(){ :|:& };:", // fork bomb
		}
	}

	return func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		// Parse parameters
		var params Params
		if err := types.ParseToolArguments(call, &params); err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("invalid parameters: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		// Validate command is not empty
		if strings.TrimSpace(params.Command) == "" {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         "command cannot be empty",
				ExecutionTime: time.Since(start),
			}
		}

		// Check blocked commands
		for _, blocked := range config.BlockedCommands {
			if strings.Contains(params.Command, blocked) {
				return types.ToolResult{
					ToolCallID:    call.ID,
					ToolName:      call.Function.Name,
					Error:         fmt.Sprintf("command blocked for safety: contains '%s'", blocked),
					ExecutionTime: time.Since(start),
				}
			}
		}

		// Check allowed commands if configured
		if len(config.AllowedCommands) > 0 {
			allowed := false
			cmdParts := strings.Fields(params.Command)
			if len(cmdParts) > 0 {
				baseCmd := cmdParts[0]
				for _, allowedCmd := range config.AllowedCommands {
					if baseCmd == allowedCmd || strings.HasPrefix(baseCmd, allowedCmd+"/") {
						allowed = true
						break
					}
				}
			}
			if !allowed {
				return types.ToolResult{
					ToolCallID:    call.ID,
					ToolName:      call.Function.Name,
					Error:         "command not in allowed list",
					ExecutionTime: time.Since(start),
				}
			}
		}

		// Determine timeout
		timeout := config.DefaultTimeout
		if params.Timeout > 0 {
			if params.Timeout > 300 {
				params.Timeout = 300 // max 5 minutes
			}
			timeout = time.Duration(params.Timeout) * time.Second
		}

		// Create command context with timeout
		cmdCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Create command using shell for proper shell parsing
		// #nosec G204 - Command execution is the intended purpose of this tool with safety constraints
		cmd := exec.CommandContext(cmdCtx, "sh", "-c", params.Command)

		// Set working directory
		workingDir := params.WorkingDir
		if workingDir == "" && config.DefaultWorkingDir != "" {
			workingDir = config.DefaultWorkingDir
		}
		if workingDir != "" {
			cmd.Dir = workingDir
		}

		// Set environment variables
		if len(params.Env) > 0 {
			// Start with current environment
			cmd.Env = append([]string{}, cmd.Environ()...)
			// Add/override with provided variables
			for key, value := range params.Env {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
			}
		}

		// Capture stdout and stderr
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// Execute command
		err := cmd.Run()

		// Get actual working directory used
		actualWorkingDir := workingDir
		if actualWorkingDir == "" {
			actualWorkingDir = "." // current directory
		}

		// Determine exit code
		exitCode := 0
		var errMsg string
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				// Non-exit error (e.g., command not found, timeout)
				exitCode = -1
				errMsg = err.Error()
			}
		}

		// Check for context cancellation
		if cmdCtx.Err() != nil {
			errMsg = fmt.Sprintf("command timeout after %v", timeout)
			exitCode = -1
		}

		// Truncate output if needed
		stdoutStr := stdout.String()
		stderrStr := stderr.String()
		truncated := false

		if len(stdoutStr) > config.MaxOutputSize {
			stdoutStr = stdoutStr[:config.MaxOutputSize] + "\n... (output truncated)"
			truncated = true
		}
		if len(stderrStr) > config.MaxOutputSize {
			stderrStr = stderrStr[:config.MaxOutputSize] + "\n... (output truncated)"
			truncated = true
		}

		// Create response object
		resp := Response{
			Stdout:     stdoutStr,
			Stderr:     stderrStr,
			ExitCode:   exitCode,
			Error:      errMsg,
			Truncated:  truncated,
			WorkingDir: actualWorkingDir,
		}

		// Marshal response to JSON
		resultJSON, err := json.Marshal(resp)
		if err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("failed to marshal response: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		return types.ToolResult{
			ToolCallID:    call.ID,
			ToolName:      call.Function.Name,
			Content:       string(resultJSON),
			ExecutionTime: time.Since(start),
		}
	}
}
