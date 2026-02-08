package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Constitution validation tests ---

func validConstitution() *Constitution {
	return &Constitution{
		Governance: Governance{
			Voters: []VoterRef{
				{Role: "techlead"},
				{Role: "architect"},
			},
			Quorum: QuorumConfig{
				Type:      "two_thirds",
				Threshold: 0.66,
			},
			ForbidSelfApproval: true,
			ProposalTTLDays:    30,
		},
		Roles: map[string]Role{
			"techlead": {
				Members: []RoleMember{{Email: "lead@company.com"}},
			},
			"architect": {
				Members: []RoleMember{{Email: "arch@company.com"}},
			},
		},
		LLM: LLMConfig{
			Provider: "openai",
		},
	}
}

func TestValidateConstitution_Valid(t *testing.T) {
	c := validConstitution()
	err := ValidateConstitution(c)
	assert.NoError(t, err)
}

func TestValidateConstitution_Nil(t *testing.T) {
	err := ValidateConstitution(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "constitution is nil")
}

func TestValidateConstitution_EmptyVoters(t *testing.T) {
	c := validConstitution()
	c.Governance.Voters = nil
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "governance.voters must not be empty")
}

func TestValidateConstitution_EmptyVoterRole(t *testing.T) {
	c := validConstitution()
	c.Governance.Voters = append(c.Governance.Voters, VoterRef{Role: ""})
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "governance.voters[2].role must not be empty")
}

func TestValidateConstitution_InvalidQuorumType(t *testing.T) {
	c := validConstitution()
	c.Governance.Quorum.Type = "invalid"
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "governance.quorum.type \"invalid\" is invalid")
}

func TestValidateConstitution_EmptyQuorumType(t *testing.T) {
	c := validConstitution()
	c.Governance.Quorum.Type = ""
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "governance.quorum.type must not be empty")
}

func TestValidateConstitution_AllQuorumTypes(t *testing.T) {
	types := []string{"majority", "two_thirds", "unanimous"}
	for _, qt := range types {
		t.Run(qt, func(t *testing.T) {
			c := validConstitution()
			c.Governance.Quorum.Type = qt
			err := ValidateConstitution(c)
			assert.NoError(t, err)
		})
	}
}

func TestValidateConstitution_CustomQuorumRequiresThreshold(t *testing.T) {
	c := validConstitution()
	c.Governance.Quorum.Type = "custom"
	c.Governance.Quorum.Threshold = 0
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "threshold must be between 0")
}

func TestValidateConstitution_CustomQuorumThresholdTooHigh(t *testing.T) {
	c := validConstitution()
	c.Governance.Quorum.Type = "custom"
	c.Governance.Quorum.Threshold = 1.5
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "threshold must be between 0")
}

func TestValidateConstitution_CustomQuorumValidThreshold(t *testing.T) {
	c := validConstitution()
	c.Governance.Quorum.Type = "custom"
	c.Governance.Quorum.Threshold = 0.75
	err := ValidateConstitution(c)
	assert.NoError(t, err)
}

func TestValidateConstitution_CustomQuorumThresholdExactlyOne(t *testing.T) {
	c := validConstitution()
	c.Governance.Quorum.Type = "custom"
	c.Governance.Quorum.Threshold = 1.0
	err := ValidateConstitution(c)
	assert.NoError(t, err)
}

func TestValidateConstitution_NegativeTTL(t *testing.T) {
	c := validConstitution()
	c.Governance.ProposalTTLDays = -5
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "proposal_ttl_days must not be negative")
}

func TestValidateConstitution_ZeroTTL(t *testing.T) {
	c := validConstitution()
	c.Governance.ProposalTTLDays = 0
	err := ValidateConstitution(c)
	assert.NoError(t, err)
}

func TestValidateConstitution_EmptyRoles(t *testing.T) {
	c := validConstitution()
	c.Roles = nil
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "roles must not be empty")
}

func TestValidateConstitution_EmptyRoleMembers(t *testing.T) {
	c := validConstitution()
	c.Roles["techlead"] = Role{Members: nil}
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "members must not be empty")
}

