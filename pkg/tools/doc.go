// Package tools provides a comprehensive tool registration and execution
// framework for AI agents. Tools enable agents to interact with external
// systems, APIs, databases, and execute code during their reasoning loop.
//
// # Core Concepts
//
// A Tool consists of three parts:
//   - Definition: Name, description, and JSON Schema for parameters
//   - Handler: Go function that executes the tool logic
//   - Registry: Central location where tools are registered and looked up
//
// # Basic Usage
//
// Create a tool using the Builder:
//
//	weatherTool := tools.NewBuilder(
//	    "get_weather",
//	    "Get current weather for a location",
//	).
//	    StringParam("location", "City name", true).
//	    StringParam("units", "celsius or fahrenheit", false).
//	    Build()
//
// Implement the handler:
//
//	handler := func(ctx context.Context, call types.ToolCall) types.ToolResult {
//	    // Parse parameters
//	    var params struct {
//	        Location string `json:"location"`
//	        Units    string `json:"units"`
//	    }
//	    if err := types.ParseToolArguments(call, &params); err != nil {
//	        return types.ToolResult{
//	            ToolCallID: call.ID,
//	            ToolName:   call.Function.Name,
//	            Error:      err.Error(),
//	        }
//	    }
//
//	    // Execute tool logic
//	    weather := getWeather(params.Location, params.Units)
//
//	    // Return result
//	    return types.ToolResult{
//	        ToolCallID: call.ID,
//	        ToolName:   call.Function.Name,
//	        Content:    weather,
//	    }
//	}
//
// Register with a registry:
//
//	registry := tools.NewRegistry()
//	err := registry.Register(weatherTool, handler)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Parameter Validation
//
// The tool system validates parameters against JSON Schema before execution.
// Validation covers:
//   - Type checking (string, number, integer, boolean, object, array)
//   - Required field enforcement
//   - Nested object validation
//   - Array element validation
//
// Validate explicitly using:
//
//	if err := tools.ValidateParameters(tool, call); err != nil {
//	    // Handle validation error
//	}
//
// # Thread Safety
//
// The Registry is thread-safe and can be used concurrently from multiple
// goroutines. Tools can be registered, unregistered, and executed safely
// in parallel.
//
// # Error Handling
//
// Handlers should return ToolResult with the Error field set on failure:
//
//	return types.ToolResult{
//	    ToolCallID: call.ID,
//	    ToolName:   call.Function.Name,
//	    Error:      "failed to fetch weather: " + err.Error(),
//	    Content:    "",  // Optional error context
//	}
//
// The agent will see the error and can reason about how to recover.
//
// # Built-in Tools
//
// The SDK includes several built-in tools:
//   - HTTP: Make REST API calls
//   - Shell: Execute shell commands
//   - Kubernetes: Query Kubernetes resources
//
// See their respective packages for usage details.
package tools
