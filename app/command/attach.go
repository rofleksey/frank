package command

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/dto"
	"log/slog"
	"strings"
)

type AttachCommand struct {
	actor    Actor
	reasoner Reasoner
}

func NewAttachCommand(actor Actor, reasoner Reasoner) *AttachCommand {
	return &AttachCommand{
		actor:    actor,
		reasoner: reasoner,
	}
}

type AttachSubcommand struct {
	Name       string          `json:"name"`
	Subcommand json.RawMessage `json:"subcommand"`
}

type AttachCommandData struct {
	NewPrompt string             `json:"new_prompt"`
	List      []AttachSubcommand `json:"list"`
}

func (c *AttachCommand) Execute(ctx context.Context, prompt dto.Prompt) (string, error) {
	slog.Info("Executing attach command",
		slog.String("text", prompt.Text),
	)

	var data AttachCommandData

	if err := json.Unmarshal([]byte(prompt.Text), &data); err != nil {
		return "", fmt.Errorf("json unmarshal: %w", err)
	}

	for i, cmd := range data.List {
		slog.Info("Executing attached command",
			slog.Int("index", i),
			slog.String("name", cmd.Name),
			slog.String("cmd", string(cmd.Subcommand)),
		)

		output, err := c.actor.Handle(ctx, prompt.BranchWithNewText(string(cmd.Subcommand)))
		if err != nil {
			return "", fmt.Errorf("failed to handle attachment subcommand '%s' %s: %w", cmd.Name, string(cmd.Subcommand), err)
		}

		prompt = prompt.BranchWithNewAttachment(dto.Attachment{
			Name:    cmd.Name,
			Content: output,
		})
	}

	c.reasoner.Handle(prompt.BranchWithNewText(data.NewPrompt))

	return "", nil
}

func (c *AttachCommand) Name() string {
	return "attach"
}

func (c *AttachCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    required:
      - command
      - new_prompt
      - list
    properties:
      command:
        type: string
        enum: 
          = attach
      new_prompt:
        type: string
        description: The text of the new prompt
      list:
        type: array
        description: List of subcommands whose results need to be attached
        items:
          type: object
          required:
            - name
            - subcommand
          properties:
            name:
              type: string
              description: The name of the subcommand
            subcommand:
              <JSON of the command to schedule, must have a result defined in the spec>
    description: executes a new prompt with the results of the subcommands attached to it
  `)
}