func TestValidateConstitution_EmptyMemberEmail(t *testing.T) {
	c := validConstitution()
	c.Roles["techlead"] = Role{Members: []RoleMember{{Email: ""}}}
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email must not be empty")
}

func TestValidateConstitution_VoterRoleNotDefined(t *testing.T) {
	c := validConstitution()
	c.Governance.Voters = append(c.Governance.Voters, VoterRef{Role: "nonexistent"})
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "references role \"nonexistent\" which is not defined")
}

func TestValidateConstitution_InvalidLLMProvider(t *testing.T) {
	c := validConstitution()
	c.LLM.Provider = "gpt5"
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "llm.provider \"gpt5\" is invalid")
}

func TestValidateConstitution_ValidLLMProviders(t *testing.T) {
	providers := []string{"deepseek", "openai", "claude", "custom"}
	for _, p := range providers {
		t.Run(p, func(t *testing.T) {
			c := validConstitution()
			c.LLM.Provider = p
			if p == "custom" {
				c.LLM.Endpoint = "http://localhost:11434/v1"
			}
			err := ValidateConstitution(c)
			assert.NoError(t, err)
		})
	}
}

func TestValidateConstitution_CustomProviderRequiresEndpoint(t *testing.T) {
	c := validConstitution()
	c.LLM.Provider = "custom"
	c.LLM.Endpoint = ""
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "llm.endpoint is required when provider is custom")
}

func TestValidateConstitution_EmptyLLMProvider(t *testing.T) {
	c := validConstitution()
	c.LLM.Provider = ""
	err := ValidateConstitution(c)
	assert.NoError(t, err) // Empty provider is allowed (unconfigured)
}

func TestValidateConstitution_PerRuleOverrideInvalidQuorum(t *testing.T) {
	c := validConstitution()
	c.Governance.PerRuleOverrides = map[string]RuleOverride{
		"some_rule": {
			Quorum: QuorumConfig{Type: "invalid"},
		},
	}
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "per_rule_overrides[some_rule].quorum.type \"invalid\" is invalid")
}

func TestValidateConstitution_PerRuleOverrideEmptyQuorum(t *testing.T) {
	c := validConstitution()
	c.Governance.PerRuleOverrides = map[string]RuleOverride{
		"some_rule": {
			Quorum: QuorumConfig{Type: ""},
		},
	}
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "per_rule_overrides[some_rule].quorum.type must not be empty")
}

func TestValidateConstitution_PerRuleOverrideCustomBadThreshold(t *testing.T) {
	c := validConstitution()
	c.Governance.PerRuleOverrides = map[string]RuleOverride{
		"critical": {
			Quorum: QuorumConfig{Type: "custom", Threshold: 2.0},
		},
	}
	err := ValidateConstitution(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "per_rule_overrides[critical].quorum.threshold")
}

func TestValidateConstitution_MultipleErrors(t *testing.T) {
	c := &Constitution{
		Governance: Governance{
			Quorum:          QuorumConfig{Type: "invalid"},
			ProposalTTLDays: -1,
		},
		LLM: LLMConfig{Provider: "unknown"},
	}
	err := ValidateConstitution(c)
	require.Error(t, err)
	errStr := err.Error()
	assert.Contains(t, errStr, "governance.voters must not be empty")
	assert.Contains(t, errStr, "governance.quorum.type")
	assert.Contains(t, errStr, "proposal_ttl_days must not be negative")
	assert.Contains(t, errStr, "roles must not be empty")
	assert.Contains(t, errStr, "llm.provider")
}

// --- Rules validation tests ---

func validRulesFile() *RulesFile {
	return &RulesFile{
		Rules: []Rule{
			{
				ID:          "rule1",
				Description: "Test rule",
				Type:        "imports_forbidden",
				Severity:    "error",
			},
		},
	}
}

func TestValidateRules_Valid(t *testing.T) {
	r := validRulesFile()
	err := ValidateRules(r)
	assert.NoError(t, err)
}

