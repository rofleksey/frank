package bothub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	ModelDeepseekR1_0528 = "deepseek-r1-0528"
	ModelDeepseekR1      = "deepseek-r1"
	ModelDeepseekChatV3  = "deepseek-chat-v3-0324"
)

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
		model = ModelDeepseekR1_0528
	}

	requestBody := apiRequest{
		Model:       model,
		MaxTokens:   100000,
		Temperature: 0.2,
		Messages:    messages,
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
