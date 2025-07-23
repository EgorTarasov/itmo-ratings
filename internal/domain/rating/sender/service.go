package sender

import (
	"context"
	"fmt"
	"itmo-ratings/internal/domain/rating"
)

type Service struct {
	sender sender
	parser parser
}

func New(sender sender, parser parser) *Service {
	return &Service{
		sender: sender,
		parser: parser,
	}
}

func (s *Service) UpdateStudentRating(ctx context.Context, studentID string, telegramUserID int64, programID int) error {
	entries, err := s.parser.GetEntries(ctx, int64(programID))
	if err != nil {
		return fmt.Errorf("failed to get rating entries")
	}
	var studentEntry rating.Entry

	for _, entry := range entries {
		if entry.SSPVOID == studentID {
			studentEntry = entry
			break
		}
	}

	return s.sender.SendMessage(ctx, telegramUserID, s.formatMessage(studentEntry))
}

func (s *Service) formatMessage(studentEntry rating.Entry) string {
	msg := `
Обновление места в списке поступления на программу:

Твой номер заявления: %s
Твое место в списке: %d
Твои баллы: %.2f

`
	return fmt.Sprintf(msg, studentEntry.SSPVOID, studentEntry.Position, studentEntry.ExamScores)
}