func TestValidateRules_Nil(t *testing.T) {
	err := ValidateRules(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rules file is nil")
}

func TestValidateRules_EmptyRules(t *testing.T) {
	r := &RulesFile{Rules: []Rule{}}
	err := ValidateRules(r)
	assert.NoError(t, err) // Empty rules list is valid (no rules defined yet)
}

func TestValidateRules_EmptyID(t *testing.T) {
	r := validRulesFile()
	r.Rules[0].ID = ""
	err := ValidateRules(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rules[0].id must not be empty")
}

func TestValidateRules_DuplicateID(t *testing.T) {
	r := &RulesFile{
		Rules: []Rule{
			{ID: "dup", Description: "first", Type: "t1", Severity: "error"},
			{ID: "dup", Description: "second", Type: "t2", Severity: "warning"},
		},
	}
	err := ValidateRules(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rules[1].id \"dup\" is duplicated")
}

func TestValidateRules_EmptyDescription(t *testing.T) {
	r := validRulesFile()
	r.Rules[0].Description = ""
	err := ValidateRules(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rules[0].description must not be empty")
}

func TestValidateRules_EmptyType(t *testing.T) {
	r := validRulesFile()
	r.Rules[0].Type = ""
	err := ValidateRules(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rules[0].type must not be empty")
}

func TestValidateRules_EmptySeverity(t *testing.T) {
	r := validRulesFile()
	r.Rules[0].Severity = ""
	err := ValidateRules(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rules[0].severity must not be empty")
}

func TestValidateRules_InvalidSeverity(t *testing.T) {
	r := validRulesFile()
	r.Rules[0].Severity = "critical"
	err := ValidateRules(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rules[0].severity \"critical\" is invalid")
}

func TestValidateRules_ValidSeverities(t *testing.T) {
	for _, sev := range []string{"error", "warning"} {
		t.Run(sev, func(t *testing.T) {
			r := validRulesFile()
			r.Rules[0].Severity = sev
			err := ValidateRules(r)
			assert.NoError(t, err)
		})
	}
}

func TestValidateRules_MultipleErrors(t *testing.T) {
	r := &RulesFile{
		Rules: []Rule{
			{ID: "", Description: "", Type: "", Severity: "critical"},
		},
	}
	err := ValidateRules(r)
	require.Error(t, err)
	errStr := err.Error()
	assert.Contains(t, errStr, "rules[0].id must not be empty")
	assert.Contains(t, errStr, "rules[0].description must not be empty")
	assert.Contains(t, errStr, "rules[0].type must not be empty")
	assert.Contains(t, errStr, "rules[0].severity \"critical\" is invalid")
}

func TestValidateRules_MultipleRulesWithErrors(t *testing.T) {
	r := &RulesFile{
		Rules: []Rule{
			{ID: "ok_rule", Description: "Good", Type: "check", Severity: "error"},
			{ID: "", Description: "No ID", Type: "check", Severity: "error"},
			{ID: "bad_sev", Description: "Bad severity", Type: "check", Severity: "fatal"},
		},
	}
	err := ValidateRules(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rules[1].id must not be empty")
	assert.Contains(t, err.Error(), "rules[2].severity \"fatal\" is invalid")
}

// --- Proposal validation tests ---

func validProposal() *Proposal {
	return &Proposal{
		ID:           "2024-01-15-test",
		RuleID:       "test_rule",
		ProposalType: "modify",
		Change: ProposalChange{
			Description: "Test change",
		},
		Reason:    "Test reason",
		CreatedBy: "user@test.com",
		CreatedAt: time.Now(),
		Status:    "proposed",
	}
}

func TestValidateProposal_Valid(t *testing.T) {
	p := validProposal()
	err := ValidateProposal(p)
	assert.NoError(t, err)
}

func TestValidateProposal_Nil(t *testing.T) {
	err := ValidateProposal(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "proposal is nil")
}

func TestValidateProposal_EmptyID(t *testing.T) {
	p := validProposal()
	p.ID = ""
	err := ValidateProposal(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id must not be empty")
}

func TestValidateProposal_EmptyRuleID(t *testing.T) {
	p := validProposal()
	p.RuleID = ""
	err := ValidateProposal(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule_id must not be empty")
}

func TestValidateProposal_EmptyProposalType(t *testing.T) {
	p := validProposal()
	p.ProposalType = ""
	err := ValidateProposal(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "proposal_type must not be empty")
}

func TestValidateProposal_InvalidProposalType(t *testing.T) {
	p := validProposal()
	p.ProposalType = "update"
	err := ValidateProposal(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "proposal_type \"update\" is invalid")
}

func TestValidateProposal_ValidProposalTypes(t *testing.T) {
	for _, pt := range []string{"modify", "add", "remove"} {
		t.Run(pt, func(t *testing.T) {
			p := validProposal()
			p.ProposalType = pt
			err := ValidateProposal(p)
			assert.NoError(t, err)
		})
	}
}

func TestValidateProposal_EmptyChangeDescription(t *testing.T) {
	p := validProposal()
	p.Change.Description = ""
	err := ValidateProposal(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "change.description must not be empty")
}

func TestValidateProposal_EmptyReason(t *testing.T) {
	p := validProposal()
	p.Reason = ""
	err := ValidateProposal(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reason must not be empty")
}

func TestValidateProposal_EmptyCreatedBy(t *testing.T) {
	p := validProposal()
	p.CreatedBy = ""
	err := ValidateProposal(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "created_by must not be empty")
}

func TestValidateProposal_ZeroCreatedAt(t *testing.T) {
	p := validProposal()
	p.CreatedAt = time.Time{}
	err := ValidateProposal(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "created_at must not be zero")
}

func TestValidateProposal_EmptyStatus(t *testing.T) {
	p := validProposal()
	p.Status = ""
	err := ValidateProposal(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status must not be empty")
}

func TestValidateProposal_InvalidStatus(t *testing.T) {
	p := validProposal()
	p.Status = "pending"
	err := ValidateProposal(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status \"pending\" is invalid")
}

func TestValidateProposal_ValidStatuses(t *testing.T) {
	for _, s := range []string{"proposed", "accepted", "rejected", "withdrawn", "expired"} {
		t.Run(s, func(t *testing.T) {
			p := validProposal()
			p.Status = s
			err := ValidateProposal(p)
			assert.NoError(t, err)
		})
	}
}

func TestValidateProposal_MultipleErrors(t *testing.T) {
	p := &Proposal{}
	err := ValidateProposal(p)
	require.Error(t, err)
	errStr := err.Error()
	assert.Contains(t, errStr, "id must not be empty")
	assert.Contains(t, errStr, "rule_id must not be empty")
	assert.Contains(t, errStr, "proposal_type must not be empty")
	assert.Contains(t, errStr, "change.description must not be empty")
	assert.Contains(t, errStr, "reason must not be empty")
	assert.Contains(t, errStr, "created_by must not be empty")
	assert.Contains(t, errStr, "created_at must not be zero")
	assert.Contains(t, errStr, "status must not be empty")
}

// --- Vote validation tests ---

func validVote() *Vote {
	return &Vote{
		ProposalID: "2024-01-15-test",
		VoterEmail: "voter@test.com",
		Decision:   "yes",
		VotedAt:    time.Now(),
	}
}

func TestValidateVote_Valid(t *testing.T) {
	v := validVote()
	err := ValidateVote(v)
	assert.NoError(t, err)
}

func TestValidateVote_Nil(t *testing.T) {
	err := ValidateVote(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vote is nil")
}

func TestValidateVote_EmptyProposalID(t *testing.T) {
	v := validVote()
	v.ProposalID = ""
	err := ValidateVote(v)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "proposal_id must not be empty")
}

func TestValidateVote_EmptyVoterEmail(t *testing.T) {
	v := validVote()
	v.VoterEmail = ""
	err := ValidateVote(v)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "voter_email must not be empty")
}

func TestValidateVote_EmptyDecision(t *testing.T) {
	v := validVote()
	v.Decision = ""
	err := ValidateVote(v)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decision must not be empty")
}

func TestValidateVote_InvalidDecision(t *testing.T) {
	v := validVote()
	v.Decision = "abstain"
	err := ValidateVote(v)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decision \"abstain\" is invalid")
}

func TestValidateVote_ValidDecisions(t *testing.T) {
	for _, d := range []string{"yes", "no"} {
		t.Run(d, func(t *testing.T) {
			v := validVote()
			v.Decision = d
			err := ValidateVote(v)
			assert.NoError(t, err)
		})
	}
}

func TestValidateVote_ZeroVotedAt(t *testing.T) {
	v := validVote()
	v.VotedAt = time.Time{}
	err := ValidateVote(v)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "voted_at must not be zero")
}

func TestValidateVote_EmptyCommentIsValid(t *testing.T) {
	v := validVote()
	v.Comment = ""
	err := ValidateVote(v)
	assert.NoError(t, err)
}

func TestValidateVote_MultipleErrors(t *testing.T) {
	v := &Vote{}
	err := ValidateVote(v)
	require.Error(t, err)
	errStr := err.Error()
	assert.Contains(t, errStr, "proposal_id must not be empty")
	assert.Contains(t, errStr, "voter_email must not be empty")
	assert.Contains(t, errStr, "decision must not be empty")
	assert.Contains(t, errStr, "voted_at must not be zero")
}

// --- Exception validation tests ---

func validException() *Exception {
	return &Exception{
		ID:        "exc-test",
		RuleID:    "test_rule",
		Paths:     []string{"path/to/file.go"},
		Reason:    "Legacy code",
		CreatedBy: "dev@test.com",
		CreatedAt: time.Now(),
	}
}

func TestValidateException_Valid(t *testing.T) {
	e := validException()
	err := ValidateException(e)
	assert.NoError(t, err)
}

func TestValidateException_Nil(t *testing.T) {
	err := ValidateException(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exception is nil")
}

func TestValidateException_EmptyID(t *testing.T) {
	e := validException()
	e.ID = ""
	err := ValidateException(e)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id must not be empty")
}

func TestValidateException_EmptyRuleID(t *testing.T) {
	e := validException()
	e.RuleID = ""
	err := ValidateException(e)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule_id must not be empty")
}

func TestValidateException_EmptyPaths(t *testing.T) {
	e := validException()
	e.Paths = nil
	err := ValidateException(e)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "paths must not be empty")
}

func TestValidateException_EmptyReason(t *testing.T) {
	e := validException()
	e.Reason = ""
	err := ValidateException(e)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reason must not be empty")
}

func TestValidateException_EmptyCreatedBy(t *testing.T) {
	e := validException()
	e.CreatedBy = ""
	err := ValidateException(e)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "created_by must not be empty")
}

func TestValidateException_ZeroCreatedAt(t *testing.T) {
	e := validException()
	e.CreatedAt = time.Time{}
	err := ValidateException(e)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "created_at must not be zero")
}

func TestValidateException_ExpiresAtBeforeCreatedAt(t *testing.T) {
	e := validException()
	e.CreatedAt = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	e.ExpiresAt = &expiresAt
	err := ValidateException(e)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expires_at must not be before created_at")
}

func TestValidateException_NilExpiresAtIsValid(t *testing.T) {
	e := validException()
	e.ExpiresAt = nil
	err := ValidateException(e)
	assert.NoError(t, err)
}

func TestValidateException_ExpiresAtAfterCreatedAt(t *testing.T) {
	e := validException()
	e.CreatedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
	e.ExpiresAt = &expiresAt
	err := ValidateException(e)
	assert.NoError(t, err)
}

func TestValidateException_MultipleErrors(t *testing.T) {
	e := &Exception{}
	err := ValidateException(e)
	require.Error(t, err)
	errStr := err.Error()
	assert.Contains(t, errStr, "id must not be empty")
	assert.Contains(t, errStr, "rule_id must not be empty")
	assert.Contains(t, errStr, "paths must not be empty")
	assert.Contains(t, errStr, "reason must not be empty")
	assert.Contains(t, errStr, "created_by must not be empty")
	assert.Contains(t, errStr, "created_at must not be zero")
}
