package tools

import (
	"github.com/oskarhane/goagent/pkg/types"
)

// Builder provides a fluent interface for constructing tool definitions.
// It simplifies creating tools with proper JSON Schema validation.
//
// Example:
//
//	tool := NewBuilder("get_weather", "Get current weather for a location").
//	    StringParam("location", "City name", true).
//	    StringParam("units", "Temperature units (celsius/fahrenheit)", false).
//	    Build()
type Builder struct {
	name        string
	description string
	properties  map[string]any
	required    []string
}

// NewBuilder creates a new tool builder with the given name and description.
func NewBuilder(name, description string) *Builder {
	return &Builder{
		name:        name,
		description: description,
		properties:  make(map[string]any),
		required:    make([]string, 0),
	}
}

// StringParam adds a string parameter to the tool.
// If required is true, the parameter must be provided when calling the tool.
func (b *Builder) StringParam(name, description string, required bool) *Builder {
	b.properties[name] = map[string]any{
		"type":        "string",
		"description": description,
	}
	if required {
		b.required = append(b.required, name)
	}
	return b
}

// StringParamWithEnum adds a string parameter with allowed values.
func (b *Builder) StringParamWithEnum(name, description string, required bool, enum []string) *Builder {
	b.properties[name] = map[string]any{
		"type":        "string",
		"description": description,
		"enum":        enum,
	}
	if required {
		b.required = append(b.required, name)
	}
	return b
}

// NumberParam adds a number parameter to the tool.
func (b *Builder) NumberParam(name, description string, required bool) *Builder {
	b.properties[name] = map[string]any{
		"type":        "number",
		"description": description,
	}
	if required {
		b.required = append(b.required, name)
	}
	return b
}

// IntegerParam adds an integer parameter to the tool.
func (b *Builder) IntegerParam(name, description string, required bool) *Builder {
	b.properties[name] = map[string]any{
		"type":        "integer",
		"description": description,
	}
	if required {
		b.required = append(b.required, name)
	}
	return b
}

// BooleanParam adds a boolean parameter to the tool.
func (b *Builder) BooleanParam(name, description string, required bool) *Builder {
	b.properties[name] = map[string]any{
		"type":        "boolean",
		"description": description,
	}
	if required {
		b.required = append(b.required, name)
	}
	return b
}

// ObjectParam adds an object parameter to the tool.
// The properties map should contain the nested schema definition.
func (b *Builder) ObjectParam(
	name, description string,
	required bool,
	properties map[string]any,
	requiredFields []string,
) *Builder {
	schema := map[string]any{
		"type":        "object",
		"description": description,
		"properties":  properties,
	}
	if len(requiredFields) > 0 {
		schema["required"] = requiredFields
	}
	b.properties[name] = schema
	if required {
		b.required = append(b.required, name)
	}
	return b
}

// ArrayParam adds an array parameter to the tool.
// The items schema defines what each array element should look like.
func (b *Builder) ArrayParam(name, description string, required bool, items map[string]any) *Builder {
	b.properties[name] = map[string]any{
		"type":        "array",
		"description": description,
		"items":       items,
	}
	if required {
		b.required = append(b.required, name)
	}
	return b
}

// ArrayOfStrings is a convenience method for adding a string array parameter.
func (b *Builder) ArrayOfStrings(name, description string, required bool) *Builder {
	return b.ArrayParam(name, description, required, map[string]any{
		"type": "string",
	})
}

// ArrayOfObjects is a convenience method for adding an object array parameter.
func (b *Builder) ArrayOfObjects(
	name, description string,
	required bool,
	objectProperties map[string]any,
	objectRequired []string,
) *Builder {
	items := map[string]any{
		"type":       "object",
		"properties": objectProperties,
	}
	if len(objectRequired) > 0 {
		items["required"] = objectRequired
	}
	return b.ArrayParam(name, description, required, items)
}

// Build creates the final Tool definition.
// This method should be called after adding all parameters.
func (b *Builder) Build() types.Tool {
	parameters := map[string]any{
		"type":       "object",
		"properties": b.properties,
	}

	if len(b.required) > 0 {
		parameters["required"] = b.required
	}

	return types.Tool{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        b.name,
			Description: b.description,
			Parameters:  parameters,
		},
	}
}
