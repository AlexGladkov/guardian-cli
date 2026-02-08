package inbox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AlexGladkov/guardian-cli/internal/config"
)

// testConstitution creates a constitution with standard test data.
func testConstitution() *config.Constitution {
	return &config.Constitution{
		Governance: config.Governance{
			Voters: []config.VoterRef{
				{Role: "techlead"},
				{Role: "architect"},
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
			"developer": {
				Members: []config.RoleMember{
					{Email: "dev@company.com"},
				},
			},
		},
	}
}

// testProposal creates a test proposal with the given parameters.
func testProposal(id, ruleID, status, createdBy string, createdAt time.Time) *config.Proposal {
	return &config.Proposal{
		ID:           id,
		RuleID:       ruleID,
		ProposalType: "modify",
		Change: config.ProposalChange{
			Description: "Test change for " + ruleID,
		},
		Reason:    "Test reason",
		CreatedBy: createdBy,
		CreatedAt: createdAt,
		Status:    status,
	}
}

func TestGetInbox_UserNeedsToVote(t *testing.T) {
	constitution := testConstitution()
	now := time.Now()

	proposals := []*config.Proposal{
		testProposal("2024-01-15-rule1", "rule1", "proposed", "someone@company.com", now.Add(-2*24*time.Hour)),
	}
	votes := map[string][]*config.Vote{}

	items, err := GetInbox(proposals, votes, constitution, "ivan@company.com", nil)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "2024-01-15-rule1", items[0].Proposal.ID)
	assert.False(t, items[0].IsOld)
}

func TestGetInbox_UserAlreadyVoted(t *testing.T) {
	constitution := testConstitution()
	now := time.Now()

	proposals := []*config.Proposal{
		testProposal("2024-01-15-rule1", "rule1", "proposed", "someone@company.com", now.Add(-2*24*time.Hour)),
	}
	votes := map[string][]*config.Vote{
		"2024-01-15-rule1": {
			{
				ProposalID: "2024-01-15-rule1",
				VoterEmail: "ivan@company.com",
				Decision:   "yes",
				VotedAt:    now.Add(-1 * 24 * time.Hour),
			},
		},
	}

	items, err := GetInbox(proposals, votes, constitution, "ivan@company.com", nil)
	require.NoError(t, err)
	assert.Len(t, items, 0)
}

func TestGetInbox_UserIsNotAVoter(t *testing.T) {
	constitution := testConstitution()
	now := time.Now()

	proposals := []*config.Proposal{
		testProposal("2024-01-15-rule1", "rule1", "proposed", "someone@company.com", now.Add(-2*24*time.Hour)),
	}
	votes := map[string][]*config.Vote{}

	// dev@company.com has role "developer" which is not in governance.voters
	items, err := GetInbox(proposals, votes, constitution, "dev@company.com", nil)
	require.NoError(t, err)
	assert.Len(t, items, 0)
}

func TestGetInbox_UnknownUserIsNotAVoter(t *testing.T) {
	constitution := testConstitution()
	now := time.Now()

	proposals := []*config.Proposal{
		testProposal("2024-01-15-rule1", "rule1", "proposed", "someone@company.com", now.Add(-2*24*time.Hour)),
	}
	votes := map[string][]*config.Vote{}

	items, err := GetInbox(proposals, votes, constitution, "stranger@other.com", nil)
	require.NoError(t, err)
	assert.Len(t, items, 0)
}

func TestGetInbox_ProposalExpired(t *testing.T) {
	constitution := testConstitution()
	constitution.Governance.ProposalTTLDays = 30

	// Create a proposal that is 31 days old
	createdAt := time.Now().Add(-31 * 24 * time.Hour)
	proposals := []*config.Proposal{
		testProposal("2024-01-15-rule1", "rule1", "proposed", "someone@company.com", createdAt),
	}
	votes := map[string][]*config.Vote{}

	items, err := GetInbox(proposals, votes, constitution, "ivan@company.com", nil)
	require.NoError(t, err)
	assert.Len(t, items, 0)
}

