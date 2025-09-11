package command

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/dto"
	"log/slog"
	"strings"
)

type ChainCommand struct {
	actor Actor
}

func NewChainCommand(actor Actor) *ChainCommand {
	return &ChainCommand{
		actor: actor,
	}
}

type ChainCommandData struct {
	List []json.RawMessage `json:"list"`
}

func (c *ChainCommand) Handle(ctx context.Context, prompt dto.Prompt) error {
	slog.Info("Executing chain command",
		slog.Any("prompt", prompt),
	)

	var data ChainCommandData

	if err := json.Unmarshal([]byte(prompt.Text), &data); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}

	for i, subCommand := range data.List {
		slog.Info("Executing chained command",
			slog.Int("index", i),
			slog.String("text", string(subCommand)),
		)

		if err := c.actor.Handle(ctx, prompt.BranchWithNewText(string(subCommand))); err != nil {
			return fmt.Errorf("failed to handle subcommand %s: %w", string(subCommand), err)
		}
	}

	return nil
}

func (c *ChainCommand) Name() string {
	return "chain"
}

func (c *ChainCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    required:
      - command
      - list
    properties:
      command:
        type: string
        enum: 
          - chain
      list:
        type: array
        items:
          <JSONs of commands to execute>
        description: |
          The list of commands to execute.
    description: executes multiple commands sequentially.
  `)
}
