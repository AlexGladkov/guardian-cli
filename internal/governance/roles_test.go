package governance

import (
	"sort"
	"testing"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/stretchr/testify/assert"
)

func makeConstitution() *config.Constitution {
	return &config.Constitution{
		Governance: config.Governance{
			Voters: []config.VoterRef{
				{Role: "techlead"},
				{Role: "architect"},
				{Role: "product"},
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

func TestGetEligibleVoters_Basic(t *testing.T) {
	c := makeConstitution()
	voters := GetEligibleVoters(c)

	sort.Strings(voters)
	assert.Equal(t, []string{"alex@company.com", "ivan@company.com", "maria@company.com"}, voters)
}

func TestGetEligibleVoters_Deduplication(t *testing.T) {
	c := &config.Constitution{
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
					{Email: "shared@company.com"},
				},
			},
			"architect": {
				Members: []config.RoleMember{
					{Email: "maria@company.com"},
					{Email: "shared@company.com"}, // duplicate
				},
			},
		},
	}

	voters := GetEligibleVoters(c)

	// Should have 3 unique voters, not 4.
	assert.Len(t, voters, 3)

	sort.Strings(voters)
	assert.Equal(t, []string{"ivan@company.com", "maria@company.com", "shared@company.com"}, voters)
}

func TestGetEligibleVoters_MissingRole(t *testing.T) {
	c := &config.Constitution{
		Governance: config.Governance{
			Voters: []config.VoterRef{
				{Role: "techlead"},
				{Role: "nonexistent"},
			},
		},
		Roles: map[string]config.Role{
			"techlead": {
				Members: []config.RoleMember{
					{Email: "ivan@company.com"},
				},
			},
		},
	}

	voters := GetEligibleVoters(c)
	assert.Len(t, voters, 1)
	assert.Equal(t, "ivan@company.com", voters[0])
}

func TestGetEligibleVoters_Empty(t *testing.T) {
	c := &config.Constitution{
		Governance: config.Governance{
			Voters: []config.VoterRef{},
		},
		Roles: map[string]config.Role{},
	}

	voters := GetEligibleVoters(c)
	assert.Empty(t, voters)
}

func TestGetUserRoles_SingleRole(t *testing.T) {
	c := makeConstitution()
	roles := GetUserRoles(c, "ivan@company.com")

	assert.Len(t, roles, 1)
	assert.Contains(t, roles, "techlead")
}

func TestGetUserRoles_MultipleRoles(t *testing.T) {
	c := &config.Constitution{
		Roles: map[string]config.Role{
			"techlead": {
				Members: []config.RoleMember{
					{Email: "superuser@company.com"},
				},
			},
			"architect": {
				Members: []config.RoleMember{
					{Email: "superuser@company.com"},
				},
			},
			"product": {
				Members: []config.RoleMember{
					{Email: "other@company.com"},
				},
			},
		},
	}

	roles := GetUserRoles(c, "superuser@company.com")

	sort.Strings(roles)
	assert.Equal(t, []string{"architect", "techlead"}, roles)
}

func TestGetUserRoles_NoRoles(t *testing.T) {
	c := makeConstitution()
	roles := GetUserRoles(c, "unknown@company.com")
	assert.Empty(t, roles)
}

func TestIsVoter_True(t *testing.T) {
	c := makeConstitution()
	assert.True(t, IsVoter(c, "ivan@company.com"))
	assert.True(t, IsVoter(c, "maria@company.com"))
	assert.True(t, IsVoter(c, "alex@company.com"))
}

func TestIsVoter_False(t *testing.T) {
	c := makeConstitution()
	assert.False(t, IsVoter(c, "unknown@company.com"))
}

func TestIsVoter_NonVoterRole(t *testing.T) {
	c := &config.Constitution{
		Governance: config.Governance{
			Voters: []config.VoterRef{
				{Role: "techlead"},
			},
		},
		Roles: map[string]config.Role{
			"techlead": {
				Members: []config.RoleMember{
					{Email: "ivan@company.com"},
				},
			},
			"developer": {
				Members: []config.RoleMember{
					{Email: "dev@company.com"},
				},
			},
		},
	}

	assert.True(t, IsVoter(c, "ivan@company.com"))
	assert.False(t, IsVoter(c, "dev@company.com"), "developer role is not a voter role")
}

func TestIsVoter_MissingVoterRole(t *testing.T) {
	c := &config.Constitution{
		Governance: config.Governance{
			Voters: []config.VoterRef{
				{Role: "nonexistent"},
			},
		},
		Roles: map[string]config.Role{},
	}

	assert.False(t, IsVoter(c, "anyone@company.com"))
}
