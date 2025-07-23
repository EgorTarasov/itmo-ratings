package rating

// Add this to your rating package
type Entry struct {
	Contest                   string   `json:"contest"`
	ExamType                  string   `json:"exam_type"`
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
	Status                    string   `json:"status"`
	IsSpecialBCategory        *bool    `json:"is_special_b_category"`
	SSPVOID                   string   `json:"sspvo_id"`
	MainTopPriority           bool     `json:"main_top_priority"`
	HighestPassagewayPriority bool     `json:"highest_passageway_priority"`
	IsPublishedInWorkInRussia *bool    `json:"is_published_in_work_in_russia"`
	OfferNumber               *string  `json:"offer_number"`
	TargetOrganizationNumber  *string  `json:"target_organization_number"`
	IsDetailedTargetQuota     *bool    `json:"is_detailed_target_quota"`
	TargetAchievements        *float64 `json:"target_achievements"`
	HasApprovedContract       *bool    `json:"has_approved_contract"`
}
