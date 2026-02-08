package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintCheckReportJSON_NoViolations(t *testing.T) {
	r := &CheckReport{
		Violations: []ViolationReport{},
		Summary:    ReportSummary{Passed: true},
	}

	var buf bytes.Buffer
	err := PrintCheckReportJSON(&buf, r)
	require.NoError(t, err)

	var decoded CheckReport
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Empty(t, decoded.Violations)
	assert.True(t, decoded.Summary.Passed)
	assert.Equal(t, 0, decoded.Summary.Errors)
	assert.Equal(t, 0, decoded.Summary.Warnings)
}

func TestPrintCheckReportJSON_WithViolations(t *testing.T) {
	r := &CheckReport{
		Violations: []ViolationReport{
			{
				RuleID:         "domain_no_infra",
				Severity:       "error",
				Description:    "Domain layer must not depend on infra",
				FilePath:       "domain/service/UserService.kt",
				DiffSnippet:    "+ import infra",
				LLMExplanation: "Bad dependency",
			},
			{
				RuleID:      "money_minor_units",
				Severity:    "warning",
				Description: "Use int for money",
			},
		},
		Summary: ReportSummary{
			Errors:   1,
			Warnings: 1,
			Passed:   false,
		},
	}

	var buf bytes.Buffer
	err := PrintCheckReportJSON(&buf, r)
	require.NoError(t, err)

	var decoded CheckReport
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Violations, 2)
	assert.Equal(t, "domain_no_infra", decoded.Violations[0].RuleID)
	assert.Equal(t, "error", decoded.Violations[0].Severity)
	assert.Equal(t, "Domain layer must not depend on infra", decoded.Violations[0].Description)
	assert.Equal(t, "domain/service/UserService.kt", decoded.Violations[0].FilePath)
	assert.Equal(t, "+ import infra", decoded.Violations[0].DiffSnippet)
	assert.Equal(t, "Bad dependency", decoded.Violations[0].LLMExplanation)

	assert.Equal(t, "money_minor_units", decoded.Violations[1].RuleID)
	assert.Equal(t, "warning", decoded.Violations[1].Severity)

	assert.Equal(t, 1, decoded.Summary.Errors)
	assert.Equal(t, 1, decoded.Summary.Warnings)
	assert.False(t, decoded.Summary.Passed)
}

func TestPrintCheckReportJSON_ValidJSON(t *testing.T) {
	r := &CheckReport{
		Violations: []ViolationReport{
			{
				RuleID:      "test",
				Severity:    "error",
				Description: `Contains "quotes" and special chars: <>&`,
			},
		},
		Summary: ReportSummary{Errors: 1, Passed: false},
	}

	var buf bytes.Buffer
	err := PrintCheckReportJSON(&buf, r)
	require.NoError(t, err)

	assert.True(t, json.Valid(buf.Bytes()))
}

func TestPrintCheckReportJSON_Indented(t *testing.T) {
	r := &CheckReport{
		Violations: []ViolationReport{
			{RuleID: "test", Severity: "error", Description: "test"},
		},
		Summary: ReportSummary{Errors: 1, Passed: false},
	}

	var buf bytes.Buffer
	err := PrintCheckReportJSON(&buf, r)
	require.NoError(t, err)

	// Should be indented
	assert.Contains(t, buf.String(), "  ")
}

// --- Tally report JSON ---

func TestPrintTallyReportJSON(t *testing.T) {
	r := &TallyReport{
		ProposalID:     "2024-01-15-test",
		RuleID:         "test_rule",
		EligibleVoters: []string{"a@t.com", "b@t.com"},
		Votes: []VoteEntry{
			{Email: "a@t.com", Decision: "yes", Comment: "ok"},
			{Email: "b@t.com", Decision: "no", Comment: "nope"},
		},
		Result:   "PENDING",
		YesCount: 1,
		NoCount:  1,
		Required: 2,
	}

	var buf bytes.Buffer
	err := PrintTallyReportJSON(&buf, r)
	require.NoError(t, err)

	var decoded TallyReport
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Equal(t, "2024-01-15-test", decoded.ProposalID)
	assert.Equal(t, "test_rule", decoded.RuleID)
	assert.Equal(t, []string{"a@t.com", "b@t.com"}, decoded.EligibleVoters)
	assert.Len(t, decoded.Votes, 2)
	assert.Equal(t, "a@t.com", decoded.Votes[0].Email)
	assert.Equal(t, "yes", decoded.Votes[0].Decision)
	assert.Equal(t, "ok", decoded.Votes[0].Comment)
	assert.Equal(t, "PENDING", decoded.Result)
	assert.Equal(t, 1, decoded.YesCount)
	assert.Equal(t, 1, decoded.NoCount)
	assert.Equal(t, 2, decoded.Required)
}

