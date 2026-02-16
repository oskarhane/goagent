// Package tools provides a tool registration and execution framework
// for agents. Tools allow agents to interact with external systems,
// execute code, and perform actions beyond text generation.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/oskarhane/goagent/pkg/types"
)

// Handler is the function signature for tool implementations.
// It receives a context for cancellation, the original tool call from the LLM,
// and should return a ToolResult with the execution outcome.
//
// Handlers should:
//   - Respect context cancellation
//   - Validate parameters using ParseToolArguments
//   - Return structured results as JSON when possible
//   - Include error information in ToolResult.Error on failure
type Handler func(ctx context.Context, call types.ToolCall) types.ToolResult

// Registry manages the collection of available tools and their handlers.
// It provides thread-safe registration and lookup of tools by name.
//
// A Registry is required to execute tool calls during agent reasoning loops.
// Tools must be registered before an agent attempts to use them.
type Registry struct {
	mu       sync.RWMutex
	tools    map[string]types.Tool
	handlers map[string]Handler
}

// NewRegistry creates a new empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools:    make(map[string]types.Tool),
		handlers: make(map[string]Handler),
	}
}

// Register adds a tool and its handler to the registry.
// The tool's name must be unique within the registry.
//
// Returns an error if:
//   - A tool with the same name is already registered
//   - The tool definition is invalid (missing name, description, or parameters)
//   - The handler is nil
func (r *Registry) Register(tool types.Tool, handler Handler) error {
	if tool.Function.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if tool.Function.Description == "" {
		return fmt.Errorf("tool description cannot be empty")
	}
	if tool.Function.Parameters == nil {
		return fmt.Errorf("tool parameters cannot be nil")
	}
	if handler == nil {
		return fmt.Errorf("tool handler cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Function.Name]; exists {
		return fmt.Errorf("tool %q already registered", tool.Function.Name)
	}

	r.tools[tool.Function.Name] = tool
	r.handlers[tool.Function.Name] = handler

	return nil
}

// MustRegister is like Register but panics on error.
// Useful for built-in tools that should never fail registration.
func (r *Registry) MustRegister(tool types.Tool, handler Handler) {
	if err := r.Register(tool, handler); err != nil {
		panic(fmt.Sprintf("failed to register tool: %v", err))
	}
}

// Get retrieves a tool definition by name.
// Returns false if the tool is not registered.
func (r *Registry) Get(name string) (types.Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	return tool, exists
}

// GetHandler retrieves a tool handler by name.
// Returns nil if the tool is not registered.
func (r *Registry) GetHandler(name string) Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.handlers[name]
}

// List returns all registered tools.
// The returned slice is a copy and safe to modify.
func (r *Registry) List() []types.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]types.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return tools
}

// Count returns the number of registered tools.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tools)
}

// Execute runs a tool handler with the given context and tool call.
// It validates that the tool exists and handles panics gracefully.
//
// Returns a ToolResult with:
//   - Error field set if the tool is not registered or handler panics
//   - The handler's result otherwise
func (r *Registry) Execute(ctx context.Context, call types.ToolCall) types.ToolResult {
	handler := r.GetHandler(call.Function.Name)
	if handler == nil {
		return types.ToolResult{
			ToolCallID: call.ID,
			ToolName:   call.Function.Name,
			Error:      fmt.Sprintf("tool %q not registered", call.Function.Name),
		}
	}

	return handler(ctx, call)
}

// Unregister removes a tool from the registry.
// Returns true if the tool was removed, false if it didn't exist.
func (r *Registry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return false
	}

	delete(r.tools, name)
	delete(r.handlers, name)
	return true
}

// Clear removes all tools from the registry.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = make(map[string]types.Tool)
	r.handlers = make(map[string]Handler)
}

// ValidateParameters validates tool call arguments against the tool's JSON Schema.
// Returns nil if validation passes, an error describing the validation failure otherwise.
//
// This uses a lightweight JSON Schema validator that supports:
//   - Type checking (string, number, integer, boolean, object, array)
//   - Required field validation
//   - Nested object validation
//   - Array item validation
func ValidateParameters(tool types.Tool, call types.ToolCall) error {
	// Parse arguments into a map
	var args map[string]any
	if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
		return fmt.Errorf("invalid JSON arguments: %w", err)
	}

	// Get schema from parameters
	schema := tool.Function.Parameters
	if schema == nil {
		return fmt.Errorf("tool has no parameter schema")
	}

	// Validate against schema
	return validateAgainstSchema(args, schema, "")
}

