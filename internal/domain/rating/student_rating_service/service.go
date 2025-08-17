package sender

import (
	"context"
	"fmt"

	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"itmo-ratings/internal/domain/rating"

	"github.com/samber/lo"
)

type Service struct {
	parser     parser
	programMap map[int]rating.ProgramData
	students   map[string][]rating.StudentEntry
	mu         sync.RWMutex
}

func New(parser parser) *Service {
	s := &Service{
		parser:     parser,
		programMap: make(map[int]rating.ProgramData),
		students:   make(map[string][]rating.StudentEntry),
	}
	return s
}

func (s *Service) Enrich(ctx context.Context) error {
	programs, err := s.parser.GetAllPrograms(ctx)
	if err != nil {
		return fmt.Errorf("failed to get available programs: %w", err)
	}

	programMap := make(map[int]rating.ProgramData)
	students := make(map[string][]rating.StudentEntry)

	for _, program := range programs {
		entries, lastUpdated, err := s.parser.GetEntries(ctx, int64(program.CompetitiveGroupID))
		if err != nil {
			slog.Info("failed to get rating entries",
				"err", err.Error(),
				"programID", program.CompetitiveGroupID,
			)
			continue
		}
		pd := rating.ProgramData{
			Data:        &program,
			Entries:     entries,
			LastUpdated: lastUpdated,
		}
		programMap[program.CompetitiveGroupID] = pd

		for _, student := range entries {
			students[student.SSPVOID] = append(students[student.SSPVOID], rating.StudentEntry{
				StudentID: student.SSPVOID,
				Entry:     &student,
				Program:   &pd,
			})
		}
	}

	s.mu.Lock()
	s.programMap = programMap
	s.students = students
	s.mu.Unlock()
	return nil
}

func (s *Service) getStudents(ctx context.Context) map[string][]rating.StudentEntry {
	s.mu.RLock()
	students := s.students
	s.mu.RUnlock()
	if len(students) == 0 {
		if err := s.Enrich(ctx); err != nil {
			slog.Error("failed to update cache", "err", err.Error())
			return nil
		}
		s.mu.RLock()
		students = s.students
		s.mu.RUnlock()
	}
	return students
}

func (s *Service) GetStudentSummary(
	ctx context.Context,
	studentID string,
) (string, error) {
	students := s.getStudents(ctx)
	if students == nil {
		return "", fmt.Errorf("failed to find students")
	}
	requestedStudentEntries, ok := students[studentID]
	if !ok {
		return "", fmt.Errorf("failed to find student in all programs")
	}
	return studentSummary(requestedStudentEntries), nil
}

func (s *Service) GetStudentSummaryRaw(
	ctx context.Context,
	studentID string,
) (*rating.StudentSummary, error) {
	students := s.getStudents(ctx)
	if students == nil {
		return nil, fmt.Errorf("failed to find students")
	}
	requestedStudentEntries, ok := students[studentID]
	if !ok {
		return nil, fmt.Errorf("failed to find student in all programs")
	}

	summary := buildStudentSummary(studentID, requestedStudentEntries)
	return &summary, nil
}

func buildStudentSummary(studentID string, data []rating.StudentEntry) rating.StudentSummary {
	slices.SortFunc(data, func(a, b rating.StudentEntry) int {
		return a.Entry.Priority - b.Entry.Priority
	})

	out := rating.StudentSummary{
		StudentID: studentID,
		Entries:   make([]rating.StudentSummaryEntry, 0, len(data)),
	}

	for _, row := range data {
		withLowerPriority := lo.Filter(row.Program.Entries, func(v rating.Entry, _ int) bool {
			return v.Priority > row.Entry.Priority && v.Position < row.Entry.Position
		})

		out.Entries = append(out.Entries, rating.StudentSummaryEntry{
			Priority:             row.Entry.Priority,
			Program:              formatProgram(row.Program.Data),
			Position:             row.Entry.Position,
			BudgetMin:            row.Program.Data.BudgetMin,
			TotalApplications:    len(row.Program.Entries),
			LowerPriorityAhead:   len(withLowerPriority),
			LastUpdatedFormatted: row.Program.LastUpdated.Format(time.RFC822),
		})
	}

	return out
}

func studentSummary(data []rating.StudentEntry) string {
	slices.SortFunc(data, func(a, b rating.StudentEntry) int {
		return a.Entry.Priority - b.Entry.Priority
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
		withLowerPriority := lo.Filter(row.Program.Entries, func(v rating.Entry, _ int) bool {
			if v.Priority > row.Entry.Priority && v.Position < row.Entry.Position {
				return true
			}
			return false
		})

		msgBuilder.WriteString(fmt.Sprintf(msgRow,
			row.Entry.Priority,
			formatProgram(row.Program.Data),
			row.Entry.Position,
			row.Program.Data.BudgetMin,
			len(row.Program.Entries),
			len(withLowerPriority),
			row.Program.LastUpdated.Format(time.RFC822),
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
