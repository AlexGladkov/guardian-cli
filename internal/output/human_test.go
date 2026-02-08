package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintCheckReportHuman_NoViolations(t *testing.T) {
	r := &CheckReport{
		Summary: ReportSummary{Passed: true},
	}
	var buf bytes.Buffer
	PrintCheckReportHuman(&buf, r)

	out := buf.String()
	assert.Contains(t, out, "Guardian Check Report")
	assert.Contains(t, out, "=====================")
	assert.Contains(t, out, "No violations found.")
	assert.Contains(t, out, "Result: PASSED")
}

func TestPrintCheckReportHuman_WithViolations(t *testing.T) {
	r := &CheckReport{
		Violations: []ViolationReport{
			{
				RuleID:         "domain_no_infra",
				Severity:       "error",
				Description:    "Domain layer must not depend on infra",
				FilePath:       "domain/service/UserService.kt",
				DiffSnippet:    "+ import com.myapp.infra.database.UserRepository",
				LLMExplanation: "This import creates a direct dependency.",
			},
			{
				RuleID:         "money_minor_units",
				Severity:       "warning",
				Description:    "Money must use int minor units",
				FilePath:       "domain/model/Price.kt",
				DiffSnippet:    "+ val amount: Double",
				LLMExplanation: "Using Double for monetary values can cause precision issues.",
			},
		},
		Summary: ReportSummary{
			Errors:   1,
			Warnings: 1,
			Passed:   false,
		},
	}

	var buf bytes.Buffer
	PrintCheckReportHuman(&buf, r)

	out := buf.String()

	// Check header
	assert.Contains(t, out, "Guardian Check Report")

	// Check first violation
	assert.Contains(t, out, "VIOLATION [error] domain_no_infra")
	assert.Contains(t, out, "Domain layer must not depend on infra")
	assert.Contains(t, out, "File: domain/service/UserService.kt")
	assert.Contains(t, out, "+ import com.myapp.infra.database.UserRepository")
	assert.Contains(t, out, "AI: This import creates a direct dependency.")

	// Check second violation (warning label)
	assert.Contains(t, out, "WARNING [warning] money_minor_units")
	assert.Contains(t, out, "Money must use int minor units")
	assert.Contains(t, out, "File: domain/model/Price.kt")

	// Check summary
	assert.Contains(t, out, "1 error(s), 1 warning(s) - FAILED")
}

func TestPrintCheckReportHuman_OnlyWarnings(t *testing.T) {
	r := &CheckReport{
		Violations: []ViolationReport{
			{
				RuleID:      "test_rule",
				Severity:    "warning",
				Description: "Minor issue",
			},
		},
		Summary: ReportSummary{
			Errors:   0,
			Warnings: 1,
			Passed:   true,
		},
	}

	var buf bytes.Buffer
	PrintCheckReportHuman(&buf, r)
	out := buf.String()

	assert.Contains(t, out, "WARNING [warning] test_rule")
	assert.Contains(t, out, "0 error(s), 1 warning(s) - PASSED")
}

func TestPrintCheckReportHuman_NoFilePath(t *testing.T) {
	r := &CheckReport{
		Violations: []ViolationReport{
			{
				RuleID:      "meta_check",
				Severity:    "error",
				Description: "Unauthorized changes to .agreements/",
			},
		},
		Summary: ReportSummary{Errors: 1, Passed: false},
	}

	var buf bytes.Buffer
	PrintCheckReportHuman(&buf, r)
	out := buf.String()

	assert.NotContains(t, out, "File:")
}

func TestPrintCheckReportHuman_NoDiffSnippet(t *testing.T) {
	r := &CheckReport{
		Violations: []ViolationReport{
			{
				RuleID:      "some_rule",
				Severity:    "error",
				Description: "Missing something",
				FilePath:    "src/main.go",
			},
		},
		Summary: ReportSummary{Errors: 1, Passed: false},
	}

	var buf bytes.Buffer
	PrintCheckReportHuman(&buf, r)
	out := buf.String()

	assert.NotContains(t, out, "Diff:")
}