func TestPrintTallyReportJSON_EmptyVotes(t *testing.T) {
	r := &TallyReport{
		ProposalID:     "test",
		RuleID:         "rule",
		EligibleVoters: []string{"a@t.com"},
		Votes:          []VoteEntry{},
		Result:         "PENDING",
	}

	var buf bytes.Buffer
	err := PrintTallyReportJSON(&buf, r)
	require.NoError(t, err)

	var decoded TallyReport
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Empty(t, decoded.Votes)
}

// --- Inbox report JSON ---

func TestPrintInboxReportJSON(t *testing.T) {
	r := &InboxReport{
		Items: []InboxItem{
			{
				ProposalID:   "p1",
				RuleID:       "r1",
				ProposalType: "modify",
				CreatedBy:    "user@test.com",
				CreatedAt:    "2024-01-15T10:30:00Z",
				Age:          "5 days",
			},
		},
		Total: 1,
	}

	var buf bytes.Buffer
	err := PrintInboxReportJSON(&buf, r)
	require.NoError(t, err)

	var decoded InboxReport
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Equal(t, 1, decoded.Total)
	require.Len(t, decoded.Items, 1)
	assert.Equal(t, "p1", decoded.Items[0].ProposalID)
	assert.Equal(t, "r1", decoded.Items[0].RuleID)
	assert.Equal(t, "modify", decoded.Items[0].ProposalType)
	assert.Equal(t, "user@test.com", decoded.Items[0].CreatedBy)
	assert.Equal(t, "5 days", decoded.Items[0].Age)
}

func TestPrintInboxReportJSON_Empty(t *testing.T) {
	r := &InboxReport{
		Items: []InboxItem{},
		Total: 0,
	}

	var buf bytes.Buffer
	err := PrintInboxReportJSON(&buf, r)
	require.NoError(t, err)

	var decoded InboxReport
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Equal(t, 0, decoded.Total)
	assert.Empty(t, decoded.Items)
}

// --- History report JSON ---

func TestPrintHistoryReportJSON(t *testing.T) {
	r := &HistoryReport{
		Items: []HistoryItem{
			{
				ProposalID:   "2024-01-15-rule1",
				RuleID:       "rule1",
				ProposalType: "modify",
				FinalizedAt:  "2024-01-20T15:00:00Z",
				Summary:      "Changed restrictions",
			},
		},
	}

	var buf bytes.Buffer
	err := PrintHistoryReportJSON(&buf, r)
	require.NoError(t, err)

	var decoded HistoryReport
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	require.Len(t, decoded.Items, 1)
	assert.Equal(t, "2024-01-15-rule1", decoded.Items[0].ProposalID)
	assert.Equal(t, "rule1", decoded.Items[0].RuleID)
	assert.Equal(t, "modify", decoded.Items[0].ProposalType)
	assert.Equal(t, "2024-01-20T15:00:00Z", decoded.Items[0].FinalizedAt)
	assert.Equal(t, "Changed restrictions", decoded.Items[0].Summary)
}

func TestPrintHistoryReportJSON_Empty(t *testing.T) {
	r := &HistoryReport{
		Items: []HistoryItem{},
	}

	var buf bytes.Buffer
	err := PrintHistoryReportJSON(&buf, r)
	require.NoError(t, err)

	var decoded HistoryReport
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Empty(t, decoded.Items)
}

func TestPrintHistoryReportJSON_MultipleItems(t *testing.T) {
	r := &HistoryReport{
		Items: []HistoryItem{
			{ProposalID: "p1", RuleID: "r1", ProposalType: "modify", FinalizedAt: "2024-01-01"},
			{ProposalID: "p2", RuleID: "r2", ProposalType: "add", FinalizedAt: "2024-02-01"},
			{ProposalID: "p3", RuleID: "r3", ProposalType: "remove", FinalizedAt: "2024-03-01"},
		},
	}

	var buf bytes.Buffer
	err := PrintHistoryReportJSON(&buf, r)
	require.NoError(t, err)

	var decoded HistoryReport
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Items, 3)
}

