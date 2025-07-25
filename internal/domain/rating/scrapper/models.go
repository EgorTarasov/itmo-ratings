package scrapper

import "time"

// RatingsNextJSData represents the structure of __NEXT_DATA__ content
type RatingsNextJSData struct {
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
				UpdateTime time.Time `json:"update_time"`
			} `json:"programList"`
		} `json:"pageProps"`
	} `json:"props"`
}

// RatingEntry represents a single rating entry
type RatingEntry struct {
	Contest                   *string `json:"contest"`
	ExamType                  *string `json:"exam_type"`
	DiplomaAverage            float64 `json:"diploma_average"`
	Position                  int     `json:"position"`
	Priority                  int     `json:"priority"`
	IAScores                  float64 `json:"ia_scores"`
	ExamScores                float64 `json:"exam_scores"`
	TotalScores               float64 `json:"total_scores"`
	IsSendAgreement           bool    `json:"is_send_agreement"`
	SNILS                     string  `json:"snils"`
	CaseNumber                string  `json:"case_number"`
	Link                      string  `json:"link"`
	Status                    *string `json:"status"`
	IsSpecialBCategory        *bool   `json:"is_special_b_category"`
	SSPVO                     string  `json:"sspvo_id"`
	MainTopPriority           bool    `json:"main_top_priority"`
	HighestPassagewayPriority bool    `json:"highest_passageway_priority"`
	IsPublishedInWorkInRussia *bool   `json:"is_published_in_work_in_russia"`
	OfferNumber               *string `json:"offer_number"`
	// TargetOrganizationNumber  *string  `json:"target_organization_number"`
	IsDetailedTargetQuota *bool    `json:"is_detailed_target_quota"`
	TargetAchievements    *float64 `json:"target_achievements"`
	HasApprovedContract   *bool    `json:"has_approved_contract"`
}

// ProgramsAPIResponse represents the response from the programs API
type ProgramsAPIResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Result  struct {
		Items []ProgramDirection `json:"items"`
	} `json:"result"`
}

// ProgramDirection represents a single program direction from the API
type ProgramDirection struct {
	DirectionTitle     string `json:"direction_title"`
	BudgetMin          int    `json:"budget_min"`
	Contract           int    `json:"contract"`
	TargetReception    int    `json:"target_reception"`
	IsuID              *int   `json:"isu_id"`
	Invalid            int    `json:"invalid"`
	SpecialQuota       int    `json:"special_quota"`
	CompetitiveGroupID int    `json:"competitive_group_id"`
}
