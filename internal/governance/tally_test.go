package governance

import (
	"testing"
	"time"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestConstitution() *config.Constitution {
	return &config.Constitution{
		Governance: config.Governance{
			Voters: []config.VoterRef{
				{Role: "techlead"},
				{Role: "architect"},
				{Role: "product"},
			},
			Quorum: config.QuorumConfig{
				Type: "two_thirds",
			},
			ProposalTTLDays: 30,
		},
		Roles: map[string]config.Role{
			"techlead": {
				Members: []config.RoleMember{
					{Email: "ivan@company.com"},
				},
			},
			"architect": {
				Members: []config.RoleMember{
					{Email: "maria@company.com"},
				},
			},
			"product": {
				Members: []config.RoleMember{
					{Email: "alex@company.com"},
				},
			},
		},
	}
}

func TestComputeTally_Accepted(t *testing.T) {
	c := makeTestConstitution()
	proposal := &config.Proposal{
		ID:        "2024-01-15-domain_no_infra",
		RuleID:    "domain_no_infra",
		CreatedAt: time.Now().Add(-24 * time.Hour), // 1 day ago
	}

	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "yes"},
		{ProposalID: proposal.ID, VoterEmail: "maria@company.com", Decision: "yes"},
	}

	tally := ComputeTally(proposal, votes, c)

	assert.Equal(t, proposal.ID, tally.ProposalID)
	assert.Equal(t, "domain_no_infra", tally.RuleID)
	assert.Len(t, tally.EligibleVoters, 3)
	assert.Equal(t, 2, tally.QuorumResult.Required) // ceil(3 * 2/3) = 2
	assert.Equal(t, 2, tally.QuorumResult.YesVotes)
	assert.Equal(t, 0, tally.QuorumResult.NoVotes)
	assert.Equal(t, "ACCEPTED", tally.QuorumResult.Result)
	assert.False(t, tally.IsExpired)
}

func TestComputeTally_Pending(t *testing.T) {
	c := makeTestConstitution()
	proposal := &config.Proposal{
		ID:        "2024-01-15-domain_no_infra",
		RuleID:    "domain_no_infra",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "yes"},
	}

	tally := ComputeTally(proposal, votes, c)

	assert.Equal(t, "PENDING", tally.QuorumResult.Result)
	assert.Equal(t, 1, tally.QuorumResult.YesVotes)
}

func TestComputeTally_Rejected(t *testing.T) {
	c := makeTestConstitution()
	proposal := &config.Proposal{
		ID:        "2024-01-15-domain_no_infra",
		RuleID:    "domain_no_infra",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "no"},
		{ProposalID: proposal.ID, VoterEmail: "maria@company.com", Decision: "no"},
	}

	tally := ComputeTally(proposal, votes, c)

	assert.Equal(t, "REJECTED", tally.QuorumResult.Result)
	assert.Equal(t, 2, tally.QuorumResult.NoVotes)
}

func TestComputeTally_TTLExpired(t *testing.T) {
	c := makeTestConstitution()
	c.Governance.ProposalTTLDays = 7

	proposal := &config.Proposal{
		ID:        "2024-01-01-domain_no_infra",
		RuleID:    "domain_no_infra",
		CreatedAt: time.Now().Add(-10 * 24 * time.Hour), // 10 days ago, TTL is 7
	}

	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "yes"},
		{ProposalID: proposal.ID, VoterEmail: "maria@company.com", Decision: "yes"},
	}

	tally := ComputeTally(proposal, votes, c)

	assert.True(t, tally.IsExpired)
	assert.Equal(t, "EXPIRED", tally.QuorumResult.Result)
}

func TestComputeTally_TTLNotExpired(t *testing.T) {
	c := makeTestConstitution()
	c.Governance.ProposalTTLDays = 30

	proposal := &config.Proposal{
		ID:        "2024-01-15-domain_no_infra",
		RuleID:    "domain_no_infra",
		CreatedAt: time.Now().Add(-5 * 24 * time.Hour), // 5 days ago, TTL is 30
	}

	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "yes"},
		{ProposalID: proposal.ID, VoterEmail: "maria@company.com", Decision: "yes"},
	}

	tally := ComputeTally(proposal, votes, c)

	assert.False(t, tally.IsExpired)
	assert.Equal(t, "ACCEPTED", tally.QuorumResult.Result)
}

func TestComputeTally_NoTTL(t *testing.T) {
	c := makeTestConstitution()
	c.Governance.ProposalTTLDays = 0 // No TTL

	proposal := &config.Proposal{
		ID:        "2023-01-01-domain_no_infra",
		RuleID:    "domain_no_infra",
		CreatedAt: time.Now().Add(-365 * 24 * time.Hour), // 1 year ago
	}

	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "yes"},
	}

	tally := ComputeTally(proposal, votes, c)

	assert.False(t, tally.IsExpired)
	assert.Equal(t, "PENDING", tally.QuorumResult.Result)
}

