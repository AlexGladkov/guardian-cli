package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaCheck_NoProtectedFilesChanged(t *testing.T) {
	checker := &MetaChecker{}
	changedFiles := []string{
		"domain/service/UserService.kt",
		"build.gradle",
	}

	violations, err := checker.Check(changedFiles, "/nonexistent")
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestMetaCheck_ConstitutionChangedWithoutProposal(t *testing.T) {
	tmpDir := t.TempDir()
	proposalsDir := filepath.Join(tmpDir, "proposals")
	require.NoError(t, os.MkdirAll(proposalsDir, 0755))

	checker := &MetaChecker{}
	changedFiles := []string{
		".agreements/constitution.yml",
		"domain/service/UserService.kt",
	}

	violations, err := checker.Check(changedFiles, proposalsDir)
	require.NoError(t, err)

	assert.Len(t, violations, 1)
	assert.Equal(t, "meta_check", violations[0].RuleID)
	assert.Equal(t, "error", violations[0].Severity)
	assert.Contains(t, violations[0].Description, "constitution.yml")
}

func TestMetaCheck_RulesChangedWithoutProposal(t *testing.T) {
	tmpDir := t.TempDir()
	proposalsDir := filepath.Join(tmpDir, "proposals")
	require.NoError(t, os.MkdirAll(proposalsDir, 0755))

	checker := &MetaChecker{}
	changedFiles := []string{".agreements/rules.yml"}

	violations, err := checker.Check(changedFiles, proposalsDir)
	require.NoError(t, err)

	assert.Len(t, violations, 1)
	assert.Contains(t, violations[0].Description, "rules.yml")
}

func TestMetaCheck_BothFilesChangedWithoutProposal(t *testing.T) {
	tmpDir := t.TempDir()
	proposalsDir := filepath.Join(tmpDir, "proposals")
	require.NoError(t, os.MkdirAll(proposalsDir, 0755))

	checker := &MetaChecker{}
	changedFiles := []string{
		".agreements/constitution.yml",
		".agreements/rules.yml",
	}

	violations, err := checker.Check(changedFiles, proposalsDir)
	require.NoError(t, err)

	assert.Len(t, violations, 2)
}

func TestMetaCheck_ConstitutionChangedWithAcceptedProposal(t *testing.T) {
	tmpDir := t.TempDir()
	proposalsDir := filepath.Join(tmpDir, "proposals")
	require.NoError(t, os.MkdirAll(proposalsDir, 0755))

	// Create an accepted proposal file.
	proposalContent := `id: "2024-01-15-constitution-change"
rule_id: constitution
proposal_type: modify
status: accepted
`
	require.NoError(t, os.WriteFile(
		filepath.Join(proposalsDir, "2024-01-15-constitution-change.yml"),
		[]byte(proposalContent),
		0644,
	))

	checker := &MetaChecker{}
	changedFiles := []string{".agreements/constitution.yml"}

	violations, err := checker.Check(changedFiles, proposalsDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestMetaCheck_ProposalExistsButNotAccepted(t *testing.T) {
	tmpDir := t.TempDir()
	proposalsDir := filepath.Join(tmpDir, "proposals")
	require.NoError(t, os.MkdirAll(proposalsDir, 0755))

	proposalContent := `id: "2024-01-15-some-change"
rule_id: some_rule
proposal_type: modify
status: proposed
`
	require.NoError(t, os.WriteFile(
		filepath.Join(proposalsDir, "2024-01-15-some-change.yml"),
		[]byte(proposalContent),
		0644,
	))

	checker := &MetaChecker{}
	changedFiles := []string{".agreements/constitution.yml"}

	violations, err := checker.Check(changedFiles, proposalsDir)
	require.NoError(t, err)

	assert.Len(t, violations, 1)
}

func TestMetaCheck_NonexistentProposalsDir(t *testing.T) {
	checker := &MetaChecker{}
	changedFiles := []string{".agreements/constitution.yml"}

	violations, err := checker.Check(changedFiles, "/nonexistent/proposals")
	require.NoError(t, err)

	assert.Len(t, violations, 1)
}

func TestMetaCheck_OtherAgreementsFilesNotProtected(t *testing.T) {
	checker := &MetaChecker{}
	changedFiles := []string{
		".agreements/proposals/some-proposal.yml",
		".agreements/exceptions/some-exception.yml",
	}

	violations, err := checker.Check(changedFiles, "/nonexistent")
	require.NoError(t, err)
	assert.Empty(t, violations)
}
