package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadProposal_FullFile(t *testing.T) {
	content := `
id: "2024-01-15-domain_no_infra"
rule_id: domain_no_infra
proposal_type: modify
change:
  description: "Allow infra imports in domain/adapters/"
  details: "Domain adapters need to implement infra interfaces"
reason: "Domain adapters need to implement infra interfaces"
impact: "domain/adapters/ files can now import from infra/"
created_by: ivan@company.com
created_at: "2024-01-15T10:30:00Z"
status: proposed
`
	path := writeTestFile(t, "proposal.yml", content)

	p, err := LoadProposal(path)
	require.NoError(t, err)
	require.NotNil(t, p)

	assert.Equal(t, "2024-01-15-domain_no_infra", p.ID)
	assert.Equal(t, "domain_no_infra", p.RuleID)
	assert.Equal(t, "modify", p.ProposalType)
	assert.Equal(t, "Allow infra imports in domain/adapters/", p.Change.Description)
	assert.Equal(t, "Domain adapters need to implement infra interfaces", p.Change.Details)
	assert.Equal(t, "Domain adapters need to implement infra interfaces", p.Reason)
	assert.Equal(t, "domain/adapters/ files can now import from infra/", p.Impact)
	assert.Equal(t, "ivan@company.com", p.CreatedBy)
	assert.Equal(t, "proposed", p.Status)
	assert.False(t, p.CreatedAt.IsZero())
}

func TestLoadProposal_AllStatuses(t *testing.T) {
	statuses := []string{"proposed", "accepted", "rejected", "withdrawn", "expired"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			content := `
id: "test-proposal"
rule_id: test_rule
proposal_type: modify
change:
  description: "test"
reason: "test"
created_by: test@test.com
created_at: "2024-01-01T00:00:00Z"
status: ` + status

			path := writeTestFile(t, "proposal.yml", content)
			p, err := LoadProposal(path)
			require.NoError(t, err)
			assert.Equal(t, status, p.Status)
		})
	}
}

func TestLoadProposal_AllTypes(t *testing.T) {
	types := []string{"modify", "add", "remove"}

	for _, pt := range types {
		t.Run(pt, func(t *testing.T) {
			content := `
id: "test-proposal"
rule_id: test_rule
proposal_type: ` + pt + `
change:
  description: "test"
reason: "test"
created_by: test@test.com
created_at: "2024-01-01T00:00:00Z"
status: proposed`

			path := writeTestFile(t, "proposal.yml", content)
			p, err := LoadProposal(path)
			require.NoError(t, err)
			assert.Equal(t, pt, p.ProposalType)
		})
	}
}

func TestLoadProposal_FileNotFound(t *testing.T) {
	_, err := LoadProposal("/nonexistent/proposal.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading proposal file")
}

