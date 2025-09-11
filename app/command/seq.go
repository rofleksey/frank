package command

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

type SeqCommand struct {
	actor Actor
}

func NewSeqCommand(actor Actor) *SeqCommand {
	return &SeqCommand{
		actor: actor,
	}
}

type SeqCommandData struct {
	SubCommands []json.RawMessage `json:"subcommands"`
}

func (c *SeqCommand) Handle(ctx context.Context, dataBytes []byte) error {
	slog.Info("Executing seq command",
		slog.String("text", string(dataBytes)),
	)

	var data SeqCommandData

	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}

	for _, subCommand := range data.SubCommands {
		if err := c.actor.Handle(ctx, subCommand); err != nil {
			return fmt.Errorf("failed to handle subcommand %s: %w", string(subCommand), err)
		}
	}

	return nil
}

func (c *SeqCommand) Name() string {
	return "seq"
}

func (c *SeqCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    required:
      - commands
    properties:
      commands:
        type: array
        items:
          <JSONs of commands to execute>
        description: |
          The list of commands to execute.
    description: executes multiple commands sequentially. this useful for example when you need to execute command with no output (e.g. 'schedule') and still notify a user (e.g. via 'reply')
  `)
}
