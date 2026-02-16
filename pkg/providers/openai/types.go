package openai

import "github.com/oskarhane/goagent/pkg/types"

// openAIRequest matches the OpenAI Chat Completions API request format.
type openAIRequest struct {
	Model               string          `json:"model"`
	Messages            []types.Message `json:"messages"`
	Tools               []types.Tool    `json:"tools,omitempty"`
	Temperature         float64         `json:"temperature,omitempty"`
	MaxTokens           int             `json:"max_tokens,omitempty"`
	MaxCompletionTokens int             `json:"max_completion_tokens,omitempty"`
	TopP                float64         `json:"top_p,omitempty"`
}

// openAIResponse matches the OpenAI Chat Completions API response format.
type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int           `json:"index"`
		Message      types.Message `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// openAIErrorResponse matches the OpenAI API error response format.
type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}

// convertToOpenAIRequest converts our standard CompletionRequest to OpenAI's format.
func convertToOpenAIRequest(req *types.CompletionRequest) openAIRequest {
	oaiReq := openAIRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Tools:       req.Tools,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	// Use max_completion_tokens for newer models (gpt-5+), max_tokens for older ones
	if req.MaxTokens > 0 {
		if isNewerModel(req.Model) {
			oaiReq.MaxCompletionTokens = req.MaxTokens
		} else {
			oaiReq.MaxTokens = req.MaxTokens
		}
	}

	return oaiReq
}

// isNewerModel returns true if the model uses max_completion_tokens instead of max_tokens.
func isNewerModel(model string) bool {
	// gpt-5+ models use max_completion_tokens
	if len(model) >= 5 && model[:5] == "gpt-5" {
		return true
	}
	// Add other newer model prefixes as needed
	return false
}

// convertFromOpenAIResponse converts OpenAI's response to our standard format.
func convertFromOpenAIResponse(resp *openAIResponse) types.CompletionResponse {
	var message types.Message
	var finishReason string

	if len(resp.Choices) > 0 {
		message = resp.Choices[0].Message
		finishReason = resp.Choices[0].FinishReason
	}

	return types.CompletionResponse{
		ID:           resp.ID,
		Model:        resp.Model,
		Created:      resp.Created,
		Message:      message,
		FinishReason: finishReason,
		Usage: types.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
}
