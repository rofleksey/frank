package telegram_sender

import (
	"context"
	"frank/pkg/util"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samber/do"
)

type Service struct {
	tgBot *bot.Bot
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		tgBot: do.MustInvoke[*bot.Bot](di),
	}, nil
}

func (s *Service) SendMessage(ctx context.Context, chatID int64, text string) error {
	_, err := s.tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: util.ToPtr(true),
		},
	})
	return err
}
