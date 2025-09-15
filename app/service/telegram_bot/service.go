package telegram_bot

import (
	"context"
	"frank/app/service/prompt_manager"
	"frank/app/service/reason"
	"frank/app/service/telegram_reply"
	"frank/pkg/config"
	"frank/pkg/database"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samber/do"
)

type Service struct {
	tgBot         *bot.Bot
	cfg           *config.Config
	queries       *database.Queries
	replyService  *telegram_reply.Service
	reasonService *reason.Service
	promptManager *prompt_manager.Service
}

func New(di *do.Injector) (*Service, error) {
	tgBot := do.MustInvoke[*bot.Bot](di)

	service := &Service{
		cfg:           do.MustInvoke[*config.Config](di),
		tgBot:         tgBot,
		queries:       do.MustInvoke[*database.Queries](di),
		replyService:  do.MustInvoke[*telegram_reply.Service](di),
		reasonService: do.MustInvoke[*reason.Service](di),
		promptManager: do.MustInvoke[*prompt_manager.Service](di),
	}

	tgBot.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return true
	}, service.handleUpdates)

	return service, nil
}

func (s *Service) initCommands(ctx context.Context) {
	cmds := []models.BotCommand{
		{
			Command:     "/cancel",
			Description: "Отменить текущее действие",
		},
	}

	if _, err := s.tgBot.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: cmds,
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to set commands",
			slog.Any("error", err),
		)
	}
}

func (s *Service) Run(ctx context.Context) {
	s.initCommands(ctx)
}
