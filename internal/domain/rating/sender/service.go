package sender

import (
	"context"
	"fmt"
	"itmo-ratings/internal/domain/rating"
	"log/slog"
	"slices"
	"time"
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

func (s *Service) UpdateStudentRating(ctx context.Context, studentID string, telegramUserID int64, _ int) error {
	// получение всех программ и списка поступающих на все программы:

	programs, err := s.parser.GetAllPrograms(ctx)
	if err != nil {
		return fmt.Errorf("failed to get avaliable programs: %w", err)
	}

	for _, p := range programs {
		entries, lastUpdate, err := s.parser.GetEntries(ctx, int64(p.CompetitiveGroupID))
		if err != nil {
			slog.Info("failed to get rating entries",
				"err", err.Error(),
				"programID", p.CompetitiveGroupID,
			)
			continue
		}
		rating := calculateOurRating(entries, p.BudgetMin)

		msg, err := s.formatMessage(studentID, rating, programInfo{
			budgetSlots: p.BudgetMin,
			programID:   p.CompetitiveGroupID,
			name:        p.DirectionTitle,
			lastUpdate:  lastUpdate,
		})
		if err != nil {
			slog.Info("failed to format msg", "err", fmt.Errorf("failed to format msg: %w", err), "programID", p.CompetitiveGroupID)
			continue
		}
		err = s.sender.SendMessage(ctx, telegramUserID, msg)
		if err != nil {
			return fmt.Errorf("failed to calculate ratings stats for program: %d", p.CompetitiveGroupID)
		}
	}

	return nil
}

// slots - доступное количество бюджетных мест
// slots - доступное количество бюджетных мест
// entries - вхождения студентов на программе
func calculateOurRating(entries []rating.Entry, slots int) []rating.Entry {
	if len(entries) == 0 {
		return entries
	}

	// Safeguard: ensure slots is not negative
	if slots < 0 {
		slots = 0
	}

	// Safeguard: ensure we don't exceed the actual entries length
	maxIndex := slots + 3
	if maxIndex > len(entries) {
		maxIndex = len(entries)
	}

	// 1) берем в наш расчет только количество <= максимального + 3
	// +3 добавляет маленькую вероятность что кто-то пойдет на другую программу
	relatedEntries := entries[:maxIndex]
	var other []rating.Entry
	if maxIndex < len(entries) {
		other = entries[maxIndex:]
	}

	// 2) сортируем по:
	// количество баллов за ВИ + ИД, приоритет, есть соглансие или нет, средний балл диплома
	slices.SortFunc(relatedEntries, func(a, b rating.Entry) int {
		// More robust sorting logic
		if a.TotalScores != b.TotalScores {
			if a.TotalScores > b.TotalScores {
				return -1 // a should come before b (higher score first)
			}
			return 1
		}

		// If scores are equal, check priority (lower priority comes first)
		if a.Priority != b.Priority {
			if a.Priority < b.Priority {
				return -1
			}
			return 1
		}

		// If priority is equal, check agreement status
		if a.IsSendAgreement != b.IsSendAgreement {
			if a.IsSendAgreement && !b.IsSendAgreement {
				return -1 // agreement comes first
			}
			return 1
		}

		// If everything else is equal, compare diploma average
		if a.DiplomaAverage > b.DiplomaAverage {
			return -1
		} else if a.DiplomaAverage < b.DiplomaAverage {
			return 1
		}

		return 0 // completely equal
	})

	// Combine sorted relatedEntries with other entries
	result := make([]rating.Entry, 0, len(entries))
	result = append(result, relatedEntries...)
	result = append(result, other...)

	return result
}

type programInfo struct {
	budgetSlots int
	programID   int
	name        string
	lastUpdate  time.Time
}

func (s *Service) formatMessage(studentID string, entries []rating.Entry, program programInfo) (string, error) {
	var studentEntry rating.Entry
	var studentIndex int = -1

	for i, entry := range entries {
		if entry.SSPVOID == studentID {
			studentEntry = entry
			studentIndex = i
			break
		}
	}

	if studentIndex == -1 {
		return "", fmt.Errorf("Студент с номером заявления %s не найден в списке поступающих", studentID)
	}

	totalStudents := len(entries)
	studentsAbove := studentIndex
	studentsBelow := totalStudents - studentIndex - 1

	// Calculate percentage position
	percentilePosition := float64(studentsAbove) / float64(totalStudents) * 100

	// Check if student is in budget zone
	var budgetStatus string
	if studentIndex < program.budgetSlots {
		budgetStatus = "🟢 В зоне бюджетных мест"
	} else if studentIndex < program.budgetSlots+5 {
		budgetStatus = "🟡 Близко к бюджетной зоне"
	} else {
		budgetStatus = "🔴 Вне бюджетной зоны"
	}

	// Calculate distance to budget zone
	var distanceToBudget string
	if studentIndex >= program.budgetSlots {
		distance := studentIndex - program.budgetSlots + 1
		distanceToBudget = fmt.Sprintf("📏 До бюджета: %d мест", distance)
	} else {
		distanceToBudget = "🎉 В бюджетной зоне!"
	}

	// Calculate score difference with budget cutoff
	var scoreDiffWithBudget float64
	if program.budgetSlots > 0 && program.budgetSlots <= len(entries) {
		budgetCutoffScore := entries[program.budgetSlots-1].TotalScores
		scoreDiffWithBudget = budgetCutoffScore - studentEntry.TotalScores
	}

	msg := `
⌛ Последнее обновление: %s
📊 Обновление места в списке поступления на программу: %d, %s

👤 Твой номер заявления: %s
📍 Твое место в нашем рейтинге: %d из %d
📈 Официальная позиция: %d / %d
🎯 Твои баллы: %.2f
📊 Общий балл: %.2f

📈 Статистика позиции:
• Студентов выше тебя: %d
• Студентов ниже тебя: %d
• Процентиль: %.1f%%

🎯 Статус: %s
%s

📉 Отставание от бюджетного порога: %.2f баллов
🔄 Приоритет: %d
%s
`

	// Agreement status
	agreementStatus := ""
	if studentEntry.IsSendAgreement {
		agreementStatus = "✅ Согласие подано"
	} else {
		agreementStatus = "❌ Согласие не подано"
	}

	return fmt.Sprintf(msg,
		program.lastUpdate.String(),
		program.programID,
		program.name,
		studentEntry.SSPVOID,
		studentIndex+1, // +1 because index is 0-based
		totalStudents,
		studentEntry.Position,
		program.budgetSlots,
		studentEntry.ExamScores,
		studentEntry.TotalScores,
		studentsAbove,
		studentsBelow,
		percentilePosition,
		budgetStatus,
		distanceToBudget,
		scoreDiffWithBudget,
		studentEntry.Priority,
		agreementStatus,
	), nil
}
