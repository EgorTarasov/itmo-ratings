package sender

import (
	"context"
	"itmo-ratings/internal/domain/rating"
	"time"
)

type (
	sender interface {
		SendMessage(ctx context.Context, userID int64, content string) error
	}
	parser interface {
		GetEntries(ctx context.Context, programID int64) ([]rating.Entry, time.Time, error)
		GetAllPrograms(ctx context.Context) ([]rating.ProgramDirection, error)
	}
)
