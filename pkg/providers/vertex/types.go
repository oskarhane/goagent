package vertex

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oskarhane/goagent/pkg/types"
)

// vertexRequest matches the Vertex AI Gemini API request format.
type vertexRequest struct {
	Contents          []vertexContent         `json:"contents"`
	SystemInstruction *vertexContent          `json:"systemInstruction,omitempty"`
	Tools             []vertexTool            `json:"tools,omitempty"`
	GenerationConfig  *vertexGenerationConfig `json:"generationConfig,omitempty"`
}

// vertexContent represents a message in Vertex AI format.
type vertexContent struct {
	Role  string       `json:"role"`
	Parts []vertexPart `json:"parts"`
}

// vertexPart represents a part of a message (text or function call/response).
type vertexPart struct {
	Text         string                `json:"text,omitempty"`
	FunctionCall *vertexFunctionCall   `json:"functionCall,omitempty"`
	FunctionResp *vertexFunctionResult `json:"functionResponse,omitempty"`
}

// vertexFunctionCall represents a function call from the model.
type vertexFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

// vertexFunctionResult represents the result of a function execution.
type vertexFunctionResult struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

// vertexTool represents a tool definition in Vertex AI format.
type vertexTool struct {
	FunctionDeclarations []vertexFunctionDeclaration `json:"functionDeclarations"`
}

// vertexFunctionDeclaration represents a function declaration.
type vertexFunctionDeclaration struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// vertexGenerationConfig contains generation parameters.
type vertexGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
}

// vertexResponse matches the Vertex AI Gemini API response format.
type vertexResponse struct {
	Candidates []struct {
		Content struct {
			Role  string       `json:"role"`
			Parts []vertexPart `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// vertexErrorResponse matches the Vertex AI API error response format.
type vertexErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// convertToVertexRequest converts our standard CompletionRequest to Vertex AI's format.
func convertToVertexRequest(req *types.CompletionRequest) (*vertexRequest, error) {
	vertexReq := &vertexRequest{
		Contents: make([]vertexContent, 0, len(req.Messages)),
	}

	// Separate system messages from other messages
	// Vertex AI handles system instructions separately
	for i := range req.Messages {
		if req.Messages[i].Role == types.RoleSystem {
			// Use first system message as systemInstruction
			if vertexReq.SystemInstruction == nil {
				content, err := convertMessageToVertex(&req.Messages[i])
				if err != nil {
					return nil, err
				}
				vertexReq.SystemInstruction = &content
			}
			// Subsequent system messages are skipped (only one systemInstruction allowed)
			continue
		}

		content, err := convertMessageToVertex(&req.Messages[i])
		if err != nil {
			return nil, err
		}
		vertexReq.Contents = append(vertexReq.Contents, content)
	}

	// Convert tools
	if len(req.Tools) > 0 {
		funcDecls := make([]vertexFunctionDeclaration, 0, len(req.Tools))
		for _, tool := range req.Tools {
			funcDecls = append(funcDecls, vertexFunctionDeclaration{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			})
		}
		vertexReq.Tools = []vertexTool{{FunctionDeclarations: funcDecls}}
	}

	// Convert generation config
	var genConfig vertexGenerationConfig
	hasConfig := false

	if req.Temperature != 0 {
		temp := req.Temperature
		genConfig.Temperature = &temp
		hasConfig = true
	}

	if req.MaxTokens != 0 {
		maxTokens := req.MaxTokens
		genConfig.MaxOutputTokens = &maxTokens
		hasConfig = true
	}

	if req.TopP != 0 {
		topP := req.TopP
		genConfig.TopP = &topP
		hasConfig = true
	}

	if hasConfig {
		vertexReq.GenerationConfig = &genConfig
	}

	return vertexReq, nil
}

// convertMessageToVertex converts a types.Message to Vertex AI format.
func convertMessageToVertex(msg *types.Message) (vertexContent, error) {
	// Map role names
	role := string(msg.Role)
	if role == "assistant" {
		role = "model"
	} else if role == "tool" {
		role = "function"
	}

	content := vertexContent{
		Role:  role,
		Parts: make([]vertexPart, 0),
	}

	// Add text content if present
	if msg.Content != "" {
		content.Parts = append(content.Parts, vertexPart{
			Text: msg.Content,
		})
	}

	// Add tool calls if present
	for _, tc := range msg.ToolCalls {
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return vertexContent{}, fmt.Errorf("failed to parse tool call arguments: %w", err)
		}

		content.Parts = append(content.Parts, vertexPart{
			FunctionCall: &vertexFunctionCall{
				Name: tc.Function.Name,
				Args: args,
			},
		})
	}

	// Add tool result if present (for role=tool messages)
	if msg.Role == types.RoleTool && msg.ToolCallID != "" {
		// Parse content as JSON if possible, otherwise use as-is
		var response map[string]any
		if err := json.Unmarshal([]byte(msg.Content), &response); err != nil {
			// If content is not JSON, wrap it in a simple structure
			response = map[string]any{"result": msg.Content}
		}

		content.Parts = []vertexPart{{
			FunctionResp: &vertexFunctionResult{
				Name:     msg.Name,
				Response: response,
			},
		}}
	}

	return content, nil
}

// convertFromVertexResponse converts Vertex AI's response to our standard format.
func convertFromVertexResponse(resp *vertexResponse, model string) (types.CompletionResponse, error) {
	result := types.CompletionResponse{
		ID:      fmt.Sprintf("vertex-%d", time.Now().Unix()),
		Model:   model,
		Created: time.Now().Unix(),
		Usage: types.Usage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		},
	}

	if len(resp.Candidates) == 0 {
		return result, fmt.Errorf("no candidates in response")
	}

	candidate := resp.Candidates[0]
	result.FinishReason = convertFinishReason(candidate.FinishReason)

	// Convert the content back to our Message format
	msg := types.Message{
		Role: types.RoleAssistant,
	}

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			msg.Content += part.Text
		}

		if part.FunctionCall != nil {
			// Convert function call to tool call
			argsJSON, err := json.Marshal(part.FunctionCall.Args)
			if err != nil {
				return result, fmt.Errorf("failed to marshal function args: %w", err)
			}

			msg.ToolCalls = append(msg.ToolCalls, types.ToolCall{
				ID:   generateToolCallID(),
				Type: "function",
				Function: types.FunctionCall{
					Name:      part.FunctionCall.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	result.Message = msg
	return result, nil
}

// convertFinishReason maps Vertex AI finish reasons to our standard format.
func convertFinishReason(reason string) string {
	switch reason {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "SAFETY":
		return "content_filter"
	case "RECITATION":
		return "content_filter"
	default:
		return reason
	}
}

// generateToolCallID creates a unique ID for tool calls using crypto/rand.
func generateToolCallID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based ID if random generation fails
		return fmt.Sprintf("call_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("call_%s", hex.EncodeToString(b))
}
