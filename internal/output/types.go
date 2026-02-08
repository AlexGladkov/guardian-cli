// Package output provides formatting for Guardian CLI output in both
// human-readable and JSON formats.
package output

// CheckReport contains the results of a guardian check operation.
type CheckReport struct {
	Violations []ViolationReport `json:"violations"`
	Summary    ReportSummary     `json:"summary"`
}

// ViolationReport describes a single rule violation found during a check.
type ViolationReport struct {
	RuleID         string `json:"rule_id"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	FilePath       string `json:"file_path"`
	DiffSnippet    string `json:"diff_snippet"`
	LLMExplanation string `json:"llm_explanation"`
}

// ReportSummary summarizes the check results.
type ReportSummary struct {
	Errors   int  `json:"errors"`
	Warnings int  `json:"warnings"`
	Passed   bool `json:"passed"`
}

// TallyReport contains the voting results for a proposal.
type TallyReport struct {
	ProposalID     string      `json:"proposal_id"`
	RuleID         string      `json:"rule_id"`
	EligibleVoters []string    `json:"eligible_voters"`
	Votes          []VoteEntry `json:"votes"`
	Result         string      `json:"result"`
	YesCount       int         `json:"yes_count"`
	NoCount        int         `json:"no_count"`
	Required       int         `json:"required"`
}

// VoteEntry represents a single vote in a tally report.
type VoteEntry struct {
	Email    string `json:"email"`
	Decision string `json:"decision"`
	Comment  string `json:"comment"`
}

// InboxReport lists proposals awaiting the current user's vote.
type InboxReport struct {
	Items []InboxItem `json:"items"`
	Total int         `json:"total"`
}

// InboxItem represents a single pending proposal in the inbox.
type InboxItem struct {
	ProposalID   string `json:"proposal_id"`
	RuleID       string `json:"rule_id"`
	ProposalType string `json:"proposal_type"`
	CreatedBy    string `json:"created_by"`
	CreatedAt    string `json:"created_at"`
	Age          string `json:"age"`
}

// HistoryReport lists finalized proposals.
type HistoryReport struct {
	Items []HistoryItem `json:"items"`
}

// HistoryItem represents a finalized proposal in the history.
type HistoryItem struct {
	ProposalID   string `json:"proposal_id"`
	RuleID       string `json:"rule_id"`
	ProposalType string `json:"proposal_type"`
	FinalizedAt  string `json:"finalized_at"`
	Summary      string `json:"summary"`
}
