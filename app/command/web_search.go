package command

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/dto"
	"log/slog"
	"strings"
)

type WebSearchCommand struct {
	replier      Replier
	searchEngine WebSearchEngine
}

func NewWebSearchCommand(replier Replier, searchEngine WebSearchEngine) *WebSearchCommand {
	return &WebSearchCommand{
		replier:      replier,
		searchEngine: searchEngine,
	}
}

type WebSearchCommandData struct {
	Query string `json:"query"`
}

func (c *WebSearchCommand) Execute(ctx context.Context, prompt dto.Prompt) (string, error) {
	logger := slog.With(
		slog.String("command", c.Name()),
		slog.String("prompt_id", prompt.ID.String()),
	)

	logger.InfoContext(ctx, "Executing web_search command",
		slog.Any("prompt", prompt),
	)

	var requestData WebSearchCommandData
	if err := json.Unmarshal([]byte(prompt.Text), &requestData); err != nil {
		return "", fmt.Errorf("json unmarshal: %w", err)
	}

	if requestData.Query == "" {
		return "", fmt.Errorf("empty query")
	}

	_ = c.replier.Reply(ctx, fmt.Sprintf("Web searching query '%s'...", requestData.Query))

	result, err := c.searchEngine.WebSearch(ctx, requestData.Query)
	if err != nil {
		return "", fmt.Errorf("WebSearch: %w", err)
	}

	return result, nil
}

func (c *WebSearchCommand) Name() string {
	return "web_search"
}

func (c *WebSearchCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    required:
      - command
      - query
    properties:
      command:
        type: string
        enum: 
          - web_search
      query:
        type: string
        description: Search query
    description: Executes a web search query. Returns result as an XML string.
  `)
}
