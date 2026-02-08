package governance

import (
	"time"

	"github.com/AlexGladkov/guardian-cli/internal/config"
)

// TallyResult is the full tally of a proposal including voter information,
// vote counts, quorum calculation, and expiry status.
type TallyResult struct {
	ProposalID     string
	RuleID         string
	EligibleVoters []string
	Votes          []*config.Vote
	QuorumResult   *QuorumResult
	QuorumConfig   config.QuorumConfig
	IsExpired      bool
}

// ComputeTally calculates the full tally for a proposal given its votes and
// the constitution. It handles:
//   - Eligible voter determination (deduplicated by email)
//   - Per-rule quorum overrides
//   - Vote counting (only from eligible voters)
//   - TTL expiry checking
//   - Final quorum computation
func ComputeTally(
	proposal *config.Proposal,
	votes []*config.Vote,
	constitution *config.Constitution,
) *TallyResult {
	// Get eligible voters.
	eligibleVoters := GetEligibleVoters(constitution)

	// Build a set of eligible emails for fast lookup.
	eligibleSet := make(map[string]bool, len(eligibleVoters))
	for _, email := range eligibleVoters {
		eligibleSet[email] = true
	}

	// Determine quorum config: check per-rule overrides first, then use default.
	qc := constitution.Governance.Quorum
	if override, ok := constitution.Governance.PerRuleOverrides[proposal.RuleID]; ok {
		qc = override.Quorum
	}

	// Count votes from eligible voters only.
	yesVotes := 0
	noVotes := 0
	for _, vote := range votes {
		if !eligibleSet[vote.VoterEmail] {
			continue
		}
		switch vote.Decision {
		case "yes":
			yesVotes++
		case "no":
			noVotes++
		}
	}

	// Check TTL expiry.
	isExpired := false
	ttlDays := constitution.Governance.ProposalTTLDays
	if ttlDays > 0 {
		expiry := proposal.CreatedAt.Add(time.Duration(ttlDays) * 24 * time.Hour)
		if time.Now().After(expiry) {
			isExpired = true
		}
	}

	// Calculate quorum.
	quorumResult := CalculateQuorum(qc, len(eligibleVoters), yesVotes, noVotes)

	// Override result if expired.
	if isExpired {
		quorumResult.Result = "EXPIRED"
	}

	return &TallyResult{
		ProposalID:     proposal.ID,
		RuleID:         proposal.RuleID,
		EligibleVoters: eligibleVoters,
		Votes:          votes,
		QuorumResult:   quorumResult,
		QuorumConfig:   qc,
		IsExpired:      isExpired,
	}
}