func TestPrintCheckReportJSON_NilViolations(t *testing.T) {
	r := &CheckReport{
		Violations: nil,
		Summary:    ReportSummary{Passed: true},
	}

	var buf bytes.Buffer
	err := PrintCheckReportJSON(&buf, r)
	require.NoError(t, err)

	assert.True(t, json.Valid(buf.Bytes()))

	// Nil slice serializes as null in JSON
	var raw map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &raw)
	require.NoError(t, err)
}

func TestAllJSONOutputsAreValidJSON(t *testing.T) {
	tests := []struct {
		name string
		fn   func() (*bytes.Buffer, error)
	}{
		{
			name: "CheckReport",
			fn: func() (*bytes.Buffer, error) {
				var buf bytes.Buffer
				err := PrintCheckReportJSON(&buf, &CheckReport{
					Violations: []ViolationReport{{RuleID: "r", Severity: "error", Description: "d"}},
					Summary:    ReportSummary{Errors: 1},
				})
				return &buf, err
			},
		},
		{
			name: "TallyReport",
			fn: func() (*bytes.Buffer, error) {
				var buf bytes.Buffer
				err := PrintTallyReportJSON(&buf, &TallyReport{ProposalID: "p", RuleID: "r"})
				return &buf, err
			},
		},
		{
			name: "InboxReport",
			fn: func() (*bytes.Buffer, error) {
				var buf bytes.Buffer
				err := PrintInboxReportJSON(&buf, &InboxReport{Total: 0})
				return &buf, err
			},
		},
		{
			name: "HistoryReport",
			fn: func() (*bytes.Buffer, error) {
				var buf bytes.Buffer
				err := PrintHistoryReportJSON(&buf, &HistoryReport{})
				return &buf, err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf, err := tc.fn()
			require.NoError(t, err)
			assert.True(t, json.Valid(buf.Bytes()), "output should be valid JSON")
		})
	}
}

func TestJSONFieldNames(t *testing.T) {
	r := &CheckReport{
		Violations: []ViolationReport{
			{
				RuleID:         "test_rule",
				Severity:       "error",
				Description:    "Test desc",
				FilePath:       "test/file.go",
				DiffSnippet:    "+ code",
				LLMExplanation: "explanation",
			},
		},
		Summary: ReportSummary{
			Errors:   1,
			Warnings: 0,
			Passed:   false,
		},
	}

	var buf bytes.Buffer
	err := PrintCheckReportJSON(&buf, r)
	require.NoError(t, err)

	out := buf.String()
	// Check JSON field names match the spec
	assert.Contains(t, out, `"rule_id"`)
	assert.Contains(t, out, `"severity"`)
	assert.Contains(t, out, `"description"`)
	assert.Contains(t, out, `"file_path"`)
	assert.Contains(t, out, `"diff_snippet"`)
	assert.Contains(t, out, `"llm_explanation"`)
	assert.Contains(t, out, `"errors"`)
	assert.Contains(t, out, `"warnings"`)
	assert.Contains(t, out, `"passed"`)
}

func TestTallyReportJSONFieldNames(t *testing.T) {
	r := &TallyReport{
		ProposalID:     "pid",
		RuleID:         "rid",
		EligibleVoters: []string{"a@t.com"},
		Votes:          []VoteEntry{{Email: "a@t.com", Decision: "yes", Comment: "ok"}},
		Result:         "ACCEPTED",
		YesCount:       1,
		NoCount:        0,
		Required:       1,
	}

	var buf bytes.Buffer
	err := PrintTallyReportJSON(&buf, r)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, `"proposal_id"`)
	assert.Contains(t, out, `"rule_id"`)
	assert.Contains(t, out, `"eligible_voters"`)
	assert.Contains(t, out, `"votes"`)
	assert.Contains(t, out, `"result"`)
	assert.Contains(t, out, `"yes_count"`)
	assert.Contains(t, out, `"no_count"`)
	assert.Contains(t, out, `"required"`)
	assert.Contains(t, out, `"email"`)
	assert.Contains(t, out, `"decision"`)
	assert.Contains(t, out, `"comment"`)
}
