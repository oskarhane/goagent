package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder("test_tool", "A test tool")

	assert.NotNil(t, builder)
	assert.Equal(t, "test_tool", builder.name)
	assert.Equal(t, "A test tool", builder.description)
}

func TestBuilder_StringParam(t *testing.T) {
	tool := NewBuilder("test", "desc").
		StringParam("name", "Name parameter", true).
		Build()

	assert.Equal(t, "test", tool.Function.Name)
	assert.Equal(t, "desc", tool.Function.Description)

	// Check parameter schema
	params := tool.Function.Parameters
	assert.Equal(t, "object", params["type"])

	props, ok := params["properties"].(map[string]any)
	require.True(t, ok)

	nameParam, ok := props["name"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", nameParam["type"])
	assert.Equal(t, "Name parameter", nameParam["description"])

	// Check required
	required, ok := params["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "name")
}

func TestBuilder_IntegerParam(t *testing.T) {
	tool := NewBuilder("test", "desc").
		IntegerParam("count", "Count parameter", false).
		Build()

	props, ok := tool.Function.Parameters["properties"].(map[string]any)
	require.True(t, ok)

	countParam, ok := props["count"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "integer", countParam["type"])

	// Not required
	required := tool.Function.Parameters["required"]
	assert.Nil(t, required)
}

func TestBuilder_NumberParam(t *testing.T) {
	tool := NewBuilder("test", "desc").
		NumberParam("price", "Price parameter", true).
		Build()

	props, ok := tool.Function.Parameters["properties"].(map[string]any)
	require.True(t, ok)

	priceParam, ok := props["price"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "number", priceParam["type"])

	required, ok := tool.Function.Parameters["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "price")
}

func TestBuilder_BooleanParam(t *testing.T) {
	tool := NewBuilder("test", "desc").
		BooleanParam("enabled", "Enabled flag", false).
		Build()

	props, ok := tool.Function.Parameters["properties"].(map[string]any)
	require.True(t, ok)

	enabledParam, ok := props["enabled"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "boolean", enabledParam["type"])
}

func TestBuilder_ArrayParam(t *testing.T) {
	itemSchema := map[string]any{"type": "string"}
	tool := NewBuilder("test", "desc").
		ArrayParam("items", "List of items", true, itemSchema).
		Build()

	props, ok := tool.Function.Parameters["properties"].(map[string]any)
	require.True(t, ok)

	itemsParam, ok := props["items"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "array", itemsParam["type"])
	assert.Equal(t, itemSchema, itemsParam["items"])
}

func TestBuilder_ObjectParam(t *testing.T) {
	objProps := map[string]any{
		"key": map[string]any{"type": "string"},
	}
	tool := NewBuilder("test", "desc").
		ObjectParam("config", "Configuration object", false, objProps, nil).
		Build()

	props, ok := tool.Function.Parameters["properties"].(map[string]any)
	require.True(t, ok)

	configParam, ok := props["config"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "object", configParam["type"])
}

func TestBuilder_StringParamWithEnum(t *testing.T) {
	tool := NewBuilder("test", "desc").
		StringParamWithEnum("method", "HTTP method", true, []string{"GET", "POST", "PUT"}).
		Build()

	props, ok := tool.Function.Parameters["properties"].(map[string]any)
	require.True(t, ok)

	methodParam, ok := props["method"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", methodParam["type"])

	enum, ok := methodParam["enum"].([]string)
	require.True(t, ok)
	assert.Equal(t, []string{"GET", "POST", "PUT"}, enum)
}

func TestBuilder_MultipleParams(t *testing.T) {
	tool := NewBuilder("complex_tool", "A complex tool").
		StringParam("name", "Name", true).
		IntegerParam("age", "Age", false).
		BooleanParam("active", "Active status", true).
		StringParamWithEnum("role", "User role", false, []string{"admin", "user"}).
		Build()

	assert.Equal(t, "complex_tool", tool.Function.Name)
	assert.Equal(t, "function", tool.Type)

	props, ok := tool.Function.Parameters["properties"].(map[string]any)
	require.True(t, ok)
	assert.Len(t, props, 4) // 4 parameters

	required, ok := tool.Function.Parameters["required"].([]string)
	require.True(t, ok)
	assert.Len(t, required, 2) // name and active are required
	assert.Contains(t, required, "name")
	assert.Contains(t, required, "active")
}

func TestBuilder_BuildWithNoParams(t *testing.T) {
	tool := NewBuilder("simple_tool", "A simple tool with no params").
		Build()

	assert.Equal(t, "simple_tool", tool.Function.Name)
	assert.Equal(t, "function", tool.Type)

	params := tool.Function.Parameters
	assert.Equal(t, "object", params["type"])

	props, ok := params["properties"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, props)
}

func TestBuilder_ChainedCalls(t *testing.T) {
	// Test that builder pattern allows chaining
	tool := NewBuilder("chain", "Chained tool").
		StringParam("a", "A", true).
		StringParam("b", "B", true).
		StringParam("c", "C", false).
		Build()

	required, ok := tool.Function.Parameters["required"].([]string)
	require.True(t, ok)
	assert.Len(t, required, 2)
}
