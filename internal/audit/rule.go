package audit

// Severity defines the level of impact of an audit finding.
type Severity string

const (
	SeverityInfo Severity = "INFO"
	SeverityWarn Severity = "WARN"
	SeverityFail Severity = "FAIL"
)

// Finding represents a single compliance or security violation discovered during audit.
type Finding struct {
	Resource       string   `json:"resource"`
	Severity       Severity `json:"severity"`
	Message        string   `json:"message"`
	Recommendation string   `json:"recommendation"`
	File           string   `json:"file,omitempty"`
}

// AuditReport holds the aggregated results of the infrastructure audit.
type AuditReport struct {
	Directory string    `json:"directory"`
	Passed    bool      `json:"passed"`
	Findings  []Finding `json:"findings"`
	Summary   struct {
		Total int `json:"total"`
		Fail  int `json:"fail"`
		Warn  int `json:"warn"`
		Info  int `json:"info"`
	} `json:"summary"`
	AIRecommendations string `json:"ai_recommendations,omitempty"`
}
