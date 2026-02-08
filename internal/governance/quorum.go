package governance

import (
	"math"

	"github.com/AlexGladkov/guardian-cli/internal/config"
)

// QuorumResult holds the result of a quorum calculation.
type QuorumResult struct {
	Required      int
	YesVotes      int
	NoVotes       int
	TotalEligible int
	Result        string // ACCEPTED|REJECTED|PENDING|EXPIRED
}

// CalculateQuorum computes the quorum result based on the given configuration,
// total eligible voters, and the current vote counts.
//
// Quorum types:
//   - majority:   required = totalEligible/2 + 1
//   - two_thirds: required = ceil(totalEligible * 2/3)
//   - unanimous:  required = totalEligible
//   - custom:     required = ceil(totalEligible * threshold)
//
// Result determination:
//   - ACCEPTED if yesVotes >= required
//   - REJECTED if noVotes > (totalEligible - required), meaning there are not enough
//     remaining voters for yes to reach the threshold
//   - PENDING otherwise
func CalculateQuorum(qc config.QuorumConfig, totalEligible int, yesVotes int, noVotes int) *QuorumResult {
	required := calculateRequired(qc, totalEligible)

	result := "PENDING"
	if yesVotes >= required {
		result = "ACCEPTED"
	} else if noVotes > (totalEligible - required) {
		result = "REJECTED"
	}

	return &QuorumResult{
		Required:      required,
		YesVotes:      yesVotes,
		NoVotes:       noVotes,
		TotalEligible: totalEligible,
		Result:        result,
	}
}

// calculateRequired computes the number of yes votes required for a quorum
// to be met, based on the quorum type and total eligible voters.
func calculateRequired(qc config.QuorumConfig, totalEligible int) int {
	switch qc.Type {
	case "majority":
		return totalEligible/2 + 1
	case "two_thirds":
		return int(math.Ceil(float64(totalEligible) * 2.0 / 3.0))
	case "unanimous":
		return totalEligible
	case "custom":
		return int(math.Ceil(float64(totalEligible) * qc.Threshold))
	default:
		// Fallback to majority if unknown type.
		return totalEligible/2 + 1
	}
}
