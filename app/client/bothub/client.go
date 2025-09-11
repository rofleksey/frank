package bothub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"frank/pkg/config"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/samber/do"
)

// Client represents the Bothub chat API client
type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
}

// Prompt represents the input for generating completions
type Prompt struct {
	SystemText string `json:"systemText"`
	UserText   string `json:"userText"`
	Model      string `json:"model"`
}

// apiRequest represents the request payload for the Bothub API
type apiRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
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

// NewClient creates a new Bothub client instance
func NewClient(di *do.Injector) (*Client, error) {
	cfg := do.MustInvoke[*config.Config](di)

	if cfg.Bothub.Token == "" {
		return nil, fmt.Errorf("bothub token is required")
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Minute,
					KeepAlive: 5 * time.Minute,
				}).DialContext,
				TLSHandshakeTimeout:   5 * time.Minute,
				ResponseHeaderTimeout: 5 * time.Minute,
				ExpectContinueTimeout: 5 * time.Minute,
			},
		},
		token:   cfg.Bothub.Token,
		baseURL: "https://bothub.chat/api/v2/openai/v1/chat/completions",
	}, nil
}

// Process sends a prompt to the Bothub API and returns the generated completion
func (c *Client) Process(ctx context.Context, prompt Prompt) (string, error) {
	messages := make([]Message, 0, 2)

	if prompt.SystemText != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: prompt.SystemText,
		})
	}

	messages = append(messages, Message{
		Role:    "user",
		Content: prompt.UserText,
	})

	model := prompt.Model
	if model == "" {
		model = "deepseek-chat-v3-0324"
	}

	requestBody := apiRequest{
		Model:    model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("API error: %s (type: %s, code: %d)",
			apiResp.Error.Message, apiResp.Error.Type, apiResp.Error.Code)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	return apiResp.Choices[0].Message.Content, nil
}
