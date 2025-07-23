package scrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"itmo-ratings/internal/domain/rating"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	pageUrl    = "https://abit.itmo.ru/rating/master/budget/%d"
	timeout    = time.Minute * 5
	maxRetries = 3
	retryDelay = time.Second * 2
)

// NextJSData represents the structure of __NEXT_DATA__ content
type NextJSData struct {
	Props struct {
		PageProps struct {
			ProgramList struct {
				ByTargetQuota      []RatingEntry `json:"by_target_quota"`
				GeneralCompetition []RatingEntry `json:"general_competition"`
				Direction          struct {
					DirectionTitle     string `json:"direction_title"`
					BudgetMin          int    `json:"budget_min"`
					Contract           int    `json:"contract"`
					TargetReception    int    `json:"target_reception"`
					IsuID              *int   `json:"isu_id"`
					Invalid            int    `json:"invalid"`
					SpecialQuota       int    `json:"special_quota"`
					CompetitiveGroupID int    `json:"competitive_group_id"`
				} `json:"direction"`
				UpdateTime string `json:"update_time"`
			} `json:"programList"`
		} `json:"pageProps"`
	} `json:"props"`
}

// RatingEntry represents a single rating entry
type RatingEntry struct {
	Contest                   *string  `json:"contest"`
	ExamType                  *string  `json:"exam_type"`
	DiplomaAverage            float64  `json:"diploma_average"`
	Position                  int      `json:"position"`
	Priority                  int      `json:"priority"`
	IAScores                  float64  `json:"ia_scores"`
	ExamScores                float64  `json:"exam_scores"`
	TotalScores               float64  `json:"total_scores"`
	IsSendAgreement           bool     `json:"is_send_agreement"`
	SNILS                     string   `json:"snils"`
	CaseNumber                string   `json:"case_number"`
	Link                      string   `json:"link"`
	Status                    *string  `json:"status"`
	IsSpecialBCategory        *bool    `json:"is_special_b_category"`
	SSPVO                     string   `json:"sspvo_id"`
	MainTopPriority           bool     `json:"main_top_priority"`
	HighestPassagewayPriority bool     `json:"highest_passageway_priority"`
	IsPublishedInWorkInRussia *bool    `json:"is_published_in_work_in_russia"`
	OfferNumber               *string  `json:"offer_number"`
	TargetOrganizationNumber  *string  `json:"target_organization_number"`
	IsDetailedTargetQuota     *bool    `json:"is_detailed_target_quota"`
	TargetAchievements        *float64 `json:"target_achievements"`
	HasApprovedContract       *bool    `json:"has_approved_contract"`
}

type Service struct {
	client client
}

func New(httpClient client) *Service {
	return &Service{client: httpClient}
}

func (s *Service) GetEntries(ctx context.Context, programID int64) ([]rating.Entry, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	htmlContent, err := s.getPageWithRetries(ctx, fmt.Sprintf(pageUrl, programID))
	if err != nil {
		return nil, fmt.Errorf("failed to download html page: %d err: %v", programID, err)
	}

	nextData, err := s.extractNextData(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to extract __NEXT_DATA__: %w", err)
	}

	entries := s.convertToRatingEntries(nextData)

	return entries, nil
}

func (s *Service) getPageWithRetries(ctx context.Context, url string) (string, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		content, err := s.getPage(ctx, url)
		if err == nil {
			return content, nil
		}

		lastErr = err

		// Don't retry on context cancellation/timeout
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		// Wait before retry (except for last attempt)
		if attempt < maxRetries-1 {
			select {
			case <-time.After(retryDelay * time.Duration(attempt+1)): // Exponential backoff
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}

	return "", fmt.Errorf("failed after %d attempts, last error: %w", maxRetries, lastErr)
}

func (s *Service) getPage(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to assemble request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body content: %w", err)
	}

	return string(content), nil
}

func (s *Service) extractNextData(htmlContent string) (*NextJSData, error) {
	re := regexp.MustCompile(`<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)
	matches := re.FindStringSubmatch(htmlContent)

	if len(matches) < 2 {
		return nil, fmt.Errorf("__NEXT_DATA__ script tag not found")
	}

	jsonContent := strings.TrimSpace(matches[1])

	var nextData NextJSData
	if err := json.Unmarshal([]byte(jsonContent), &nextData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &nextData, nil
}

func (s *Service) convertToRatingEntries(nextData *NextJSData) []rating.Entry {
	var entries []rating.Entry

	// Process general competition entries
	for _, entry := range nextData.Props.PageProps.ProgramList.GeneralCompetition {
		ratingEntry := rating.Entry{
			Position:       entry.Position,
			Priority:       entry.Priority,
			DiplomaAverage: entry.DiplomaAverage,
			ExamScores:     entry.ExamScores,
			TotalScores:    entry.TotalScores,
			SNILS:          entry.SNILS,
			CaseNumber:     entry.CaseNumber,
			SSPVOID:        entry.SSPVO,
			// Add other fields as needed for your rating.Entry struct
		}

		// Handle nullable fields
		if entry.Contest != nil {
			ratingEntry.Contest = *entry.Contest
		}
		if entry.ExamType != nil {
			ratingEntry.ExamType = *entry.ExamType
		}
		if entry.Status != nil {
			ratingEntry.Status = *entry.Status
		}

		entries = append(entries, ratingEntry)
	}

	// for _, entry := range nextData.Props.PageProps.ProgramList.ByTargetQuota {
	// 	ratingEntry := rating.Entry{
	// 		Position:       entry.Position,
	// 		Priority:       entry.Priority,
	// 		DiplomaAverage: entry.DiplomaAverage,
	// 		ExamScores:     entry.ExamScores,
	// 		TotalScores:    entry.TotalScores,
	// 		SNILS:          entry.SNILS,
	// 		CaseNumber:     entry.CaseNumber,
	// 		// Add other fields as needed
	// 	}

	// 	// Handle nullable fields
	// 	if entry.Contest != nil {
	// 		ratingEntry.Contest = *entry.Contest
	// 	}
	// 	if entry.ExamType != nil {
	// 		ratingEntry.ExamType = *entry.ExamType
	// 	}
	// 	if entry.Status != nil {
	// 		ratingEntry.Status = *entry.Status
	// 	}

	// 	entries = append(entries, ratingEntry)
	// }

	return entries
}
