package main

import (
	"context"
	"itmo-ratings/internal/domain/rating/scrapper"
	"itmo-ratings/internal/domain/rating/sender"
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
	programRawID := os.Getenv("PROGRAM_ID")
	studentID := os.Getenv("STUDENT_ID")
	telegramStudentID := os.Getenv("TELEGRAM_USER_ID")
	ctx := context.Background()

	programID, _ := strconv.ParseInt(programRawID, 10, 64)
	telegramID, _ := strconv.ParseInt(telegramStudentID, 10, 64)

	bot := bot.New(telegramToken)

	parser := scrapper.New(http.DefaultClient)

	runner := sender.New(bot, parser)

	if err := runner.UpdateStudentRating(ctx, studentID, telegramID, int(programID)); err != nil {
		slog.Error("failed to update status", "err", err)
		return fail
	}

	return success
}