func TestComputeTally_PerRuleOverride(t *testing.T) {
	c := makeTestConstitution()
	c.Governance.PerRuleOverrides = map[string]config.RuleOverride{
		"payment_state_machine": {
			Quorum: config.QuorumConfig{
				Type: "unanimous",
			},
		},
	}

	proposal := &config.Proposal{
		ID:        "2024-01-15-payment_state_machine",
		RuleID:    "payment_state_machine",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "yes"},
		{ProposalID: proposal.ID, VoterEmail: "maria@company.com", Decision: "yes"},
	}

	tally := ComputeTally(proposal, votes, c)

	// Unanimous requires 3, only 2 yes.
	assert.Equal(t, 3, tally.QuorumResult.Required)
	assert.Equal(t, "PENDING", tally.QuorumResult.Result)
	assert.Equal(t, "unanimous", tally.QuorumConfig.Type)
}

func TestComputeTally_PerRuleOverride_Accepted(t *testing.T) {
	c := makeTestConstitution()
	c.Governance.PerRuleOverrides = map[string]config.RuleOverride{
		"payment_state_machine": {
			Quorum: config.QuorumConfig{
				Type: "unanimous",
			},
		},
	}

	proposal := &config.Proposal{
		ID:        "2024-01-15-payment_state_machine",
		RuleID:    "payment_state_machine",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "yes"},
		{ProposalID: proposal.ID, VoterEmail: "maria@company.com", Decision: "yes"},
		{ProposalID: proposal.ID, VoterEmail: "alex@company.com", Decision: "yes"},
	}

	tally := ComputeTally(proposal, votes, c)

	assert.Equal(t, 3, tally.QuorumResult.Required)
	assert.Equal(t, "ACCEPTED", tally.QuorumResult.Result)
}

func TestComputeTally_IneligibleVoterIgnored(t *testing.T) {
	c := makeTestConstitution()
	proposal := &config.Proposal{
		ID:        "2024-01-15-domain_no_infra",
		RuleID:    "domain_no_infra",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "yes"},
		{ProposalID: proposal.ID, VoterEmail: "outsider@other.com", Decision: "yes"}, // not eligible
	}

	tally := ComputeTally(proposal, votes, c)

	// Only ivan's vote counts.
	assert.Equal(t, 1, tally.QuorumResult.YesVotes)
	assert.Equal(t, "PENDING", tally.QuorumResult.Result)
}

func TestComputeTally_NoVotes(t *testing.T) {
	c := makeTestConstitution()
	proposal := &config.Proposal{
		ID:        "2024-01-15-domain_no_infra",
		RuleID:    "domain_no_infra",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	tally := ComputeTally(proposal, nil, c)

	assert.Equal(t, 0, tally.QuorumResult.YesVotes)
	assert.Equal(t, 0, tally.QuorumResult.NoVotes)
	assert.Equal(t, "PENDING", tally.QuorumResult.Result)
}

func TestComputeTally_DefaultQuorum(t *testing.T) {
	c := makeTestConstitution()
	// No per-rule override for this rule, should use default two_thirds.

	proposal := &config.Proposal{
		ID:        "2024-01-15-some_rule",
		RuleID:    "some_rule",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "yes"},
		{ProposalID: proposal.ID, VoterEmail: "maria@company.com", Decision: "yes"},
	}

	tally := ComputeTally(proposal, votes, c)

	assert.Equal(t, "two_thirds", tally.QuorumConfig.Type)
	assert.Equal(t, 2, tally.QuorumResult.Required) // ceil(3 * 2/3) = 2
	assert.Equal(t, "ACCEPTED", tally.QuorumResult.Result)
}

func TestComputeTally_VotesListPreserved(t *testing.T) {
	c := makeTestConstitution()
	proposal := &config.Proposal{
		ID:        "2024-01-15-domain_no_infra",
		RuleID:    "domain_no_infra",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "yes"},
		{ProposalID: proposal.ID, VoterEmail: "maria@company.com", Decision: "no"},
	}

	tally := ComputeTally(proposal, votes, c)

	require.Len(t, tally.Votes, 2)
	assert.Equal(t, "ivan@company.com", tally.Votes[0].VoterEmail)
	assert.Equal(t, "maria@company.com", tally.Votes[1].VoterEmail)
}

func TestComputeTally_ExpiredOverridesAccepted(t *testing.T) {
	c := makeTestConstitution()
	c.Governance.ProposalTTLDays = 1

	proposal := &config.Proposal{
		ID:        "2024-01-01-domain_no_infra",
		RuleID:    "domain_no_infra",
		CreatedAt: time.Now().Add(-3 * 24 * time.Hour), // 3 days ago, TTL 1 day
	}

	// Enough votes for acceptance.
	votes := []*config.Vote{
		{ProposalID: proposal.ID, VoterEmail: "ivan@company.com", Decision: "yes"},
		{ProposalID: proposal.ID, VoterEmail: "maria@company.com", Decision: "yes"},
		{ProposalID: proposal.ID, VoterEmail: "alex@company.com", Decision: "yes"},
	}

	tally := ComputeTally(proposal, votes, c)

	assert.True(t, tally.IsExpired)
	assert.Equal(t, "EXPIRED", tally.QuorumResult.Result,
		"expired should override even if quorum was met")
}