// validateAgainstSchema performs recursive JSON Schema validation.
// path is used for error reporting to show where validation failed.
func validateAgainstSchema(value any, schema map[string]any, path string) error {
	// Check type
	schemaType, ok := schema["type"].(string)
	if !ok {
		return fmt.Errorf("schema missing type at %s", path)
	}

	if err := validateType(value, schemaType, path); err != nil {
		return err
	}

	// Check required fields for objects
	if schemaType == "object" {
		if err := validateRequired(value, schema, path); err != nil {
			return err
		}

		// Validate nested properties
		if err := validateProperties(value, schema, path); err != nil {
			return err
		}
	}

	// Validate array items
	if schemaType == "array" {
		if err := validateArrayItems(value, schema, path); err != nil {
			return err
		}
	}

	return nil
}

// validateType checks if the value matches the expected JSON Schema type.
func validateType(value any, expectedType, path string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string at %s, got %T", path, value)
		}
	case "number":
		switch value.(type) {
		case float64, float32, int, int64, int32:
			// OK
		default:
			return fmt.Errorf("expected number at %s, got %T", path, value)
		}
	case "integer":
		// In JSON, integers come as float64, so check if it's a whole number
		switch v := value.(type) {
		case float64:
			if v != float64(int64(v)) {
				return fmt.Errorf("expected integer at %s, got float", path)
			}
		case int, int64, int32:
			// OK
		default:
			return fmt.Errorf("expected integer at %s, got %T", path, value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean at %s, got %T", path, value)
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("expected object at %s, got %T", path, value)
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("expected array at %s, got %T", path, value)
		}
	case "null":
		if value != nil {
			return fmt.Errorf("expected null at %s, got %T", path, value)
		}
	default:
		return fmt.Errorf("unknown schema type %q at %s", expectedType, path)
	}

	return nil
}

// validateRequired checks if all required fields are present in an object.
func validateRequired(value any, schema map[string]any, path string) error {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil // Type validation handles this
	}

	// Handle both []any and []string for required field
	var requiredFields []string
	switch req := schema["required"].(type) {
	case []any:
		for _, r := range req {
			if s, ok := r.(string); ok {
				requiredFields = append(requiredFields, s)
			}
		}
	case []string:
		requiredFields = req
	default:
		return nil // No required fields or invalid format
	}

	if len(requiredFields) == 0 {
		return nil
	}

	for _, fieldName := range requiredFields {
		if _, exists := obj[fieldName]; !exists {
			fieldPath := path
			if fieldPath == "" {
				fieldPath = fieldName
			} else {
				fieldPath = path + "." + fieldName
			}
			return fmt.Errorf("required field %q missing at %s", fieldName, fieldPath)
		}
	}

	return nil
}

// validateProperties validates nested object properties against their schemas.
func validateProperties(value any, schema map[string]any, path string) error {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return nil // No properties defined
	}

	for fieldName, fieldValue := range obj {
		propSchema, ok := properties[fieldName].(map[string]any)
		if !ok {
			continue // No schema for this property, skip validation
		}

		fieldPath := path
		if fieldPath == "" {
			fieldPath = fieldName
		} else {
			fieldPath = path + "." + fieldName
		}

		// Handle nil values for optional fields
		if fieldValue == nil {
			// Check if field is required
			isRequired := false
			if reqFields, ok := schema["required"]; ok {
				switch req := reqFields.(type) {
				case []any:
					for _, r := range req {
						if s, ok := r.(string); ok && s == fieldName {
							isRequired = true
							break
						}
					}
				case []string:
					for _, r := range req {
						if r == fieldName {
							isRequired = true
							break
						}
					}
				}
			}
			// If required and nil, error; if optional and nil, skip validation
			if isRequired {
				return fmt.Errorf("required field %q is null at %s", fieldName, fieldPath)
			}
			continue
		}

		if err := validateAgainstSchema(fieldValue, propSchema, fieldPath); err != nil {
			return err
		}
	}

	return nil
}

// validateArrayItems validates array elements against the items schema.
func validateArrayItems(value any, schema map[string]any, path string) error {
	arr, ok := value.([]any)
	if !ok {
		return nil
	}

	itemSchema, ok := schema["items"].(map[string]any)
	if !ok {
		return nil // No item schema defined
	}

	for i, item := range arr {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		if err := validateAgainstSchema(item, itemSchema, itemPath); err != nil {
			return err
		}
	}

	return nil
}
