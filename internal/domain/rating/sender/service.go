package sender

import (
	"context"
	"fmt"
	"itmo-ratings/internal/domain/rating"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/samber/lo"
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

type studentEntry struct {
	studentID string
	entry     *rating.Entry
	program   *programData
}

type programData struct {
	data        *rating.ProgramDirection
	entries     []rating.Entry
	lastUpdated time.Time
}

// информация для форматирования сообщения о текущем состоянии студента
type programInfo struct {
	budgetSlots int
	programID   int
	name        string
	lastUpdate  time.Time
}

func (s *Service) UpdateStudentRating(
	ctx context.Context,
	studentID string,
	telegramUserID int64,
	_ int,
) error {
	// feature: отображать позиции студентов на других программах выше указанного studentID
	students := make(map[string][]studentEntry)
	// получение всех программ и списка поступающих на все программы:
	programsMap := make(map[int]programData)

	programs, err := s.parser.GetAllPrograms(ctx)
	if err != nil {
		return fmt.Errorf("failed to get avaliable programs: %w", err)
	}

	for _, program := range programs {
		entries, lastUpdated, err := s.parser.GetEntries(ctx, int64(program.CompetitiveGroupID)) // _ - lastUpdateTime
		if err != nil {
			slog.Info("failed to get rating entries",
				"err", err.Error(),
				"programID", program.CompetitiveGroupID,
			)
			continue
		}
		programData := programData{
			data:        &program,
			entries:     entries,
			lastUpdated: lastUpdated,
		}

		programsMap[program.CompetitiveGroupID] = programData

		for _, student := range entries {
			if _, ok := students[student.SSPVOID]; !ok {
				students[student.SSPVOID] = make([]studentEntry, 0, 2)
			}

			students[student.SSPVOID] = append(students[student.SSPVOID], studentEntry{
				studentID: student.SSPVOID,
				entry:     &student,
				program:   &programData,
			})
		}
	}
	// находим требуемого студента:
	requestedStudentEntries, ok := students[studentID]
	if !ok {
		return fmt.Errorf("failed to find student in all programs: %w", err)
	}

	err = s.sender.SendMessage(ctx, telegramUserID, studentSummary(requestedStudentEntries))
	if err != nil {
		return fmt.Errorf("failed to send student summary to userID: %d", telegramUserID)
	}

	return nil
}

func studentSummary(data []studentEntry) string {
	slices.SortFunc(data, func(a, b studentEntry) int {
		return a.entry.Priority - b.entry.Priority
	})

	msgRow := `
Приоритет: %d
Программа: %s
Позиция: %d / %d (всего подано заявлений: %d)
Студентов с приоритетов ниже чем у студента: %d
Последнее обновление: %s
`

	msgBuilder := strings.Builder{}
	for _, row := range data {
		withLowerPriority := lo.Filter(row.program.entries, func(v rating.Entry, _ int) bool {
			if v.Priority > row.entry.Priority && v.Position < row.entry.Position {
				return true
			}
			return false
		})

		msgBuilder.WriteString(fmt.Sprintf(msgRow,
			row.entry.Priority,
			formatProgram(row.program.data),
			row.entry.Position,
			row.program.data.BudgetMin,
			len(row.program.entries),
			len(withLowerPriority),
			row.program.lastUpdated.Format(time.RFC822),
		),
		)
	}

	return msgBuilder.String()
}

func formatProgram(program *rating.ProgramDirection) string {
	if program == nil {
		return ""
	}
	return fmt.Sprintf("[%s](https://abit.itmo.ru/rating/master/budget/%d)", program.DirectionTitle, program.CompetitiveGroupID)
}
