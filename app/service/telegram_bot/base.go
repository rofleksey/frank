package telegram_bot

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (s *Service) handleUpdates(ctx context.Context, _ *bot.Bot, update *models.Update) {
	if update.Message != nil {
		s.handleMessage(ctx, update.Message)
	}
}

func (s *Service) handleMessage(ctx context.Context, msg *models.Message) {
	if msg.Chat.ID != s.cfg.Telegram.ChatID {
		slog.WarnContext(ctx, "Got message from unexpected chat id",
			slog.Int64("chat_id", msg.Chat.ID),
		)
		return
	}

	switch strings.TrimSpace(msg.Text) {
	case "/cancel":
		s.handleCancel(ctx)
	default:
		s.handleUnknownMessage(ctx, msg.Text)
	}
}
