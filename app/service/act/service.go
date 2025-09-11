package act

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/command"
	"frank/app/service/scheduler"
	"frank/app/service/telegram_sender"
	"frank/pkg/config"
	"frank/pkg/database"

	"github.com/samber/do"
)

type Command interface {
	Handle(ctx context.Context, data []byte) error
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

	service := &Service{
		cfg:     cfg,
		queries: do.MustInvoke[*database.Queries](di),
	}

	commands := []Command{
		command.NewNoopCommand(),
		command.NewReplyCommand(cfg.Telegram.ChatID, telegramSender),
		command.NewScheduleCommand(cfg.Telegram.ChatID, telegramSender, schedulerService),
		command.NewSeqCommand(service),
	}

	service.commands = commands
	service.description = generateDescription(commands)

	return service, nil
}

type GenericCommandData struct {
	Command string `json:"command"`
}

func (s *Service) Handle(ctx context.Context, dataBytes []byte) error {
	var data GenericCommandData

	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}

	if data.Command == "" {
		return fmt.Errorf("command is empty")
	}

	var cmd Command

	for _, c := range s.commands {
		if c.Name() == data.Command {
			cmd = c
			break
		}
	}

	if cmd == nil {
		return fmt.Errorf("command not found: %s", data.Command)
	}

	if err := cmd.Handle(ctx, dataBytes); err != nil {
		return fmt.Errorf("command.Handle failed for command %s: %w", cmd.Name(), err)
	}

	return nil
}

func (s *Service) CommandsDescription() string {
	return s.description
}
