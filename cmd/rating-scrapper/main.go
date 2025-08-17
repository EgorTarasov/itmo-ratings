package main

import (
	"context"
	"itmo-ratings/internal/domain/rating/scrapper"
	sender "itmo-ratings/internal/domain/rating/student_rating_service"
	"itmo-ratings/internal/infrustructure/bot"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

const (
	success int = 0
	fail        = 1
)

func main() {
	os.Exit(run())
}

func run() int {
	telegramToken := os.Getenv("TELEGRAM_API_TOKEN")
	studentID := os.Getenv("STUDENT_ID")
	telegramStudentID := os.Getenv("TELEGRAM_USER_ID")
	ctx := context.Background()

	telegramID, _ := strconv.ParseInt(telegramStudentID, 10, 64)

	telegram := bot.New(telegramToken)

	parser := scrapper.New(http.DefaultClient)

	runner := sender.New(parser)

	summary, err := runner.GetStudentSummary(ctx, studentID)
	if err != nil {
		slog.Error("failed to update status", "err", err)
		return fail
	}

	if err = telegram.SendMessage(ctx, telegramID, summary); err != nil {
		slog.Error("failed to send message", "err", err)
		return fail
	}

	return success
}