func TestGetInbox_ProposalNotExpiredWhenTTLZero(t *testing.T) {
	constitution := testConstitution()
	constitution.Governance.ProposalTTLDays = 0 // no TTL

	// Create a proposal that is 100 days old
	createdAt := time.Now().Add(-100 * 24 * time.Hour)
	proposals := []*config.Proposal{
		testProposal("2024-01-15-rule1", "rule1", "proposed", "someone@company.com", createdAt),
	}
	votes := map[string][]*config.Vote{}

	items, err := GetInbox(proposals, votes, constitution, "ivan@company.com", nil)
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestGetInbox_SinceLastCheckFiltering(t *testing.T) {
	constitution := testConstitution()
	now := time.Now()
	since := now.Add(-3 * 24 * time.Hour) // 3 days ago

	proposals := []*config.Proposal{
		// Created 2 days ago (after sinceLastCheck) - should be included
		testProposal("new-proposal", "rule1", "proposed", "someone@company.com", now.Add(-2*24*time.Hour)),
		// Created 5 days ago (before sinceLastCheck) - should be excluded
		testProposal("old-proposal", "rule2", "proposed", "someone@company.com", now.Add(-5*24*time.Hour)),
	}
	votes := map[string][]*config.Vote{}

	items, err := GetInbox(proposals, votes, constitution, "ivan@company.com", &since)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "new-proposal", items[0].Proposal.ID)
}

func TestGetInbox_SinceLastCheckNil(t *testing.T) {
	constitution := testConstitution()
	constitution.Governance.ProposalTTLDays = 0 // no TTL for this test
	now := time.Now()

	proposals := []*config.Proposal{
		testProposal("new-proposal", "rule1", "proposed", "someone@company.com", now.Add(-2*24*time.Hour)),
		testProposal("old-proposal", "rule2", "proposed", "someone@company.com", now.Add(-50*24*time.Hour)),
	}
	votes := map[string][]*config.Vote{}

	// nil sinceLastCheck means show all
	items, err := GetInbox(proposals, votes, constitution, "ivan@company.com", nil)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestGetInbox_MultipleProposalsMixed(t *testing.T) {
	constitution := testConstitution()
	now := time.Now()

	proposals := []*config.Proposal{
		// Active, user hasn't voted - should be included
		testProposal("active-1", "rule1", "proposed", "someone@company.com", now.Add(-2*24*time.Hour)),
		// Active, user already voted - should be excluded
		testProposal("voted-1", "rule2", "proposed", "someone@company.com", now.Add(-3*24*time.Hour)),
		// Not proposed status - should be excluded
		testProposal("accepted-1", "rule3", "accepted", "someone@company.com", now.Add(-1*24*time.Hour)),
		// Active, user hasn't voted - should be included
		testProposal("active-2", "rule4", "proposed", "someone@company.com", now.Add(-1*24*time.Hour)),
		// Withdrawn - should be excluded
		testProposal("withdrawn-1", "rule5", "withdrawn", "someone@company.com", now.Add(-1*24*time.Hour)),
	}

	votes := map[string][]*config.Vote{
		"voted-1": {
			{
				ProposalID: "voted-1",
				VoterEmail: "ivan@company.com",
				Decision:   "yes",
				VotedAt:    now,
			},
		},
	}

	items, err := GetInbox(proposals, votes, constitution, "ivan@company.com", nil)
	require.NoError(t, err)
	require.Len(t, items, 2)

	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.Proposal.ID
	}
	assert.Contains(t, ids, "active-1")
	assert.Contains(t, ids, "active-2")
}

func TestGetInbox_OldProposal(t *testing.T) {
	constitution := testConstitution()

	// Create a proposal that is 10 days old (> 7 day threshold)
	createdAt := time.Now().Add(-10 * 24 * time.Hour)
	proposals := []*config.Proposal{
		testProposal("old-proposal", "rule1", "proposed", "someone@company.com", createdAt),
	}
	votes := map[string][]*config.Vote{}

	items, err := GetInbox(proposals, votes, constitution, "ivan@company.com", nil)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.True(t, items[0].IsOld)
	assert.Greater(t, items[0].Age, 7*24*time.Hour)
}

