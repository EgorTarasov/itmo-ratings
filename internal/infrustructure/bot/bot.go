package bot

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	*tgbotapi.BotAPI
}

func New(apiToken string) *Bot {
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

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
