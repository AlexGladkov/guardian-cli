package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConstitution_FullFile(t *testing.T) {
	content := `
governance:
  voters:
    - role: techlead
    - role: architect
  quorum:
    type: two_thirds
    threshold: 0.66
  forbid_self_approval: true
  allow_vote_change: false
  proposal_ttl_days: 30
  per_rule_overrides:
    payment_rule:
      quorum:
        type: unanimous
  exceptions:
    require_approval: false
identity:
  allowed_domains: ["company.com", "corp.io"]
  require_signed_commits: false
roles:
  techlead:
    members:
      - email: ivan@company.com
  architect:
    members:
      - email: maria@company.com
      - email: alex@company.com
llm:
  provider: deepseek
  endpoint: ""
  model: "deepseek-chat"
  prompts:
    check_system: "custom check prompt"
    propose_system: "custom propose prompt"
`
	path := writeTestFile(t, "constitution.yml", content)

	c, err := LoadConstitution(path)
	require.NoError(t, err)
	require.NotNil(t, c)

	// Governance
	assert.Len(t, c.Governance.Voters, 2)
	assert.Equal(t, "techlead", c.Governance.Voters[0].Role)
	assert.Equal(t, "architect", c.Governance.Voters[1].Role)
	assert.Equal(t, "two_thirds", c.Governance.Quorum.Type)
	assert.InDelta(t, 0.66, c.Governance.Quorum.Threshold, 0.001)
	assert.True(t, c.Governance.ForbidSelfApproval)
	assert.False(t, c.Governance.AllowVoteChange)
	assert.Equal(t, 30, c.Governance.ProposalTTLDays)

	// Per-rule overrides
	require.Contains(t, c.Governance.PerRuleOverrides, "payment_rule")
	assert.Equal(t, "unanimous", c.Governance.PerRuleOverrides["payment_rule"].Quorum.Type)

	// Exceptions
	assert.False(t, c.Governance.Exceptions.RequireApproval)

	// Identity
	assert.Equal(t, []string{"company.com", "corp.io"}, c.Identity.AllowedDomains)
	assert.False(t, c.Identity.RequireSignedCommits)

	// Roles
	require.Len(t, c.Roles, 2)
	assert.Len(t, c.Roles["techlead"].Members, 1)
	assert.Equal(t, "ivan@company.com", c.Roles["techlead"].Members[0].Email)
	assert.Len(t, c.Roles["architect"].Members, 2)

	// LLM
	assert.Equal(t, "deepseek", c.LLM.Provider)
	assert.Equal(t, "deepseek-chat", c.LLM.Model)
	assert.Equal(t, "custom check prompt", c.LLM.Prompts.CheckSystem)
	assert.Equal(t, "custom propose prompt", c.LLM.Prompts.ProposeSystem)
}

func TestLoadConstitution_MinimalFile(t *testing.T) {
	content := `
governance:
  voters:
    - role: admin
  quorum:
    type: majority
roles:
  admin:
    members:
      - email: admin@test.com
llm:
  provider: openai
`
	path := writeTestFile(t, "constitution.yml", content)

	c, err := LoadConstitution(path)
	require.NoError(t, err)
	require.NotNil(t, c)

	assert.Equal(t, "majority", c.Governance.Quorum.Type)
	assert.False(t, c.Governance.ForbidSelfApproval)
	assert.False(t, c.Governance.AllowVoteChange)
	assert.Equal(t, 0, c.Governance.ProposalTTLDays)
	assert.Empty(t, c.Identity.AllowedDomains)
	assert.Equal(t, "openai", c.LLM.Provider)
}

func TestLoadConstitution_FileNotFound(t *testing.T) {
	_, err := LoadConstitution("/nonexistent/path/constitution.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading constitution file")
}

