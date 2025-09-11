package act

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/command"
	"frank/app/dto"
	"frank/app/service/reason"
	"frank/app/service/scheduler"
	"frank/app/service/telegram_sender"
	"frank/pkg/config"
	"frank/pkg/database"

	"github.com/samber/do"
)

type Command interface {
	Execute(ctx context.Context, prompt dto.Prompt) (string, error)
	Name() string
	Description() string
}

type Service struct {
	cfg         *config.Config
	queries     *database.Queries
	commands    []Command
	description string
}

func New(di *do.Injector) (*Service, error) {
	cfg := do.MustInvoke[*config.Config](di)
	telegramSender := do.MustInvoke[*telegram_sender.Service](di)
	schedulerService := do.MustInvoke[*scheduler.Service](di)
	reasonService := do.MustInvoke[*reason.Service](di)

	actService := &Service{
		cfg:     cfg,
		queries: do.MustInvoke[*database.Queries](di),
	}

	commands := []Command{
		command.NewNoopCommand(),
		command.NewReplyCommand(cfg.Telegram.ChatID, telegramSender),
		command.NewScheduleCommand(cfg.Telegram.ChatID, telegramSender, schedulerService),
		command.NewChainCommand(actService),
		command.NewHTTPRequestCommand(),
		command.NewAttachCommand(actService, reasonService),
	}

	actService.commands = commands
	actService.description = generateDescription(commands)

	return actService, nil
}

type GenericCommandData struct {
	Command string `json:"command"`
}

func (s *Service) Handle(ctx context.Context, prompt dto.Prompt) (string, error) {
	var data GenericCommandData

	if err := json.Unmarshal([]byte(prompt.Text), &data); err != nil {
		return "", fmt.Errorf("json unmarshal: %w", err)
	}

	if data.Command == "" {
		return "", fmt.Errorf("command is empty")
	}

	var cmd Command

	for _, c := range s.commands {
		if c.Name() == data.Command {
			cmd = c
			break
		}
	}

	if cmd == nil {
		return "", fmt.Errorf("command not found: %s", data.Command)
	}

	output, err := cmd.Execute(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("command.Handle failed for command %s: %w", cmd.Name(), err)
	}

	return output, nil
}

func (s *Service) CommandsDescription() string {
	return s.description
}
