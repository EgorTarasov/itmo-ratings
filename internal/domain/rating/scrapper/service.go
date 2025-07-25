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

	"github.com/samber/lo"
)

const (
	programsPageUrl = "https://abit.itmo.ru/ratings/master"
	programPageUrl  = "https://abit.itmo.ru/rating/master/budget/%d"
	programsAPIUrl  = "https://abitlk.itmo.ru/api/v1/rating/directions?degree=master"
	timeout         = time.Minute * 5
	maxRetries      = 3
	retryDelay      = time.Second * 2
)

type Service struct {
	client client
}

func New(httpClient client) *Service {
	return &Service{client: httpClient}
}

func (s *Service) GetEntries(ctx context.Context, programID int64) ([]rating.Entry, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	htmlContent, err := s.getPageWithRetries(ctx, fmt.Sprintf(programPageUrl, programID))
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

func (s *Service) GetAllPrograms(ctx context.Context) ([]rating.ProgramDirection, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, programsAPIUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "ru")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("DNT", "1")
	req.Header.Set("Origin", "https://abit.itmo.ru")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", "https://abit.itmo.ru/")
	req.Header.Set("Sec-CH-UA", `"Not)A;Brand";v="8", "Chromium";v="138"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse ProgramsAPIResponse
	if err := json.Unmarshal(content, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	if !apiResponse.OK {
		return nil, fmt.Errorf("API returned error: %s", apiResponse.Message)
	}

	return lo.Map(apiResponse.Result.Items, func(v ProgramDirection, _ int) rating.ProgramDirection {
		return rating.ProgramDirection{
			DirectionTitle:     v.DirectionTitle,
			BudgetMin:          v.BudgetMin,
			Contract:           v.Contract,
			TargetReception:    v.TargetReception,
			IsuID:              v.IsuID,
			Invalid:            v.Invalid,
			SpecialQuota:       v.SpecialQuota,
			CompetitiveGroupID: v.CompetitiveGroupID,
		}
	}), nil
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

func (s *Service) extractNextData(htmlContent string) (*RatingsNextJSData, error) {
	re := regexp.MustCompile(`<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)
	matches := re.FindStringSubmatch(htmlContent)

	if len(matches) < 2 {
		return nil, fmt.Errorf("__NEXT_DATA__ script tag not found")
	}

	jsonContent := strings.TrimSpace(matches[1])

	var nextData RatingsNextJSData
	if err := json.Unmarshal([]byte(jsonContent), &nextData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &nextData, nil
}

func (s *Service) convertToRatingEntries(nextData *RatingsNextJSData) []rating.Entry {
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
