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
		// GetEntries получение списка студентов в рейтинговом списке на зачисление на данную программу.
		//
		// Parameters:
		//   - ctx: контекст выполнения
		//   - programID: идентификатор программы (competitive_group_id)
		//
		// Returns:
		//   - entries: список студентов, отсортированный по рейтингу
		//   - lastUpdate: время последнего обновления рейтинга на сайте
		//   - error: ошибка получения или парсинга данных
		GetEntries(ctx context.Context, programID int64) ([]rating.Entry, time.Time, error)

		// GetAllPrograms получение всех доступных программ магистратуры ИТМО.
		//
		// Returns:
		//   - programs: список программ с информацией о количестве мест
		//   - error: ошибка при запросе к API или парсинге ответа
		GetAllPrograms(ctx context.Context) ([]rating.ProgramDirection, error)
	}
)
