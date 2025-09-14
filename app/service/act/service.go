package act

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/client/yandex"
	"frank/app/command"
	"frank/app/dto"
	"frank/app/service/reason"
	"frank/app/service/scheduler"
	"frank/app/service/secret"
	"frank/app/service/telegram_reply"
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
	cfg                   *config.Config
	queries               *database.Queries
	commands              []Command
	rootDescription       string
	additionalDescription string
}

func New(di *do.Injector) (*Service, error) {
	cfg := do.MustInvoke[*config.Config](di)
	yandexClient := do.MustInvoke[*yandex.Client](di)
	replyService := do.MustInvoke[*telegram_reply.Service](di)
	schedulerService := do.MustInvoke[*scheduler.Service](di)
	reasonService := do.MustInvoke[*reason.Service](di)
	secretsService := do.MustInvoke[*secret.Service](di)

	actService := &Service{
		cfg:     cfg,
		queries: do.MustInvoke[*database.Queries](di),
	}

	rootCommands := []Command{
		command.NewNoopCommand(),
		command.NewReplyCommand(replyService),
		command.NewAttachCommand(actService, reasonService),
		command.NewChainCommand(actService),
	}

	additionalCommands := []Command{
		command.NewScheduleCommand(replyService, schedulerService),
		command.NewListScheduleCommand(schedulerService),
		command.NewCancelScheduleCommand(replyService, schedulerService),
		command.NewHTTPRequestCommand(replyService, secretsService),
		command.NewWebSearchCommand(replyService, yandexClient),
	}

	allCommands := make([]Command, 0, len(additionalCommands)+len(rootCommands))
	allCommands = append(allCommands, rootCommands...)
	allCommands = append(allCommands, additionalCommands...)

	actService.commands = allCommands
	actService.rootDescription = generateDescription(rootCommands)
	actService.additionalDescription = generateDescription(additionalCommands)

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

func (s *Service) RootCommandsDescription() string {
	return s.rootDescription
}

func (s *Service) AdditionalCommandsDescription() string {
	return s.additionalDescription
}
