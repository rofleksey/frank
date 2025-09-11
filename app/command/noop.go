package command

import (
	"context"
	"frank/app/dto"
	"log/slog"
	"strings"
)

type NoopCommand struct{}

func NewNoopCommand() *NoopCommand {
	return &NoopCommand{}
}

func (c *NoopCommand) Execute(ctx context.Context, prompt dto.Prompt) (string, error) {
	slog.Info("Executing noop command",
		slog.Any("prompt", prompt),
	)

	return "", nil
}

func (c *NoopCommand) Name() string {
	return "noop"
}

func (c *NoopCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    required:
      - command
    properties: 
      command:
        type: string
        enum: 
          - noop
    description: do nothing
  `)
}