func TestPrintCheckReportHuman_NoLLMExplanation(t *testing.T) {
	r := &CheckReport{
		Violations: []ViolationReport{
			{
				RuleID:      "some_rule",
				Severity:    "error",
				Description: "Issue found",
				FilePath:    "src/main.go",
				DiffSnippet: "+ bad code",
			},
		},
		Summary: ReportSummary{Errors: 1, Passed: false},
	}

	var buf bytes.Buffer
	PrintCheckReportHuman(&buf, r)
	out := buf.String()

	assert.NotContains(t, out, "AI:")
}

func TestPrintCheckReportHuman_MultilineDiff(t *testing.T) {
	r := &CheckReport{
		Violations: []ViolationReport{
			{
				RuleID:      "test",
				Severity:    "error",
				Description: "Test",
				DiffSnippet: "+ line 1\n+ line 2\n+ line 3",
			},
		},
		Summary: ReportSummary{Errors: 1, Passed: false},
	}

	var buf bytes.Buffer
	PrintCheckReportHuman(&buf, r)
	out := buf.String()

	assert.Contains(t, out, "    + line 1")
	assert.Contains(t, out, "    + line 2")
	assert.Contains(t, out, "    + line 3")
}

// --- Tally report ---

func TestPrintTallyReportHuman_WithVotes(t *testing.T) {
	r := &TallyReport{
		ProposalID:     "2024-01-15-domain_no_infra",
		RuleID:         "domain_no_infra",
		EligibleVoters: []string{"alice@test.com", "bob@test.com", "carol@test.com"},
		Votes: []VoteEntry{
			{Email: "alice@test.com", Decision: "yes", Comment: "LGTM"},
			{Email: "bob@test.com", Decision: "no", Comment: "Need more discussion"},
		},
		Result:   "PENDING",
		YesCount: 1,
		NoCount:  1,
		Required: 2,
	}

	var buf bytes.Buffer
	PrintTallyReportHuman(&buf, r)
	out := buf.String()

	assert.Contains(t, out, "Guardian Tally Report")
	assert.Contains(t, out, "Proposal: 2024-01-15-domain_no_infra")
	assert.Contains(t, out, "Rule:     domain_no_infra")
	assert.Contains(t, out, "Eligible voters (3):")
	assert.Contains(t, out, "- alice@test.com")
	assert.Contains(t, out, "- bob@test.com")
	assert.Contains(t, out, "- carol@test.com")
	assert.Contains(t, out, "Votes (2):")
	assert.Contains(t, out, "alice@test.com: YES")
	assert.Contains(t, out, `"LGTM"`)
	assert.Contains(t, out, "bob@test.com: NO")
	assert.Contains(t, out, "Yes: 1 / No: 1 / Required: 2")
	assert.Contains(t, out, "Result: PENDING")
}

func TestPrintTallyReportHuman_NoVotes(t *testing.T) {
	r := &TallyReport{
		ProposalID:     "test-proposal",
		RuleID:         "test_rule",
		EligibleVoters: []string{"voter@test.com"},
		Votes:          nil,
		Result:         "PENDING",
		YesCount:       0,
		NoCount:        0,
		Required:       1,
	}

	var buf bytes.Buffer
	PrintTallyReportHuman(&buf, r)
	out := buf.String()

	assert.Contains(t, out, "(no votes yet)")
}

func TestPrintTallyReportHuman_Accepted(t *testing.T) {
	r := &TallyReport{
		ProposalID:     "test",
		RuleID:         "rule",
		EligibleVoters: []string{"a@t.com"},
		Votes:          []VoteEntry{{Email: "a@t.com", Decision: "yes"}},
		Result:         "ACCEPTED",
		YesCount:       1,
		NoCount:        0,
		Required:       1,
	}

	var buf bytes.Buffer
	PrintTallyReportHuman(&buf, r)
	out := buf.String()

	assert.Contains(t, out, "Result: ACCEPTED")
}

