package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadVote_FullFile(t *testing.T) {
	content := `
proposal_id: "2024-01-15-domain_no_infra"
voter_email: maria@company.com
decision: "yes"
comment: "Makes sense for adapter pattern"
voted_at: "2024-01-16T14:00:00Z"
`
	path := writeTestFile(t, "vote.yml", content)

	v, err := LoadVote(path)
	require.NoError(t, err)
	require.NotNil(t, v)

	assert.Equal(t, "2024-01-15-domain_no_infra", v.ProposalID)
	assert.Equal(t, "maria@company.com", v.VoterEmail)
	assert.Equal(t, "yes", v.Decision)
	assert.Equal(t, "Makes sense for adapter pattern", v.Comment)
	assert.False(t, v.VotedAt.IsZero())
}

func TestLoadVote_NoDecision(t *testing.T) {
	content := `
proposal_id: "2024-01-15-domain_no_infra"
voter_email: maria@company.com
decision: "no"
comment: "I disagree with this change"
voted_at: "2024-01-16T14:00:00Z"
`
	path := writeTestFile(t, "vote.yml", content)

	v, err := LoadVote(path)
	require.NoError(t, err)
	assert.Equal(t, "no", v.Decision)
}

func TestLoadVote_EmptyComment(t *testing.T) {
	content := `
proposal_id: "test-proposal"
voter_email: test@test.com
decision: "yes"
comment: ""
voted_at: "2024-01-16T14:00:00Z"
`
	path := writeTestFile(t, "vote.yml", content)

	v, err := LoadVote(path)
	require.NoError(t, err)
	assert.Equal(t, "", v.Comment)
}

func TestLoadVote_FileNotFound(t *testing.T) {
	_, err := LoadVote("/nonexistent/vote.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading vote file")
}

func TestLoadVote_InvalidYAML(t *testing.T) {
	path := writeTestFile(t, "bad.yml", "{{invalid")
	_, err := LoadVote(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing vote file")
}

func TestSaveVote(t *testing.T) {
	dir := t.TempDir()
	votesDir := filepath.Join(dir, "votes", "proposal-1")
	path := filepath.Join(votesDir, "voter@test.com.yml")

	now := time.Now().UTC().Truncate(time.Second)
	v := &Vote{
		ProposalID: "proposal-1",
		VoterEmail: "voter@test.com",
		Decision:   "yes",
		Comment:    "LGTM",
		VotedAt:    now,
	}

	err := SaveVote(path, v)
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(votesDir)
	require.NoError(t, err)

	// Round-trip verification
	loaded, err := LoadVote(path)
	require.NoError(t, err)
	assert.Equal(t, v.ProposalID, loaded.ProposalID)
	assert.Equal(t, v.VoterEmail, loaded.VoterEmail)
	assert.Equal(t, v.Decision, loaded.Decision)
	assert.Equal(t, v.Comment, loaded.Comment)
	assert.True(t, v.VotedAt.Equal(loaded.VotedAt))
}

func TestLoadVotesForProposal(t *testing.T) {
	dir := t.TempDir()
	proposalID := "2024-01-15-test"
	proposalVotesDir := filepath.Join(dir, proposalID)

	now := time.Now().UTC().Truncate(time.Second)

	votes := []*Vote{
		{
			ProposalID: proposalID,
			VoterEmail: "alice@test.com",
			Decision:   "yes",
			Comment:    "Approve",
			VotedAt:    now,
		},
		{
			ProposalID: proposalID,
			VoterEmail: "bob@test.com",
			Decision:   "no",
			Comment:    "Reject",
			VotedAt:    now,
		},
		{
			ProposalID: proposalID,
			VoterEmail: "carol@test.com",
			Decision:   "yes",
			Comment:    "",
			VotedAt:    now,
		},
	}

	for _, v := range votes {
		err := SaveVote(filepath.Join(proposalVotesDir, v.VoterEmail+".yml"), v)
		require.NoError(t, err)
	}

	// Add a non-YAML file to be skipped
	err := os.WriteFile(filepath.Join(proposalVotesDir, "notes.txt"), []byte("skip"), 0644)
	require.NoError(t, err)

	loaded, err := LoadVotesForProposal(dir, proposalID)
	require.NoError(t, err)
	assert.Len(t, loaded, 3)

	emails := make(map[string]string)
	for _, v := range loaded {
		emails[v.VoterEmail] = v.Decision
	}
	assert.Equal(t, "yes", emails["alice@test.com"])
	assert.Equal(t, "no", emails["bob@test.com"])
	assert.Equal(t, "yes", emails["carol@test.com"])
}

func TestLoadVotesForProposal_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	proposalID := "empty-proposal"
	err := os.MkdirAll(filepath.Join(dir, proposalID), 0755)
	require.NoError(t, err)

	loaded, err := LoadVotesForProposal(dir, proposalID)
	require.NoError(t, err)
	assert.Empty(t, loaded)
}

func TestLoadVotesForProposal_NonexistentDir(t *testing.T) {
	loaded, err := LoadVotesForProposal("/nonexistent/votes", "some-proposal")
	require.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestLoadVotesForProposal_SkipsSubdirectories(t *testing.T) {
	dir := t.TempDir()
	proposalID := "test-proposal"
	proposalVotesDir := filepath.Join(dir, proposalID)

	now := time.Now().UTC().Truncate(time.Second)
	v := &Vote{
		ProposalID: proposalID,
		VoterEmail: "alice@test.com",
		Decision:   "yes",
		VotedAt:    now,
	}
	err := SaveVote(filepath.Join(proposalVotesDir, "alice@test.com.yml"), v)
	require.NoError(t, err)

	// Create a subdirectory
	err = os.MkdirAll(filepath.Join(proposalVotesDir, "subdir"), 0755)
	require.NoError(t, err)

	loaded, err := LoadVotesForProposal(dir, proposalID)
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
}

func TestSaveVote_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vote.yml")

	now := time.Now().UTC().Truncate(time.Second)
	original := &Vote{
		ProposalID: "round-trip-test",
		VoterEmail: "roundtrip@test.com",
		Decision:   "no",
		Comment:    "This needs more discussion",
		VotedAt:    now,
	}

	err := SaveVote(path, original)
	require.NoError(t, err)

	loaded, err := LoadVote(path)
	require.NoError(t, err)

	assert.Equal(t, original.ProposalID, loaded.ProposalID)
	assert.Equal(t, original.VoterEmail, loaded.VoterEmail)
	assert.Equal(t, original.Decision, loaded.Decision)
	assert.Equal(t, original.Comment, loaded.Comment)
	assert.True(t, original.VotedAt.Equal(loaded.VotedAt))
}
