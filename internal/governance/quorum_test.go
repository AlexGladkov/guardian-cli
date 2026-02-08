package governance

import (
	"testing"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCalculateQuorum_Majority_Accepted(t *testing.T) {
	qc := config.QuorumConfig{Type: "majority"}
	result := CalculateQuorum(qc, 5, 3, 2)

	assert.Equal(t, 3, result.Required)     // 5/2 + 1 = 3
	assert.Equal(t, 3, result.YesVotes)
	assert.Equal(t, 2, result.NoVotes)
	assert.Equal(t, 5, result.TotalEligible)
	assert.Equal(t, "ACCEPTED", result.Result)
}

func TestCalculateQuorum_Majority_Rejected(t *testing.T) {
	qc := config.QuorumConfig{Type: "majority"}
	result := CalculateQuorum(qc, 5, 1, 3)

	assert.Equal(t, 3, result.Required)
	assert.Equal(t, "REJECTED", result.Result) // noVotes(3) > totalEligible(5) - required(3) = 2
}

func TestCalculateQuorum_Majority_Pending(t *testing.T) {
	qc := config.QuorumConfig{Type: "majority"}
	result := CalculateQuorum(qc, 5, 1, 1)

	assert.Equal(t, 3, result.Required)
	assert.Equal(t, "PENDING", result.Result) // 1 yes, 1 no, 3 remaining => still possible
}

func TestCalculateQuorum_TwoThirds_Accepted(t *testing.T) {
	qc := config.QuorumConfig{Type: "two_thirds"}
	result := CalculateQuorum(qc, 3, 2, 1)

	assert.Equal(t, 2, result.Required) // ceil(3 * 2/3) = 2
	assert.Equal(t, "ACCEPTED", result.Result)
}

func TestCalculateQuorum_TwoThirds_Rejected(t *testing.T) {
	qc := config.QuorumConfig{Type: "two_thirds"}
	result := CalculateQuorum(qc, 3, 0, 2)

	assert.Equal(t, 2, result.Required) // ceil(3 * 2/3) = 2
	assert.Equal(t, "REJECTED", result.Result) // noVotes(2) > 3-2 = 1
}

func TestCalculateQuorum_TwoThirds_Pending(t *testing.T) {
	qc := config.QuorumConfig{Type: "two_thirds"}
	result := CalculateQuorum(qc, 6, 3, 1)

	assert.Equal(t, 4, result.Required) // ceil(6 * 2/3) = 4
	assert.Equal(t, "PENDING", result.Result) // 3 yes < 4, noVotes(1) <= 6-4=2
}

func TestCalculateQuorum_Unanimous_Accepted(t *testing.T) {
	qc := config.QuorumConfig{Type: "unanimous"}
	result := CalculateQuorum(qc, 3, 3, 0)

	assert.Equal(t, 3, result.Required)
	assert.Equal(t, "ACCEPTED", result.Result)
}

func TestCalculateQuorum_Unanimous_Rejected(t *testing.T) {
	qc := config.QuorumConfig{Type: "unanimous"}
	result := CalculateQuorum(qc, 3, 2, 1)

	assert.Equal(t, 3, result.Required)
	assert.Equal(t, "REJECTED", result.Result) // noVotes(1) > 3-3 = 0
}

func TestCalculateQuorum_Unanimous_Pending(t *testing.T) {
	qc := config.QuorumConfig{Type: "unanimous"}
	result := CalculateQuorum(qc, 3, 2, 0)

	assert.Equal(t, 3, result.Required)
	assert.Equal(t, "PENDING", result.Result) // 2 yes, 0 no, 1 remaining
}

func TestCalculateQuorum_Custom_Accepted(t *testing.T) {
	qc := config.QuorumConfig{Type: "custom", Threshold: 0.75}
	result := CalculateQuorum(qc, 4, 3, 1)

	assert.Equal(t, 3, result.Required) // ceil(4 * 0.75) = 3
	assert.Equal(t, "ACCEPTED", result.Result)
}

func TestCalculateQuorum_Custom_Rejected(t *testing.T) {
	qc := config.QuorumConfig{Type: "custom", Threshold: 0.75}
	result := CalculateQuorum(qc, 4, 0, 2)

	assert.Equal(t, 3, result.Required) // ceil(4 * 0.75) = 3
	assert.Equal(t, "REJECTED", result.Result) // noVotes(2) > 4-3 = 1
}

func TestCalculateQuorum_Custom_Pending(t *testing.T) {
	qc := config.QuorumConfig{Type: "custom", Threshold: 0.5}
	result := CalculateQuorum(qc, 4, 1, 1)

	assert.Equal(t, 2, result.Required) // ceil(4 * 0.5) = 2
	assert.Equal(t, "PENDING", result.Result)
}

func TestCalculateQuorum_SingleVoter_Majority(t *testing.T) {
	qc := config.QuorumConfig{Type: "majority"}
	result := CalculateQuorum(qc, 1, 1, 0)

	assert.Equal(t, 1, result.Required) // 1/2 + 1 = 1
	assert.Equal(t, "ACCEPTED", result.Result)
}

func TestCalculateQuorum_TwoVoters_TwoThirds(t *testing.T) {
	qc := config.QuorumConfig{Type: "two_thirds"}
	result := CalculateQuorum(qc, 2, 2, 0)

	assert.Equal(t, 2, result.Required) // ceil(2 * 2/3) = ceil(1.33) = 2
	assert.Equal(t, "ACCEPTED", result.Result)
}

func TestCalculateQuorum_TwoVoters_TwoThirds_OnlyOneYes(t *testing.T) {
	qc := config.QuorumConfig{Type: "two_thirds"}
	result := CalculateQuorum(qc, 2, 1, 1)

	assert.Equal(t, 2, result.Required)
	// noVotes(1) > 2-2 = 0, so REJECTED.
	assert.Equal(t, "REJECTED", result.Result)
}

func TestCalculateQuorum_ZeroVoters(t *testing.T) {
	qc := config.QuorumConfig{Type: "majority"}
	result := CalculateQuorum(qc, 0, 0, 0)

	assert.Equal(t, 1, result.Required) // 0/2 + 1 = 1
	// With zero eligible voters, quorum can never be met, so REJECTED.
	assert.Equal(t, "REJECTED", result.Result)
}

func TestCalculateQuorum_AllNoVotes(t *testing.T) {
	qc := config.QuorumConfig{Type: "majority"}
	result := CalculateQuorum(qc, 3, 0, 3)

	assert.Equal(t, 2, result.Required)
	assert.Equal(t, "REJECTED", result.Result) // noVotes(3) > 3-2 = 1
}

func TestCalculateQuorum_UnknownTypeFallsBackToMajority(t *testing.T) {
	qc := config.QuorumConfig{Type: "unknown_type"}
	result := CalculateQuorum(qc, 5, 3, 0)

	assert.Equal(t, 3, result.Required) // majority: 5/2 + 1 = 3
	assert.Equal(t, "ACCEPTED", result.Result)
}

func TestCalculateQuorum_Custom_HighThreshold(t *testing.T) {
	qc := config.QuorumConfig{Type: "custom", Threshold: 0.9}
	result := CalculateQuorum(qc, 10, 9, 1)

	assert.Equal(t, 9, result.Required) // ceil(10 * 0.9) = 9
	assert.Equal(t, "ACCEPTED", result.Result)
}

func TestCalculateQuorum_Custom_LowThreshold(t *testing.T) {
	qc := config.QuorumConfig{Type: "custom", Threshold: 0.1}
	result := CalculateQuorum(qc, 10, 1, 0)

	assert.Equal(t, 1, result.Required) // ceil(10 * 0.1) = 1
	assert.Equal(t, "ACCEPTED", result.Result)
}

func TestCalculateQuorum_Majority_EvenVoters(t *testing.T) {
	qc := config.QuorumConfig{Type: "majority"}
	result := CalculateQuorum(qc, 4, 3, 1)

	assert.Equal(t, 3, result.Required) // 4/2 + 1 = 3
	assert.Equal(t, "ACCEPTED", result.Result)
}

func TestCalculateQuorum_Majority_ExactlyRequired(t *testing.T) {
	qc := config.QuorumConfig{Type: "majority"}
	result := CalculateQuorum(qc, 4, 3, 0)

	assert.Equal(t, 3, result.Required)
	assert.Equal(t, "ACCEPTED", result.Result)
}
