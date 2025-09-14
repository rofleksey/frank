package bothub

// Prompt represents the input for generating completions
type Prompt struct {
	SystemText string `json:"systemText"`
	UserText   string `json:"userText"`
	Model      string `json:"model"`
}

// apiRequest represents the request payload for the Bothub API
type apiRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature"`
	Messages    []Message `json:"messages"`
}

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// apiResponse represents the response from the Bothub API
type apiResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *apiError `json:"error,omitempty"`
}

// apiError represents an error response from the API
type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    int    `json:"code"`
}
