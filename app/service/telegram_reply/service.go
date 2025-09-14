package telegram_reply

import (
	"context"
	"frank/pkg/config"
	"frank/pkg/util"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samber/do"
)

type Service struct {
	cfg   *config.Config
	tgBot *bot.Bot
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		cfg:   do.MustInvoke[*config.Config](di),
		tgBot: do.MustInvoke[*bot.Bot](di),
	}, nil
}

func (s *Service) Reply(ctx context.Context, text string) {
	_, err := s.tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    s.cfg.Telegram.ChatID,
		Text:      text,
		ParseMode: "Markdown",
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: util.ToPtr(true),
		},
	})
	if err != nil {
		_, _ = s.tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: s.cfg.Telegram.ChatID,
			Text:   text,
			LinkPreviewOptions: &models.LinkPreviewOptions{
				IsDisabled: util.ToPtr(true),
			},
		})
	}
}