func TestGetInbox_RecentProposalNotOld(t *testing.T) {
	constitution := testConstitution()

	// Create a proposal that is 2 days old (< 7 day threshold)
	createdAt := time.Now().Add(-2 * 24 * time.Hour)
	proposals := []*config.Proposal{
		testProposal("recent-proposal", "rule1", "proposed", "someone@company.com", createdAt),
	}
	votes := map[string][]*config.Vote{}

	items, err := GetInbox(proposals, votes, constitution, "ivan@company.com", nil)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.False(t, items[0].IsOld)
}

func TestGetInbox_NilConstitution(t *testing.T) {
	proposals := []*config.Proposal{}
	votes := map[string][]*config.Vote{}

	items, err := GetInbox(proposals, votes, nil, "ivan@company.com", nil)
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "constitution is nil")
}

func TestGetInbox_EmptyProposals(t *testing.T) {
	constitution := testConstitution()
	votes := map[string][]*config.Vote{}

	items, err := GetInbox(nil, votes, constitution, "ivan@company.com", nil)
	require.NoError(t, err)
	assert.Len(t, items, 0)
}

func TestGetInbox_NilProposalInSlice(t *testing.T) {
	constitution := testConstitution()
	now := time.Now()

	proposals := []*config.Proposal{
		nil,
		testProposal("valid-1", "rule1", "proposed", "someone@company.com", now.Add(-1*24*time.Hour)),
		nil,
	}
	votes := map[string][]*config.Vote{}

	items, err := GetInbox(proposals, votes, constitution, "ivan@company.com", nil)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "valid-1", items[0].Proposal.ID)
}

func TestGetInbox_MultipleRolesForSameUser(t *testing.T) {
	// User has both techlead and architect roles
	constitution := &config.Constitution{
		Governance: config.Governance{
			Voters: []config.VoterRef{
				{Role: "techlead"},
				{Role: "architect"},
			},
		},
		Roles: map[string]config.Role{
			"techlead": {
				Members: []config.RoleMember{
					{Email: "ivan@company.com"},
				},
			},
			"architect": {
				Members: []config.RoleMember{
					{Email: "ivan@company.com"}, // same user in multiple roles
				},
			},
		},
	}

	now := time.Now()
	proposals := []*config.Proposal{
		testProposal("proposal-1", "rule1", "proposed", "someone@company.com", now.Add(-1*24*time.Hour)),
	}
	votes := map[string][]*config.Vote{}

	// Should still only appear once in inbox
	items, err := GetInbox(proposals, votes, constitution, "ivan@company.com", nil)
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestIsEligibleVoter(t *testing.T) {
	constitution := testConstitution()

	tests := []struct {
		name      string
		email     string
		wantVoter bool
	}{
		{"techlead is voter", "ivan@company.com", true},
		{"architect is voter", "maria@company.com", true},
		{"developer is not voter", "dev@company.com", false},
		{"unknown is not voter", "unknown@company.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEligibleVoter(constitution, tt.email)
			assert.Equal(t, tt.wantVoter, result)
		})
	}
}

func TestHasVoted(t *testing.T) {
	now := time.Now()

	votes := []*config.Vote{
		{VoterEmail: "ivan@company.com", Decision: "yes", VotedAt: now},
		{VoterEmail: "maria@company.com", Decision: "no", VotedAt: now},
	}

	assert.True(t, hasVoted(votes, "ivan@company.com"))
	assert.True(t, hasVoted(votes, "maria@company.com"))
	assert.False(t, hasVoted(votes, "unknown@company.com"))
	assert.False(t, hasVoted(nil, "ivan@company.com"))
	assert.False(t, hasVoted([]*config.Vote{}, "ivan@company.com"))
}