func TestPrintTallyReportHuman_VoteWithoutComment(t *testing.T) {
	r := &TallyReport{
		ProposalID:     "test",
		RuleID:         "rule",
		EligibleVoters: []string{"a@t.com"},
		Votes:          []VoteEntry{{Email: "a@t.com", Decision: "yes", Comment: ""}},
		Result:         "ACCEPTED",
		YesCount:       1,
		Required:       1,
	}

	var buf bytes.Buffer
	PrintTallyReportHuman(&buf, r)
	out := buf.String()

	// Should not show empty comment
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "a@t.com") && strings.Contains(line, "YES") {
			assert.NotContains(t, line, `""`)
		}
	}
}

// --- Inbox report ---

func TestPrintInboxReportHuman_WithItems(t *testing.T) {
	r := &InboxReport{
		Items: []InboxItem{
			{
				ProposalID:   "2024-01-15-rule1",
				RuleID:       "rule1",
				ProposalType: "modify",
				CreatedBy:    "author@test.com",
				CreatedAt:    "2024-01-15T10:30:00Z",
				Age:          "5 days",
			},
			{
				ProposalID:   "2024-01-10-rule2",
				RuleID:       "rule2",
				ProposalType: "add",
				CreatedBy:    "dev@test.com",
				CreatedAt:    "2024-01-10T08:00:00Z",
				Age:          "10 days",
			},
		},
		Total: 2,
	}

	var buf bytes.Buffer
	PrintInboxReportHuman(&buf, r)
	out := buf.String()

	assert.Contains(t, out, "Guardian Inbox")
	assert.Contains(t, out, "2 proposal(s) awaiting your vote:")
	assert.Contains(t, out, "[modify] 2024-01-15-rule1 (rule1)")
	assert.Contains(t, out, "Created by: author@test.com")
	assert.Contains(t, out, "age: 5 days")
	assert.Contains(t, out, "[add] 2024-01-10-rule2 (rule2)")
}

func TestPrintInboxReportHuman_Empty(t *testing.T) {
	r := &InboxReport{
		Items: nil,
		Total: 0,
	}

	var buf bytes.Buffer
	PrintInboxReportHuman(&buf, r)
	out := buf.String()

	assert.Contains(t, out, "No pending proposals require your vote.")
}

// --- History report ---

func TestPrintHistoryReportHuman_WithItems(t *testing.T) {
	r := &HistoryReport{
		Items: []HistoryItem{
			{
				ProposalID:   "2024-01-15-rule1",
				RuleID:       "rule1",
				ProposalType: "modify",
				FinalizedAt:  "2024-01-20T15:00:00Z",
				Summary:      "Changed import restrictions",
			},
			{
				ProposalID:   "2024-01-10-rule2",
				RuleID:       "rule2",
				ProposalType: "remove",
				FinalizedAt:  "2024-01-12T10:00:00Z",
				Summary:      "Removed deprecated rule",
			},
		},
	}

	var buf bytes.Buffer
	PrintHistoryReportHuman(&buf, r)
	out := buf.String()

	assert.Contains(t, out, "Guardian History")
	assert.Contains(t, out, "[modify] 2024-01-15-rule1 (rule1)")
	assert.Contains(t, out, "Finalized: 2024-01-20T15:00:00Z")
	assert.Contains(t, out, "Summary: Changed import restrictions")
	assert.Contains(t, out, "[remove] 2024-01-10-rule2 (rule2)")
}

func TestPrintHistoryReportHuman_Empty(t *testing.T) {
	r := &HistoryReport{Items: nil}

	var buf bytes.Buffer
	PrintHistoryReportHuman(&buf, r)
	out := buf.String()

	assert.Contains(t, out, "No finalized proposals.")
}

func TestPrintHistoryReportHuman_NoSummary(t *testing.T) {
	r := &HistoryReport{
		Items: []HistoryItem{
			{
				ProposalID:   "p1",
				RuleID:       "r1",
				ProposalType: "add",
				FinalizedAt:  "2024-01-01",
				Summary:      "",
			},
		},
	}

	var buf bytes.Buffer
	PrintHistoryReportHuman(&buf, r)
	out := buf.String()

	assert.NotContains(t, out, "Summary:")
}
