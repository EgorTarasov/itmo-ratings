package rating

import "time"

// Add this to your rating package
type Entry struct {
	Contest                   string  `json:"contest"`
	ExamType                  string  `json:"exam_type"`
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
	Status                    string  `json:"status"`
	IsSpecialBCategory        *bool   `json:"is_special_b_category"`
	SSPVOID                   string  `json:"sspvo_id"`
	MainTopPriority           bool    `json:"main_top_priority"`
	HighestPassagewayPriority bool    `json:"highest_passageway_priority"`
	IsPublishedInWorkInRussia *bool   `json:"is_published_in_work_in_russia"`
	OfferNumber               *string `json:"offer_number"`
	// TargetOrganizationNumber  *string  `json:"target_organization_number"`
	IsDetailedTargetQuota *bool    `json:"is_detailed_target_quota"`
	TargetAchievements    *float64 `json:"target_achievements"`
	HasApprovedContract   *bool    `json:"has_approved_contract"`
}

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

type ProgramData struct {
	Data        *ProgramDirection
	Entries     []Entry
	LastUpdated time.Time
}

type StudentEntry struct {
	StudentID string
	Entry     *Entry
	Program   *ProgramData
}

type StudentSummaryEntry struct {
	Priority             int    `json:"priority"`
	Program              string `json:"program"` // formatted link like "[Title](url)"
	Position             int    `json:"position"`
	BudgetMin            int    `json:"budgetMin"`
	TotalApplications    int    `json:"totalApplications"`
	LowerPriorityAhead   int    `json:"lowerPriorityAhead"`
	LastUpdatedFormatted string `json:"lastUpdated"` // RFC822 to match msgRow
}

type StudentSummary struct {
	StudentID string                `json:"studentId"`
	Entries   []StudentSummaryEntry `json:"entries"`
}
