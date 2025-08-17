package bot

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	*tgbotapi.BotAPI
}

type Option func(*tgbotapi.BotAPI)

func WithDebug() func(*tgbotapi.BotAPI) {
	return func(b *tgbotapi.BotAPI) {
		b.Debug = true
	}
}

func New(apiToken string, options ...Option) *Bot {
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		slog.Error("failed to init telegram bot", "token", apiToken, "err", err.Error())
		log.Panic(err)
	}

	for _, opt := range options {
		opt(bot)
	}

	return &Bot{
		BotAPI: bot,
	}
}

func (b *Bot) SendMessage(ctx context.Context, userID int64, content string) error {
	if userID == 0 {
		return fmt.Errorf("invalid userID: %v", userID)
	}
	if content == "" {
		return fmt.Errorf("invalid message content must be not empty string")
	}
	msg := tgbotapi.NewMessage(userID, content)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := b.BotAPI.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to user: %w", err)
	}
	return nil
}