func TestLoadConstitution_InvalidYAML(t *testing.T) {
	path := writeTestFile(t, "bad.yml", "{{invalid yaml content")

	_, err := LoadConstitution(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing constitution file")
}

func TestSaveConstitution(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "constitution.yml")

	c := &Constitution{
		Governance: Governance{
			Voters: []VoterRef{
				{Role: "techlead"},
			},
			Quorum: QuorumConfig{
				Type:      "majority",
				Threshold: 0.5,
			},
			ForbidSelfApproval: true,
			ProposalTTLDays:    14,
		},
		Roles: map[string]Role{
			"techlead": {
				Members: []RoleMember{
					{Email: "test@example.com"},
				},
			},
		},
		LLM: LLMConfig{
			Provider: "openai",
			Model:    "gpt-4",
		},
	}

	err := SaveConstitution(path, c)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(path)
	require.NoError(t, err)

	// Load it back and verify
	loaded, err := LoadConstitution(path)
	require.NoError(t, err)
	assert.Equal(t, c.Governance.Quorum.Type, loaded.Governance.Quorum.Type)
	assert.Equal(t, c.Governance.ForbidSelfApproval, loaded.Governance.ForbidSelfApproval)
	assert.Equal(t, c.Governance.ProposalTTLDays, loaded.Governance.ProposalTTLDays)
	assert.Equal(t, c.LLM.Provider, loaded.LLM.Provider)
	assert.Equal(t, c.LLM.Model, loaded.LLM.Model)
	assert.Equal(t, "test@example.com", loaded.Roles["techlead"].Members[0].Email)
}

func TestSaveConstitution_InvalidPath(t *testing.T) {
	err := SaveConstitution("/nonexistent/deep/path/constitution.yml", &Constitution{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "writing constitution file")
}

func TestLoadConstitution_RoundTrip(t *testing.T) {
	original := &Constitution{
		Governance: Governance{
			Voters: []VoterRef{
				{Role: "lead"},
				{Role: "arch"},
			},
			Quorum: QuorumConfig{
				Type:      "custom",
				Threshold: 0.75,
			},
			ForbidSelfApproval: true,
			AllowVoteChange:    true,
			ProposalTTLDays:    60,
			PerRuleOverrides: map[string]RuleOverride{
				"critical_rule": {
					Quorum: QuorumConfig{Type: "unanimous"},
				},
			},
			Exceptions: ExceptionPolicy{RequireApproval: true},
		},
		Identity: Identity{
			AllowedDomains:       []string{"example.com"},
			RequireSignedCommits: true,
		},
		Roles: map[string]Role{
			"lead": {Members: []RoleMember{{Email: "lead@example.com"}}},
			"arch": {Members: []RoleMember{{Email: "arch@example.com"}}},
		},
		LLM: LLMConfig{
			Provider: "custom",
			Endpoint: "http://localhost:11434/v1",
			Model:    "llama3",
			Prompts: LLMPrompts{
				CheckSystem:   "check prompt",
				ProposeSystem: "propose prompt",
			},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "constitution.yml")

	err := SaveConstitution(path, original)
	require.NoError(t, err)

	loaded, err := LoadConstitution(path)
	require.NoError(t, err)

	assert.Equal(t, original.Governance.Quorum, loaded.Governance.Quorum)
	assert.Equal(t, original.Governance.ForbidSelfApproval, loaded.Governance.ForbidSelfApproval)
	assert.Equal(t, original.Governance.AllowVoteChange, loaded.Governance.AllowVoteChange)
	assert.Equal(t, original.Identity.AllowedDomains, loaded.Identity.AllowedDomains)
	assert.Equal(t, original.Identity.RequireSignedCommits, loaded.Identity.RequireSignedCommits)
	assert.Equal(t, original.LLM.Endpoint, loaded.LLM.Endpoint)
	assert.Equal(t, original.LLM.Model, loaded.LLM.Model)
	assert.Equal(t, original.Governance.Exceptions.RequireApproval, loaded.Governance.Exceptions.RequireApproval)
}

func TestLoadConstitution_EmptyFile(t *testing.T) {
	path := writeTestFile(t, "empty.yml", "")

	c, err := LoadConstitution(path)
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Nil(t, c.Roles)
	assert.Nil(t, c.Governance.Voters)
}

// writeTestFile creates a temp file with the given content and returns its path.
func writeTestFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}
