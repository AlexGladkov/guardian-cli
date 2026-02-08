// Package inbox provides inbox logic for the Guardian CLI tool.
// It manages proposal notifications, state persistence, and OS notifications.
package inbox

import (
	"fmt"
	"time"

	"github.com/AlexGladkov/guardian-cli/internal/config"
)

// InboxItem represents a proposal that needs the user's vote.
type InboxItem struct {
	Proposal *config.Proposal
	Age      time.Duration
	IsOld    bool // true if age > 7 days
}

// oldThreshold is the duration after which a proposal is considered old.
const oldThreshold = 7 * 24 * time.Hour

// GetInbox returns proposals that need the given user's vote.
//
// It filters proposals based on:
//   - Status must be "proposed"
//   - Not expired (if ProposalTTLDays > 0)
//   - User must be an eligible voter (has a role in governance.voters)
//   - User has not already voted for this proposal
//   - If sinceLastCheck is provided, only proposals created after that time
//
// The votes parameter maps proposal IDs to their votes.
func GetInbox(
	proposals []*config.Proposal,
	votes map[string][]*config.Vote,
	constitution *config.Constitution,
	userEmail string,
	sinceLastCheck *time.Time,
) ([]InboxItem, error) {
	if constitution == nil {
		return nil, fmt.Errorf("constitution is nil")
	}

	// Check if user is an eligible voter
	if !isEligibleVoter(constitution, userEmail) {
		return nil, nil
	}

	now := time.Now()
	var items []InboxItem

	for _, p := range proposals {
		if p == nil {
			continue
		}

		// Only consider proposals with status "proposed"
		if p.Status != "proposed" {
			continue
		}

		// Check TTL: skip if expired
		if constitution.Governance.ProposalTTLDays > 0 {
			ttl := time.Duration(constitution.Governance.ProposalTTLDays) * 24 * time.Hour
			if now.Sub(p.CreatedAt) > ttl {
				continue
			}
		}

		// If sinceLastCheck is provided, filter proposals created after that time
		if sinceLastCheck != nil && !p.CreatedAt.After(*sinceLastCheck) {
			continue
		}

		// Check if user has already voted for this proposal
		if hasVoted(votes[p.ID], userEmail) {
			continue
		}

		// Calculate age
		age := now.Sub(p.CreatedAt)
		isOld := age > oldThreshold

		items = append(items, InboxItem{
			Proposal: p,
			Age:      age,
			IsOld:    isOld,
		})
	}

	return items, nil
}

// isEligibleVoter checks if the user has a role listed in governance.voters.
func isEligibleVoter(constitution *config.Constitution, userEmail string) bool {
	for _, voterRef := range constitution.Governance.Voters {
		role, exists := constitution.Roles[voterRef.Role]
		if !exists {
			continue
		}
		for _, member := range role.Members {
			if member.Email == userEmail {
				return true
			}
		}
	}
	return false
}

// hasVoted checks if the user has already voted in the given votes slice.
func hasVoted(votes []*config.Vote, userEmail string) bool {
	for _, v := range votes {
		if v != nil && v.VoterEmail == userEmail {
			return true
		}
	}
	return false
}
