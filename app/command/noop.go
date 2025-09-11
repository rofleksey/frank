package command

import (
	"context"
	"log/slog"
	"strings"
)

type NoopCommand struct{}

func NewNoopCommand() *NoopCommand {
	return &NoopCommand{}
}

func (c *NoopCommand) Handle(ctx context.Context, dataBytes []byte) error {
	slog.Info("Executing noop command",
		slog.String("text", string(dataBytes)),
	)

	return nil
}

func (c *NoopCommand) Name() string {
	return "noop"
}

func (c *NoopCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    properties: {}
    description: do nothing
  `)
}
