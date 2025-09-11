package command

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/dto"
	"log/slog"
	"strings"
)

type ReplyCommand struct {
	replier Replier
}

func NewReplyCommand(replier Replier) *ReplyCommand {
	return &ReplyCommand{
		replier: replier,
	}
}

type ReplyCommandData struct {
	Text string `json:"text"`
}

func (c *ReplyCommand) Execute(ctx context.Context, prompt dto.Prompt) (string, error) {
	slog.Info("Executing reply command",
		slog.String("text", prompt.Text),
	)

	var data ReplyCommandData

	if err := json.Unmarshal([]byte(prompt.Text), &data); err != nil {
		return "", fmt.Errorf("json unmarshal: %w", err)
	}

	if data.Text == "" {
		return "", fmt.Errorf("empty text")
	}

	if err := c.replier.Reply(ctx, data.Text); err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	return "", nil
}

func (c *ReplyCommand) Name() string {
	return "reply"
}

func (c *ReplyCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    required:
      - command
      - text
    properties:
      command:
        type: string
        enum: 
          - reply
      text:
        type: string
        description: The text content of the reply
    description: sends a message to the user
  `)
}