func TestLoadProposal_InvalidYAML(t *testing.T) {
	path := writeTestFile(t, "bad.yml", "{{invalid")
	_, err := LoadProposal(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing proposal file")
}

func TestSaveProposal(t *testing.T) {
	dir := t.TempDir()
	proposalsDir := filepath.Join(dir, "proposals")
	path := filepath.Join(proposalsDir, "test-proposal.yml")

	now := time.Now().UTC().Truncate(time.Second)
	p := &Proposal{
		ID:           "2024-01-15-test",
		RuleID:       "test_rule",
		ProposalType: "modify",
		Change: ProposalChange{
			Description: "Test change",
			Details:     "Detailed description",
		},
		Reason:    "Testing save",
		Impact:    "No impact",
		CreatedBy: "test@example.com",
		CreatedAt: now,
		Status:    "proposed",
	}

	err := SaveProposal(path, p)
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(proposalsDir)
	require.NoError(t, err)

	// Verify round-trip
	loaded, err := LoadProposal(path)
	require.NoError(t, err)
	assert.Equal(t, p.ID, loaded.ID)
	assert.Equal(t, p.RuleID, loaded.RuleID)
	assert.Equal(t, p.ProposalType, loaded.ProposalType)
	assert.Equal(t, p.Change.Description, loaded.Change.Description)
	assert.Equal(t, p.Change.Details, loaded.Change.Details)
	assert.Equal(t, p.Reason, loaded.Reason)
	assert.Equal(t, p.Impact, loaded.Impact)
	assert.Equal(t, p.CreatedBy, loaded.CreatedBy)
	assert.Equal(t, p.Status, loaded.Status)
	assert.True(t, p.CreatedAt.Equal(loaded.CreatedAt))
}

func TestLoadAllProposals(t *testing.T) {
	dir := t.TempDir()

	now := time.Now().UTC().Truncate(time.Second)

	proposals := []*Proposal{
		{
			ID: "proposal-1", RuleID: "rule1", ProposalType: "modify",
			Change: ProposalChange{Description: "Change 1"},
			Reason: "R1", CreatedBy: "a@test.com", CreatedAt: now, Status: "proposed",
		},
		{
			ID: "proposal-2", RuleID: "rule2", ProposalType: "add",
			Change: ProposalChange{Description: "Change 2"},
			Reason: "R2", CreatedBy: "b@test.com", CreatedAt: now, Status: "accepted",
		},
		{
			ID: "proposal-3", RuleID: "rule3", ProposalType: "remove",
			Change: ProposalChange{Description: "Change 3"},
			Reason: "R3", CreatedBy: "c@test.com", CreatedAt: now, Status: "rejected",
		},
	}

	for _, p := range proposals {
		err := SaveProposal(filepath.Join(dir, p.ID+".yml"), p)
		require.NoError(t, err)
	}

	// Also add a non-YAML file that should be skipped
	err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("ignore me"), 0644)
	require.NoError(t, err)

	loaded, err := LoadAllProposals(dir)
	require.NoError(t, err)
	assert.Len(t, loaded, 3)

	// Verify the proposals were loaded (order depends on ReadDir which is alphabetical)
	ids := make(map[string]bool)
	for _, p := range loaded {
		ids[p.ID] = true
	}
	assert.True(t, ids["proposal-1"])
	assert.True(t, ids["proposal-2"])
	assert.True(t, ids["proposal-3"])
}

func TestLoadAllProposals_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	loaded, err := LoadAllProposals(dir)
	require.NoError(t, err)
	assert.Empty(t, loaded)
}

func TestLoadAllProposals_NonexistentDir(t *testing.T) {
	loaded, err := LoadAllProposals("/nonexistent/proposals")
	require.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestLoadAllProposals_SkipsSubdirectories(t *testing.T) {
	dir := t.TempDir()

	// Create a proposal file
	now := time.Now().UTC().Truncate(time.Second)
	p := &Proposal{
		ID: "proposal-1", RuleID: "rule1", ProposalType: "modify",
		Change: ProposalChange{Description: "Change 1"},
		Reason: "R1", CreatedBy: "a@test.com", CreatedAt: now, Status: "proposed",
	}
	err := SaveProposal(filepath.Join(dir, "proposal-1.yml"), p)
	require.NoError(t, err)

	// Create a subdirectory
	err = os.MkdirAll(filepath.Join(dir, "subdir"), 0755)
	require.NoError(t, err)

	loaded, err := LoadAllProposals(dir)
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
}

func TestLoadAllProposals_YamlExtension(t *testing.T) {
	dir := t.TempDir()

	now := time.Now().UTC().Truncate(time.Second)

	// Create .yml file
	p1 := &Proposal{
		ID: "p1", RuleID: "r1", ProposalType: "modify",
		Change: ProposalChange{Description: "C1"},
		Reason: "R1", CreatedBy: "a@test.com", CreatedAt: now, Status: "proposed",
	}
	err := SaveProposal(filepath.Join(dir, "p1.yml"), p1)
	require.NoError(t, err)

	// Create .yaml file manually
	p2 := &Proposal{
		ID: "p2", RuleID: "r2", ProposalType: "add",
		Change: ProposalChange{Description: "C2"},
		Reason: "R2", CreatedBy: "b@test.com", CreatedAt: now, Status: "proposed",
	}
	err = SaveProposal(filepath.Join(dir, "p2.yaml"), p2)
	require.NoError(t, err)

	loaded, err := LoadAllProposals(dir)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
}
